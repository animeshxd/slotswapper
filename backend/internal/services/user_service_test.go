package services

import (
	"context"
	"testing"

	"slotswapper/internal/crypto"
	"slotswapper/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func TestUserCreator(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		testQueries := repository.SetupTestDB(t)
		userRepo := repository.NewUserRepository(testQueries)
		passwordCrypto := crypto.NewPassword()
		userCreator := NewUserCreator(userRepo, passwordCrypto)

		password := "password123"
		input := CreateUserInput{
			Name:     "test user",
			Email:    "service-test@example.com",
			Password: password,
		}

		user, err := userCreator.CreateUser(context.Background(), input)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		if user == nil {
			t.Error("expected user to be non-nil")
		}

		if user.ID == 0 {
			t.Error("expected user ID to be non-zero")
		}

		if user.Name != input.Name {
			t.Errorf("expected user name to be %q, got %q", input.Name, user.Name)
		}

		if user.Email != input.Email {
			t.Errorf("expected user email to be %q, got %q", input.Email, user.Email)
		}

		err = passwordCrypto.Verify(user.Password, password)
		if err != nil {
			t.Errorf("password should be hashed correctly")
		}
	})

	t.Run("CreateUser_ValidationErrors", func(t *testing.T) {
		testQueries := repository.SetupTestDB(t)
		userRepo := repository.NewUserRepository(testQueries)
		passwordCrypto := crypto.NewPassword()
		userCreator := NewUserCreator(userRepo, passwordCrypto)

		testCases := []struct {
			name  string
			input CreateUserInput
		}{
			{
				name: "missing name",
				input: CreateUserInput{
					Email:    "test@example.com",
					Password: "password123",
				},
			},
			{
				name: "missing email",
				input: CreateUserInput{
					Name:     "test user",
					Password: "password123",
				},
			},
			{
				name: "invalid email",
				input: CreateUserInput{
					Name:     "test user",
					Email:    "invalid-email",
					Password: "password123",
				},
			},
			{
				name: "short password",
				input: CreateUserInput{
					Name:     "test user",
					Email:    "test@example.com",
					Password: "123",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := userCreator.CreateUser(context.Background(), tc.input)
				if err == nil {
					t.Fatal("expected a validation error, got nil")
				}
			})
		}
	})

	t.Run("CreateUser_DuplicateEmail", func(t *testing.T) {
		testQueries := repository.SetupTestDB(t)
		userRepo := repository.NewUserRepository(testQueries)
		passwordCrypto := crypto.NewPassword()
		userCreator := NewUserCreator(userRepo, passwordCrypto)

		arg1 := CreateUserInput{
			Name:     "test user 2",
			Email:    "service-test2@example.com",
			Password: "password",
		}
		_, err := userCreator.CreateUser(context.Background(), arg1)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		arg2 := CreateUserInput{
			Name:     "test user 3",
			Email:    "service-test2@example.com",
			Password: "password",
		}

		_, err = userCreator.CreateUser(context.Background(), arg2)
		if err == nil {
			t.Fatal("expected an error when creating user with duplicate email, got nil")
		}
		if err.Error() != "UNIQUE constraint failed: users.email" {
			t.Errorf("expected unique constraint error, got %v", err)
		}
	})
}
