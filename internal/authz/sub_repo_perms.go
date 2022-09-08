package authz

import (
	"context"
	"io/fs"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/atomic"
	"golang.org/x/sync/singleflight"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RepoContent specifies data existing in a repo. It currently only supports
// paths but will be extended in future to support other pieces of metadata, for
// example branch.
type RepoContent struct {
	Repo api.RepoName
	Path string
}

// FilePermissionFunc is a function which returns the Perm of path. This
// function is associated with a user and repository and should not be used
// beyond the lifetime of a single request. It exists to amortize the costs of
// setup when checking many files in a repository.
type FilePermissionFunc func(path string) (Perms, error)

// SubRepoPermissionChecker is the interface exposed by the SubRepoPermsClient and is
// exposed to allow consumers to mock out the client.
type SubRepoPermissionChecker interface {
	// Permissions returns the level of access the provided user has for the requested
	// content.
	//
	// If the userID represents an anonymous user, ErrUnauthenticated is returned.
	Permissions(ctx context.Context, userID int32, content RepoContent) (Perms, error)

	// FilePermissionFunc returns a FilePermissionFunc for userID in repo.
	// This function should only be used during the lifetime of a request. It
	// exists to amortize the cost of checking many files in a repo.
	//
	// If the userID represents an anonymous user, ErrUnauthenticated is returned.
	FilePermissionsFunc(ctx context.Context, userID int32, repo api.RepoName) (FilePermissionFunc, error)

	// Enabled indicates whether sub-repo permissions are enabled.
	Enabled() bool

	// EnabledForRepoId indicates whether sub-repo permissions are enabled for the given repoID
	EnabledForRepoId(ctx context.Context, repoId api.RepoID) (bool, error)

	// EnabledForRepo indicates whether sub-repo permissions are enabled for the given repo
	EnabledForRepo(ctx context.Context, repo api.RepoName) (bool, error)
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

func (*noopPermsChecker) FilePermissionsFunc(ctx context.Context, userID int32, repo api.RepoName) (FilePermissionFunc, error) {
	return func(path string) (Perms, error) {
		return None, nil
	}, nil
}

func (*noopPermsChecker) Enabled() bool {
	return false
}

func (*noopPermsChecker) EnabledForRepoId(ctx context.Context, repoId api.RepoID) (bool, error) {
	return false, nil
}

func (*noopPermsChecker) EnabledForRepo(ctx context.Context, repo api.RepoName) (bool, error) {
	return false, nil
}

var _ SubRepoPermissionChecker = &SubRepoPermsClient{}

// SubRepoPermissionsGetter allows getting sub repository permissions.
type SubRepoPermissionsGetter interface {
	// GetByUser returns the known sub repository permissions rules known for a user.
	GetByUser(ctx context.Context, userID int32) (map[api.RepoName]SubRepoPermissions, error)

	// RepoIdSupported returns true if repo with the given ID has sub-repo permissions
	RepoIdSupported(ctx context.Context, repoId api.RepoID) (bool, error)

	// RepoSupported returns true if repo with the given name has sub-repo permissions
	RepoSupported(ctx context.Context, repo api.RepoName) (bool, error)
}

// SubRepoPermsClient is a concrete implementation of SubRepoPermissionChecker.
// Always use NewSubRepoPermsClient to instantiate an instance.
type SubRepoPermsClient struct {
	permissionsGetter SubRepoPermissionsGetter
	clock             func() time.Time
	since             func(time.Time) time.Duration

	group   *singleflight.Group
	cache   *lru.Cache
	enabled *atomic.Bool
}

const defaultCacheSize = 1000
const defaultCacheTTL = 10 * time.Second

// cachedRules caches the perms rules known for a particular user by repo.
type cachedRules struct {
	rules     map[api.RepoName]compiledRules
	timestamp time.Time
}

type compiledRules struct {
	includes    []glob.Glob
	excludes    []glob.Glob
	dirIncludes []glob.Glob
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
	}, nil
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
func (s *SubRepoPermsClient) Permissions(ctx context.Context, userID int32, content RepoContent) (perms Perms, err error) {
	// Are sub-repo permissions enabled at the site level
	if !s.Enabled() {
		return Read, nil
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
		return None, err
	}
	return f(content.Path)
}

// filePermissionsFuncAllRead is a FilePermissionFunc which _always_ returns
// Read. Only use in cases that sub repo permission checks should not be done.
func filePermissionsFuncAllRead(_ string) (Perms, error) {
	return Read, nil
}

func (s *SubRepoPermsClient) FilePermissionsFunc(ctx context.Context, userID int32, repo api.RepoName) (FilePermissionFunc, error) {
	// Are sub-repo permissions enabled at the site level
	if !s.Enabled() {
		return filePermissionsFuncAllRead, nil
	}

	if s.permissionsGetter == nil {
		return nil, errors.New("PermissionsGetter is nil")
	}

	if userID == 0 {
		return nil, &ErrUnauthenticated{}
	}

	var (
		once       sync.Once
		rules      compiledRules
		rulesExist bool
		rulesErr   error
	)

	return func(path string) (Perms, error) {
		// An empty path is equivalent to repo permissions so we can assume it has
		// already been checked at that level.
		if path == "" {
			return Read, nil
		}

		once.Do(func() {
			repoRules, err := s.getCompiledRules(ctx, userID)
			if err != nil {
				rulesErr = errors.Wrap(err, "compiling match rules")
				return
			}

			rules, rulesExist = repoRules[repo]
		})

		if rulesErr != nil {
			return None, rulesErr
		} else if !rulesExist {
			// If we make it this far it implies that we have access at the repo level.
			// Having any empty set of rules here implies that we can access the whole repo.
			// Repos that support sub-repo permissions will only have an entry in our
			// repo_permissions table after all sub-repo permissions have been processed.
			return Read, nil
		}

		// TODO(keegancsmith) I am preserving behaviour with all the code
		// above here in this function. IE we are using a Once to lazily
		// initialize rules, just in case passing in an empty path happens.
		// I'd prefer to not check for empty path + remove lazily evaluation.
		// Lets discuss if we can change this behaviour on the PR and
		// follow-up. For now I'd rather preserve behaviour to prevent
		// regressions.

		// The current path needs to either be included or NOT excluded and we'll give
		// preference to exclusion.
		for _, rule := range rules.excludes {
			if rule.Match(path) {
				return None, nil
			}
		}
		for _, rule := range rules.includes {
			if rule.Match(path) {
				return Read, nil
			}
		}

		// We also want to match any directories above paths that we include so that we
		// can browse down the file hierarchy.
		if strings.HasSuffix(path, "/") {
			for _, rule := range rules.dirIncludes {
				if rule.Match(path) {
					return Read, nil
				}
			}
		}

		// Return None if no rule matches to be safe
		return None, nil
	}, nil
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
			rules:     make(map[api.RepoName]compiledRules, len(repoPerms)),
			timestamp: time.Time{},
		}
		for repo, perms := range repoPerms {
			includes := make([]glob.Glob, 0, len(perms.PathIncludes))
			dirIncludes := make([]glob.Glob, 0)
			dirSeen := make(map[string]struct{})
			for _, rule := range perms.PathIncludes {
				g, err := glob.Compile(rule, '/')
				if err != nil {
					return nil, errors.Wrap(err, "building include matcher")
				}
				includes = append(includes, g)

				// We should include all directories above an include rule
				dirs := expandDirs(rule)
				for _, dir := range dirs {
					if _, ok := dirSeen[dir]; ok {
						continue
					}
					g, err := glob.Compile(dir, '/')
					if err != nil {
						return nil, errors.Wrap(err, "building include matcher for dir")
					}
					dirIncludes = append(dirIncludes, g)
					dirSeen[dir] = struct{}{}
				}
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
				includes:    includes,
				excludes:    excludes,
				dirIncludes: dirIncludes,
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

func (s *SubRepoPermsClient) EnabledForRepoId(ctx context.Context, id api.RepoID) (bool, error) {
	return s.permissionsGetter.RepoIdSupported(ctx, id)
}

func (s *SubRepoPermsClient) EnabledForRepo(ctx context.Context, repo api.RepoName) (bool, error) {
	return s.permissionsGetter.RepoSupported(ctx, repo)
}

// expandDirs will return rules that match all parent directories of the given
// rule.
func expandDirs(rule string) []string {
	dirs := make([]string, 0)

	// We can't support rules that start with a wildcard because we can only
	// see one level of the tree at a time so we have no way of knowing which path leads
	// to a file the user is allowed to see.
	if strings.HasPrefix(rule, "*") {
		return dirs
	}

	for {
		lastSlash := strings.LastIndex(rule, "/")
		if lastSlash == -1 {
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
func NewSimpleChecker(repo api.RepoName, includes []string, excludes []string) (SubRepoPermissionChecker, error) {
	getter := NewMockSubRepoPermissionsGetter()
	getter.GetByUserFunc.SetDefaultHook(func(ctx context.Context, i int32) (map[api.RepoName]SubRepoPermissions, error) {
		return map[api.RepoName]SubRepoPermissions{
			repo: {
				PathIncludes: includes,
				PathExcludes: excludes,
			},
		}, nil
	})
	getter.RepoSupportedFunc.SetDefaultReturn(true, nil)
	getter.RepoIdSupportedFunc.SetDefaultReturn(true, nil)
	return NewSubRepoPermsClient(getter)
}

// ActorPermissions returns the level of access the given actor has for the requested
// content.
//
// If the context is unauthenticated, ErrUnauthenticated is returned. If the context is
// internal, Read permissions is granted.
func ActorPermissions(ctx context.Context, s SubRepoPermissionChecker, a *actor.Actor, content RepoContent) (Perms, error) {
	// Check config here, despite checking again in the s.Permissions implementation,
	// because we also make some permissions decisions here.
	if doCheck, err := actorSubRepoEnabled(s, a); err != nil {
		return None, err
	} else if !doCheck {
		return Read, nil
	}

	perms, err := s.Permissions(ctx, a.UID, content)
	if err != nil {
		return None, errors.Wrapf(err, "getting actor permissions for actor: %d", a.UID)
	}
	return perms, nil
}

// actorSubRepoEnabled returns true if you should do sub repo permission
// checks with s for actor a. If false, you can skip sub repo checks.
//
// If the actor represents an anonymous user, ErrUnauthenticated is returned.
func actorSubRepoEnabled(s SubRepoPermissionChecker, a *actor.Actor) (bool, error) {
	if !SubRepoEnabled(s) {
		return false, nil
	}
	if a.IsInternal() {
		return false, nil
	}
	if !a.IsAuthenticated() {
		return false, &ErrUnauthenticated{}
	}
	return true, nil
}

// SubRepoEnabled takes a SubRepoPermissionChecker and returns true if the checker is not nil and is enabled
func SubRepoEnabled(checker SubRepoPermissionChecker) bool {
	return checker != nil && checker.Enabled()
}

// SubRepoEnabledForRepoID takes a SubRepoPermissionChecker and repoID and returns true if sub-repo
// permissions are enabled for a repo with given repoID
func SubRepoEnabledForRepoID(ctx context.Context, checker SubRepoPermissionChecker, repoID api.RepoID) (bool, error) {
	if !SubRepoEnabled(checker) {
		return false, nil
	}
	return checker.EnabledForRepoId(ctx, repoID)
}

// SubRepoEnabledForRepo takes a SubRepoPermissionChecker and repo name and returns true if sub-repo
// permissions are enabled for the given repo
func SubRepoEnabledForRepo(ctx context.Context, checker SubRepoPermissionChecker, repo api.RepoName) (bool, error) {
	if !SubRepoEnabled(checker) {
		return false, nil
	}
	return checker.EnabledForRepo(ctx, repo)
}

var (
	metricCanReadPathsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "authz_sub_repo_perms_can_read_paths_duration_seconds",
		Help: "Time spent checking permissions for files for an actor.",
	}, []string{"any", "result", "error"})
	metricCanReadPathsLenTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "authz_sub_repo_perms_can_read_paths_len_total",
		Help: "The total number of paths considered for permissions checking.",
	}, []string{"any", "result"})
)

func canReadPaths(ctx context.Context, checker SubRepoPermissionChecker, repo api.RepoName, paths []string, any bool) (result bool, err error) {
	a := actor.FromContext(ctx)
	if doCheck, err := actorSubRepoEnabled(checker, a); err != nil {
		return false, err
	} else if !doCheck {
		return true, nil
	}

	start := time.Now()
	var checkPathPermsCount int
	defer func() {
		anyS := strconv.FormatBool(any)
		resultS := strconv.FormatBool(result)
		errS := strconv.FormatBool(err != nil)
		metricCanReadPathsLenTotal.WithLabelValues(anyS, resultS).Add(float64(checkPathPermsCount))
		metricCanReadPathsDuration.WithLabelValues(anyS, resultS, errS).Observe(time.Since(start).Seconds())
	}()

	checkPathPerms, err := checker.FilePermissionsFunc(ctx, a.UID, repo)
	if err != nil {
		return false, err
	}

	for _, p := range paths {
		checkPathPermsCount++
		perms, err := checkPathPerms(p)
		if err != nil {
			return false, err
		}
		if !perms.Include(Read) && !any {
			return false, nil
		} else if perms.Include(Read) && any {
			return true, nil
		}
	}

	return !any, nil
}

// CanReadAllPaths returns true if the actor can read all paths.
func CanReadAllPaths(ctx context.Context, checker SubRepoPermissionChecker, repo api.RepoName, paths []string) (bool, error) {
	return canReadPaths(ctx, checker, repo, paths, false)
}

// CanReadAnyPath returns true if the actor can read any path in the list of paths.
func CanReadAnyPath(ctx context.Context, checker SubRepoPermissionChecker, repo api.RepoName, paths []string) (bool, error) {
	return canReadPaths(ctx, checker, repo, paths, true)
}

var (
	metricFilterActorPathsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "authz_sub_repo_perms_filter_actor_paths_duration_seconds",
		Help: "Time spent checking permissions for files for an actor.",
	}, []string{"error"})
	metricFilterActorPathsLenTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "authz_sub_repo_perms_filter_actor_paths_len_total",
		Help: "The total number of paths considered for permissions filtering.",
	})
)

// FilterActorPaths will filter the given list of paths for the given actor
// returning on paths they are allowed to read.
func FilterActorPaths(ctx context.Context, checker SubRepoPermissionChecker, a *actor.Actor, repo api.RepoName, paths []string) (_ []string, err error) {
	if doCheck, err := actorSubRepoEnabled(checker, a); err != nil {
		return nil, errors.Wrap(err, "checking sub-repo permissions")
	} else if !doCheck {
		return paths, nil
	}

	start := time.Now()
	var checkPathPermsCount int
	defer func() {
		metricFilterActorPathsLenTotal.Add(float64(checkPathPermsCount))
		metricFilterActorPathsDuration.WithLabelValues(strconv.FormatBool(err != nil)).Observe(time.Since(start).Seconds())
	}()

	checkPathPerms, err := checker.FilePermissionsFunc(ctx, a.UID, repo)
	if err != nil {
		return nil, errors.Wrap(err, "checking sub-repo permissions")
	}

	filtered := make([]string, 0, len(paths))
	for _, p := range paths {
		checkPathPermsCount++
		perms, err := checkPathPerms(p)
		if err != nil {
			return nil, errors.Wrap(err, "checking sub-repo permissions")
		}
		if perms.Include(Read) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// FilterActorPath will filter the given path for the given actor
// returning true if the path is allowed to read.
func FilterActorPath(ctx context.Context, checker SubRepoPermissionChecker, a *actor.Actor, repo api.RepoName, path string) (bool, error) {
	if !SubRepoEnabled(checker) {
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

func FilterActorFileInfos(ctx context.Context, checker SubRepoPermissionChecker, a *actor.Actor, repo api.RepoName, fis []fs.FileInfo) (_ []fs.FileInfo, err error) {
	if doCheck, err := actorSubRepoEnabled(checker, a); err != nil {
		return nil, errors.Wrap(err, "checking sub-repo permissions")
	} else if !doCheck {
		return fis, nil
	}

	start := time.Now()
	var checkPathPermsCount int
	defer func() {
		// we intentionally use the same metric, since we are essentially
		// measuring the same operation.
		metricFilterActorPathsLenTotal.Add(float64(checkPathPermsCount))
		metricFilterActorPathsDuration.WithLabelValues(strconv.FormatBool(err != nil)).Observe(time.Since(start).Seconds())
	}()

	checkPathPerms, err := checker.FilePermissionsFunc(ctx, a.UID, repo)
	if err != nil {
		return nil, errors.Wrap(err, "checking sub-repo permissions")
	}

	filtered := make([]fs.FileInfo, 0, len(fis))
	for _, fi := range fis {
		checkPathPermsCount++
		perms, err := checkPathPerms(fileInfoPath(fi))
		if err != nil {
			return nil, err
		}
		if perms.Include(Read) {
			filtered = append(filtered, fi)
		}
	}
	return filtered, nil
}

func FilterActorFileInfo(ctx context.Context, checker SubRepoPermissionChecker, a *actor.Actor, repo api.RepoName, fi fs.FileInfo) (bool, error) {
	rc := RepoContent{
		Repo: repo,
		Path: fileInfoPath(fi),
	}
	perms, err := ActorPermissions(ctx, checker, a, rc)
	if err != nil {
		return false, errors.Wrap(err, "checking sub-repo permissions")
	}
	return perms.Include(Read), nil
}

// fileInfoPath returns path for a fi as used by our sub repo filtering. If fi
// is a dir, the path has a trailing slash.
func fileInfoPath(fi fs.FileInfo) string {
	if fi.IsDir() {
		return fi.Name() + "/"
	}
	return fi.Name()
}
