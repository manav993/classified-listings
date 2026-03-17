package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"classified-listings/internal/model"
)

const maxBodyBytes = 1 << 20 // 1 MB

// parsePagination reads ?limit= and ?offset= from the query string.
// Applies defaults and clamps limit to maxLimit.
func parsePagination(r *http.Request) (model.PaginationParams, error) {
	params := model.PaginationParams{
		Limit:  defaultLimit,
		Offset: 0,
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return params, errors.New("limit must be a positive integer")
		}
		if n > maxLimit {
			n = maxLimit
		}
		params.Limit = n
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return params, errors.New("offset must be a non-negative integer")
		}
		params.Offset = n
	}

	return params, nil
}

// parseFilters reads the optional filter query params.
// Supported: ?search=, ?category=, ?status=
func parseFilters(r *http.Request) (model.ListingFilter, error) {
	q := r.URL.Query()
	var filter model.ListingFilter

	if v := strings.TrimSpace(q.Get("search")); v != "" {
		filter.Search = &v
	}
	if v := strings.TrimSpace(q.Get("category")); v != "" {
		filter.Category = &v
	}
	if v := strings.TrimSpace(q.Get("status")); v != "" {
		filter.Status = &v
	}

	return filter, nil
}

// parseID extracts and validates the {id} URL parameter.
func parseID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id must be a positive integer")
	}
	return id, nil
}

// decodeListingInput decodes the JSON request body into a ListingInput.
// Enforces a 1 MB body limit and requires Content-Type: application/json.
// Unknown fields are silently ignored so clients don't break on extra metadata.
func decodeListingInput(w http.ResponseWriter, r *http.Request) (model.ListingInput, error) {
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return model.ListingInput{}, errors.New("Content-Type must be application/json")
	}

	// MaxBytesReader rejects bodies larger than 1 MB, preventing unbounded memory use.
	body := http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer body.Close()

	var input model.ListingInput
	dec := json.NewDecoder(body)
	if err := dec.Decode(&input); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return model.ListingInput{}, errors.New("request body too large (max 1 MB)")
		}
		return model.ListingInput{}, errors.New("invalid JSON body")
	}
	return input, nil
}
