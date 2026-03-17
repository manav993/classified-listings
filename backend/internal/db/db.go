// Package db handles database initialisation and schema setup.
package db

import (
	"database/sql"
)

// EnsureSchema creates the listings table and indexes if they do not already
// exist. All statements are idempotent - safe to call on every startup.
// In production, a migration tool such as golang-migrate would replace this.
func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS listings (
	id          INTEGER  PRIMARY KEY AUTOINCREMENT,
	title       TEXT     NOT NULL,
	description TEXT     NOT NULL,
	price       REAL     NOT NULL,
	category    TEXT     NOT NULL,
	date_posted DATETIME NOT NULL,
	status      TEXT     NOT NULL,
	image_url   TEXT
);`)
	if err != nil {
		return err
	}

	// Indexes on the most-filtered columns so category/status/date queries
	// don't degrade into full table scans as the listings table grows.
	for _, idx := range []string{
		`CREATE INDEX IF NOT EXISTS idx_listings_category    ON listings(category)`,
		`CREATE INDEX IF NOT EXISTS idx_listings_status      ON listings(status)`,
		`CREATE INDEX IF NOT EXISTS idx_listings_date_posted ON listings(date_posted)`,
	} {
		if _, err := db.Exec(idx); err != nil {
			return err
		}
	}

	return nil
}
