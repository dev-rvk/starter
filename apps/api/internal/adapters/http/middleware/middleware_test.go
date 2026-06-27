package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/starterpack/api/internal/platform/jwtutil"
)

func TestBearerToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "valid bearer",
			header: "Bearer some-token-value",
			want:   "some-token-value",
		},
		{
			name:   "case insensitive bearer",
			header: "bearer some-token-value",
			want:   "some-token-value",
		},
		{
			name:   "empty",
			header: "",
			want:   "",
		},
		{
			name:   "invalid scheme",
			header: "Basic some-token-value",
			want:   "",
		},
		{
			name:   "missing token part",
			header: "Bearer",
			want:   "",
		},
		{
			name:   "extra spaces",
			header: "Bearer   spaced-token  ",
			want:   "spaced-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				c.Request.Header.Set("Authorization", tt.header)
			}

			got := bearerToken(c)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		method         string
		wantStatus     int
		wantCORS       bool
	}{
		{
			name:           "allowed origin",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://localhost:3000",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       true,
		},
		{
			name:           "disallowed origin",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://malicious.com",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       false,
		},
		{
			name:           "wildcard allows any",
			allowedOrigins: []string{"*"},
			origin:         "http://anywhere.com",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       true,
		},
		{
			name:           "preflight options request",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://localhost:3000",
			method:         http.MethodOptions,
			wantStatus:     http.StatusNoContent,
			wantCORS:       true,
		},
		{
			name:           "no origin header",
			allowedOrigins: []string{"*"},
			origin:         "",
			method:         http.MethodGet,
			wantStatus:     http.StatusOK,
			wantCORS:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(CORS(tt.allowedOrigins))
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			acao := w.Header().Get("Access-Control-Allow-Origin")
			if tt.wantCORS {
				if acao != tt.origin {
					t.Errorf("expected Access-Control-Allow-Origin to be %q, got %q", tt.origin, acao)
				}
			} else {
				if acao != "" {
					t.Errorf("expected no Access-Control-Allow-Origin header, got %q", acao)
				}
			}
		})
	}
}

func TestLocalAuth(t *testing.T) {
	t.Parallel()

	jwtMgr := jwtutil.New("test-secret-key-that-is-long-enough", time.Hour)

	// Generate a valid token
	validToken, err := jwtMgr.Sign("user-123")
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Generate an expired token
	expiredJwtMgr := jwtutil.New("test-secret-key-that-is-long-enough", time.Millisecond)
	expiredToken, err := expiredJwtMgr.Sign("user-123")
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}
	time.Sleep(5 * time.Millisecond)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantUser   string
	}{
		{
			name:       "valid token",
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusOK,
			wantUser:   "user-123",
		},
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantUser:   "",
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-garbage",
			wantStatus: http.StatusUnauthorized,
			wantUser:   "",
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + expiredToken,
			wantStatus: http.StatusUnauthorized,
			wantUser:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(LocalAuth(jwtMgr))

			var gotUser string
			r.GET("/", func(c *gin.Context) {
				if val, exists := c.Get(ContextUserIDKey); exists {
					gotUser = val.(string)
				}
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if gotUser != tt.wantUser {
				t.Errorf("expected user ID in context %q, got %q", tt.wantUser, gotUser)
			}
		})
	}
}
