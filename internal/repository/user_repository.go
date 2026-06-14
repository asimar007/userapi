package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/userapi/db/sqlc"
)

// ErrNotFound is returned when a requested user does not exist.
var ErrNotFound = errors.New("user not found")

// UserRepository abstracts persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error)
	GetByID(ctx context.Context, id int32) (sqlc.User, error)
	List(ctx context.Context, limit, offset int32) ([]sqlc.User, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error)
	Delete(ctx context.Context, id int32) error
}

type userRepository struct {
	q *sqlc.Queries
}

// NewUserRepository builds a UserRepository backed by a pgx connection pool.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{q: sqlc.New(pool)}
}

func (r *userRepository) Create(ctx context.Context, name string, dob time.Time) (sqlc.User, error) {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{Name: name, Dob: dob})
}

func (r *userRepository) GetByID(ctx context.Context, id int32) (sqlc.User, error) {
	u, err := r.q.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, ErrNotFound
		}
		return sqlc.User{}, err
	}
	return u, nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int32) ([]sqlc.User, error) {
	return r.q.ListUsers(ctx, sqlc.ListUsersParams{Limit: limit, Offset: offset})
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}

func (r *userRepository) Update(ctx context.Context, id int32, name string, dob time.Time) (sqlc.User, error) {
	u, err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{ID: id, Name: name, Dob: dob})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, ErrNotFound
		}
		return sqlc.User{}, err
	}
	return u, nil
}

func (r *userRepository) Delete(ctx context.Context, id int32) error {
	return r.q.DeleteUser(ctx, id)
}
