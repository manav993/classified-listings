package validator

import (
	"strings"
	"testing"

	"classified-listings/internal/model"
)

// repeat builds a string of n copies of s — used to construct boundary-length inputs.
func repeat(s string, n int) string { return strings.Repeat(s, n) }

// validInput returns a fully valid ListingInput baseline so each test case
// only needs to mutate the single field under test.
func validInput() model.ListingInput {
	return model.ListingInput{
		Title:       "Test Listing",
		Description: "A valid, detailed description.",
		Price:       10,
		Category:    model.CategoryVehicle,
		Status:      model.StatusActive,
	}
}

// hasFieldError reports whether errs contains an error for the given field name.
func hasFieldError(errs []model.FieldError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

// TestValidateListingInput covers every single-field rule through a table of
// mutate functions applied to the validInput baseline.
// Adding a new rule = adding one row; no new function needed.
func TestValidateListingInput(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*model.ListingInput)
		wantField string // empty = expect zero errors
	}{
		{
			name:      "valid input",
			mutate:    nil,
			wantField: "",
		},
		{
			name:      "missing title",
			mutate:    func(i *model.ListingInput) { i.Title = "" },
			wantField: "title",
		},
		{
			name:      "whitespace-only title",
			mutate:    func(i *model.ListingInput) { i.Title = "   " },
			wantField: "title",
		},
		{
			// Boundary: one character under the minimum must be rejected.
			name:      "title too short",
			mutate:    func(i *model.ListingInput) { i.Title = repeat("a", minTitleLen-1) },
			wantField: "title",
		},
		{
			// Boundary: exactly at the minimum must be accepted.
			name:      "title at min length",
			mutate:    func(i *model.ListingInput) { i.Title = repeat("a", minTitleLen) },
			wantField: "",
		},
		{
			name:      "missing description",
			mutate:    func(i *model.ListingInput) { i.Description = "" },
			wantField: "description",
		},
		{
			name:      "whitespace-only description",
			mutate:    func(i *model.ListingInput) { i.Description = "   " },
			wantField: "description",
		},
		{
			// Boundary: one character under the minimum must be rejected.
			name:      "description too short",
			mutate:    func(i *model.ListingInput) { i.Description = repeat("a", minDescriptionLen-1) },
			wantField: "description",
		},
		{
			// Boundary: exactly at the minimum must be accepted.
			name:      "description at min length",
			mutate:    func(i *model.ListingInput) { i.Description = repeat("a", minDescriptionLen) },
			wantField: "",
		},
		{
			name:      "negative price",
			mutate:    func(i *model.ListingInput) { i.Price = -1 },
			wantField: "price",
		},
		{
			// Boundary: the rule is price > 0, so exactly 0 must be rejected.
			name:      "zero price boundary",
			mutate:    func(i *model.ListingInput) { i.Price = 0 },
			wantField: "price",
		},
		{
			// Boundary: 0.01 is the smallest strictly-positive value - must be accepted.
			name:      "minimum valid price (0.01)",
			mutate:    func(i *model.ListingInput) { i.Price = 0.01 },
			wantField: "",
		},
		{
			// Boundary: exactly at the cap must pass.
			name:      "price at max",
			mutate:    func(i *model.ListingInput) { i.Price = maxPrice },
			wantField: "",
		},
		{
			// One cent over the cap must be rejected.
			name:      "price over max",
			mutate:    func(i *model.ListingInput) { i.Price = maxPrice + 0.01 },
			wantField: "price",
		},
		{
			// Boundary: exactly at the limit must pass.
			name:      "title at max length",
			mutate:    func(i *model.ListingInput) { i.Title = repeat("a", maxTitleLen) },
			wantField: "",
		},
		{
			// One character over the limit must be rejected.
			name:      "title over max length",
			mutate:    func(i *model.ListingInput) { i.Title = repeat("a", maxTitleLen+1) },
			wantField: "title",
		},
		{
			// Boundary: exactly at the limit must pass.
			name:      "description at max length",
			mutate:    func(i *model.ListingInput) { i.Description = repeat("a", maxDescriptionLen) },
			wantField: "",
		},
		{
			// One character over the limit must be rejected.
			name:      "description over max length",
			mutate:    func(i *model.ListingInput) { i.Description = repeat("a", maxDescriptionLen+1) },
			wantField: "description",
		},
		{
			name:      "invalid category",
			mutate:    func(i *model.ListingInput) { i.Category = "Cars" },
			wantField: "category",
		},
		{
			name:      "invalid status",
			mutate:    func(i *model.ListingInput) { i.Status = "Unknown" },
			wantField: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput()
			if tt.mutate != nil {
				tt.mutate(&input)
			}

			errs := ValidateListingInput(input)

			if tt.wantField == "" {
				if len(errs) != 0 {
					t.Errorf("expected no errors, got %v", errs)
				}
				return
			}

			if len(errs) == 0 {
				t.Fatalf("expected validation error on field %q, got none", tt.wantField)
			}
			if !hasFieldError(errs, tt.wantField) {
				t.Errorf("expected error on field %q, got %v", tt.wantField, errs)
			}
		})
	}
}

// TestValidateListingInput_AllFieldsInvalid verifies that all errors are
// collected and returned together rather than stopping at the first failure.
func TestValidateListingInput_AllFieldsInvalid(t *testing.T) {
	errs := ValidateListingInput(model.ListingInput{Price: -5})
	if len(errs) < 4 {
		t.Errorf("expected at least 4 errors for fully invalid input, got %d: %v", len(errs), errs)
	}
}
