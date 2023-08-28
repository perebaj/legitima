package mysql

import (
	"github.com/birdie-ai/legitima/api"
	"github.com/google/uuid"
)

// User represents a user in the database.
type User struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func newUser(gUsr api.GoogleUser) (u *User) {
	return &User{
		ID:    uuid.New().String(),
		Name:  gUsr.Name,
		Email: gUsr.Email,
	}
}
