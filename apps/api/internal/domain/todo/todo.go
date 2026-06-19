package todo

import "time"

// Todo is the domain entity.
type Todo struct {
	ID        string    `validate:"required"`
	Title     string    `validate:"required,min=1,max=50"`
	Completed bool
	CreatedAt time.Time `validate:"required"`
	UpdatedAt time.Time `validate:"required"`
}

// New builds a Todo.
func New(id, title string, completed bool, now time.Time) *Todo {
	return &Todo{
		ID:        id,
		Title:     title,
		Completed: completed,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
