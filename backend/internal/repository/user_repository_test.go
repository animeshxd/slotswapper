package repository

import (
	"context"
	"database/sql"
	"testing"

	"slotswapper/internal/db"
)

func TestUserRepository(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		testQueries := SetupTestDB(t)
		userRepo := NewUserRepository(testQueries)

		arg := db.CreateUserParams{
			Name:     "test user",
			Email:    "test@example.com",
			Password: "password",
		}

		user, err := userRepo.CreateUser(context.Background(), arg)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		if user.ID == 0 {
			t.Error("expected user ID to be non-zero")
		}

		if user.Name != arg.Name {
			t.Errorf("expected user name to be %q, got %q", arg.Name, user.Name)
		}

		if user.Email != arg.Email {
			t.Errorf("expected user email to be %q, got %q", arg.Email, user.Email)
		}
	})

	t.Run("CreateUser_DuplicateEmail", func(t *testing.T) {
		testQueries := SetupTestDB(t)
		userRepo := NewUserRepository(testQueries)

		arg := db.CreateUserParams{
			Name:     "duplicate user",
			Email:    "test@example.com", // Use the same email as a previously created user
			Password: "password",
		}

		_, err := userRepo.CreateUser(context.Background(), arg)
		if err != nil {
			t.Fatalf("failed to create initial user: %v", err)
		}

		_, err = userRepo.CreateUser(context.Background(), arg)
		if err == nil {
			t.Fatal("expected an error when creating user with duplicate email, got nil")
		}
		if err.Error() != "UNIQUE constraint failed: users.email" {
			t.Errorf("expected unique constraint error, got %v", err)
		}
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		testQueries := SetupTestDB(t)
		userRepo := NewUserRepository(testQueries)

		createArg := db.CreateUserParams{
			Name:     "test user 2",
			Email:    "test2@example.com",
			Password: "password",
		}
		_, err := userRepo.CreateUser(context.Background(), createArg)
		if err != nil {
			t.Fatalf("failed to create user for testing GetUserByEmail: %v", err)
		}

		user, err := userRepo.GetUserByEmail(context.Background(), "test2@example.com")
		if err != nil {
			t.Fatalf("failed to get user by email: %v", err)
		}

		if user.ID == 0 {
			t.Error("expected user ID to be non-zero")
		}

		if user.Email != "test2@example.com" {
			t.Errorf("expected user email to be %q, got %q", "test2@example.com", user.Email)
		}
	})

	t.Run("GetUserByEmail_NotFound", func(t *testing.T) {
		testQueries := SetupTestDB(t)
		userRepo := NewUserRepository(testQueries)

		_, err := userRepo.GetUserByEmail(context.Background(), "nonexistent@example.com")
		if err == nil {
			t.Fatal("expected an error when getting non-existent user, got nil")
		}
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}
