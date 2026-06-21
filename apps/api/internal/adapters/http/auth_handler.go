package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authapp "github.com/starterpack/api/internal/application/auth"
)

// AuthHandler is the HTTP adapter for local auth use cases. Handlers stay thin:
// decode -> call use case -> encode. No business logic lives here.
type AuthHandler struct {
	svc authapp.AuthService
}

// NewAuthHandler constructs the handler from the application service interface.
func NewAuthHandler(svc authapp.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) registerPublic(r *gin.RouterGroup) {
	r.POST("/auth/register", h.register)
	r.POST("/auth/login", h.login)
}

func (h *AuthHandler) registerProtected(r *gin.RouterGroup) {
	r.GET("/auth/me", h.me)
}

// register godoc
// @Summary      Register a new account
// @Description  Creates an account with email/password and a user profile.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      registerRequest  true  "Registration data"
// @Success      201      {object}  authResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      422      {object}  ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.svc.Register(c.Request.Context(), authapp.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, authResponse{
		Token: result.Token,
		User:  toAuthUserResponse(result.User),
	})
}

// login godoc
// @Summary      Login with email and password
// @Description  Validates credentials and returns a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      loginRequest  true  "Login credentials"
// @Success      200      {object}  authResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.svc.Login(c.Request.Context(), authapp.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// Map "not found" to 401 for auth endpoints.
		status, msg := mapDomainError(err)
		if status == http.StatusNotFound {
			status = http.StatusUnauthorized
			msg = "invalid email or password"
		}
		if status >= 500 {
			_ = c.Error(err)
		}
		respondError(c, status, msg)
		return
	}
	c.JSON(http.StatusOK, authResponse{
		Token: result.Token,
		User:  toAuthUserResponse(result.User),
	})
}

// me godoc
// @Summary      Get current user
// @Description  Returns the authenticated user's profile.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  authUserResponse
// @Failure      401  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /auth/me [get]
func (h *AuthHandler) me(c *gin.Context) {
	accountID, exists := c.Get("userID")
	if !exists {
		respondError(c, http.StatusUnauthorized, "not authenticated")
		return
	}
	u, err := h.svc.Me(c.Request.Context(), accountID.(string))
	if err != nil {
		handleDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAuthUserResponse(u))
}
