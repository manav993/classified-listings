# ─── Stage 1: Build ───────────────────────────────────────────────────────────
# Use the full Go image so we have gcc (needed for go-sqlite3's CGO).
FROM golang:1.22 AS builder

WORKDIR /app

# Download modules first — this layer is cached as long as go.mod / go.sum don't change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build a statically-linked binary.
# CGO_ENABLED=1 is required for go-sqlite3 (it wraps the C SQLite library).
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
# debian:bookworm-slim includes the C runtime go-sqlite3 needs at runtime.
FROM debian:bookworm-slim

WORKDIR /app

# curl is used by the Docker healthcheck; keep the layer small with no cache.
RUN apt-get update -qq && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*

# Create a non-root user for security.
RUN useradd -u 10001 appuser

# Copy the compiled binary from the build stage.
COPY --from=builder /app/server /app/server

# Create persistent directories for the database and uploaded images.
RUN mkdir -p /app/data /app/uploads && chown -R appuser:appuser /app

USER appuser

ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/server"]
