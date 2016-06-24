package store

import (
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// Users defines the interface for getting and listing users and their
// email addresses. It may be implemented by local user DBs as well as
// external services (GitHub, etc.).
type Users interface {
	Get(ctx context.Context, user sourcegraph.UserSpec) (*sourcegraph.User, error)
	GetWithEmail(ctx context.Context, emailAddr sourcegraph.EmailAddr) (*sourcegraph.User, error)
	List(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error)
	ListEmails(context.Context, sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error)

	// GetUIDByGitHubID gets the Sourcegraph UID for the user given
	// the GitHub UID that they've authorized tokens with (in
	// ExternalAuthTokens).
	GetUIDByGitHubID(ctx context.Context, githubUID int) (int32, error)
}

// Accounts manages user accounts that can be registered and
// created.
type Accounts interface {
	Create(ctx context.Context, newUser *sourcegraph.User, email *sourcegraph.EmailAddr) (*sourcegraph.User, error)
	Update(context.Context, *sourcegraph.User) error
	UpdateEmails(context.Context, sourcegraph.UserSpec, []*sourcegraph.EmailAddr) error
	RequestPasswordReset(context.Context, *sourcegraph.User) (*sourcegraph.PasswordResetToken, error)
	ResetPassword(context.Context, *sourcegraph.NewPassword) error
	Delete(context.Context, int32) error
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
	GetUserToken(ctx context.Context, user int, host, clientID string) (*ExternalAuthToken, error)

	// SetUserToken sets the user's auth token for a host (e.g.,
	// "github.com").
	//
	// If the user already has an auth token for the host, it is
	// overwritten by the new tok. If user, host, and clientID do not
	// match tok.User, tok.Host, and tok.ClientID, an error is
	// returned.
	SetUserToken(ctx context.Context, tok *ExternalAuthToken) error

	// DeleteToken deletes the token from the database.
	DeleteToken(ctx context.Context, tok *sourcegraph.ExternalTokenSpec) error

	// ListExternalUsers returns the list of external tokens corresponding to
	// the given external user ids.
	ListExternalUsers(ctx context.Context, extUIDs []int, host, clientID string) ([]*ExternalAuthToken, error)
}

// UserNotFoundError occurs when a user is not found.
type UserNotFoundError struct {
	// At least one of the following fields must be set.

	Login string // the requested login
	UID   int    // the requested UID
	Email string // the requested primary email
}

func (e *UserNotFoundError) Error() string {
	if e.Login != "" {
		return fmt.Sprintf("user %s not found", e.Login)
	}
	if e.Email != "" {
		return fmt.Sprintf("user with email %s not found", e.Email)
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
	UID   int32  // the requested UID
}

func (e *AccountAlreadyExistsError) Error() string {
	var uidStr string
	if e.UID != 0 {
		uidStr = fmt.Sprintf("(UID %v) ", e.UID)
	}
	return fmt.Sprintf("account %q %salready exists", e.Login, uidStr)
}
