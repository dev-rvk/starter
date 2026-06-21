package http

import (
	"time"

	userdomain "github.com/starterpack/api/internal/domain/user"
)

// registerRequest is the inbound payload for registration.
type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRequest is the inbound payload for login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// authResponse is the outbound payload after login/register.
type authResponse struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

// authUserResponse is the user data included in auth responses.
type authUserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toAuthUserResponse(u *userdomain.User) authUserResponse {
	return authUserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
