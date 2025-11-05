package services

import (
	"context"

	"slotswapper/internal/crypto"
	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/validation"
)

type CreateUserInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserCreator interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*db.User, error)
}

type userCreator struct {
	userRepo repository.UserRepository
	password crypto.Password
}

func NewUserCreator(userRepo repository.UserRepository, password crypto.Password) UserCreator {
	return &userCreator{userRepo: userRepo, password: password}
}

func (s *userCreator) CreateUser(ctx context.Context, input CreateUserInput) (*db.User, error) {
	if err := validation.Validate.Struct(input); err != nil {
		return nil, err
	}

	hashedPassword, err := s.password.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	arg := db.CreateUserParams{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
	}

	user, err := s.userRepo.CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
