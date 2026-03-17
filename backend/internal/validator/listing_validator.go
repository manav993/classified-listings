package validator

import (
	"fmt"
	"strings"

	"classified-listings/internal/model"
)

// Field limits — keep in sync with frontend/src/constants/listing.js.
const (
	minTitleLen       = 3
	maxTitleLen       = 100
	minDescriptionLen = 20
	maxDescriptionLen = 1000
	maxPrice          = 10_000_000
)

// validCategories and validStatuses use maps so that adding a new allowed value
// requires only one addition here, without touching any conditional logic in the function.
var validCategories = map[string]bool{
	model.CategoryProperty:    true,
	model.CategoryVehicle:     true,
	model.CategoryElectronics: true,
}

var validStatuses = map[string]bool{
	model.StatusActive:   true,
	model.StatusInactive: true,
}

// formatPrice formats an integer price as a comma-separated string (e.g. 10000000 -> "10,000,000").
func formatPrice(n int) string {
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

// ValidateListingInput checks all fields and returns every validation error found.
// Collecting all errors at once lets callers show the user everything wrong in a single response.
// Returns nil if the input is valid.
func ValidateListingInput(input model.ListingInput) []model.FieldError {
	var errs []model.FieldError

	trimmedTitle := strings.TrimSpace(input.Title)
	if trimmedTitle == "" {
		errs = append(errs, model.FieldError{Field: "title", Message: "title is required"})
	} else if len(trimmedTitle) < minTitleLen {
		errs = append(errs, model.FieldError{Field: "title", Message: fmt.Sprintf("title must be at least %d characters", minTitleLen)})
	} else if len(input.Title) > maxTitleLen {
		errs = append(errs, model.FieldError{Field: "title", Message: fmt.Sprintf("title must be %d characters or fewer", maxTitleLen)})
	}

	trimmedDesc := strings.TrimSpace(input.Description)
	if trimmedDesc == "" {
		errs = append(errs, model.FieldError{Field: "description", Message: "description is required"})
	} else if len(trimmedDesc) < minDescriptionLen {
		errs = append(errs, model.FieldError{Field: "description", Message: fmt.Sprintf("description must be at least %d characters", minDescriptionLen)})
	} else if len(input.Description) > maxDescriptionLen {
		errs = append(errs, model.FieldError{Field: "description", Message: fmt.Sprintf("description must be %d characters or fewer", maxDescriptionLen)})
	}
	if input.Price <= 0 {
		errs = append(errs, model.FieldError{Field: "price", Message: "price must be greater than zero"})
	} else if input.Price > maxPrice {
		errs = append(errs, model.FieldError{Field: "price", Message: fmt.Sprintf("price must not exceed £%s", formatPrice(maxPrice))})
	}
	if !validCategories[input.Category] {
		errs = append(errs, model.FieldError{Field: "category", Message: "category must be one of Property, Vehicle, Electronics"})
	}
	if !validStatuses[input.Status] {
		errs = append(errs, model.FieldError{Field: "status", Message: "status must be one of Active, Inactive"})
	}

	return errs
}
