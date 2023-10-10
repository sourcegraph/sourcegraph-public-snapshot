package subrepoperms

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gobwas/glob"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/atomic"
	"golang.org/x/sync/singleflight"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SubRepoPermissionsGetter allows getting sub repository permissions.
type SubRepoPermissionsGetter interface {
	// GetByUser returns the sub repository permissions rules known for a user.
	GetByUser(ctx context.Context, userID int32) (map[api.RepoName]authz.SubRepoPermissions, error)

	// RepoIDSupported returns true if repo with the given ID has sub-repo permissions.
	RepoIDSupported(ctx context.Context, repoID api.RepoID) (bool, error)

	// RepoSupported returns true if repo with the given name has sub-repo permissions.
	RepoSupported(ctx context.Context, repo api.RepoName) (bool, error)
}

// SubRepoPermsClient is a concrete implementation of SubRepoPermissionChecker.
// Always use NewSubRepoPermsClient to instantiate an instance.
type SubRepoPermsClient struct {
	permissionsGetter SubRepoPermissionsGetter
	clock             func() time.Time
	since             func(time.Time) time.Duration

	group   *singleflight.Group
	cache   *lru.Cache[int32, cachedRules]
	enabled *atomic.Bool
}

const (
	defaultCacheSize = 1000
	defaultCacheTTL  = 10 * time.Second
)

// cachedRules caches the perms rules known for a particular user by repo.
type cachedRules struct {
	rules     map[api.RepoName]compiledRules
	timestamp time.Time
}

type path struct {
	globPath  glob.Glob
	exclusion bool
	// the original rule before it was compiled into a glob matcher
	original string
}

type compiledRules struct {
	paths []path
}

// GetPermissionsForPath tries to match a given path to a list of rules.
// Since the last applicable rule is the one that applies, the list is
// traversed in reverse, and the function returns as soon as a match is found.
// If no match is found, None is returned.
func (rules compiledRules) GetPermissionsForPath(path string) authz.Perms {
	for i := len(rules.paths) - 1; i >= 0; i-- {
		if rules.paths[i].globPath.Match(path) {
			if rules.paths[i].exclusion {
				return authz.None
			}
			return authz.Read
		}
	}

	// Return None if no rule matches
	return authz.None
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
func NewSubRepoPermsClient(permissionsGetter SubRepoPermissionsGetter) *SubRepoPermsClient {
	cache, err := lru.New[int32, cachedRules](defaultCacheSize)
	if err != nil {
		// Errors should only ever occur if we change the value of defaultCacheSize
		// to be negative.
		panic(fmt.Sprintf("failed to create LRU cache for sub repo perms client: %v", err))
	}

	enabled := atomic.NewBool(false)

	conf.Watch(func() {
		c := conf.Get()
		if c.ExperimentalFeatures == nil || c.ExperimentalFeatures.SubRepoPermissions == nil {
			enabled.Store(false)
			return
		}

		cacheSize := c.ExperimentalFeatures.SubRepoPermissions.UserCacheSize
		if cacheSize == 0 {
			cacheSize = defaultCacheSize
		}
		cache.Resize(cacheSize)
		enabled.Store(c.ExperimentalFeatures.SubRepoPermissions.Enabled)
	})

	return &SubRepoPermsClient{
		permissionsGetter: permissionsGetter,
		clock:             time.Now,
		since:             time.Since,
		group:             &singleflight.Group{},
		cache:             cache,
		enabled:           enabled,
	}
}

var (
	metricSubRepoPermsPermissionsDurationSuccess prometheus.Observer
	metricSubRepoPermsPermissionsDurationError   prometheus.Observer
)

func init() {
	// We cache the result of WithLabelValues since we call them in
	// performance sensitive code. See BenchmarkFilterActorPaths.
	metric := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "authz_sub_repo_perms_permissions_duration_seconds",
		Help: "Time spent calculating permissions of a file for an actor.",
	}, []string{"error"})
	metricSubRepoPermsPermissionsDurationSuccess = metric.WithLabelValues("false")
	metricSubRepoPermsPermissionsDurationError = metric.WithLabelValues("true")
}

var (
	metricSubRepoPermCacheHit  prometheus.Counter
	metricSubRepoPermCacheMiss prometheus.Counter
)

func init() {
	// We cache the result of WithLabelValues since we call them in
	// performance sensitive code. See BenchmarkFilterActorPaths.
	metric := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "authz_sub_repo_perms_permissions_cache_count",
		Help: "The number of sub-repo perms cache hits or misses",
	}, []string{"hit"})
	metricSubRepoPermCacheHit = metric.WithLabelValues("true")
	metricSubRepoPermCacheMiss = metric.WithLabelValues("false")
}

// Permissions return the current permissions granted to the given user on the
// given content. If sub-repo permissions are disabled, it is a no-op that return
// Read.
func (s *SubRepoPermsClient) Permissions(ctx context.Context, userID int32, content authz.RepoContent) (perms authz.Perms, err error) {
	// Are sub-repo permissions enabled at the site level
	if !s.Enabled() {
		return authz.Read, nil
	}

	began := time.Now()
	defer func() {
		took := time.Since(began).Seconds()
		if err == nil {
			metricSubRepoPermsPermissionsDurationSuccess.Observe(took)
		} else {
			metricSubRepoPermsPermissionsDurationError.Observe(took)
		}
	}()

	f, err := s.FilePermissionsFunc(ctx, userID, content.Repo)
	if err != nil {
		return authz.None, err
	}
	return f(content.Path)
}

// filePermissionsFuncAllRead is a FilePermissionFunc which _always_ returns
// Read. Only use in cases that sub repo permission checks should not be done.
func filePermissionsFuncAllRead(_ string) (authz.Perms, error) {
	return authz.Read, nil
}

func (s *SubRepoPermsClient) FilePermissionsFunc(ctx context.Context, userID int32, repo api.RepoName) (authz.FilePermissionFunc, error) {
	// Are sub-repo permissions enabled at the site level
	if !s.Enabled() {
		return filePermissionsFuncAllRead, nil
	}

	if s.permissionsGetter == nil {
		return nil, errors.New("permissionsGetter is nil")
	}

	if userID == 0 {
		return nil, &authz.ErrUnauthenticated{}
	}

	repoRules, err := s.getCompiledRules(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "compiling match rules")
	}

	rules, rulesExist := repoRules[repo]
	if !rulesExist {
		// If we make it this far it implies that we have access at the repo level.
		// Having any empty set of rules here implies that we can access the whole repo.
		// Repos that support sub-repo permissions will only have an entry in our
		// repo_permissions table after all sub-repo permissions have been processed.
		return filePermissionsFuncAllRead, nil
	}

	return func(path string) (authz.Perms, error) {
		// An empty path is equivalent to repo permissions so we can assume it has
		// already been checked at that level.
		if path == "" {
			return authz.Read, nil
		}

		// Prefix path with "/", otherwise suffix rules like "**/file.txt" won't match
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		// Iterate through all rules for the current path, and the final match takes
		// preference.
		return rules.GetPermissionsForPath(path), nil
	}, nil
}

// getCompiledRules fetches rules for the given repo with caching.
func (s *SubRepoPermsClient) getCompiledRules(ctx context.Context, userID int32) (map[api.RepoName]compiledRules, error) {
	// Fast path for cached rules
	cached, _ := s.cache.Get(userID)

	ttl := defaultCacheTTL
	if c := conf.Get(); c.ExperimentalFeatures != nil && c.ExperimentalFeatures.SubRepoPermissions != nil && c.ExperimentalFeatures.SubRepoPermissions.UserCacheTTLSeconds > 0 {
		ttl = time.Duration(c.ExperimentalFeatures.SubRepoPermissions.UserCacheTTLSeconds) * time.Second
	}

	if s.since(cached.timestamp) <= ttl {
		metricSubRepoPermCacheHit.Inc()
		return cached.rules, nil
	}
	metricSubRepoPermCacheMiss.Inc()

	// Slow path on cache miss or expiry. Ensure that only one goroutine is doing the
	// work
	groupKey := strconv.FormatInt(int64(userID), 10)
	result, err, _ := s.group.Do(groupKey, func() (any, error) {
		repoPerms, err := s.permissionsGetter.GetByUser(ctx, userID)
		if err != nil {
			return nil, errors.Wrap(err, "fetching rules")
		}
		toCache := cachedRules{
			rules: make(map[api.RepoName]compiledRules, len(repoPerms)),
		}
		for repo, perms := range repoPerms {
			paths := make([]path, 0, len(perms.Paths))
			for _, rule := range perms.Paths {
				exclusion := strings.HasPrefix(rule, "-")
				rule = strings.TrimPrefix(rule, "-")

				if !strings.HasPrefix(rule, "/") {
					rule = "/" + rule
				}

				g, err := glob.Compile(rule, '/')
				if err != nil {
					return nil, errors.Wrap(err, "building include matcher")
				}

				paths = append(paths, path{globPath: g, exclusion: exclusion, original: rule})

				// Special case. Our glob package does not handle rules starting with a double
				// wildcard correctly. For example, we would expect `/**/*.java` to match all
				// java files, but it does not match files at the root, eg `/foo.java`. To get
				// around this we add an extra rule to cover this case.
				if strings.HasPrefix(rule, "/**/") {
					trimmed := rule
					for {
						trimmed = strings.TrimPrefix(trimmed, "/**")
						if strings.HasPrefix(trimmed, "/**/") {
							// Keep trimming
							continue
						}
						g, err := glob.Compile(trimmed, '/')
						if err != nil {
							return nil, errors.Wrap(err, "building include matcher")
						}
						paths = append(paths, path{globPath: g, exclusion: exclusion, original: trimmed})
						break
					}
				}

				// We should include all directories above an include rule so that we can browse
				// to the included items.
				if exclusion {
					// Not required for an exclude rule
					continue
				}

				dirs := expandDirs(rule)
				for _, dir := range dirs {
					g, err := glob.Compile(dir, '/')
					if err != nil {
						return nil, errors.Wrap(err, "building include matcher for dir")
					}
					paths = append(paths, path{globPath: g, exclusion: false, original: dir})
				}
			}

			toCache.rules[repo] = compiledRules{
				paths: paths,
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
	return s.enabled.Load()
}

func (s *SubRepoPermsClient) EnabledForRepoID(ctx context.Context, id api.RepoID) (bool, error) {
	return s.permissionsGetter.RepoIDSupported(ctx, id)
}

func (s *SubRepoPermsClient) EnabledForRepo(ctx context.Context, repo api.RepoName) (bool, error) {
	return s.permissionsGetter.RepoSupported(ctx, repo)
}

// expandDirs will return a new set of rules that will match all directories
// above the supplied rule. As a special case, if the rule starts with a wildcard
// we return a rule to match all directories.
func expandDirs(rule string) []string {
	dirs := make([]string, 0)

	// Make sure the rule starts with a slash
	if !strings.HasPrefix(rule, "/") {
		rule = "/" + rule
	}

	// If a rule starts with a wildcard it can match at any level in the tree
	// structure so there's no way of walking up the tree and expand out to the list
	// of valid directories. Instead, we just return a rule that matches any
	// directory
	if strings.HasPrefix(rule, "/*") {
		dirs = append(dirs, "**/")
		return dirs
	}

	for {
		lastSlash := strings.LastIndex(rule, "/")
		if lastSlash <= 0 { // we have to ignore the slash at index 0
			break
		}
		// Drop anything after the last slash
		rule = rule[:lastSlash]

		dirs = append(dirs, rule+"/")
	}

	return dirs
}

// NewSimpleChecker is exposed for testing and allows creation of a simple
// checker based on the rules provided. The rules are expected to be in glob
// format.
func NewSimpleChecker(repo api.RepoName, paths []string) authz.SubRepoPermissionChecker {
	getter := NewMockSubRepoPermissionsGetter()
	getter.GetByUserFunc.SetDefaultHook(func(ctx context.Context, i int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
		return map[api.RepoName]authz.SubRepoPermissions{
			repo: {
				Paths: paths,
			},
		}, nil
	})
	getter.RepoSupportedFunc.SetDefaultReturn(true, nil)
	getter.RepoIDSupportedFunc.SetDefaultReturn(true, nil)
	return NewSubRepoPermsClient(getter)
}
