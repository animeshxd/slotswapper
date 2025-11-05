package services

import (
	"context"
	"testing"
	"time"

	"slotswapper/internal/crypto"
	"slotswapper/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func TestAuthService(t *testing.T) {
	jwtSecret := "test-jwt-secret"
	jwtTTL := time.Minute

	t.Run("Register and Login", func(t *testing.T) {
		testQueries := repository.SetupTestDB(t)
		userRepo := repository.NewUserRepository(testQueries)
		passwordCrypto := crypto.NewPassword()
		jwtManager := crypto.NewJWT(jwtSecret, jwtTTL)
		authService := NewAuthService(userRepo, passwordCrypto, jwtManager)

		password := "password123"
		registerInput := RegisterUserInput{
			Name:     "test user",
			Email:    "auth-test@example.com",
			Password: password,
		}

		user, token, err := authService.Register(context.Background(), registerInput)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		if user == nil {
			t.Fatal("expected user to be non-nil")
		}

		if user.ID == 0 {
			t.Error("expected user ID to be non-zero")
		}

		if token == "" {
			t.Error("expected token to be non-empty")
		}

		verifiedUserID, err := jwtManager.Verify(token)
		if err != nil {
			t.Fatalf("failed to verify token after registration: %v", err)
		}
		if verifiedUserID != user.ID {
			t.Errorf("expected verified user ID %d, got %d", user.ID, verifiedUserID)
		}

		loginInput := LoginInput{
			Email:    "auth-test@example.com",
			Password: password,
		}

		loggedInUser, loggedInToken, err := authService.Login(context.Background(), loginInput)
		if err != nil {
			t.Fatalf("failed to login user: %v", err)
		}

		if loggedInUser == nil {
			t.Fatal("expected logged in user to be non-nil")
		}

		if loggedInUser.ID != user.ID {
			t.Errorf("expected logged in user ID to be %d, got %d", user.ID, loggedInUser.ID)
		}

		if loggedInToken == "" {
			t.Error("expected logged in token to be non-empty")
		}

		verifiedLoggedInUserID, err := jwtManager.Verify(loggedInToken)
		if err != nil {
			t.Fatalf("failed to verify token after login: %v", err)
		}
		if verifiedLoggedInUserID != loggedInUser.ID {
			t.Errorf("expected verified logged in user ID %d, got %d", loggedInUser.ID, verifiedLoggedInUserID)
		}

		incorrectLoginInput := LoginInput{
			Email:    "auth-test@example.com",
			Password: "wrongpassword",
		}

		_, _, err = authService.Login(context.Background(), incorrectLoginInput)
		if err == nil {
			t.Fatal("expected an error for incorrect password, got nil")
		}

		if err.Error() != "invalid email or password" {
			t.Errorf("expected 'invalid email or password' error, got %v", err)
		}
	})
}
