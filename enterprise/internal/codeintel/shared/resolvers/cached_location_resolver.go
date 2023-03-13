package sharedresolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CachedLocationResolver resolves repositories, commits, and git tree entries and caches the resulting
// resolvers so that the same request does not re-resolve the same repository, commit, or path multiple
// times during execution. This cache reduces the number duplicate of database and gitserver queries for
// definition, reference, and diagnostic queries, which return collections of results that often refer
// to a small set of repositories, commits, and paths with a large multiplicity.
//
// This resolver maintains a hierarchy of caches as a way to decrease lock contention. Resolution of a
// repository holds the top-level lock. Resolution of a commit holds a lock associated with the parent
// repository. Similarly, resolution of a path holds a lock associated with the parent commit.
type CachedLocationResolver struct {
	sync.RWMutex
	repositoryResolvers map[api.RepoID]*cachedRepositoryResolver
	cloneURLToRepoName  CloneURLToRepoNameFunc
	repoStore           database.RepoStore
	gitserverClient     gitserver.Client
	logger              log.Logger
}

type cachedRepositoryResolver struct {
	sync.RWMutex
	repositoryResolver *RepositoryResolver
	commitResolvers    map[string]*cachedCommitResolver
}

type cachedCommitResolver struct {
	sync.RWMutex
	commitResolver *GitCommitResolver
	pathResolvers  map[string]*GitTreeEntryResolver
}

type CachedLocationResolverFactory struct {
	cloneURLToRepoName CloneURLToRepoNameFunc
	repoStore          database.RepoStore
	gitserverClient    gitserver.Client
}

func NewCachedLocationResolverFactory(cloneURLToRepoName CloneURLToRepoNameFunc, repoStore database.RepoStore, gitserverClient gitserver.Client) *CachedLocationResolverFactory {
	return &CachedLocationResolverFactory{
		cloneURLToRepoName: cloneURLToRepoName,
		repoStore:          repoStore,
		gitserverClient:    gitserverClient,
	}
}

func (f *CachedLocationResolverFactory) Create() *CachedLocationResolver {
	return NewCachedLocationResolver(f.cloneURLToRepoName, f.repoStore, f.gitserverClient)
}

// NewCachedLocationResolver creates a location resolver with an empty cache.
func NewCachedLocationResolver(cloneURLToRepoName CloneURLToRepoNameFunc, repoStore database.RepoStore, gitserverClient gitserver.Client) *CachedLocationResolver {
	return &CachedLocationResolver{
		logger:              log.Scoped("CachedLocationResolver", ""),
		cloneURLToRepoName:  cloneURLToRepoName,
		repoStore:           repoStore,
		gitserverClient:     gitserverClient,
		repositoryResolvers: map[api.RepoID]*cachedRepositoryResolver{},
	}
}

// Repository resolves the repository with the given identifier. This method may return a nil resolver
// if the repository is not known by gitserver - this happens if there is exists still a bundle for a
// repo that has since been deleted.
func (r *CachedLocationResolver) Repository(ctx context.Context, id api.RepoID) (*RepositoryResolver, error) {
	resolver, err := r.cachedRepository(ctx, id)
	if err != nil || resolver == nil {
		return nil, err
	}
	return resolver.repositoryResolver, nil
}

// Commit resolves the git commit with the given repository identifier and commit hash. This method may
// return a nil resolver if the commit is not known by gitserver.
func (r *CachedLocationResolver) Commit(ctx context.Context, id api.RepoID, commit string) (*GitCommitResolver, error) {
	resolver, err := r.cachedCommit(ctx, id, commit)
	if err != nil || resolver == nil {
		return nil, err
	}
	return resolver.commitResolver, nil
}

// Path resolves the git tree entry with the given repository identifier, commit hash, and relative path.
// This method may return a nil resolver if the commit is not known by gitserver.
func (r *CachedLocationResolver) Path(ctx context.Context, id api.RepoID, commit, path string, isDir bool) (*GitTreeEntryResolver, error) {
	return r.cachedPath(ctx, id, commit, path, isDir)
}

// cachedRepository resolves the repository with the given identifier if the resulting resolver does not
// already exist in the cache. The cache is tested/populated with double-checked locking, which ensures
// that the resolver is created exactly once per GraphQL request.
//
// See https://en.wikipedia.org/wiki/Double-checked_locking.
func (r *CachedLocationResolver) cachedRepository(ctx context.Context, id api.RepoID) (*cachedRepositoryResolver, error) {
	// Fast-path cache check
	r.RLock()
	cachedResolver, ok := r.repositoryResolvers[id]
	r.RUnlock()
	if ok {
		return cachedResolver, nil
	}

	r.Lock()
	defer r.Unlock()

	// Check again once locked to avoid race
	if cachedResolver, ok := r.repositoryResolvers[id]; ok {
		return cachedResolver, nil
	}

	// Resolve new value and store in cache
	repositoryResolver, err := r.resolveRepository(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ensure value written to the cache is nil and not a nil resolver wrapped in a non-nil cached
	// commit resolver. Otherwise, a subsequent resolution of a path may result in a nil dereference.
	if repositoryResolver != nil {
		cachedResolver = &cachedRepositoryResolver{
			repositoryResolver: repositoryResolver,
			commitResolvers:    map[string]*cachedCommitResolver{},
		}
	}
	r.repositoryResolvers[id] = cachedResolver
	return cachedResolver, nil
}

// cachedCommit resolves the commit with the given repository identifier and commit hash if the resulting
// resolver does not already exist in the cache. The cache is tested/populated with double-checked locking,
// which ensures that the resolver is created exactly once per GraphQL request.
//
// See https://en.wikipedia.org/wiki/Double-checked_locking.
func (r *CachedLocationResolver) cachedCommit(ctx context.Context, id api.RepoID, commit string) (*cachedCommitResolver, error) {
	repositoryResolver, err := r.cachedRepository(ctx, id)
	if err != nil || repositoryResolver == nil {
		return nil, err
	}

	// Fast-path cache check
	repositoryResolver.RLock()
	cachedResolver, ok := repositoryResolver.commitResolvers[commit]
	repositoryResolver.RUnlock()
	if ok {
		return cachedResolver, nil
	}

	repositoryResolver.Lock()
	defer repositoryResolver.Unlock()

	// Check again once locked to avoid race
	if cachedResolver, ok := repositoryResolver.commitResolvers[commit]; ok {
		return cachedResolver, nil
	}

	// Resolve new value and store in cache
	commitResolver, err := r.resolveCommit(ctx, repositoryResolver.repositoryResolver, commit)
	if err != nil {
		return nil, err
	}

	// Ensure value written to the cache is nil and not a nil resolver wrapped in a non-nil cached
	// commit resolver. Otherwise, a subsequent resolution of a path may result in a nil dereference.
	if commitResolver != nil {
		cachedResolver = &cachedCommitResolver{
			commitResolver: commitResolver,
			pathResolvers:  map[string]*GitTreeEntryResolver{},
		}
	}
	repositoryResolver.commitResolvers[commit] = cachedResolver
	return cachedResolver, nil
}

// cachedPath resolves the commit with the given repository identifier, commit hash, and relative path
// if the resulting resolver does not already exist in the cache. The cache is tested/populated with
// double-checked locking, which ensures that the resolver is created exactly once per GraphQL request.
//
// See https://en.wikipedia.org/wiki/Double-checked_locking.
func (r *CachedLocationResolver) cachedPath(ctx context.Context, id api.RepoID, commit, path string, isDir bool) (*GitTreeEntryResolver, error) {
	commitResolver, err := r.cachedCommit(ctx, id, commit)
	if err != nil || commitResolver == nil {
		return nil, err
	}

	// Fast-path cache check
	commitResolver.Lock()
	cachedResolver, ok := commitResolver.pathResolvers[path]
	commitResolver.Unlock()
	if ok {
		return cachedResolver, nil
	}

	commitResolver.Lock()
	defer commitResolver.Unlock()

	// Check again once locked to avoid race
	if cachedResolver, ok := commitResolver.pathResolvers[path]; ok {
		return cachedResolver, nil
	}

	// Resolve new value and store in cache
	pathResolver := r.resolvePath(commitResolver.commitResolver, path, isDir)
	commitResolver.pathResolvers[path] = pathResolver
	return pathResolver, nil
}

// Repository resolves the repository with the given identifier. This method may return a nil resolver
// if the repository is not known by gitserver - this happens if there is exists still a bundle for a
// repo that has since been deleted. This method must be called only when constructing a resolver to
// populate the cache.
func (r *CachedLocationResolver) resolveRepository(ctx context.Context, id api.RepoID) (*RepositoryResolver, error) {
	repo, err := r.repoStore.Get(ctx, id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return NewRepositoryResolver(repo), nil
}

// Commit resolves the git commit with the given repository resolver and commit hash. This method may
// return a nil resolver if the commit is not known by gitserver. This method must be called only when
// constructing a resolver to populate the cache.
func (r *CachedLocationResolver) resolveCommit(ctx context.Context, repositoryResolver *RepositoryResolver, commit string) (*GitCommitResolver, error) {
	repo, err := repositoryResolver.Type(ctx)
	if err != nil {
		return nil, err
	}

	commitID, err := r.gitserverClient.ResolveRevision(ctx, repo.Name, commit, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}

	return repositoryResolver.commitFromID(&resolverstubs.RepositoryCommitArgs{Rev: commit}, commitID)
}

// Path resolves the git tree entry with the given commit resolver and relative path. This method must be
// called only when constructing a resolver to populate the cache.
func (r *CachedLocationResolver) resolvePath(commitResolver *GitCommitResolver, path string, isDir bool) *GitTreeEntryResolver {
	return NewGitTreeEntryResolver(r.cloneURLToRepoName, commitResolver, CreateFileInfo(path, isDir))
}
