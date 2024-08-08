// Package authz contains common logic and interfaces for authorization to
// external providers (such as GitLab).
package authz

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// SubRepoPermissions denotes access control rules within a repository's
// contents.
//
// Rules are expressed as Glob syntaxes:
//
//	pattern:
//	    { term }
//
//	term:
//	    `*`         matches any sequence of non-separator characters
//	    `**`        matches any sequence of characters
//	    `?`         matches any single non-separator character
//	    `[` [ `!` ] { character-range } `]`
//	                character class (must be non-empty)
//	    `{` pattern-list `}`
//	                pattern alternatives
//	    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//	    `\` c       matches character c
//
//	character-range:
//	    c           matches character c (c != `\\`, `-`, `]`)
//	    `\` c       matches character c
//	    lo `-` hi   matches character c for lo <= c <= hi
//
//	pattern-list:
//	    pattern { `,` pattern }
//	                comma-separated (without spaces) patterns
//
// This Glob syntax is currently from github.com/gobwas/glob:
// https://sourcegraph.com/github.com/gobwas/glob@e7a84e9525fe90abcda167b604e483cc959ad4aa/-/blob/glob.go?L39:6
//
// We use a third party library for double-wildcard support, which the standard
// library does not provide.
//
// Paths are relative to the root of the repo.
type SubRepoPermissions struct {
	Paths []string
}

// SubRepoPermissionsWithIPs denotes access control rules within a repository's
// contents.
type SubRepoPermissionsWithIPs struct {
	Paths []PathWithIP
}

// PathWithIP denotes a rule associated with the given file path and range of IP addresses that
// this rule applies to.
//
// Rules are expressed as Glob syntaxes:
//
//	pattern:
//	    { term }
//
//	term:
//	    `*`         matches any sequence of non-separator characters
//	    `**`        matches any sequence of characters
//	    `?`         matches any single non-separator character
//	    `[` [ `!` ] { character-range } `]`
//	                character class (must be non-empty)
//	    `{` pattern-list `}`
//	                pattern alternatives
//	    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//	    `\` c       matches character c
//
//	character-range:
//	    c           matches character c (c != `\\`, `-`, `]`)
//	    `\` c       matches character c
//	    lo `-` hi   matches character c for lo <= c <= hi
//
//	pattern-list:
//	    pattern { `,` pattern }
//	                comma-separated (without spaces) patterns
//
// This Glob syntax is currently from github.com/gobwas/glob:
// https://sourcegraph.com/github.com/gobwas/glob@e7a84e9525fe90abcda167b604e483cc959ad4aa/-/blob/glob.go?L39:6
//
// We use a third party library for double-wildcard support, which the standard
// library does not provide.
//
// Paths are relative to the root of the repo.
type PathWithIP struct {
	Path string
	IP   string
}

// ExternalUserPermissions is a collection of accessible repository/project IDs
// (on the code host). It contains exact IDs, as well as prefixes to both include
// and exclude IDs.
//
// 🚨 SECURITY: Every call site should evaluate all fields of this struct to
// have a complete set of IDs.
type ExternalUserPermissions struct {
	Exacts          []extsvc.RepoID
	IncludeContains []extsvc.RepoID
	ExcludeContains []extsvc.RepoID

	// SubRepoPermissions denotes sub-repository content access control rules where
	// relevant. If no corresponding entry for an Exacts repo exists in SubRepoPermissions,
	// it can be safely assumed that access to the entire repo is available.
	SubRepoPermissions map[extsvc.RepoID]*SubRepoPermissionsWithIPs
}

// FetchPermsOptions declares options when performing permissions sync.
type FetchPermsOptions struct {
	// InvalidateCaches indicates that caches added for optimization encountered during
	// this fetch should be invalidated.
	InvalidateCaches bool `json:"invalidate_caches"`
}

type UserPermissionsFetcher interface {
	Provider
	FetchUserPerms(ctx context.Context, account *extsvc.Account, opts FetchPermsOptions) (*ExternalUserPermissions, error)
}

// RepoPermissionsFetcher defines a source that can fetch permissions for
// a repository from a code host.
type RepoPermissionsFetcher interface {
	Provider
	FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts FetchPermsOptions) ([]extsvc.AccountID, error)
}

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
	//
	// The `verifiedEmails` should only contain a list of verified emails that is
	// associated to the `user`.
	FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmails []string) (mine *extsvc.Account, err error)

	// ServiceType returns the service type (e.g., "gitlab") of this authz provider.
	ServiceType() string

	// ServiceID returns the service ID (e.g., "https://gitlab.mycompany.com/") of this authz
	// provider.
	ServiceID() string

	// URN returns the unique resource identifier of external service where the authz provider
	// is defined.
	URN() string

	// ValidateConnection checks that the configuration and credentials of the authz provider
	// can establish a valid connection with the provider, and returns warnings based on any
	// issues it finds.
	ValidateConnection(ctx context.Context) error
}

// ErrUnauthenticated indicates an unauthenticated request.
type ErrUnauthenticated struct{}

func (e ErrUnauthenticated) Error() string {
	return "request is unauthenticated"
}

func (e ErrUnauthenticated) Unauthenticated() bool { return true }

// ErrUnimplemented indicates sync is unimplemented and its data should not be used.
//
// When returning this error, provide a pointer.
type ErrUnimplemented struct {
	// Feature indicates the unimplemented functionality.
	Feature string
}

func (e ErrUnimplemented) Error() string {
	return fmt.Sprintf("%s is unimplemented", e.Feature)
}

func (e ErrUnimplemented) Unimplemented() bool { return true }

func (e ErrUnimplemented) Is(err error) bool {
	_, ok := err.(*ErrUnimplemented)
	return ok
}
