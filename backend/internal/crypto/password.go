package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

type Password interface {
	Hash(password string) (string, error)
	Verify(hashedPassword, password string) error
}

type password struct {
	cost int
}

func NewPassword() Password {
	return &password{cost: bcrypt.DefaultCost}
}

func NewPasswordWithCost(cost int) Password {
	return &password{cost: cost}
}

func (p *password) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	return string(bytes), err
}

func (p *password) Verify(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
