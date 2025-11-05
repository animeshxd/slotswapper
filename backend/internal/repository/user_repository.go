package repository

import (
	"context"

	"slotswapper/internal/db"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
}

type userRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}
