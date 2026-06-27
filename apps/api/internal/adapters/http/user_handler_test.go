package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	userapp "github.com/starterpack/api/internal/application/user"
	"github.com/starterpack/api/internal/domain"
	userdomain "github.com/starterpack/api/internal/domain/user"
)

type mockUserService struct {
	CreateFn func(ctx context.Context, in userapp.CreateInput) (*userdomain.User, error)
	GetFn    func(ctx context.Context, id string) (*userdomain.User, error)
	ListFn   func(ctx context.Context, limit, offset int32) ([]*userdomain.User, error)
}

func (m *mockUserService) Create(ctx context.Context, in userapp.CreateInput) (*userdomain.User, error) {
	return m.CreateFn(ctx, in)
}

func (m *mockUserService) Get(ctx context.Context, id string) (*userdomain.User, error) {
	return m.GetFn(ctx, id)
}

func (m *mockUserService) List(ctx context.Context, limit, offset int32) ([]*userdomain.User, error) {
	return m.ListFn(ctx, limit, offset)
}

func setupUserRouter(svc userapp.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUserHandler(svc)
	api := r.Group("/api/v1")
	h.register(api)
	return r
}

func TestUserHandler_Create(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		body       string
		mockCreate func(ctx context.Context, in userapp.CreateInput) (*userdomain.User, error)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"username":"alice","email":"alice@example.com"}`,
			mockCreate: func(ctx context.Context, in userapp.CreateInput) (*userdomain.User, error) {
				return &userdomain.User{
					ID:        "user-id-123",
					Username:  in.Username,
					Email:     in.Email,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":"user-id-123"`,
		},
		{
			name:       "bad_json",
			body:       `{invalid-json`,
			mockCreate: nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error"`,
		},
		{
			name: "conflict_already_exists",
			body: `{"username":"alice","email":"alice@example.com"}`,
			mockCreate: func(ctx context.Context, in userapp.CreateInput) (*userdomain.User, error) {
				return nil, domain.AlreadyExists("user")
			},
			wantStatus: http.StatusConflict,
			wantBody:   `"error":"user already exists"`,
		},
		{
			name: "validation_error",
			body: `{"username":"a","email":"alice@example.com"}`,
			mockCreate: func(ctx context.Context, in userapp.CreateInput) (*userdomain.User, error) {
				return nil, domain.ValidationError("user", "Username", "too short")
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `"error":"too short"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{CreateFn: tt.mockCreate}
			r := setupUserRouter(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.body))
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

func TestUserHandler_Get(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		userID     string
		mockGet    func(ctx context.Context, id string) (*userdomain.User, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "success",
			userID: "user-id-123",
			mockGet: func(ctx context.Context, id string) (*userdomain.User, error) {
				return &userdomain.User{
					ID:        id,
					Username:  "alice",
					Email:     "alice@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `"username":"alice"`,
		},
		{
			name:   "not_found",
			userID: "missing",
			mockGet: func(ctx context.Context, id string) (*userdomain.User, error) {
				return nil, domain.NotFound("user")
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"user not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{GetFn: tt.mockGet}
			r := setupUserRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID, nil)
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

func TestUserHandler_List(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		query      string
		mockList   func(ctx context.Context, limit, offset int32) ([]*userdomain.User, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:  "success_empty",
			query: "",
			mockList: func(ctx context.Context, limit, offset int32) ([]*userdomain.User, error) {
				return []*userdomain.User{}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:  "success_items",
			query: "?limit=5&offset=2",
			mockList: func(ctx context.Context, limit, offset int32) ([]*userdomain.User, error) {
				if limit != 5 || offset != 2 {
					return nil, domain.ValidationError("user", "limit", "limit/offset mismatch")
				}
				return []*userdomain.User{
					{
						ID:        "user-1",
						Username:  "alice",
						Email:     "alice@example.com",
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `"username":"alice"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{ListFn: tt.mockList}
			r := setupUserRouter(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users"+tt.query, nil)
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

func TestParseInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		fallback int32
		want     int32
	}{
		{
			name:     "empty",
			raw:      "",
			fallback: 20,
			want:     20,
		},
		{
			name:     "valid",
			raw:      "42",
			fallback: 20,
			want:     42,
		},
		{
			name:     "invalid",
			raw:      "abc",
			fallback: 20,
			want:     20,
		},
		{
			name:     "negative",
			raw:      "-5",
			fallback: 20,
			want:     -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInt32(tt.raw, tt.fallback)
			if got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}
}
