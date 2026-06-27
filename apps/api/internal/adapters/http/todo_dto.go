package http

import (
	"time"

	"github.com/starterpack/api/internal/domain/todo"
)

// createTodoRequest is the inbound payload for creating a todo.
type createTodoRequest struct {
	Title string `json:"title"`
}

// updateTodoRequest is the inbound payload for updating a todo.
type updateTodoRequest struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// todoResponse is the outbound representation of a todo.
type todoResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toTodoResponse(t *todo.Todo) todoResponse {
	return todoResponse{
		ID:        t.ID,
		Title:     t.Title,
		Completed: t.Completed,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
