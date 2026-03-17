package model

import (
	"fmt"
	"strings"
	"time"
)

const (
	CategoryProperty    = "Property"
	CategoryVehicle     = "Vehicle"
	CategoryElectronics = "Electronics"
)

const (
	StatusActive   = "Active"
	StatusInactive = "Inactive"
)

type Listing struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	DatePosted  time.Time `json:"date_posted"`
	Status      string    `json:"status"`
	ImageURL    string    `json:"image_url"`
}

type ListingInput struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
	// ImageURL is optional. Empty string means no image.
	ImageURL string `json:"image_url"`
}

// PaginationParams holds the limit and offset values parsed from query parameters.
type PaginationParams struct {
	Limit  int
	Offset int
}

// ListingFilter holds optional filters for GET /api/listings.
// Only non-nil fields are applied to the query.
type ListingFilter struct {
	Search   *string // matches title or description (case-insensitive)
	Category *string // Property, Vehicle, Electronics
	Status   *string // Active, Inactive
}

// FieldError names the field that failed validation and explains why.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError is returned by the service when one or more input fields fail
// validation. Carrying all failures at once lets the caller report everything in
// a single response.
type ValidationError struct {
	Fields []FieldError
}

func (e ValidationError) Error() string {
	msgs := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		msgs[i] = fmt.Sprintf("%s: %s", f.Field, f.Message)
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

// PaginatedResponse is the envelope returned by GET /api/listings.
// It always wraps the results with total count and pagination info.
type PaginatedResponse struct {
	Listings []Listing `json:"listings"`
	Total    int64     `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}
