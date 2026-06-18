// Package userapp holds the user use cases (application layer). It orchestrates
// the domain and depends only on the domain's ports — not on HTTP or SQL.
package userapp

import (
	"context"
	"time"

	"github.com/google/uuid"
	userdomain "github.com/starterpack/api/internal/domain/user"
	"github.com/starterpack/api/internal/platform/validator"
)

// Service implements the user use cases.
type Service struct {
	repo userdomain.Repository
}

// NewService wires a use-case service to a repository port.
func NewService(repo userdomain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput is the raw (unvalidated) input for creating a user.
type CreateInput struct {
	Username string
	Email    string
}

// Create validates input using ValidateAndMap, then persists the user.
func (s *Service) Create(ctx context.Context, in CreateInput) (*userdomain.User, error) {
	u := userdomain.New(uuid.NewString(), in.Username, in.Email, time.Now().UTC())
	if err := validator.ValidateAndMap("user", u); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Get returns a single user by id.
func (s *Service) Get(ctx context.Context, id string) (*userdomain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a page of users with sane bounds.
func (s *Service) List(ctx context.Context, limit, offset int32) ([]*userdomain.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}
