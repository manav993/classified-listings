# Marketplace Listings API

A RESTful API for marketplace listings built with Go, chi, and SQLite.

## Tech Stack

| Layer      | Technology                        |
|------------|-----------------------------------|
| Language   | Go 1.22+                          |
| Router     | github.com/go-chi/chi/v5          |
| Database   | SQLite via database/sql + go-sqlite3 |
| Logging    | log/slog (Go standard library)    |

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # entrypoint: wires dependencies and starts server
├── internal/
│   ├── db/
│   │   └── db.go                # schema initialisation
│   ├── e2e/
│   │   └── listings_e2e_test.go # full HTTP-stack integration tests
│   ├── handler/
│   │   ├── listing_handler.go        # HTTP layer: routing, response writing, error dispatch
│   │   ├── upload_handler.go         # POST /api/upload: file validation, storage, URL response
│   │   ├── params.go                 # Request parsing: pagination, filters, body decode
│   │   └── listing_handler_test.go
│   ├── model/
│   │   └── listing.go           # shared data types and constants
│   ├── repository/
│   │   ├── listing_repository.go       # CRUD interface + ErrNotFound sentinel
│   │   └── sqlite_listing_repository.go # SQLite implementation
│   ├── response/
│   │   └── json.go              # WriteJSON, WriteError, WriteValidationErrors
│   ├── service/
│   │   └── listing_service.go   # validation and business logic
│   └── validator/
│       ├── listing_validator.go
│       └── listing_validator_test.go
├── migrations/
│   └── 001_create_listings.sql  # reference schema
├── data/
│   └── listings.db              # created at runtime (git-ignored)
├── go.mod
└── README.md
```

## Architecture

The application follows a layered design with clear separation of concerns:

- **handler** — parses HTTP requests, maps errors to status codes, and writes JSON responses. It depends on a `ListingService` interface, making it straightforward to test with a mock.
- **service** — validates input and enforces business rules before delegating to the repository. Validation errors are returned as structured field-level errors so the caller can display all problems at once.
- **repository** — encapsulates all SQL. The `ListingRepository` interface isolates the handler and service from the database driver, so switching from SQLite to Postgres only requires a new implementation.
- **model** — shared data types (`Listing`, `ListingInput`) and domain constants (`Category`, `Status`).

## Setup

**Requirements:** Go 1.22 or later (no external tools needed).

```bash
git clone <repo-url>
cd classified-listings
go mod tidy
go run ./cmd/server
```

The server starts on `http://localhost:8080`. The SQLite database is created automatically at `data/listings.db`.

## Environment Variables

| Variable         | Default                   | Description                                                  |
|------------------|---------------------------|--------------------------------------------------------------|
| `PORT`           | `8080`                    | Port the HTTP server listens on                              |
| `ALLOWED_ORIGIN` | `http://localhost:3000`   | Single origin allowed by CORS. Set to your frontend domain in production. |
| `DB_PATH`        | `./data/listings.db`      | Path to the SQLite database file                             |
| `UPLOADS_DIR`    | `./uploads`               | Directory where uploaded images are stored and served from   |

To use a different port:
```bash
PORT=9000 go run ./cmd/server
```

## Running Tests

```bash
go test ./...
```

## SQL Schema

The schema is applied automatically on startup. The reference file is at `migrations/001_create_listings.sql`.

```sql
CREATE TABLE IF NOT EXISTS listings (
  id          INTEGER  PRIMARY KEY AUTOINCREMENT,
  title       TEXT     NOT NULL,
  description TEXT     NOT NULL,
  price       REAL     NOT NULL,
  category    TEXT     NOT NULL,
  date_posted DATETIME NOT NULL,
  status      TEXT     NOT NULL,
  image_url   TEXT
);

-- Indexes applied automatically on startup
CREATE INDEX IF NOT EXISTS idx_listings_category    ON listings(category);
CREATE INDEX IF NOT EXISTS idx_listings_status      ON listings(status);
CREATE INDEX IF NOT EXISTS idx_listings_date_posted ON listings(date_posted);
```

## API Reference

Base URL: `http://localhost:8080`

### Listing object

```json
{
  "id": 1,
  "title": "Apartment downtown",
  "description": "Spacious 2 bed, 1 bath apartment.",
  "price": 1200.00,
  "category": "Property",
  "date_posted": "2026-03-14T12:00:00Z",
  "status": "Active",
  "image_url": "/uploads/abc123.jpg"
}
```

`image_url` is `""` when no image has been uploaded.

### Endpoints

#### GET /api/listings
Returns listings. Supports filtering, search, and pagination via query parameters.

| Parameter   | Type    | Description                                              |
|-------------|---------|----------------------------------------------------------|
| `search`    | string  | Case-insensitive search across title and description     |
| `category`  | string  | One of `Property`, `Vehicle`, `Electronics`              |
| `status`    | string  | One of `Active`, `Inactive`                              |
| `offset`    | integer | Number of results to skip for pagination (default: `0`)  |
| `limit`     | integer | Results per page (default: `9`, max: `100`)              |

**Example:** `GET /api/listings?category=Vehicle&status=Active&search=bike&offset=10&limit=10`

**Response:** `200 OK`
```json
{
  "listings": [{ "id": 1, "title": "...", ... }],
  "total": 42,
  "limit": 10,
  "offset": 10
}
```
Returns `listings: []` when no listings match.

---

#### GET /api/listings/{id}
Returns a single listing.

**Response:** `200 OK` or `404 Not Found`

---

#### POST /api/listings
Creates a new listing. `date_posted` is set by the server (UTC).

**Request body:**
```json
{
  "title": "Used bike",
  "description": "Well-maintained, barely ridden.",
  "price": 120,
  "category": "Vehicle",
  "status": "Active",
  "image_url": "/uploads/abc123.jpg"
}
```

To include an image, upload the file via `POST /api/upload` first and pass the returned URL as `image_url`. Omit the field or send `""` for no image.

**Response:** `201 Created` with the full listing object including `id` and `date_posted`.

---

#### POST /api/upload
Uploads an image file and returns the URL to use in `image_url`.

**Request:** `multipart/form-data` with a `file` field containing the image.

Accepted types: JPEG, PNG, WebP, GIF. Maximum size: 5 MB.

**Response:** `200 OK`
```json
{ "url": "/uploads/abc123.jpg" }
```

**Error responses:** `400` for invalid file type, `413` for file too large.

---

#### PUT /api/listings/{id}
Replaces a listing. All fields are required.

**Response:** `200 OK` or `404 Not Found`

---

#### DELETE /api/listings/{id}
Deletes a listing.

**Response:** `204 No Content` or `404 Not Found`

---

### Validation rules

| Field       | Rule                                                        |
|-------------|-------------------------------------------------------------|
| title       | required, 3–100 characters                                  |
| description | required, 20–1000 characters                                |
| price       | required, must be > 0 and <= 10,000,000 (GBP, stored as a decimal) |
| category    | must be one of `Property`, `Vehicle`, `Electronics`         |
| status      | must be one of `Active`, `Inactive`                         |
| image_url   | optional; empty string and `null` are both treated as "no image" |
| date_posted | set by server on create; unchanged on update                |

### Error responses

All errors return JSON. Non-validation errors:
```json
{ "error": "listing not found" }
```

Validation errors return all failing fields at once:
```json
{
  "errors": [
    { "field": "price",    "message": "price must be greater than zero" },
    { "field": "category", "message": "category must be one of Property, Vehicle, Electronics" }
  ]
}
```

### HTTP status codes

| Status | When                                    |
|--------|-----------------------------------------|
| 200    | Successful GET or PUT                   |
| 201    | Successful POST (created)               |
| 204    | Successful DELETE (no content)          |
| 400    | Invalid JSON, invalid path param, or validation failure |
| 404    | Listing not found                       |
| 500    | Unexpected server error                 |

## Assumptions

- All fields are required for both create and update; partial updates (PATCH) are out of scope.
- `date_posted` is set by the server on create and is never modified by update.
- Prices are in **GBP (£)**. The `price` field is a decimal number; no currency conversion is performed.
- SQLite is used for local development convenience; no external database setup is required.
- Category and status are constrained at the application layer for clear error messages; the database stores them as plain text.

## What I Would Do Differently in Production

**Database**
- Replace SQLite with PostgreSQL. SQLite serialises writes, which becomes a bottleneck under concurrent load. PostgreSQL supports row-level locking, connection pooling, and horizontal read replicas.
- Use a proper migrations tool (e.g. golang-migrate) so schema changes are versioned, repeatable, and safe to run in CI/CD pipelines.

**API design**
- Support `PATCH /api/listings/{id}` for partial updates, so callers don't have to resend every field to change one.

**Reliability and observability**
- Add structured request logging with a request ID, latency, and status code on every response.
- Add metrics (Prometheus) and distributed tracing (OpenTelemetry) so slow queries and error rates can be monitored.

**Image upload**
- The current implementation has a working `POST /api/upload` endpoint that validates file type (via magic bytes, not just extension), enforces a 5 MB size limit, and stores images on the local filesystem. The frontend uses a file picker and uploads on selection.
- In production this would be extended to:
  1. Stream files directly to object storage (AWS S3, GCS, or Cloudflare R2) rather than local disk, so uploads survive container restarts and scale horizontally.
  2. Return a CDN URL so images are served from an edge network rather than the application server.
  3. Delete the stored file when the listing is deleted or its image is replaced.

**Security**
- Add authentication (e.g. JWT) and authorisation so only listing owners can modify or delete their own listings.
- Rate-limit the API to prevent abuse.

**Testing**
- Add service-level tests covering edge cases such as concurrent updates.
- Add Playwright/Cypress end-to-end browser tests that drive the full stack.
