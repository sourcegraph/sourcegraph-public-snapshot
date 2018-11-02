package authz

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

// Perm is a type of permission (e.g., "read").
type Perm string

const (
	Read Perm = "read"
)

// AuthzProvider defines a source of truth of which repositories a user is authorized to view. The
// user is identified by a AuthzID. In most cases, the AuthzID is equivalent to the UserID supplied
// by the AuthnProvider, but in some cases, the authz provider has its own internal definition of
// identity which must be mapped from the authn provider identity. The IdentityToAuthzID interface
// handles this mapping. Examples of authz providers include the following:
//
// * Code host
// * LDAP groups
// * SAML identity provider (via SAML group permissions)
//
// In most cases, the code host is the authz provider, because it is the source of truth for
// repository permissions.
//
// In most cases, an AuthzID will be a username in the authz provider, but this is not strictly
// necessary.  For instance, an AuthzID could be the name of a role that is sufficient for the authz
// provider to determine permissions.
type AuthzProvider interface {
	// Repos partitions the set of repositories into two sets: the set of repositories for which
	// this AuthzProvider is the source of permissions and the set of repositories for which it is
	// not. Each repository in the input set must be represented in exactly one of the output sets.
	//
	// This should not depend on the current user. Implementations should not use the context to
	// determine anything about the current user.
	Repos(ctx context.Context, repos map[Repo]struct{}) (mine map[Repo]struct{}, others map[Repo]struct{})

	// RepoPerms accepts an external user account and a set of repos. The external user account
	// identifies the user to the authz source (e.g., the code host). The return value is a map of
	// repository permissions. If a repo in the input set is missing from the returned permissions
	// map, that means "no permissions" on that repo.
	//
	// Implementations should handle any external account whose ServiceID and ServiceType values
	// match the `ServiceID()` and `ServiceType()` return values of this authz provider. The caller
	// can call the `FetchAccount` method to compute such an account from existing accounts. (Note:
	// implementations should use only the userAccount parameter to compute permissions. They should
	// NOT use any information about the currently authenticated user, including what might be
	// inferred from the ctx parameter.) The userAccount parameter may be nil, in which case the set
	// of permissions for an code-host-unauthenticated user is returned.
	//
	// Design note: this is a better interface than ListAllRepos, because the list of all repos may
	// be very long (especially if the returned list includes public repos). RepoPerms is a
	// sufficient interface for all current use cases and leaves up to the implementation which repo
	// permissions it needs to compute.  In practice, most will probably use a combination of (1)
	// "list all private repos the user has access to", (2) a mechanism to determine which repos are
	// public/private, and (3) a cache of some sort.
	RepoPerms(ctx context.Context, userAccount *extsvc.ExternalAccount, repos map[Repo]struct{}) (map[api.RepoURI]map[Perm]bool, error)

	// FetchAccount returns the external account that identifies the user to this authz provider,
	// taking as input the current list of external accounts associated with the
	// user. Implementations should always recompute the returned account (rather than returning an
	// element of `current` if it has the correct ServiceID and ServiceType).
	//
	// Implementations should use only the `user` and `current` parameters to compute the returned
	// external account. Specifically, they should not try to get the currently authenticated user
	// from the context parameter.
	//
	// The `user` argument should always be non-nil. If no external account can be computed for the
	// provided user, implementations should return nil, nil.
	FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error)

	// ServiceType returns the service type (e.g., "gitlab") of this authz provider.
	ServiceType() string

	// ServiceID returns the service ID (e.g., "https://gitlab.mycompany.com/") of this authz
	// provider.
	ServiceID() string
}

type Repo struct {
	// URI is the unique name/path of the repo on Sourcegraph
	URI api.RepoURI

	// ExternalRepoSpec uniquely identifies the external repo that is the source of the repo.
	api.ExternalRepoSpec
}
