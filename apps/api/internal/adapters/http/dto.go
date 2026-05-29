package http

import (
	"time"

	userdomain "github.com/starterpack/api/internal/domain/user"
)

// createUserRequest is the inbound payload. The binding tags are enforced by
// go-playground/validator (Gin's built-in binding engine) at the transport
// edge; the domain re-validates authoritatively in its value constructors.
type createUserRequest struct {
	Username string `binding:"required,min=2,max=6" json:"username"`
	Email    string `binding:"required,email"      json:"email"`
}

// userResponse is the outbound representation of a user.
type userResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toUserResponse(u *userdomain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Username:  u.Username.String(),
		Email:     u.Email.String(),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
