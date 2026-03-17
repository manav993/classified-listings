package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"classified-listings/internal/model"
)

// ErrorResponse is the standard error envelope for all non-validation errors.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ValidationErrorResponse wraps field-level validation errors.
// Produces: {"errors": [{"field": "...", "message": "..."}]}
type ValidationErrorResponse struct {
	Errors []model.FieldError `json:"errors"`
}

// WriteJSON encodes payload as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(payload); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

// WriteError writes a plain {"error": "..."} JSON response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}

// WriteValidationErrors writes a 400 response with a list of field-level errors.
func WriteValidationErrors(w http.ResponseWriter, errs []model.FieldError) {
	WriteJSON(w, http.StatusBadRequest, ValidationErrorResponse{Errors: errs})
}
