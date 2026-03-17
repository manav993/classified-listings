package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"classified-listings/internal/model"
	"classified-listings/internal/repository"
)

// Stub service
type stubService struct {
	getAll  func(ctx context.Context, params model.PaginationParams, filter model.ListingFilter) ([]model.Listing, int64, error)
	getByID func(ctx context.Context, id int64) (model.Listing, error)
	create  func(ctx context.Context, input model.ListingInput) (model.Listing, error)
	update  func(ctx context.Context, id int64, input model.ListingInput) (model.Listing, error)
	delete  func(ctx context.Context, id int64) error
}

func (s stubService) GetAll(ctx context.Context, params model.PaginationParams, filter model.ListingFilter) ([]model.Listing, int64, error) {
	return s.getAll(ctx, params, filter)
}
func (s stubService) GetByID(ctx context.Context, id int64) (model.Listing, error) {
	return s.getByID(ctx, id)
}
func (s stubService) Create(ctx context.Context, input model.ListingInput) (model.Listing, error) {
	return s.create(ctx, input)
}
func (s stubService) Update(ctx context.Context, id int64, input model.ListingInput) (model.Listing, error) {
	return s.update(ctx, id, input)
}
func (s stubService) Delete(ctx context.Context, id int64) error {
	return s.delete(ctx, id)
}

// newStub returns a stubService where every method panics with a descriptive
// message if called unexpectedly. Tests override only the methods they need,
// so a missing override produces a clear failure instead of a nil-pointer panic.
func newStub(t *testing.T) stubService {
	t.Helper()
	unexpected := func(method string) {
		t.Helper()
		t.Fatalf("unexpected call to %s - set the stub field if this method should be called", method)
	}
	return stubService{
		getAll: func(_ context.Context, _ model.PaginationParams, _ model.ListingFilter) ([]model.Listing, int64, error) {
			unexpected("GetAll")
			return nil, 0, nil
		},
		getByID: func(_ context.Context, _ int64) (model.Listing, error) {
			unexpected("GetByID")
			return model.Listing{}, nil
		},
		create: func(_ context.Context, _ model.ListingInput) (model.Listing, error) {
			unexpected("Create")
			return model.Listing{}, nil
		},
		update: func(_ context.Context, _ int64, _ model.ListingInput) (model.Listing, error) {
			unexpected("Update")
			return model.Listing{}, nil
		},
		delete: func(_ context.Context, _ int64) error {
			unexpected("Delete")
			return nil
		},
	}
}

// Router helper
func newTestRouter(svc ListingService) http.Handler {
	h := NewListingHandler(svc)
	r := chi.NewRouter()
	r.Route("/api/listings", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Post("/", h.Create)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetByID)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
		})
	})
	return r
}

// validBody returns a JSON-encoded valid ListingInput for use in POST/PUT tests.
func validBody(t *testing.T) *bytes.Reader {
	t.Helper()
	b, _ := json.Marshal(model.ListingInput{
		Title:       "Title",
		Description: "Desc",
		Price:       10,
		Category:    model.CategoryVehicle,
		Status:      model.StatusActive,
	})
	return bytes.NewReader(b)
}

// jsonRequest creates an httptest.Request with Content-Type: application/json set.
// All mutating endpoint tests must use this so the Content-Type check passes.
func jsonRequest(method, url string, body *bytes.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// GET /api/listings
func TestGetAll_ReturnsPaginatedResponse(t *testing.T) {
	stub := newStub(t)
	stub.getAll = func(_ context.Context, _ model.PaginationParams, _ model.ListingFilter) ([]model.Listing, int64, error) {
		return []model.Listing{
			{ID: 1, Title: "Bike", Description: "Good", Price: 100, Category: model.CategoryVehicle, Status: model.StatusActive},
		}, 1, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/listings", nil)
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp model.PaginatedResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 1 || len(resp.Listings) != 1 {
		t.Errorf("expected total=1 and 1 listing, got total=%d listings=%d", resp.Total, len(resp.Listings))
	}
}

func TestGetAll_FiltersArePassedToService(t *testing.T) {
	var capturedFilter model.ListingFilter
	stub := newStub(t)
	stub.getAll = func(_ context.Context, _ model.PaginationParams, f model.ListingFilter) ([]model.Listing, int64, error) {
		capturedFilter = f
		return []model.Listing{}, 0, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/listings?category=Vehicle&status=Active&search=bike", nil)
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedFilter.Category == nil || *capturedFilter.Category != "Vehicle" {
		t.Errorf("expected category=Vehicle, got %v", capturedFilter.Category)
	}
	if capturedFilter.Status == nil || *capturedFilter.Status != "Active" {
		t.Errorf("expected status=Active, got %v", capturedFilter.Status)
	}
	if capturedFilter.Search == nil || *capturedFilter.Search != "bike" {
		t.Errorf("expected search=bike, got %v", capturedFilter.Search)
	}
}

// TestGetAll_InvalidQueryParams table-drives all bad query-string inputs that
// should return 400 before the service is ever called.
func TestGetAll_InvalidQueryParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"invalid limit",   "/api/listings?limit=abc"},
		{"negative offset", "/api/listings?offset=-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The stub should never be reached - parsing fails before the service call.
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			newTestRouter(newStub(t)).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rec.Code)
			}
		})
	}
}

func TestGetAll_LimitClamped(t *testing.T) {
	var capturedLimit int
	stub := newStub(t)
	stub.getAll = func(_ context.Context, params model.PaginationParams, _ model.ListingFilter) ([]model.Listing, int64, error) {
		capturedLimit = params.Limit
		return []model.Listing{}, 0, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/listings?limit=9999", nil)
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedLimit != maxLimit {
		t.Errorf("expected limit clamped to %d, got %d", maxLimit, capturedLimit)
	}
}

func TestGetAll_ServiceError_Returns500(t *testing.T) {
	stub := newStub(t)
	stub.getAll = func(_ context.Context, _ model.PaginationParams, _ model.ListingFilter) ([]model.Listing, int64, error) {
		return nil, 0, errors.New("database unavailable")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/listings", nil)
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// GET /api/listings/{id}
// TestGetByID_InvalidID table-drives non-numeric and zero IDs - all should 400.
func TestGetByID_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"non-numeric", "/api/listings/abc"},
		{"zero",        "/api/listings/0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			newTestRouter(newStub(t)).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rec.Code)
			}
		})
	}
}

func TestGetByID_NotFound(t *testing.T) {
	stub := newStub(t)
	stub.getByID = func(_ context.Context, _ int64) (model.Listing, error) {
		return model.Listing{}, repository.ErrNotFound
	}

	req := httptest.NewRequest(http.MethodGet, "/api/listings/123", nil)
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// POST /api/listings
func TestCreate_ReturnsCreatedListing(t *testing.T) {
	want := model.Listing{ID: 42, Title: "Bike", Description: "Great condition", Price: 150, Category: model.CategoryVehicle, Status: model.StatusActive}
	stub := newStub(t)
	stub.create = func(_ context.Context, _ model.ListingInput) (model.Listing, error) {
		return want, nil
	}

	req := jsonRequest(http.MethodPost, "/api/listings", validBody(t))
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var got model.Listing
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != want.ID || got.Title != want.Title {
		t.Errorf("expected {ID:%d Title:%q}, got {ID:%d Title:%q}", want.ID, want.Title, got.ID, got.Title)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/listings", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	newTestRouter(newStub(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreate_ValidationError(t *testing.T) {
	stub := newStub(t)
	stub.create = func(_ context.Context, _ model.ListingInput) (model.Listing, error) {
		return model.Listing{}, model.ValidationError{
			Fields: []model.FieldError{{Field: "title", Message: "title is required"}},
		}
	}

	req := jsonRequest(http.MethodPost, "/api/listings", validBody(t))
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["errors"]; !ok {
		t.Error("expected 'errors' key in validation error response")
	}
}

// PUT /api/listings/{id}
func TestUpdate_ReturnsUpdatedListing(t *testing.T) {
	want := model.Listing{ID: 1, Title: "Updated", Description: "Updated desc", Price: 200, Category: model.CategoryProperty, Status: model.StatusInactive}
	stub := newStub(t)
	stub.update = func(_ context.Context, _ int64, _ model.ListingInput) (model.Listing, error) {
		return want, nil
	}

	req := jsonRequest(http.MethodPut, "/api/listings/1", validBody(t))
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got model.Listing
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Title != want.Title || got.Status != want.Status {
		t.Errorf("expected {Title:%q Status:%q}, got {Title:%q Status:%q}", want.Title, want.Status, got.Title, got.Status)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	stub := newStub(t)
	stub.update = func(_ context.Context, _ int64, _ model.ListingInput) (model.Listing, error) {
		return model.Listing{}, repository.ErrNotFound
	}

	req := jsonRequest(http.MethodPut, "/api/listings/999", validBody(t))
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdate_ValidationError(t *testing.T) {
	stub := newStub(t)
	stub.update = func(_ context.Context, _ int64, _ model.ListingInput) (model.Listing, error) {
		return model.Listing{}, model.ValidationError{
			Fields: []model.FieldError{{Field: "price", Message: "price must be greater than zero"}},
		}
	}

	req := jsonRequest(http.MethodPut, "/api/listings/1", validBody(t))
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["errors"]; !ok {
		t.Error("expected 'errors' key in validation error response")
	}
}

func TestUpdate_InvalidID(t *testing.T) {
	req := jsonRequest(http.MethodPut, "/api/listings/abc", validBody(t))
	rec := httptest.NewRecorder()
	newTestRouter(newStub(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// DELETE /api/listings/{id}
func TestDelete_NoContent(t *testing.T) {
	stub := newStub(t)
	stub.delete = func(_ context.Context, _ int64) error { return nil }

	req := httptest.NewRequest(http.MethodDelete, "/api/listings/1", nil)
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	stub := newStub(t)
	stub.delete = func(_ context.Context, _ int64) error { return repository.ErrNotFound }

	req := httptest.NewRequest(http.MethodDelete, "/api/listings/999", nil)
	rec := httptest.NewRecorder()
	newTestRouter(stub).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
