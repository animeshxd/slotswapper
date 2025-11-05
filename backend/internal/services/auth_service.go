package services

import (
	"context"
	"errors"

	"slotswapper/internal/crypto"
	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/validation"
)

type RegisterUserInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthService interface {
	Register(ctx context.Context, input RegisterUserInput) (*db.User, string, error)
	Login(ctx context.Context, input LoginInput) (*db.User, string, error)
}

type authService struct {
	userRepo   repository.UserRepository
	password   crypto.Password
	jwtManager crypto.JWT
}

func NewAuthService(userRepo repository.UserRepository, password crypto.Password, jwtManager crypto.JWT) AuthService {
	return &authService{userRepo: userRepo, password: password, jwtManager: jwtManager}
}

func (s *authService) Register(ctx context.Context, input RegisterUserInput) (*db.User, string, error) {
	if err := validation.Validate.Struct(input); err != nil {
		return nil, "", err
	}

	hashedPassword, err := s.password.Hash(input.Password)
	if err != nil {
		return nil, "", err
	}

	arg := db.CreateUserParams{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
	}

	user, err := s.userRepo.CreateUser(ctx, arg)
	if err != nil {
		return nil, "", err
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return nil, "", err
	}
	user.Password = ""
	return &user, token, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (*db.User, string, error) {
	if err := validation.Validate.Struct(input); err != nil {
		return nil, "", err
	}

	user, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	if err := s.password.Verify(user.Password, input.Password); err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return nil, "", err
	}
	user.Password = ""
	return &user, token, nil
}
