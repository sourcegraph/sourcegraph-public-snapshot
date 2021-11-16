package authz

import (
	"context"
	"path"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// RepoContent specifies data existing in a repo. It currently only supports
// paths but will be extended in future to support other pieces of metadata, for
// example branch.
type RepoContent struct {
	Repo api.RepoName
	Path string
}

// SubRepoPermissionChecker is the interface exposed by the SubRepoPermsClient and is
// exposed to allow consumers to mock out the client.
//
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/authz -i SubRepoPermissionChecker -o mock_sub_repo_perms_checker.go
type SubRepoPermissionChecker interface {
	// Permissions returns the level of access the provided user has for the requested
	// content.
	//
	// If the userID represents an anonymous user, ErrUnauthenticated is returned.
	Permissions(ctx context.Context, userID int32, content RepoContent) (Perms, error)

	// Enabled indicates whether sub-repo permissions are enabled.
	Enabled() bool
}

var _ SubRepoPermissionChecker = &SubRepoPermsClient{}

// SubRepoPermissionsGetter allows getting sub repository permissions.
//
//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/authz -i SubRepoPermissionsGetter -o mock_sub_repo_perms_getter.go
type SubRepoPermissionsGetter interface {
	// GetByUser returns the known sub repository permissions rules known for a user.
	GetByUser(ctx context.Context, userID int32) (map[api.RepoName]SubRepoPermissions, error)
}

// SubRepoPermsClient is a concrete implementation of SubRepoPermissionChecker.
// Always use NewSubRepoPermsClient to instantiate an instance.
type SubRepoPermsClient struct {
	permissionsGetter SubRepoPermissionsGetter
}

// NewSubRepoPermsClient instantiates an instance of authz.SubRepoPermsClient
// which implements SubRepoPermissionChecker.
//
// SubRepoPermissionChecker is responsible for checking whether a user has access to
// data within a repo. Sub-repository permissions enforcement is on top of existing
// repository permissions, which means the user must already have access to the
// repository itself. The intention is for this client to be created once at startup
// and passed in to all places that need to check sub repo permissions.
//
// Note that sub-repo permissions are currently opt-in via the
// experimentalFeatures.enableSubRepoPermissions option.
func NewSubRepoPermsClient(permissionsGetter SubRepoPermissionsGetter) *SubRepoPermsClient {
	return &SubRepoPermsClient{
		permissionsGetter: permissionsGetter,
	}
}

// subRepoPermsPermissionsDuration tracks the behaviour and performance of Permissions()
var subRepoPermsPermissionsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "authz_sub_repo_perms_permissions_duration_seconds",
	Help: "Time spent syncing",
}, []string{"error"})

// Permissions return the current permissions granted to the given user on the
// given content. If sub-repo permissions are disabled, it is a no-op that return
// Read.
func (s *SubRepoPermsClient) Permissions(ctx context.Context, userID int32, content RepoContent) (perms Perms, err error) {
	// Are sub-repo permissions enabled at the site level
	if !s.Enabled() {
		return Read, nil
	}

	began := time.Now()
	defer func() {
		took := time.Since(began).Seconds()
		subRepoPermsPermissionsDuration.WithLabelValues(strconv.FormatBool(err != nil)).Observe(took)
	}()

	// Always default to not providing any permissions
	perms = None

	if s.permissionsGetter == nil {
		err = errors.New("PermissionsGetter is nil")
		return
	}

	if userID == 0 {
		err = &ErrUnauthenticated{}
		return
	}

	// An empty path is equivalent to repo permissions so we can assume it has
	// already been checked at that level.
	if content.Path == "" {
		return Read, nil
	}

	srp, err := s.permissionsGetter.GetByUser(ctx, userID)
	if err != nil {
		err = errors.Wrap(err, "getting permissions")
		return
	}

	// Check repo
	repoRules, ok := srp[content.Repo]
	if !ok {
		// If we make it this far it implies that we have access at the repo level.
		// Having any empty set of rules here implies that we can access the whole repo.
		// Repos that support sub-repo permissions will only have an entry in our
		// repo_permissions table if after all sub-repo permissions have been processed.
		return Read, nil
	}

	// TODO: This will be very slow until we can cache compiled rules
	includeMatchers := make([]glob.Glob, 0, len(repoRules.PathIncludes))
	for _, rule := range repoRules.PathIncludes {
		var g glob.Glob
		if g, err = glob.Compile(rule, '/'); err != nil {
			err = errors.Wrap(err, "building include matcher")
			return
		}
		includeMatchers = append(includeMatchers, g)
	}
	excludeMatchers := make([]glob.Glob, 0, len(repoRules.PathExcludes))
	for _, rule := range repoRules.PathExcludes {
		var g glob.Glob
		if g, err = glob.Compile(rule, '/'); err != nil {
			err = errors.Wrap(err, "building exclude matcher")
			return
		}
		excludeMatchers = append(excludeMatchers, g)
	}

	// Rules are created including the repo name
	toMatch := path.Join(string(content.Repo), content.Path)

	// The current path needs to either be included or NOT excluded and we'll give
	// preference to exclusion.
	for _, rule := range excludeMatchers {
		if rule.Match(toMatch) {
			return
		}
	}
	for _, rule := range includeMatchers {
		if rule.Match(toMatch) {
			return Read, nil
		}
	}

	// Return None if no rule matches to be safe
	return None, nil
}

func (s *SubRepoPermsClient) Enabled() bool {
	c := conf.Get()
	return c.ExperimentalFeatures != nil && c.ExperimentalFeatures.EnableSubRepoPermissions
}

// CurrentUserPermissions returns the level of access the authenticated user within
// the provided context has for the requested content by calling ActorPermissions.
func CurrentUserPermissions(ctx context.Context, s SubRepoPermissionChecker, content RepoContent) (Perms, error) {
	return ActorPermissions(ctx, s, actor.FromContext(ctx), content)
}

// ActorPermissions returns the level of access the given actor has for the requested
// content.
//
// If the context is unauthenticated, ErrUnauthenticated is returned. If the context is
// internal, Read permissions is granted.
func ActorPermissions(ctx context.Context, s SubRepoPermissionChecker, a *actor.Actor, content RepoContent) (Perms, error) {
	// Check config here, despite checking again in the s.Permissions implementation,
	// because we also make some permissions decisions here.
	if !s.Enabled() {
		return Read, nil
	}

	if a.IsInternal() {
		return Read, nil
	}
	if !a.IsAuthenticated() {
		return None, &ErrUnauthenticated{}
	}

	perms, err := s.Permissions(ctx, a.UID, content)
	if err != nil {
		return None, errors.Wrapf(err, "getting actor permissions for actor", a.UID)
	}
	return perms, nil
}
