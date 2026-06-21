package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	todoapp "github.com/starterpack/api/internal/application/todo"
)

// TodoHandler is the HTTP adapter for the todo use cases.
type TodoHandler struct {
	svc todoapp.TodoService
}

// NewTodoHandler constructs the handler from the application service interface.
func NewTodoHandler(svc todoapp.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

func (h *TodoHandler) register(r *gin.RouterGroup) {
	r.POST("/todos", h.create)
	r.GET("/todos", h.list)
	r.GET("/todos/:id", h.get)
	r.PUT("/todos/:id", h.update)
	r.DELETE("/todos/:id", h.delete)
}

// create godoc
// @Summary      Create a todo
// @Description  Creates a new todo item.
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        payload  body      createTodoRequest  true  "Todo to create"
// @Success      201      {object}  todoResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      422      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /todos [post]
func (h *TodoHandler) create(c *gin.Context) {
	var req createTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	todo, err := h.svc.Create(c.Request.Context(), req.Title)
	if err != nil {
		status, msg := mapDomainError(err)
		respondError(c, status, msg)
		return
	}
	c.JSON(http.StatusCreated, toTodoResponse(todo))
}

// get godoc
// @Summary      Get a todo by id
// @Tags         todos
// @Produce      json
// @Param        id   path      string  true  "Todo ID"
// @Success      200  {object}  todoResponse
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /todos/{id} [get]
func (h *TodoHandler) get(c *gin.Context) {
	todo, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		status, msg := mapDomainError(err)
		respondError(c, status, msg)
		return
	}
	c.JSON(http.StatusOK, toTodoResponse(todo))
}

// list godoc
// @Summary      List todos
// @Tags         todos
// @Produce      json
// @Param        limit   query     int  false  "Page size (default 20, max 100)"
// @Param        offset  query     int  false  "Offset (default 0)"
// @Success      200     {array}   todoResponse
// @Security     BearerAuth
// @Router       /todos [get]
func (h *TodoHandler) list(c *gin.Context) {
	limit := parseInt32(c.Query("limit"), 20)
	offset := parseInt32(c.Query("offset"), 0)
	todos, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		status, msg := mapDomainError(err)
		respondError(c, status, msg)
		return
	}
	out := make([]todoResponse, 0, len(todos))
	for _, t := range todos {
		out = append(out, toTodoResponse(t))
	}
	c.JSON(http.StatusOK, out)
}

// update godoc
// @Summary      Update a todo
// @Description  Updates a todo's title and completion status.
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        id       path      string             true  "Todo ID"
// @Param        payload  body      updateTodoRequest  true  "Todo updates"
// @Success      200      {object}  todoResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      422      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /todos/{id} [put]
func (h *TodoHandler) update(c *gin.Context) {
	var req updateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	todo, err := h.svc.Update(c.Request.Context(), c.Param("id"), req.Title, req.Completed)
	if err != nil {
		status, msg := mapDomainError(err)
		respondError(c, status, msg)
		return
	}
	c.JSON(http.StatusOK, toTodoResponse(todo))
}

// delete godoc
// @Summary      Delete a todo
// @Tags         todos
// @Param        id   path      string  true  "Todo ID"
// @Success      204  "No Content"
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /todos/{id} [delete]
func (h *TodoHandler) delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		status, msg := mapDomainError(err)
		respondError(c, status, msg)
		return
	}
	c.Status(http.StatusNoContent)
}
