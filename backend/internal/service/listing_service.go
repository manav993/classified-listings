package service

import (
	"context"

	"classified-listings/internal/model"
	"classified-listings/internal/repository"
	"classified-listings/internal/validator"
)

type ListingService struct {
	repo repository.ListingRepository
}

func NewListingService(repo repository.ListingRepository) *ListingService {
	return &ListingService{repo: repo}
}

// validateInput runs the validator and wraps any field errors in a model.ValidationError
// so Create and Update share the same validation call-site (DRY).
func (s *ListingService) validateInput(input model.ListingInput) error {
	if errs := validator.ValidateListingInput(input); len(errs) > 0 {
		return model.ValidationError{Fields: errs}
	}
	return nil
}

// GetAll passes pagination and filters to the repository.
func (s *ListingService) GetAll(ctx context.Context, params model.PaginationParams, filter model.ListingFilter) ([]model.Listing, int64, error) {
	return s.repo.GetAll(ctx, params, filter)
}

func (s *ListingService) GetByID(ctx context.Context, id int64) (model.Listing, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ListingService) Create(ctx context.Context, input model.ListingInput) (model.Listing, error) {
	if err := s.validateInput(input); err != nil {
		return model.Listing{}, err
	}
	return s.repo.Create(ctx, input)
}

func (s *ListingService) Update(ctx context.Context, id int64, input model.ListingInput) (model.Listing, error) {
	if err := s.validateInput(input); err != nil {
		return model.Listing{}, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *ListingService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
