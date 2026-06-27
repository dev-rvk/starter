package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	authapp "github.com/starterpack/api/internal/application/auth"
	"github.com/starterpack/api/internal/domain"
	userdomain "github.com/starterpack/api/internal/domain/user"
)

type mockAuthService struct {
	RegisterFn func(ctx context.Context, in authapp.RegisterInput) (*authapp.AuthResult, error)
	LoginFn    func(ctx context.Context, in authapp.LoginInput) (*authapp.AuthResult, error)
	MeFn       func(ctx context.Context, accountID string) (*userdomain.User, error)
}

func (m *mockAuthService) Register(ctx context.Context, in authapp.RegisterInput) (*authapp.AuthResult, error) {
	return m.RegisterFn(ctx, in)
}

func (m *mockAuthService) Login(ctx context.Context, in authapp.LoginInput) (*authapp.AuthResult, error) {
	return m.LoginFn(ctx, in)
}

func (m *mockAuthService) Me(ctx context.Context, accountID string) (*userdomain.User, error) {
	return m.MeFn(ctx, accountID)
}

func setupAuthRouter(svc authapp.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler(svc)
	api := r.Group("/api/v1")
	h.registerPublic(api)
	return r
}

func setupAuthMeRouter(svc authapp.AuthService, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler(svc)
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		if userID != "" {
			c.Set("userID", userID)
		}
		c.Next()
	})
	h.registerProtected(api)
	return r
}

func TestAuthHandler_Register(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name         string
		body         string
		mockRegister func(ctx context.Context, in authapp.RegisterInput) (*authapp.AuthResult, error)
		wantStatus   int
		wantBody     string
	}{
		{
			name: "success",
			body: `{"username":"alice","email":"alice@example.com","password":"pwd"}`,
			mockRegister: func(ctx context.Context, in authapp.RegisterInput) (*authapp.AuthResult, error) {
				return &authapp.AuthResult{
					Token: "token-123",
					User: &userdomain.User{
						ID:        "user-123",
						Username:  in.Username,
						Email:     in.Email,
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, nil
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"token":"token-123"`,
		},
		{
			name:         "bad_json",
			body:         `{bad-json`,
			mockRegister: nil,
			wantStatus:   http.StatusBadRequest,
			wantBody:     `"error"`,
		},
		{
			name: "conflict",
			body: `{"username":"alice","email":"alice@example.com","password":"pwd"}`,
			mockRegister: func(ctx context.Context, in authapp.RegisterInput) (*authapp.AuthResult, error) {
				return nil, domain.AlreadyExists("account")
			},
			wantStatus: http.StatusConflict,
			wantBody:   `"error":"account already exists"`,
		},
		{
			name: "validation_error",
			body: `{"username":"a","email":"alice@example.com","password":"pwd"}`,
			mockRegister: func(ctx context.Context, in authapp.RegisterInput) (*authapp.AuthResult, error) {
				return nil, domain.ValidationError("user", "Username", "too short")
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `"error":"too short"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{RegisterFn: tt.mockRegister}
			r := setupAuthRouter(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(tt.body))
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

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		body       string
		mockLogin  func(ctx context.Context, in authapp.LoginInput) (*authapp.AuthResult, error)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"email":"alice@example.com","password":"pwd"}`,
			mockLogin: func(ctx context.Context, in authapp.LoginInput) (*authapp.AuthResult, error) {
				return &authapp.AuthResult{
					Token: "token-123",
					User: &userdomain.User{
						ID:        "user-123",
						Username:  "alice",
						Email:     in.Email,
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `"token":"token-123"`,
		},
		{
			name:       "bad_json",
			body:       `{bad-json`,
			mockLogin:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error"`,
		},
		{
			name: "invalid_credentials",
			body: `{"email":"alice@example.com","password":"pwd"}`,
			mockLogin: func(ctx context.Context, in authapp.LoginInput) (*authapp.AuthResult, error) {
				return nil, domain.NotFound("account") // handler should map NotFound to StatusUnauthorized for login
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `"error":"invalid email or password"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{LoginFn: tt.mockLogin}
			r := setupAuthRouter(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
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

func TestAuthHandler_Me(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name       string
		userID     string
		mockMe     func(ctx context.Context, accountID string) (*userdomain.User, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "success",
			userID: "user-123",
			mockMe: func(ctx context.Context, accountID string) (*userdomain.User, error) {
				return &userdomain.User{
					ID:        accountID,
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
			name:       "unauthorized",
			userID:     "",
			mockMe:     nil,
			wantStatus: http.StatusUnauthorized,
			wantBody:   `"error":"not authenticated"`,
		},
		{
			name:   "not_found",
			userID: "user-123",
			mockMe: func(ctx context.Context, accountID string) (*userdomain.User, error) {
				return nil, domain.NotFound("user")
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"user not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockAuthService{MeFn: tt.mockMe}
			r := setupAuthMeRouter(svc, tt.userID)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
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
