package services

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"slotswapper/internal/crypto"
	"slotswapper/internal/repository"
)

func TestUserService(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		testQueries := repository.SetupTestDB(t)
		userRepo := repository.NewUserRepository(testQueries)
		passwordCrypto := crypto.NewPassword()
		userService := NewUserService(userRepo, passwordCrypto)

		password := "password123"
		input := CreateUserInput{
			Name:     "test user",
			Email:    "service-test@example.com",
			Password: password,
		}

		user, err := userService.CreateUser(context.Background(), input)
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
		userService := NewUserService(userRepo, passwordCrypto)

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
				_, err := userService.CreateUser(context.Background(), tc.input)
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
		userService := NewUserService(userRepo, passwordCrypto)

		// First create a user
		arg1 := CreateUserInput{
			Name:     "test user 2",
			Email:    "service-test2@example.com",
			Password: "password",
		}
		_, err := userService.CreateUser(context.Background(), arg1)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Then try to create another user with the same email
		arg2 := CreateUserInput{
			Name:     "test user 3",
			Email:    "service-test2@example.com",
			Password: "password",
		}

		_, err = userService.CreateUser(context.Background(), arg2)
		if err == nil {
			t.Fatal("expected an error when creating user with duplicate email, got nil")
		}
		// SQLite returns a specific error for unique constraint violation
		if err.Error() != "UNIQUE constraint failed: users.email" {
			t.Errorf("expected unique constraint error, got %v", err)
		}
	})

	t.Run("GetUserByID", func(t *testing.T) {
		testQueries := repository.SetupTestDB(t)
		userRepo := repository.NewUserRepository(testQueries)
		passwordCrypto := crypto.NewPassword()
		userService := NewUserService(userRepo, passwordCrypto)

		createInput := CreateUserInput{
			Name:     "user for get by id",
			Email:    "getbyid@example.com",
			Password: "password",
		}
		createdUser, err := userService.CreateUser(context.Background(), createInput)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		retrievedUser, err := userService.GetUserByID(context.Background(), createdUser.ID)
		if err != nil {
			t.Fatalf("failed to get user by ID: %v", err)
		}

		if retrievedUser == nil {
			t.Error("expected user to be non-nil")
		}
		if retrievedUser.ID != createdUser.ID {
			t.Errorf("expected user ID %d, got %d", createdUser.ID, retrievedUser.ID)
		}
		if retrievedUser.Email != createdUser.Email {
			t.Errorf("expected user email %q, got %q", createdUser.Email, retrievedUser.Email)
		}
		// Ensure password is not returned
		// This is implicitly tested by the type db.GetUserByIDRow not having a Password field
	})

	t.Run("GetPublicUserByID", func(t *testing.T) {
		testQueries := repository.SetupTestDB(t)
		userRepo := repository.NewUserRepository(testQueries)
		passwordCrypto := crypto.NewPassword()
		userService := NewUserService(userRepo, passwordCrypto)

		createInput := CreateUserInput{
			Name:     "public user for get by id",
			Email:    "publicgetbyid@example.com",
			Password: "password",
		}
		createdUser, err := userService.CreateUser(context.Background(), createInput)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		retrievedUser, err := userService.GetPublicUserByID(context.Background(), createdUser.ID)
		if err != nil {
			t.Fatalf("failed to get public user by ID: %v", err)
		}

		if retrievedUser == nil {
			t.Error("expected user to be non-nil")
		}
		if retrievedUser.ID != createdUser.ID {
			t.Errorf("expected user ID %d, got %d", createdUser.ID, retrievedUser.ID)
		}
		if retrievedUser.Name != createdUser.Name {
			t.Errorf("expected user name %q, got %q", createdUser.Name, retrievedUser.Name)
		}
		// Verify that the returned struct does not contain the email field
		// This is implicitly tested by the type db.GetPublicUserByIDRow not having an Email field
	})
}
