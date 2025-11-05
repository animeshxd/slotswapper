package crypto

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT interface {
	Generate(userID int64) (string, error)
	Verify(tokenString string) (int64, error)
}

type jwtManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWT(secret string, ttl time.Duration) JWT {
	return &jwtManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (j *jwtManager) Generate(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(j.ttl).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.secret)
}

func (j *jwtManager) Verify(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub, ok := claims["sub"].(float64); ok {
			return int64(sub), nil
		}
	}

	return 0, errors.New("invalid token")
}
