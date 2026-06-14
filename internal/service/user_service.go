package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/asimar007/userapi/db/sqlc"
	"github.com/asimar007/userapi/internal/models"
	"github.com/asimar007/userapi/internal/repository"
)

// ErrNotFound is re-exported so handlers depend on the service layer only.
var ErrNotFound = repository.ErrNotFound

// UserService contains the business logic for user operations.
type UserService struct {
	repo repository.UserRepository
	log  *zap.Logger
	now  func() time.Time // injectable clock for testability
}

// NewUserService constructs a UserService.
func NewUserService(repo repository.UserRepository, log *zap.Logger) *UserService {
	return &UserService{
		repo: repo,
		log:  log,
		now:  time.Now,
	}
}

// Create stores a new user and returns the persisted record (without age).
func (s *UserService) Create(ctx context.Context, name, dobStr string) (models.UserResponse, error) {
	dob, err := time.Parse(models.DateLayout, dobStr)
	if err != nil {
		return models.UserResponse{}, err
	}
	u, err := s.repo.Create(ctx, name, dob)
	if err != nil {
		s.log.Error("failed to create user", zap.Error(err))
		return models.UserResponse{}, err
	}
	s.log.Info("user created", zap.Int32("id", u.ID), zap.String("name", u.Name))
	return toUserResponse(u), nil
}

// GetByID fetches a single user including the dynamically calculated age.
func (s *UserService) GetByID(ctx context.Context, id int32) (models.UserWithAgeResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.UserWithAgeResponse{}, err
	}
	return s.toUserWithAge(u), nil
}

// List returns a page of users (each with age) and the total count.
func (s *UserService) List(ctx context.Context, limit, offset int32) ([]models.UserWithAgeResponse, int64, error) {
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.log.Error("failed to list users", zap.Error(err))
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		s.log.Error("failed to count users", zap.Error(err))
		return nil, 0, err
	}
	out := make([]models.UserWithAgeResponse, 0, len(users))
	for _, u := range users {
		out = append(out, s.toUserWithAge(u))
	}
	return out, total, nil
}

// Update modifies an existing user and returns the new record (without age).
func (s *UserService) Update(ctx context.Context, id int32, name, dobStr string) (models.UserResponse, error) {
	dob, err := time.Parse(models.DateLayout, dobStr)
	if err != nil {
		return models.UserResponse{}, err
	}
	u, err := s.repo.Update(ctx, id, name, dob)
	if err != nil {
		return models.UserResponse{}, err
	}
	s.log.Info("user updated", zap.Int32("id", u.ID))
	return toUserResponse(u), nil
}

// Delete removes a user by id.
func (s *UserService) Delete(ctx context.Context, id int32) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("failed to delete user", zap.Int32("id", id), zap.Error(err))
		return err
	}
	s.log.Info("user deleted", zap.Int32("id", id))
	return nil
}

func toUserResponse(u sqlc.User) models.UserResponse {
	return models.UserResponse{
		ID:   u.ID,
		Name: u.Name,
		DOB:  u.Dob.Format(models.DateLayout),
	}
}

func (s *UserService) toUserWithAge(u sqlc.User) models.UserWithAgeResponse {
	return models.UserWithAgeResponse{
		ID:   u.ID,
		Name: u.Name,
		DOB:  u.Dob.Format(models.DateLayout),
		Age:  models.CalculateAge(u.Dob, s.now()),
	}
}
