// Package mysql is responsible for storing data in the database.
package mysql

import (
	"database/sql"
	"fmt"

	"github.com/birdie-ai/legitima/api"
)

// Storage holds the database connection.
type Storage struct {
	db *sql.DB
}

// NewStorage returns a new Storage instance.
func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// SaveUser saves a user to the database.
func (s *Storage) SaveUser(gUsr api.GoogleUser) error {
	usr := newUser(gUsr)

	_, err := s.db.Exec(`INSERT INTO users (id, name, email) 
		VALUES (?, ?, ?)
		`, usr.ID, usr.Name, usr.Email)

	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}

	return nil
}

// UserByEmail returns a user from the database filtered by email.
func (s *Storage) UserByEmail(token *api.Token) (User, error) {
	var usr User

	err := s.db.QueryRow(`SELECT id, name, email FROM users WHERE email = ?`, token.Email).Scan(&usr.ID, &usr.Name, &usr.Email)
	if err != nil {
		return usr, fmt.Errorf("get user: %w", err)
	}

	return usr, nil
}
