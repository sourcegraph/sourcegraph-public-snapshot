package authz

import (
	"context"
	"io/fs"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
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

	// FilePermissionsFunc returns a FilePermissionFunc for userID in repo.
	// This function should only be used during the lifetime of a request. It
	// exists to amortize the cost of checking many files in a repo.
	//
	// If the userID represents an anonymous user, ErrUnauthenticated is returned.
	FilePermissionsFunc(ctx context.Context, userID int32, repo api.RepoName) (FilePermissionFunc, error)

	// Enabled indicates whether sub-repo permissions are enabled.
	Enabled() bool

	// EnabledForRepoID indicates whether sub-repo permissions are enabled for the given repoID
	EnabledForRepoID(ctx context.Context, repoID api.RepoID) (bool, error)

	// EnabledForRepo indicates whether sub-repo permissions are enabled for the given repo
	EnabledForRepo(ctx context.Context, repo api.RepoName) (bool, error)
}

// DefaultSubRepoPermsChecker allows us to use a single instance with a shared
// cache and database connection. Since we don't have a database connection at
// initialisation time, services that require this client should initialise it in
// their main function.
var DefaultSubRepoPermsChecker SubRepoPermissionChecker = &noopPermsChecker{}

type noopPermsChecker struct{}

func (*noopPermsChecker) Permissions(_ context.Context, _ int32, _ RepoContent) (Perms, error) {
	return None, nil
}

func (*noopPermsChecker) FilePermissionsFunc(_ context.Context, _ int32, _ api.RepoName) (FilePermissionFunc, error) {
	return func(path string) (Perms, error) {
		return None, nil
	}, nil
}

func (*noopPermsChecker) Enabled() bool {
	return false
}

func (*noopPermsChecker) EnabledForRepoID(_ context.Context, _ api.RepoID) (bool, error) {
	return false, nil
}

func (*noopPermsChecker) EnabledForRepo(_ context.Context, _ api.RepoName) (bool, error) {
	return false, nil
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
	return checker.EnabledForRepoID(ctx, repoID)
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
