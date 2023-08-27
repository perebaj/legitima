// Package mysql is responsible for storing data in the database.
package mysql

import "database/sql"

// Storage holds the database connection.
type Storage struct {
	db *sql.DB
}

// NewStorage returns a new Storage instance.
func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// SaveUser saves a user to the database.
func (s *Storage) SaveUser() error {
	return nil
}
