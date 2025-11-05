package repository

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	"slotswapper/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

func SetupTestDB(t *testing.T) *db.Queries {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	dbConn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	migrations, err := os.ReadFile("../../db/migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatalf("failed to read migrations: %v", err)
	}

	_, err = dbConn.Exec(string(migrations))
	if err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		dbConn.Close()
	})

	return db.New(dbConn)
}

func SetupTestDBWithUser(t *testing.T) (*db.Queries, db.User) {
	queries := SetupTestDB(t)

	user, err := queries.CreateUser(context.Background(), db.CreateUserParams{
		Name:     "test user",
		Email:    "test-user@example.com",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("failed to create user for event tests: %v", err)
	}

	return queries, user
}
