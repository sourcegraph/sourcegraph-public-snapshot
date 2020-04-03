// Package authz contains common logic and interfaces for authorization to
// external providers (such as GitLab).
package authz

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// Provider defines a source of truth of which repositories a user is authorized to view. The
// user is identified by an extsvc.Account instance. Examples of authz providers include the
// following:
//
// * Code host
// * LDAP groups
// * SAML identity provider (via SAML group permissions)
//
// In most cases, an authz provider represents a code host, because it is the source of truth for
// repository permissions.
type Provider interface {
	// RepoPerms accepts an external user account and a set of repos whose external service id and type
	// matches the Provider's `ServiceID()` and `ServiceType()`. The external user account identifies the
	// user to the authz source (e.g., the code host). The return value is a slice of repository permissions.
	// If a repo in the input slice is missing from the returned permissions slice, that means "no permissions"
	// on that repo.
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
	RepoPerms(ctx context.Context, userAccount *extsvc.Account, repos []*types.Repo) ([]RepoPerms, error)

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
	FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account) (mine *extsvc.Account, err error)

	// FetchUserPerms returns a list of repository/project IDs (on code host) that the
	// given account has read access on the code host. The repository ID should be the
	// same value as it would be used as api.ExternalRepoSpec.ID. The returned list
	// should only include private repositories/project IDs.
	//
	// Because permissions fetching APIs are often expensive, the implementation should
	// try to return partial but valid results in case of error, and it is up to callers
	// to decide whether to discard.
	FetchUserPerms(ctx context.Context, account *extsvc.Account) ([]extsvc.RepoID, error)

	// FetchRepoPerms returns a list of user IDs (on code host) who have read access to
	// the given repository/project on the code host. The user ID should be the same value
	// as it would be used as extsvc.Account.AccountID. The returned list should
	// include both direct access and inherited from the group/organization/team membership.
	//
	// Because permissions fetching APIs are often expensive, the implementation should
	// try to return partial but valid results in case of error, and it is up to callers
	// to decide whether to discard.
	FetchRepoPerms(ctx context.Context, repo *extsvc.Repository) ([]extsvc.AccountID, error)

	// ServiceType returns the service type (e.g., "gitlab") of this authz provider.
	ServiceType() string

	// ServiceID returns the service ID (e.g., "https://gitlab.mycompany.com/") of this authz
	// provider.
	ServiceID() string

	// Validate checks the configuration and credentials of the authz provider and returns any
	// problems.
	Validate() (problems []string)
}
