// Package mysql contains the database configuration and migration logic.
package mysql

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Config holds the configuration for the database.
type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
}

// OpenDB opens a connection to the database.
func OpenDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %v", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %v", err)
	}

	return db, nil
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate runs the migrations.
func Migrate(db *sql.DB) error {
	fs, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating iofs driver: %v", err)
	}
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("creating mysql driver: %v", err)
	}
	m, err := migrate.NewWithInstance("iofs", fs, "mysql", driver)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %v", err)
	}
	return nil
}
