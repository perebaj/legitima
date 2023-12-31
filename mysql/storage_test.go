//go:build integration
// +build integration

package mysql_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	// mysql driver
	"github.com/birdie-ai/legitima"
	"github.com/birdie-ai/legitima/mysql"
	_ "github.com/go-sql-driver/mysql"
)

type StorageSuite struct {
	db     *sql.DB
	dbName string
}

func Setup(t *testing.T) (db *sql.DB, dbName string) {
	t.Helper()
	dbName = "legitima" + time.Now().Format("2006-01-02 15:04:05")
	db, err := sql.Open("mysql", "root:mysql@tcp(localhost:3307)/mysql")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	err = db.Ping()
	if err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	_, err = db.Exec("create database `" + dbName + "`")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	dbURL := fmt.Sprintf("root:mysql@tcp(localhost:3307)/%s", dbName)
	cfg := mysql.Config{
		URL:             dbURL,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxIdleTime: 1 * time.Minute,
	}
	db, err = mysql.OpenDB(cfg)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	return db, dbName
}

func Teardown(t *testing.T, db *sql.DB, dbName string) {
	t.Helper()
	_, err := db.Exec("drop database `" + dbName + "`")
	if err != nil {
		t.Fatalf("failed to drop database: %v", err)
	}
	t.Logf("dropped database: %s", dbName)
	defer db.Close()
}

func TestSaveUser(t *testing.T) {
	db, dbName := Setup(t)
	defer Teardown(t, db, dbName)

	storage := mysql.NewStorage(db)

	gUsr := legitima.GoogleUser{
		Name:  "JojO",
		ID:    "123",
		Email: "jojo@example.com",
	}

	err := storage.SaveUser(gUsr)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	var count int
	err = db.QueryRow("select COUNT(*) from users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to select from users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestSaveUserTwice(t *testing.T) {
	db, dbName := Setup(t)
	defer Teardown(t, db, dbName)

	storage := mysql.NewStorage(db)

	gUsr := legitima.GoogleUser{
		Name:  "JojO",
		ID:    "123",
		Email: "jojo@gmail.com",
	}

	err := storage.SaveUser(gUsr)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	err = storage.SaveUser(gUsr)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	var count int
	err = db.QueryRow("select COUNT(*) from users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to select from users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestSaveUserUpdateName(t *testing.T) {
	db, dbName := Setup(t)
	defer Teardown(t, db, dbName)

	storage := mysql.NewStorage(db)

	gUsr := legitima.GoogleUser{
		Name:  "JojO",
		ID:    "123",
		Email: "jojo@gmail.com",
	}

	gUsr2 := legitima.GoogleUser{
		Name:  "JojO2",
		ID:    "123",
		Email: "jojo@gmail.com",
	}

	err := storage.SaveUser(gUsr)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	err = storage.SaveUser(gUsr2)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	var count int
	err = db.QueryRow("select COUNT(*) from users").Scan(&count)
	if err != nil {
		t.Fatalf("failed to select from users: %v", err)
	}

	usr, err := storage.UserByEmail(gUsr.Email)
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}
	if usr.Name != gUsr2.Name {
		t.Fatalf("expected name %s, got %s", gUsr2.Name, usr.Name)
	}
}

func TestUserByEmail(t *testing.T) {
	db, dbName := Setup(t)
	defer Teardown(t, db, dbName)

	storage := mysql.NewStorage(db)

	gUsr := legitima.GoogleUser{
		Name:  "JojO",
		ID:    "123",
		Email: "jojo@gmail.com",
	}

	err := storage.SaveUser(gUsr)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	usr, err := storage.UserByEmail(gUsr.Email)
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}
	if usr.Name != gUsr.Name {
		t.Fatalf("expected name %s, got %s", gUsr.Name, usr.Name)
	}

	if usr.Email != gUsr.Email {
		t.Fatalf("expected email %s, got %s", gUsr.Email, usr.Email)
	}
}
