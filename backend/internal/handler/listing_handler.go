package handler

import (
	"context"
	"errors"
	"net/http"

	"classified-listings/internal/model"
	"classified-listings/internal/repository"
	"classified-listings/internal/response"
)

const (
	defaultLimit = 9
	maxLimit     = 100

	errNotFound = "listing not found"
	errInternal = "internal server error"
)

// ListingService is the interface the handler depends on.
// Defined here so the handler package stays decoupled from the concrete service type.
type ListingService interface {
	GetAll(ctx context.Context, params model.PaginationParams, filter model.ListingFilter) ([]model.Listing, int64, error)
	GetByID(ctx context.Context, id int64) (model.Listing, error)
	Create(ctx context.Context, input model.ListingInput) (model.Listing, error)
	Update(ctx context.Context, id int64, input model.ListingInput) (model.Listing, error)
	Delete(ctx context.Context, id int64) error
}

type ListingHandler struct {
	svc ListingService
}

func NewListingHandler(svc ListingService) *ListingHandler {
	return &ListingHandler{svc: svc}
}

// handleServiceError maps a service-layer error to the appropriate HTTP response.
// Centralising the dispatch removes the repeated switch/if block from every handler method.
func (h *ListingHandler) handleServiceError(w http.ResponseWriter, err error) {
	var ve model.ValidationError
	switch {
	case errors.Is(err, repository.ErrNotFound):
		response.WriteError(w, http.StatusNotFound, errNotFound)
	case errors.As(err, &ve):
		// Return all field-level errors at once so the caller can fix everything in one shot.
		response.WriteValidationErrors(w, ve.Fields)
	default:
		response.WriteError(w, http.StatusInternalServerError, errInternal)
	}
}

// GetAll handles GET /api/listings.
// Supports pagination (?limit, ?offset) and filters (?search, ?category, ?status).
func (h *ListingHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	params, err := parsePagination(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter, err := parseFilters(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	listings, total, err := h.svc.GetAll(r.Context(), params, filter)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Always return an array, never null.
	if listings == nil {
		listings = []model.Listing{}
	}

	response.WriteJSON(w, http.StatusOK, model.PaginatedResponse{
		Listings: listings,
		Total:    total,
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
}

func (h *ListingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	listing, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, listing)
}

func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, err := decodeListingInput(w, r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	listing, err := h.svc.Create(r.Context(), input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, listing)
}

func (h *ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	input, err := decodeListingInput(w, r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	listing, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, listing)
}

func (h *ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
