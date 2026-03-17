package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"classified-listings/internal/model"
)

// Compile-time check: SQLiteListingRepository must satisfy the interface.
var _ ListingRepository = (*SQLiteListingRepository)(nil)

type SQLiteListingRepository struct {
	db *sql.DB
}

func NewSQLiteListingRepository(db *sql.DB) *SQLiteListingRepository {
	return &SQLiteListingRepository{db: db}
}

// GetAll returns a page of listings and the total count, applying any provided filters.
// Filters are optional - only non-nil fields are added to the WHERE clause.
func (r *SQLiteListingRepository) GetAll(ctx context.Context, params model.PaginationParams, filter model.ListingFilter) ([]model.Listing, int64, error) {
	var (
		whereClauses []string
		args         []any
	)

	// 1=1 lets us safely append AND clauses without special-casing the first one.
	whereClauses = append(whereClauses, "1=1")

	if filter.Category != nil {
		whereClauses = append(whereClauses, "category = ?")
		args = append(args, *filter.Category)
	}

	if filter.Status != nil {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, *filter.Status)
	}

	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		// Escape LIKE metacharacters before wrapping in % so that user input like
		// "50%" or "_item" is treated as literals, not SQL wildcards.
		escaped := escapeLike(strings.ToLower(strings.TrimSpace(*filter.Search)))
		s := "%" + escaped + "%"
		whereClauses = append(whereClauses, `(lower(title) LIKE ? ESCAPE '\' OR lower(description) LIKE ? ESCAPE '\')`)
		args = append(args, s, s)
	}

	whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

	// Count query uses the same WHERE clause so total reflects filtered results.
	var total int64
	countQuery := "SELECT COUNT(*) FROM listings " + whereSQL
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Build the data query, appending LIMIT and OFFSET at the end.
	// ORDER BY date_posted DESC reflects recency; id DESC breaks ties deterministically.
	listQuery := `
SELECT id, title, description, price, category, date_posted, status, image_url
FROM listings
` + whereSQL + `
ORDER BY date_posted DESC, id DESC
LIMIT ? OFFSET ?`

	// Copy filter args then append pagination so the original slice is not aliased.
	pageArgs := make([]any, 0, len(args)+2)
	pageArgs = append(pageArgs, args...)
	pageArgs = append(pageArgs, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, listQuery, pageArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		listing, err := scanListing(rows)
		if err != nil {
			return nil, 0, err
		}
		listings = append(listings, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return listings, total, nil
}

func (r *SQLiteListingRepository) GetByID(ctx context.Context, id int64) (model.Listing, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, title, description, price, category, date_posted, status, image_url
FROM listings
WHERE id = ?`, id)

	listing, err := scanListing(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Listing{}, ErrNotFound
		}
		return model.Listing{}, err
	}
	return listing, nil
}

func (r *SQLiteListingRepository) Create(ctx context.Context, input model.ListingInput) (model.Listing, error) {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `
INSERT INTO listings (title, description, price, category, date_posted, status, image_url)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.Title,
		input.Description,
		input.Price,
		input.Category,
		now,
		input.Status,
		toNullString(input.ImageURL),
	)
	if err != nil {
		return model.Listing{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Listing{}, err
	}
	return model.Listing{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Price:       input.Price,
		Category:    input.Category,
		DatePosted:  now,
		Status:      input.Status,
		ImageURL:    strings.TrimSpace(input.ImageURL),
	}, nil
}

// Update replaces every editable field on the listing identified by id.
// date_posted is intentionally excluded from the SET clause so it always
// reflects when the listing was originally created, not when it was last edited.
func (r *SQLiteListingRepository) Update(ctx context.Context, id int64, input model.ListingInput) (model.Listing, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE listings
SET title = ?, description = ?, price = ?, category = ?, status = ?, image_url = ?
WHERE id = ?`,
		input.Title,
		input.Description,
		input.Price,
		input.Category,
		input.Status,
		toNullString(input.ImageURL),
		id,
	)
	if err != nil {
		return model.Listing{}, err
	}
	if err := requireAffected(res); err != nil {
		return model.Listing{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *SQLiteListingRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM listings WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// toNullString trims whitespace and converts an empty/whitespace string to SQL NULL.
// A non-empty trimmed value is stored as-is. This avoids repeating the inline
// struct literal in every INSERT and UPDATE that touches the nullable image_url column.
func toNullString(s string) sql.NullString {
	s = strings.TrimSpace(s)
	return sql.NullString{String: s, Valid: s != ""}
}

// escapeLike escapes the three LIKE metacharacters recognised by SQLite so that
// user-supplied search terms are matched literally rather than as patterns.
// The backslash is chosen as the escape character and must be declared with
// ESCAPE '\' in the LIKE expression.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// requireAffected returns ErrNotFound when a statement matched zero rows.
// Update and Delete share this check, so the 6-line RowsAffected pattern only
// lives in one place.
func requireAffected(res sql.Result) error {
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows,
// allowing scanListing to be reused for single and multi-row queries.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanListing(rs rowScanner) (model.Listing, error) {
	var listing model.Listing
	// image_url is nullable in the DB (the column allows NULL when no image is set).
	// sql.NullString handles the NULL -> empty-string conversion transparently.
	var imageURL sql.NullString
	err := rs.Scan(
		&listing.ID,
		&listing.Title,
		&listing.Description,
		&listing.Price,
		&listing.Category,
		&listing.DatePosted,
		&listing.Status,
		&imageURL,
	)
	if err != nil {
		return listing, err
	}
	if imageURL.Valid {
		listing.ImageURL = imageURL.String
	}
	return listing, nil
}
