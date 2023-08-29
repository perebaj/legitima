package mysql

import (
	"github.com/birdie-ai/legitima"
	"github.com/google/uuid"
)

// User represents a user in the database.
type User struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func newUser(gUsr legitima.GoogleUser) (u *User) {
	return &User{
		ID:    uuid.New().String(),
		Name:  gUsr.Name,
		Email: gUsr.Email,
	}
}

func (uDB *User) Convert() legitima.User {
	return legitima.User{
		ID:    uDB.ID,
		Name:  uDB.Name,
		Email: uDB.Email,
	}
}

// To reproduce the tests and lint, just run respectively: `make test` and `make lint`.
