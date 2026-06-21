// Package authapp holds the authentication use cases (application layer). It
// orchestrates account + user domains and depends only on ports — not on HTTP
// or SQL.
package authapp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	accountdomain "github.com/starterpack/api/internal/domain/account"
	userdomain "github.com/starterpack/api/internal/domain/user"
	"github.com/starterpack/api/internal/platform/jwtutil"
	"github.com/starterpack/api/internal/platform/validator"
)

// AuthService is the inbound port consumed by the HTTP adapter.
type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, in LoginInput) (*AuthResult, error)
	Me(ctx context.Context, accountID string) (*userdomain.User, error)
}

// Compile-time check: Service must satisfy AuthService.
var _ AuthService = (*Service)(nil)

// Service implements the auth use cases.
type Service struct {
	accounts accountdomain.Repository
	users    userdomain.Repository
	jwt      *jwtutil.JWTManager
	v        *validator.Validator
}

// NewService wires the auth service to its dependencies.
func NewService(
	accounts accountdomain.Repository,
	users userdomain.Repository,
	jwt *jwtutil.JWTManager,
	v *validator.Validator,
) *Service {
	return &Service{
		accounts: accounts,
		users:    users,
		jwt:      jwt,
		v:        v,
	}
}

// RegisterInput is the raw (unvalidated) input for registration.
type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// LoginInput is the raw (unvalidated) input for login.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult is the output of a successful register or login.
type AuthResult struct {
	Token string
	User  *userdomain.User
}

// Register creates an account + user and returns a JWT.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	// Hash password.
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Build and validate account entity.
	acct := accountdomain.New(id, in.Email, string(hash), now)
	if err := s.v.ValidateAndMap("account", acct); err != nil {
		return nil, err
	}

	// Build and validate user entity.
	u := userdomain.New(id, in.Username, in.Email, now)
	if err := s.v.ValidateAndMap("user", u); err != nil {
		return nil, err
	}

	// Persist both. Account first — if email is taken, fail early.
	if err := s.accounts.Create(ctx, acct); err != nil {
		return nil, err
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	// Issue JWT.
	token, err := s.jwt.Sign(id)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: u}, nil
}

// Login validates credentials and returns a JWT.
func (s *Service) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	acct, err := s.accounts.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acct.PasswordHash), []byte(in.Password)); err != nil {
		return nil, accountdomain.ErrNotFound // Don't leak whether the email exists.
	}

	// Fetch the associated user.
	u, err := s.users.GetByID(ctx, acct.ID)
	if err != nil {
		return nil, err
	}

	// Issue JWT.
	token, err := s.jwt.Sign(acct.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: u}, nil
}

// Me returns the user for the authenticated account.
func (s *Service) Me(ctx context.Context, accountID string) (*userdomain.User, error) {
	return s.users.GetByID(ctx, accountID)
}
