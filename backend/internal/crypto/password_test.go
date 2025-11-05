package crypto

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	p := NewPasswordWithCost(bcrypt.MinCost) // Use minimum cost for faster tests

	t.Run("Hash and Verify", func(t *testing.T) {
		password := "my-secret-password"

		hashedPassword, err := p.Hash(password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		if hashedPassword == password {
			t.Fatal("hashed password should not be the same as the original password")
		}

		err = p.Verify(hashedPassword, password)
		if err != nil {
			t.Fatalf("failed to verify password: %v", err)
		}
	})

	t.Run("Verify_InvalidPassword", func(t *testing.T) {
		password := "my-secret-password"
		wrongPassword := "wrong-password"

		hashedPassword, err := p.Hash(password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		err = p.Verify(hashedPassword, wrongPassword)
		if err == nil {
			t.Fatal("expected an error when verifying with wrong password, got nil")
		}

		if err != bcrypt.ErrMismatchedHashAndPassword {
			t.Errorf("expected bcrypt.ErrMismatchedHashAndPassword, got %v", err)
		}
	})
}
