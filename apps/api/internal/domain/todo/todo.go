package todo

import (
	"time"
)

type Todo struct {
	ID	string `validate:"required"`
	Title string `validate:"required,min=1,max=50"`
	Completed bool 
	CreatedAt time.Time `validate:"required"`
	UpdatedAt time.Time `validate:"required"`
}

func New(id string, title string, completed bool, now time.Time ) *Todo {
	return &Todo{
		ID: id,
		Title: title,
		Completed: completed,
		CreatedAt: now,
		UpdatedAt: now,
	}
}