package services

import (
	"context"

	"slotswapper/internal/crypto"
	"slotswapper/internal/db"
	"slotswapper/internal/repository"
	"slotswapper/internal/validation"
)

// CreateUserInput defines the input for creating a user.
type CreateUserInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UserService is responsible for user-related operations.
type UserService interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*db.User, error)
	GetUserByID(ctx context.Context, id int64) (*db.GetUserByIDRow, error)
	GetPublicUserByID(ctx context.Context, id int64) (*db.GetPublicUserByIDRow, error)
}

type userService struct {
	userRepo repository.UserRepository
	password crypto.Password
}

func NewUserService(userRepo repository.UserRepository, password crypto.Password) UserService {
	return &userService{userRepo: userRepo, password: password}
}

func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (*db.User, error) {
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

func (s *userService) GetUserByID(ctx context.Context, id int64) (*db.GetUserByIDRow, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) GetPublicUserByID(ctx context.Context, id int64) (*db.GetPublicUserByIDRow, error) {
	user, err := s.userRepo.GetPublicUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
