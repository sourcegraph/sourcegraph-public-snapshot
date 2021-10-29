package authz

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// RepoContent specifies data existing in a repo. It currently only supports
// paths but will be extended in future to support other pieces of metadata, for
// example branch.
type RepoContent struct {
	Repo api.RepoID
	Path string
}

// PermissionsGetter allow getting sub repository permissions.
type PermissionsGetter interface {
	GetByUser(ctx context.Context, userID int32) (map[api.RepoID]SubRepoPermissions, error)
}

// SubRepoPermsClient is responsible for checking whether a user has access to
// data within a repo. The intention is for this client to be created once at
// startup and passed in to all places that need to check sub repo permissions.
type SubRepoPermsClient struct {
	PermissionsGetter PermissionsGetter
}

func (s *SubRepoPermsClient) CheckPermissions(ctx context.Context, userID int32, content RepoContent) (Perms, error) {
	if s.PermissionsGetter == nil {
		return None, errors.New("PermissionsGetter is nil")
	}

	srp, err := s.PermissionsGetter.GetByUser(ctx, userID)
	if err != nil {
		return None, errors.Wrap(err, "getting permissions")
	}

	// Check repo
	repoRules, ok := srp[content.Repo]
	if !ok {
		// No sub-repository rules exist so we'll assume read access has already been
		// enforced at the repo level
		return Read, nil
	}

	// TODO: This will be very slow until we can cache compiled rules
	includeMatchers := make([]glob.Glob, 0, len(repoRules.PathIncludes))
	for _, rule := range repoRules.PathIncludes {
		g, err := glob.Compile(rule, '/')
		if err != nil {
			return None, errors.Wrap(err, "building include matcher")
		}
		includeMatchers = append(includeMatchers, g)
	}
	excludeMatchers := make([]glob.Glob, 0, len(repoRules.PathExcludes))
	for _, rule := range repoRules.PathExcludes {
		g, err := glob.Compile(rule, '/')
		if err != nil {
			return None, errors.Wrap(err, "building exclude matcher")
		}
		excludeMatchers = append(excludeMatchers, g)
	}

	// The current path needs to either be included or NOT excluded and we'll give
	// preference to exclusion.
	for _, rule := range excludeMatchers {
		if rule.Match(content.Path) {
			return None, nil
		}
	}
	for _, rule := range includeMatchers {
		if rule.Match(content.Path) {
			return Read, nil
		}
	}

	// Return None if no rule matches to be safe
	return None, nil
}
