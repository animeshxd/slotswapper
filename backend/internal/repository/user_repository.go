package repository

import (
	"context"

	"slotswapper/internal/db"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id int64) (db.GetUserByIDRow, error)
	GetPublicUserByID(ctx context.Context, id int64) (db.GetPublicUserByIDRow, error)
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

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (db.GetUserByIDRow, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *userRepository) GetPublicUserByID(ctx context.Context, id int64) (db.GetPublicUserByIDRow, error) {
	return r.queries.GetPublicUserByID(ctx, id)
}
