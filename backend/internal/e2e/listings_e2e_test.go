// Package e2e exercises the full HTTP stack (handler -> service -> repository -> SQLite)
// using a shared in-process test server started once in TestMain.
// Each test calls resetDB to start with an empty listings table.
// Do not use t.Parallel() - tests share the same database.
package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"classified-listings/internal/db"
	"classified-listings/internal/handler"
	"classified-listings/internal/model"
	"classified-listings/internal/repository"
	"classified-listings/internal/service"
)

// Shared test server
// srv and sqlDB are package-level so all tests share a single HTTP server
// and a single in-memory SQLite database.
var (
	srv   *httptest.Server
	sqlDB *sql.DB
)

// TestMain starts the shared server once, runs all tests, then tears down.
func TestMain(m *testing.M) {
	var err error

	// SQLite in-memory databases are connection-scoped.  MaxOpenConns(1)
	// ensures the pool never opens a second connection that would see an
	// empty database, which would cause schema initialisation to be lost.
	sqlDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic("open sqlite: " + err.Error())
	}
	sqlDB.SetMaxOpenConns(1)

	if err = db.EnsureSchema(sqlDB); err != nil {
		panic("ensure schema: " + err.Error())
	}

	repo := repository.NewSQLiteListingRepository(sqlDB)
	svc := service.NewListingService(repo)
	h := handler.NewListingHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Route("/api/listings", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Post("/", h.Create)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetByID)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
		})
	})

	srv = httptest.NewServer(r)

	code := m.Run()

	srv.Close()
	sqlDB.Close()
	os.Exit(code)
}

// Per-test helpers
// resetDB deletes all rows from the listings table so each test starts clean.
// It must be the first call in every test function.
func resetDB(t *testing.T) {
	t.Helper()
	if _, err := sqlDB.Exec("DELETE FROM listings"); err != nil {
		t.Fatalf("resetDB: %v", err)
	}
}

// postListing creates a listing via the API and returns the created object.
// It fails immediately if the response status is not 201.
func postListing(t *testing.T, input model.ListingInput) model.Listing {
	t.Helper()
	body, _ := json.Marshal(input)
	resp, err := http.Post(srv.URL+"/api/listings", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/listings: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var created model.Listing
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created listing: %v", err)
	}
	return created
}

// seed is a concise shorthand for postListing when only core fields matter
// and the returned listing object is not needed.
func seed(t *testing.T, title, description string, price float64, category, status string) {
	t.Helper()
	postListing(t, model.ListingInput{
		Title:       title,
		Description: description,
		Price:       price,
		Category:    category,
		Status:      status,
	})
}

/*
doRequest sends a JSON request and returns the response.
body may be nil for methods that carry no payload (e.g. DELETE).
The caller is responsible for closing resp.Body.
*/
func doRequest(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, url, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

// getListings calls GET /api/listings with the given query string and decodes
// the paginated response. It fails if the status is not 200.
func getListings(t *testing.T, query string) model.PaginatedResponse {
	t.Helper()
	url := srv.URL + "/api/listings"
	if query != "" {
		url += "?" + query
	}
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var page model.PaginatedResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		t.Fatalf("decode paginated response: %v", err)
	}
	return page
}

// Create
func TestE2E_CreateAndGetByID(t *testing.T) {
	resetDB(t)

	created := postListing(t, model.ListingInput{
		Title:       "Red Bike",
		Description: "Well-maintained road bike",
		Price:       199.99,
		Category:    model.CategoryVehicle,
		Status:      model.StatusActive,
	})

	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/listings/%d", srv.URL, created.ID))
	if err != nil {
		t.Fatalf("GET by ID: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var got model.Listing
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Title != created.Title {
		t.Errorf("expected title %q, got %q", created.Title, got.Title)
	}
}

func TestE2E_CreateListing_ValidationError(t *testing.T) {
	resetDB(t)

	body, _ := json.Marshal(map[string]any{
		"title":    "",
		"price":    -10,
		"category": "InvalidCat",
		"status":   "Bad",
	})
	resp, err := http.Post(srv.URL+"/api/listings", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var errResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if _, ok := errResp["errors"]; !ok {
		t.Error("expected 'errors' key in validation error response")
	}
}

// List / filter
func TestE2E_GetAll_Empty(t *testing.T) {
	resetDB(t)

	page := getListings(t, "")
	if len(page.Listings) != 0 {
		t.Errorf("expected empty listings, got %d", len(page.Listings))
	}
	if page.Total != 0 {
		t.Errorf("expected total=0, got %d", page.Total)
	}
}

func TestE2E_GetAll_FilterByCategory(t *testing.T) {
	resetDB(t)

	seed(t, "Laptop", "Fast and reliable laptop", 999, model.CategoryElectronics, model.StatusActive)
	seed(t, "Car", "Well-maintained older car", 5000, model.CategoryVehicle, model.StatusActive)

	page := getListings(t, "category=Vehicle")
	if page.Total != 1 {
		t.Fatalf("expected 1 Vehicle listing, got %d", page.Total)
	}
	if page.Listings[0].Category != model.CategoryVehicle {
		t.Errorf("expected Vehicle, got %q", page.Listings[0].Category)
	}
}

func TestE2E_GetAll_FilterByStatus(t *testing.T) {
	resetDB(t)

	seed(t, "Active item", "Available and in great condition", 100, model.CategoryElectronics, model.StatusActive)
	seed(t, "Sold item", "No longer available for sale", 50, model.CategoryElectronics, model.StatusInactive)

	page := getListings(t, "status=Inactive")
	if page.Total != 1 {
		t.Fatalf("expected 1 inactive listing, got %d", page.Total)
	}
	if page.Listings[0].Status != model.StatusInactive {
		t.Errorf("expected Inactive, got %q", page.Listings[0].Status)
	}
	if page.Listings[0].Title != "Sold item" {
		t.Errorf("expected 'Sold item', got %q", page.Listings[0].Title)
	}
}

func TestE2E_GetAll_Search(t *testing.T) {
	resetDB(t)

	// "vortex" appears only in the title; "cobalt" appears only in the description.
	// Both searches must return exactly one result, proving that search covers both fields.
	seed(t, "Vortex Blender", "Kitchen appliance, great condition", 80, model.CategoryElectronics, model.StatusActive)
	seed(t, "City Apartment", "Cobalt blue interior, modern finish", 1500, model.CategoryProperty, model.StatusActive)

	titlePage := getListings(t, "search=vortex")
	if titlePage.Total != 1 || titlePage.Listings[0].Title != "Vortex Blender" {
		t.Errorf("title search: expected 'Vortex Blender', got %v", titlePage.Listings)
	}

	descPage := getListings(t, "search=cobalt")
	if descPage.Total != 1 || descPage.Listings[0].Title != "City Apartment" {
		t.Errorf("description search: expected 'City Apartment', got %v", descPage.Listings)
	}
}

func TestE2E_GetAll_Pagination(t *testing.T) {
	resetDB(t)

	for i := 1; i <= 5; i++ {
		seed(t, fmt.Sprintf("Item %d", i), "A simple placeholder description.", float64(i*10), model.CategoryElectronics, model.StatusActive)
	}

	page := getListings(t, "limit=2&offset=2")
	if page.Total != 5 {
		t.Errorf("expected total=5, got %d", page.Total)
	}
	if len(page.Listings) != 2 {
		t.Errorf("expected 2 listings on page, got %d", len(page.Listings))
	}
	if page.Limit != 2 || page.Offset != 2 {
		t.Errorf("expected limit=2 offset=2 in response, got limit=%d offset=%d", page.Limit, page.Offset)
	}
}

// Update
func TestE2E_UpdateListing(t *testing.T) {
	resetDB(t)

	created := postListing(t, model.ListingInput{
		Title:       "Old Title",
		Description: "Original item description text.",
		Price:       50,
		Category:    model.CategoryElectronics,
		Status:      model.StatusActive,
	})

	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/api/listings/%d", srv.URL, created.ID), model.ListingInput{
		Title:       "New Title",
		Description: "Updated item description text.",
		Price:       75,
		Category:    model.CategoryElectronics,
		Status:      model.StatusInactive,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var updated model.Listing
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected 'New Title', got %q", updated.Title)
	}
	if updated.Status != model.StatusInactive {
		t.Errorf("expected Inactive, got %q", updated.Status)
	}
}

func TestE2E_UpdateListing_NotFound(t *testing.T) {
	resetDB(t)

	resp := doRequest(t, http.MethodPut, srv.URL+"/api/listings/99999", model.ListingInput{
		Title:       "Ghost",
		Description: "This item does not exist at all.",
		Price:       1,
		Category:    model.CategoryElectronics,
		Status:      model.StatusActive,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent listing, got %d", resp.StatusCode)
	}
}

// Delete
func TestE2E_DeleteListing(t *testing.T) {
	resetDB(t)

	created := postListing(t, model.ListingInput{
		Title:       "To delete",
		Description: "A temporary listing for testing.",
		Price:       1,
		Category:    model.CategoryElectronics,
		Status:      model.StatusActive,
	})

	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/api/listings/%d", srv.URL, created.ID), nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// Subsequent GET must return 404 - confirms the row is truly gone.
	getResp, err := http.Get(fmt.Sprintf("%s/api/listings/%d", srv.URL, created.ID))
	if err != nil {
		t.Fatalf("GET after delete: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

// Not found
func TestE2E_GetByID_NotFound(t *testing.T) {
	resetDB(t)

	resp, err := http.Get(srv.URL + "/api/listings/99999")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestE2E_GetAll_SearchWithPercentLiteral(t *testing.T) {
	resetDB(t)

	seed(t, "50% off sale", "Discounted item, great value", 50, model.CategoryElectronics, model.StatusActive)
	seed(t, "Full price item", "Full price, no discount applied", 100, model.CategoryElectronics, model.StatusActive)

	// "%25" is the URL-encoded form of "%"; the search term becomes "50%".
	// escapeLike must treat "%" as a literal, not a wildcard.
	page := getListings(t, "search=50%25")
	if page.Total != 1 {
		t.Fatalf("expected 1 result for literal '50%%', got %d", page.Total)
	}
	if page.Listings[0].Title != "50% off sale" {
		t.Errorf("unexpected listing: %q", page.Listings[0].Title)
	}
}
