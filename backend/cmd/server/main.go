package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/mattn/go-sqlite3"

	"classified-listings/internal/db"
	"classified-listings/internal/handler"
	"classified-listings/internal/repository"
	"classified-listings/internal/service"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/listings.db"
	}

	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "data/uploads"
	}
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		slog.Error("failed to create uploads directory", "err", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		slog.Error("failed to create data directory", "err", err)
		os.Exit(1)
	}

	dsn := "file:" + dbPath + "?_loc=UTC&_busy_timeout=5000"
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	// SQLite supports only one concurrent writer. Limiting the pool to a single
	// connection prevents "database is locked" errors under concurrent requests.
	// _busy_timeout in the DSN handles brief contention, but the pool cap makes
	// the behaviour deterministic.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	// Run schema initialisation on every startup - safe because it is idempotent.
	if err := db.EnsureSchema(sqlDB); err != nil {
		slog.Error("failed to initialise schema", "err", err)
		os.Exit(1)
	}

	// Wire dependencies: repository -> service -> handler.
	repo := repository.NewSQLiteListingRepository(sqlDB)
	svc := service.NewListingService(repo)
	listingHandler := handler.NewListingHandler(svc)
	uploadHandler := handler.NewUploadHandler(uploadsDir)

	r := chi.NewRouter()
	r.Use(middleware.Logger)    // structured request logging
	r.Use(middleware.Recoverer) // recover from panics and return 500

	// AllowedOrigins is read from the environment so it can be tightened per
	// deployment without rebuilding the binary. Defaults to localhost for local dev.
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api/listings", func(r chi.Router) {
		r.Get("/", listingHandler.GetAll)
		r.Post("/", listingHandler.Create)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", listingHandler.GetByID)
			r.Put("/", listingHandler.Update)
			r.Delete("/", listingHandler.Delete)
		})
	})

	r.Post("/api/upload", uploadHandler.Upload)

	// Serve uploaded images as static files.
	// StripPrefix removes the /uploads/ prefix before looking up the file on disk.
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Listen for OS shutdown signals so in-flight requests can finish cleanly.
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-shutdownCh
	slog.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}
