package store

import (
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
)

// Users defines the interface for getting and listing users and their
// email addresses. It may be implemented by local user DBs as well as
// external services (GitHub, etc.).
type Users interface {
	Get(ctx context.Context, user sourcegraph.UserSpec) (*sourcegraph.User, error)
	List(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error)
	ListEmails(context.Context, sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error)
}

// Accounts manages user accounts that can be registered and
// created. External sources of users (e.g., GitHub) should probably
// only implement Users.
type Accounts interface {
	Create(ctx context.Context, newUser *sourcegraph.User) (*sourcegraph.User, error)
	GetByGitHubID(ctx context.Context, id int) (*sourcegraph.User, error)
	Update(context.Context, *sourcegraph.User) error
	UpdateEmails(context.Context, sourcegraph.UserSpec, []*sourcegraph.EmailAddr) error
	RequestPasswordReset(context.Context, *sourcegraph.User) (*sourcegraph.PasswordResetToken, error)
	ResetPassword(context.Context, *sourcegraph.NewPassword) error
}

type Directory interface {
	GetUserByEmail(ctx context.Context, email string) (*sourcegraph.UserSpec, error)
}

// ExternalAuthTokens manages per-user authentication tokens used to
// access external services.
type ExternalAuthTokens interface {
	// GetUserToken returns the user's auth token for a host (e.g.,
	// "github.com"). If none exists, ErrNoExternalAuthToken is
	// returned.
	GetUserToken(ctx context.Context, user int, host, clientID string) (*auth.ExternalAuthToken, error)

	// SetUserToken sets the user's auth token for a host (e.g.,
	// "github.com").
	//
	// If the user already has an auth token for the host, it is
	// overwritten by the new tok. If user, host, and clientID do not
	// match tok.User, tok.Host, and tok.ClientID, an error is
	// returned.
	SetUserToken(ctx context.Context, tok *auth.ExternalAuthToken) error
}

// UserNotFoundError occurs when a user is not found.
type UserNotFoundError struct {
	// At least one of the following fields must be set.

	Login string // the requested login
	UID   int    // the requested UID
}

func (e *UserNotFoundError) Error() string {
	if e.Login != "" {
		return fmt.Sprintf("user %s not found", e.Login)
	}
	return fmt.Sprintf("user #%d not found", e.UID)
}

// IsUserNotFound returns true iff err is a *UserNotFoundError.
func IsUserNotFound(err error) bool {
	_, ok := err.(*UserNotFoundError)
	return ok
}

// AccountAlreadyExistsError occurs when an account already exists
// with the requested login.
type AccountAlreadyExistsError struct {
	Login string // the requested login
}

func (e *AccountAlreadyExistsError) Error() string {
	return fmt.Sprintf("account %q already exists", e.Login)
}
