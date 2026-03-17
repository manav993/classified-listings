package repository

import (
	"context"
	"errors"

	"classified-listings/internal/model"
)

// ErrNotFound is returned when a listing does not exist in the database.
var ErrNotFound = errors.New("listing not found")

// ListingRepository defines the persistence contract for listings.
// The handler and service depend on this interface, not on the SQLite implementation.
type ListingRepository interface {
	GetAll(ctx context.Context, params model.PaginationParams, filter model.ListingFilter) ([]model.Listing, int64, error)
	GetByID(ctx context.Context, id int64) (model.Listing, error)
	Create(ctx context.Context, input model.ListingInput) (model.Listing, error)
	Update(ctx context.Context, id int64, input model.ListingInput) (model.Listing, error)
	Delete(ctx context.Context, id int64) error
}
