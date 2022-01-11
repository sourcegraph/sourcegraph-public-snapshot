package authz

import (
	"context"
	"path"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"
	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/singleflight"

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

// DefaultSubRepoPermsChecker allows us to use a single instance with a shared
// cache and database connection. Since we don't have a database connection at
// initialisation time, services that require this client should initialise it in
// their main function.
var DefaultSubRepoPermsChecker SubRepoPermissionChecker = &noopPermsChecker{}

type noopPermsChecker struct{}

func (*noopPermsChecker) Permissions(ctx context.Context, userID int32, content RepoContent) (Perms, error) {
	return None, nil
}

func (*noopPermsChecker) Enabled() bool {
	return false
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
	clock             func() time.Time
	since             func(time.Time) time.Duration

	group *singleflight.Group
	cache *lru.Cache
}

const defaultCacheSize = 1000
const defaultCacheTTL = 10 * time.Second

// cachedRules caches the perms rules known for a particular user by repo.
type cachedRules struct {
	rules     map[api.RepoName]compiledRules
	timestamp time.Time
}

type compiledRules struct {
	includes []glob.Glob
	excludes []glob.Glob
}

// NewSubRepoPermsClient instantiates an instance of authz.SubRepoPermsClient
// which implements SubRepoPermissionChecker.
//
// SubRepoPermissionChecker is responsible for checking whether a user has access
// to data within a repo. Sub-repository permissions enforcement is on top of
// existing repository permissions, which means the user must already have access
// to the repository itself. The intention is for this client to be created once
// at startup and passed in to all places that need to check sub repo
// permissions.
//
// Note that sub-repo permissions are currently opt-in via the
// experimentalFeatures.enableSubRepoPermissions option.
func NewSubRepoPermsClient(permissionsGetter SubRepoPermissionsGetter) (*SubRepoPermsClient, error) {
	cache, err := lru.New(defaultCacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "creating LRU cache")
	}

	conf.Watch(func() {
		if c := conf.Get(); c.ExperimentalFeatures != nil && c.ExperimentalFeatures.SubRepoPermissions != nil && c.ExperimentalFeatures.SubRepoPermissions.UserCacheSize > 0 {
			cache.Resize(c.ExperimentalFeatures.SubRepoPermissions.UserCacheSize)
		}
	})

	return &SubRepoPermsClient{
		permissionsGetter: permissionsGetter,
		clock:             time.Now,
		since:             time.Since,
		group:             &singleflight.Group{},
		cache:             cache,
	}, nil
}

// WithGetter returns a new instance that uses the supplied getter. The cache
// from the original instance is left intact.
func (s *SubRepoPermsClient) WithGetter(g SubRepoPermissionsGetter) *SubRepoPermsClient {
	return &SubRepoPermsClient{
		permissionsGetter: g,
		clock:             s.clock,
		since:             s.since,
		group:             s.group,
		cache:             s.cache,
	}
}

// subRepoPermsPermissionsDuration tracks the behaviour and performance of Permissions()
var subRepoPermsPermissionsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "authz_sub_repo_perms_permissions_duration_seconds",
	Help: "Time spent syncing",
}, []string{"error"})

// subRepoPermsCacheHit tracks the number of cache hits and misses for sub-repo permissions
var subRepoPermsCacheHit = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "authz_sub_repo_perms_permissions_cache_count",
	Help: "The number of sub-repo perms cache hits or misses",
}, []string{"hit"})

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

	repoRules, err := s.getCompiledRules(ctx, userID)
	if err != nil {
		return None, errors.Wrap(err, "compiling match rules")
	}

	rules, ok := repoRules[content.Repo]
	if !ok {
		// If we make it this far it implies that we have access at the repo level.
		// Having any empty set of rules here implies that we can access the whole repo.
		// Repos that support sub-repo permissions will only have an entry in our
		// repo_permissions table if after all sub-repo permissions have been processed.
		return Read, nil
	}

	// Rules are created including the repo name
	toMatch := path.Join(string(content.Repo), content.Path)

	// The current path needs to either be included or NOT excluded and we'll give
	// preference to exclusion.
	for _, rule := range rules.excludes {
		if rule.Match(toMatch) {
			return
		}
	}
	for _, rule := range rules.includes {
		if rule.Match(toMatch) {
			return Read, nil
		}
	}

	// Return None if no rule matches to be safe
	return None, nil
}

// getCompiledRules fetches rules for the given repo with caching.
func (s *SubRepoPermsClient) getCompiledRules(ctx context.Context, userID int32) (map[api.RepoName]compiledRules, error) {
	// Fast path for cached rules
	item, _ := s.cache.Get(userID)
	cached, ok := item.(cachedRules)

	ttl := defaultCacheTTL
	if c := conf.Get(); c.ExperimentalFeatures != nil && c.ExperimentalFeatures.SubRepoPermissions != nil && c.ExperimentalFeatures.SubRepoPermissions.UserCacheTTLSeconds > 0 {
		ttl = time.Duration(c.ExperimentalFeatures.SubRepoPermissions.UserCacheTTLSeconds) * time.Second
	}

	if ok && s.since(cached.timestamp) <= ttl {
		subRepoPermsCacheHit.WithLabelValues("true").Inc()
		return cached.rules, nil
	}
	subRepoPermsCacheHit.WithLabelValues("false").Inc()

	// Slow path on cache miss or expiry. Ensure that only one goroutine is doing the
	// work
	groupKey := strconv.FormatInt(int64(userID), 10)
	result, err, _ := s.group.Do(groupKey, func() (interface{}, error) {
		repoPerms, err := s.permissionsGetter.GetByUser(ctx, userID)
		if err != nil {
			return nil, errors.Wrap(err, "fetching rules")
		}
		toCache := cachedRules{
			rules:     make(map[api.RepoName]compiledRules, len(repoPerms)),
			timestamp: time.Time{},
		}
		for repo, perms := range repoPerms {
			includes := make([]glob.Glob, 0, len(perms.PathIncludes))
			for _, rule := range perms.PathIncludes {
				g, err := glob.Compile(rule, '/')
				if err != nil {
					return nil, errors.Wrap(err, "building include matcher")
				}
				includes = append(includes, g)
			}
			excludes := make([]glob.Glob, 0, len(perms.PathExcludes))
			for _, rule := range perms.PathExcludes {
				g, err := glob.Compile(rule, '/')
				if err != nil {
					return nil, errors.Wrap(err, "building exclude matcher")
				}
				excludes = append(excludes, g)
			}
			toCache.rules[repo] = compiledRules{
				includes: includes,
				excludes: excludes,
			}
		}
		toCache.timestamp = s.clock()
		s.cache.Add(userID, toCache)
		return toCache.rules, nil
	})
	if err != nil {
		return nil, err
	}

	compiled := result.(map[api.RepoName]compiledRules)
	return compiled, nil
}

func (s *SubRepoPermsClient) Enabled() bool {
	if c := conf.Get(); c.ExperimentalFeatures != nil && c.ExperimentalFeatures.SubRepoPermissions != nil {
		return c.ExperimentalFeatures.SubRepoPermissions.Enabled
	}
	return false
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
		return None, errors.Wrapf(err, "getting actor permissions for actor: %d", a.UID)
	}
	return perms, nil
}

// FilterActorPaths will filter the given list of paths for the given actor
// returning on paths they are allowed to read.
func FilterActorPaths(ctx context.Context, checker SubRepoPermissionChecker, a *actor.Actor, repo api.RepoName, paths []string) ([]string, error) {
	filtered := make([]string, 0, len(paths))
	for _, p := range paths {
		include, err := FilterActorPath(ctx, checker, a, repo, p)
		if err != nil {
			return nil, errors.Wrap(err, "checking sub-repo permissions")
		}
		if include {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// FilterActorPath will filter the given path for the given actor
// returning true if the path is allowed to read.
func FilterActorPath(ctx context.Context, checker SubRepoPermissionChecker, a *actor.Actor, repo api.RepoName, path string) (bool, error) {
	if checker == nil || !checker.Enabled() {
		return true, nil
	}
	perms, err := ActorPermissions(ctx, checker, a, RepoContent{
		Repo: repo,
		Path: path,
	})
	if err != nil {
		return false, errors.Wrap(err, "checking sub-repo permissions")
	}
	return perms.Include(Read), nil
}
