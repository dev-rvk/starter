package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	userapp "github.com/starterpack/api/internal/application/user"
)

// UserHandler is the HTTP adapter for the user use cases. Handlers stay thin:
// decode -> call use case -> encode. No business logic lives here.
type UserHandler struct {
	svc userapp.UserService
}

// NewUserHandler constructs the handler from the application service interface.
func NewUserHandler(svc userapp.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) register(r *gin.RouterGroup) {
	r.POST("/users", h.create)
	r.GET("/users", h.list)
	r.GET("/users/:id", h.get)
}

// create godoc
// @Summary      Create a user
// @Description  Creates a user. Username must be 2-6 characters.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        payload  body      createUserRequest  true  "User to create"
// @Success      201      {object}  userResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      422      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users [post]
func (h *UserHandler) create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	u, err := h.svc.Create(c.Request.Context(), userapp.CreateInput{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		handleDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(u))
}

// get godoc
// @Summary      Get a user by id
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  userResponse
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *UserHandler) get(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u))
}

// list godoc
// @Summary      List users
// @Tags         users
// @Produce      json
// @Param        limit   query     int  false  "Page size (default 20, max 100)"
// @Param        offset  query     int  false  "Offset (default 0)"
// @Success      200     {array}   userResponse
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) list(c *gin.Context) {
	limit := parseInt32(c.Query("limit"), 20)
	offset := parseInt32(c.Query("offset"), 0)
	users, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		handleDomainError(c, err)
		return
	}
	out := make([]userResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toUserResponse(u))
	}
	c.JSON(http.StatusOK, out)
}

func parseInt32(raw string, fallback int32) int32 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return fallback
	}
	return int32(v)
}
