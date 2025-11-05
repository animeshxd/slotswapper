package crypto

import (
	"testing"
	"time"
)

func TestJWT(t *testing.T) {
	secret := "my-super-secret-key"
	j := NewJWT(secret, time.Hour)

	t.Run("Generate and Verify", func(t *testing.T) {
		userID := int64(123)

		token, err := j.Generate(userID)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		if token == "" {
			t.Fatal("generated token should not be empty")
		}

		verifiedUserID, err := j.Verify(token)
		if err != nil {
			t.Fatalf("failed to verify token: %v", err)
		}

		if verifiedUserID != userID {
			t.Errorf("expected user ID %d, got %d", userID, verifiedUserID)
		}
	})

	t.Run("Verify_InvalidToken", func(t *testing.T) {
		_, err := j.Verify("invalid-token")
		if err == nil {
			t.Fatal("expected an error when verifying invalid token, got nil")
		}
	})

	t.Run("Verify_ExpiredToken", func(t *testing.T) {
		jExpired := NewJWT(secret, -time.Hour) // Expired token
		userID := int64(456)

		expiredToken, err := jExpired.Generate(userID)
		if err != nil {
			t.Fatalf("failed to generate expired token: %v", err)
		}

		_, err = j.Verify(expiredToken)
		if err == nil {
			t.Fatal("expected an error when verifying expired token, got nil")
		}
	})
}
