# Classified Listings — Full Stack Mini Application

A classified listings marketplace built with **Go** (backend API) and **Vue 3** (frontend SPA), served via **Docker Compose**.

---

## Tech Stack

| Layer     | Technology                              |
|-----------|-----------------------------------------|
| Backend   | Go 1.22, chi router, SQLite             |
| Frontend  | Vue 3, Vite, Vue Router                 |
| Container | Docker, Docker Compose, Nginx           |

---

## Quick Start (Docker — recommended)

> **Requirements:** Docker Desktop installed and running.

```bash
git clone <repo-url>
cd classified-listings
docker compose up --build
```

| Service  | URL                         |
|----------|-----------------------------|
| Frontend | http://localhost:3000        |
| API      | http://localhost:8080/api/listings |

To stop:
```bash
docker compose down
```

---

## Local Development (without Docker)

### Backend

**Requirements:** Go 1.22+

```bash
cd backend
go mod tidy
go run ./cmd/server
# API is now running on http://localhost:8080
```

Run backend tests:
```bash
cd backend
go test ./...
```

### Frontend

**Requirements:** Node.js 18+

```bash
cd frontend
npm install
npm run dev
# App is now running on http://localhost:5173
```

Build for production:
```bash
npm run build
```

---

## Project Structure

```
classified-listings/
├── backend/                  # Go REST API
│   ├── cmd/server/main.go    # Entry point
│   ├── internal/
│   │   ├── handler/          # HTTP layer
│   │   ├── service/          # Business logic & validation
│   │   ├── repository/       # SQLite data access
│   │   ├── model/            # Shared types
│   │   ├── validator/        # Input validation
│   │   ├── db/               # Schema initialisation
│   │   └── e2e/              # End-to-end tests
│   └── README.md             # Full API reference
│
├── frontend/                 # Vue 3 SPA
│   ├── src/
│   │   ├── views/            # Page components
│   │   ├── components/       # Reusable UI components
│   │   ├── composables/      # Shared state logic (useListings)
│   │   ├── api/              # API client
│   │   └── router/           # Vue Router config
│   └── README.md
│
├── docker-compose.yml
└── README.md                 # This file
```

---

## Features

- **Listings page** — responsive card grid, live search, category/status filters, URL-persisted filters, skeleton loading states
- **Create / edit form** — client-side and server-side validation, image URL support, inline error feedback
- **Single listing view** — full details page, edit in place, delete with confirmation dialog
- **Pagination** — server-side with "Showing X to Y of Z" count
- **Fully Dockerised** — multi-stage builds, Nginx SPA serving, SQLite data volume

---

## API Reference

See [`backend/README.md`](./backend/README.md) for the full endpoint reference, request/response shapes, validation rules, and HTTP status code documentation.

---

## Assumptions

- **Vue 3 + Vite used instead of Nuxt.js** — the requirements list "Vue.js/Nuxt.js" as the frontend stack. Vue 3 was chosen because:
  - The application is a pure client-side SPA with no SEO requirements, so Nuxt's SSR/SSG capabilities add complexity without benefit here.
  - Vue Router handles all navigation needs directly.
  - In a real product with public-facing listing pages that need SEO, Nuxt would be the right choice and the migration path from this codebase is straightforward (move `views/` → `pages/`, replace `useListings` with Pinia, add `useFetch`/`useAsyncData`).
- All form fields (title, description, price, category, status) are required. Image URL is optional.
- `date_posted` is set server-side on creation and never modified by updates.
- SQLite is used for local/test convenience. A production deployment would use PostgreSQL (see backend README).
- Category and status are validated at the application layer rather than via database constraints for clearer error messages.

---

## What I Would Do Differently in Production

### Architecture
- **Replace SQLite with PostgreSQL** — SQLite serialises all writes, which becomes a bottleneck under concurrent load.
- **Use a migrations tool** (e.g. `golang-migrate`) so schema changes are versioned and safe to run in CI/CD.
- **Add authentication** — JWT-based auth so only listing owners can modify or delete their own listings.
- **Replace the image URL input with a proper file upload flow.** The current implementation accepts a URL string entered by the user. In production this would be replaced with: a `POST /api/upload` endpoint that accepts a `multipart/form-data` file, streams it to object storage (S3, GCS, or Cloudflare R2), and returns the CDN URL. The frontend would present a file picker, upload on selection, and store only the returned URL in `image_url` — so the data model stays the same but the user never types a URL manually.

### API
- Support `PATCH /api/listings/{id}` for partial updates.
- Add an `Idempotency-Key` header on `POST` to prevent duplicate listings from retried requests.
- Add rate limiting to protect against abuse.

### Observability
- Structured logging with request IDs and latency on every response.
- Prometheus metrics + OpenTelemetry tracing.

### Testing
- Frontend component tests with Vitest + Vue Test Utils.
- Playwright/Cypress end-to-end browser tests.
- Repository integration tests against a real in-process SQLite instance.
