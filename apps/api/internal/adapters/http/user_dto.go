package http

import (
	"time"

	userdomain "github.com/starterpack/api/internal/domain/user"
)

// createUserRequest is the inbound payload for creating a user.
// Validation is handled by the application service, not here.
type createUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
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
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
