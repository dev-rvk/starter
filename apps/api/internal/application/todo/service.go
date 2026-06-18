package todoapp

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/starterpack/api/internal/domain/todo"
	"github.com/starterpack/api/internal/platform/validator"
)

type Service struct {
	repo todo.Repository
}

func NewService(repo todo.Repository) *Service{
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, title string) (*todo.Todo, error) {
	task := todo.New(uuid.NewString(), title, false, time.Now())
	if err := validator.ValidateAndMap("todo", task); err != nil {
		
		return nil, err
	}
	if err := s.repo.Create(ctx, task); err != nil{
		return nil, err
	}
	return task, nil
}