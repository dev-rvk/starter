package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	todoapp "github.com/starterpack/api/internal/application/todo"
	"github.com/starterpack/api/internal/domain"
	"github.com/starterpack/api/internal/domain/todo"
)

type mockTodoService struct {
	CreateFn func(ctx context.Context, title string) (*todo.Todo, error)
	GetFn    func(ctx context.Context, id string) (*todo.Todo, error)
	UpdateFn func(ctx context.Context, id, title string, completed bool) (*todo.Todo, error)
	DeleteFn func(ctx context.Context, id string) error
	ListFn   func(ctx context.Context, limit, offset int32) ([]*todo.Todo, error)
}

func (m *mockTodoService) Create(ctx context.Context, title string) (*todo.Todo, error) {
	return m.CreateFn(ctx, title)
}

func (m *mockTodoService) Get(ctx context.Context, id string) (*todo.Todo, error) {
	return m.GetFn(ctx, id)
}

func (m *mockTodoService) Update(ctx context.Context, id, title string, completed bool) (*todo.Todo, error) {
	return m.UpdateFn(ctx, id, title, completed)
}

func (m *mockTodoService) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *mockTodoService) List(ctx context.Context, limit, offset int32) ([]*todo.Todo, error) {
	return m.ListFn(ctx, limit, offset)
}

func setupTodoRouter(svc todoapp.TodoService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTodoHandler(svc)
	api := r.Group("/api/v1")
	h.register(api)
	return r
}

func TestTodoHandler_Create(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		body       string
		mockCreate func(ctx context.Context, title string) (*todo.Todo, error)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"title":"Buy milk"}`,
			mockCreate: func(ctx context.Context, title string) (*todo.Todo, error) {
				return &todo.Todo{
					ID:        "todo-123",
					Title:     title,
					Completed: false,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"title":"Buy milk"`,
		},
		{
			name:       "bad_json",
			body:       `{invalid`,
			mockCreate: nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error"`,
		},
		{
			name: "validation_error",
			body: `{"title":""}`,
			mockCreate: func(ctx context.Context, title string) (*todo.Todo, error) {
				return nil, domain.ValidationError("todo", "Title", "cannot be empty")
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `"error":"cannot be empty"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTodoService{CreateFn: tt.mockCreate}
			r := setupTodoRouter(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestTodoHandler_Get(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		todoID     string
		mockGet    func(ctx context.Context, id string) (*todo.Todo, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "success",
			todoID: "todo-123",
			mockGet: func(ctx context.Context, id string) (*todo.Todo, error) {
				return &todo.Todo{
					ID:        id,
					Title:     "Buy milk",
					Completed: false,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"Buy milk"`,
		},
		{
			name:   "not_found",
			todoID: "missing",
			mockGet: func(ctx context.Context, id string) (*todo.Todo, error) {
				return nil, todo.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"todo not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTodoService{GetFn: tt.mockGet}
			r := setupTodoRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/todos/"+tt.todoID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestTodoHandler_Update(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		todoID     string
		body       string
		mockUpdate func(ctx context.Context, id, title string, completed bool) (*todo.Todo, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "success",
			todoID: "todo-123",
			body:   `{"title":"Clean room","completed":true}`,
			mockUpdate: func(ctx context.Context, id, title string, completed bool) (*todo.Todo, error) {
				return &todo.Todo{
					ID:        id,
					Title:     title,
					Completed: completed,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `"completed":true`,
		},
		{
			name:       "bad_json",
			todoID:     "todo-123",
			body:       `{bad-json`,
			mockUpdate: nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error"`,
		},
		{
			name:   "not_found",
			todoID: "todo-missing",
			body:   `{"title":"Clean room","completed":true}`,
			mockUpdate: func(ctx context.Context, id, title string, completed bool) (*todo.Todo, error) {
				return nil, todo.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"todo not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTodoService{UpdateFn: tt.mockUpdate}
			r := setupTodoRouter(svc)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/todos/"+tt.todoID, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestTodoHandler_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		todoID     string
		mockDelete func(ctx context.Context, id string) error
		wantStatus int
	}{
		{
			name:   "success",
			todoID: "todo-123",
			mockDelete: func(ctx context.Context, id string) error {
				return nil
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "error_not_found",
			todoID: "missing",
			mockDelete: func(ctx context.Context, id string) error {
				return todo.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTodoService{DeleteFn: tt.mockDelete}
			r := setupTodoRouter(svc)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos/"+tt.todoID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestTodoHandler_List(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		query      string
		mockList   func(ctx context.Context, limit, offset int32) ([]*todo.Todo, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:  "success_empty",
			query: "",
			mockList: func(ctx context.Context, limit, offset int32) ([]*todo.Todo, error) {
				return []*todo.Todo{}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:  "success_items",
			query: "?limit=10",
			mockList: func(ctx context.Context, limit, offset int32) ([]*todo.Todo, error) {
				return []*todo.Todo{
					{
						ID:        "todo-1",
						Title:     "Todo 1",
						Completed: false,
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"Todo 1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTodoService{ListFn: tt.mockList}
			r := setupTodoRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/todos"+tt.query, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}
