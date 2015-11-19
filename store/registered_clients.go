package store

import (
	"errors"
	"fmt"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"golang.org/x/net/context"
)

// RegisteredClients stores registered API clients.
type RegisteredClients interface {
	// Get retrieves the RegisteredClient with the given ID. If none
	// is found, RegisteredClientNotFoundError is returned.
	Get(context.Context, sourcegraph.RegisteredClientSpec) (*sourcegraph.RegisteredClient, error)

	// GetByCredentials retrieves the RegisteredClient with the given
	// credentials. If none is found, RegisteredClientNotFoundError is
	// returned.
	//
	// SECURITY: GetByCredentials must use constant time comparison to
	// look up by the secret, or else there is a potential timing
	// attack.
	GetByCredentials(context.Context, sourcegraph.RegisteredClientCredentials) (*sourcegraph.RegisteredClient, error)

	// Create creates a new registered API client. The
	// RegisteredClient arg's ID and Secret must be filled in.
	Create(context.Context, sourcegraph.RegisteredClient) error

	// Update updates the registered API client with the given ID. All
	// fields are overwritten. If no client with the given ID is
	// found, RegisteredClientNotFoundError is returned.
	Update(context.Context, sourcegraph.RegisteredClient) error

	// Delete deletes the registered API client with the given ID. If
	// no client with the given ID is found,
	// RegisteredClientNotFoundError is returned.
	Delete(context.Context, sourcegraph.RegisteredClientSpec) error

	// List enumerates registered API clients according to the
	// options.
	List(context.Context, sourcegraph.RegisteredClientListOptions) (*sourcegraph.RegisteredClientList, error)
}

// UserPermissions manages per-client (i.e. local instance) user permissions
// that control which users can login to a particular instance with
// their sourcegraph.com account credentials.
type UserPermissions interface {
	Get(ctx context.Context, opt *sourcegraph.UserPermissionsOptions) (*sourcegraph.UserPermissions, error)
	Verify(ctx context.Context, perms *sourcegraph.UserPermissions) (bool, error)
	Set(ctx context.Context, perms *sourcegraph.UserPermissions) error
	List(ctx context.Context, client *sourcegraph.RegisteredClientSpec) (*sourcegraph.UserPermissionsList, error)
}

// RegisteredClientNotFoundError occurs when a RegisteredClient is not
// found with the given arguments.
type RegisteredClientNotFoundError struct {
	// ID and Secret were the query parameters for RegisteredClients
	// methods that failed to yield a result.
	ID, Secret string
}

func (e *RegisteredClientNotFoundError) Error() string {
	s := fmt.Sprintf("no such registered client with ID %q", e.ID)
	if e.Secret != "" {
		s += " and secret (redacted)"
	}
	return s
}

var (
	// ErrRegisteredClientIDExists occurs when a RegisteredClient with
	// the given ID already exists.
	ErrRegisteredClientIDExists = errors.New("registered API client already exists with given ID")
)
