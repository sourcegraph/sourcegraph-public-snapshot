package store

import (
	"fmt"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Users defines the interface for getting and listing users and their
// email addresses. It may be implemented by local user DBs as well as
// external services (GitHub, etc.).
type Users interface {
	Get(ctx context.Context, user sourcegraph.UserSpec) (*sourcegraph.User, error)
	GetWithEmail(ctx context.Context, emailAddr sourcegraph.EmailAddr) (*sourcegraph.User, error)
	List(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error)
	ListEmails(context.Context, sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error)
	Count(context.Context) (int32, error)
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
	Delete(context.Context, int32) error
}

// Invites manages pending invites to new users.
type Invites interface {
	// CreateOrUpdate creates an invite for the given email, and if one exists then
	// the invite is updated. A token is returned which can be used to retrieve the invite.
	CreateOrUpdate(ctx context.Context, invite *sourcegraph.AccountInvite) (string, error)

	// Retrieve gets the invite and marks as in use to avoid creating multiple accounts
	// from one invite. If the invite is already marked for use, this will return an error.
	Retrieve(ctx context.Context, token string) (*sourcegraph.AccountInvite, error)

	// MarkUnused marks an invite as unused. This should be called if an account could
	// not be created from this invite.
	MarkUnused(ctx context.Context, token string) error

	// Delete removes an invite. This should be called after an account is successfully
	// created from this invite, to prevent creation of multiple accounts.
	Delete(ctx context.Context, token string) error

	// DeleteByEmail removes the invite for the given email. If no
	// such invite exists, an error is returned.
	DeleteByEmail(ctx context.Context, email string) error

	// List fetches all pending invites on this server.
	List(ctx context.Context) ([]*sourcegraph.AccountInvite, error)
}

type Directory interface {
	GetUserByEmail(ctx context.Context, email string) (*sourcegraph.UserSpec, error)
}

// UserKeys defines the interface for updating a user's ssh public key,
// and looking up a user by their key.
type UserKeys interface {
	AddKey(ctx context.Context, uid int32, key sourcegraph.SSHPublicKey) error

	// LookupUser looks up user by key. The returned UserSpec will only have UID field set.
	LookupUser(ctx context.Context, key sourcegraph.SSHPublicKey) (*sourcegraph.UserSpec, error)

	DeleteKey(ctx context.Context, uid int32) error
	ListKeys(ctx context.Context, uid uint32) ([]sourcegraph.SSHPublicKey, error)
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
