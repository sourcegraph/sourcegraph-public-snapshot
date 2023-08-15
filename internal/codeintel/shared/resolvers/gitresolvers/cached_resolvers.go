package gitresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/dataloader"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CachedLocationResolver resolves repositories, commits, and git tree entries and caches the resulting
// resolvers so that the same request does not re-resolve the same repository, commit, or path multiple
// times during execution.
//
// This cache reduces duplicate database and gitserver calls when resolving repositories, commits, or
// locations (but does not batch or pre-fetch). Location resolvers generally have a small set of paths
// with large multiplicity, so the savings here can be significant.
type CachedLocationResolver struct {
	repositoryCache *dataloader.DoubleLockedCache[api.RepoID, cachedRepositoryResolver]
}

type cachedRepositoryResolver struct {
	repositoryResolver resolverstubs.RepositoryResolver
	commitCache        *dataloader.DoubleLockedCache[string, cachedCommitResolver]
}

type cachedCommitResolver struct {
	commitResolver resolverstubs.GitCommitResolver
	dirCache       *dataloader.DoubleLockedCache[string, *cachedGitTreeEntryResolver]
	pathCache      *dataloader.DoubleLockedCache[string, *cachedGitTreeEntryResolver]
}

type cachedGitTreeEntryResolver struct {
	treeEntryResolver resolverstubs.GitTreeEntryResolver
}

// DoubleLockedCache[K, V] requires V to conform to Identifier[K]
func (r cachedRepositoryResolver) RecordID() api.RepoID { return r.repositoryResolver.RepoID() }
func (r cachedCommitResolver) RecordID() string         { return string(r.commitResolver.OID()) }
func (r *cachedGitTreeEntryResolver) RecordID() string  { return r.treeEntryResolver.Path() }

func newCachedLocationResolver(
	repoStore database.RepoStore,
	gitserverClient gitserver.Client,
) *CachedLocationResolver {
	resolveRepo := func(ctx context.Context, repoID api.RepoID) (resolverstubs.RepositoryResolver, error) {
		resolver, err := NewRepositoryFromID(ctx, repoStore, int(repoID))
		if errcode.IsNotFound(err) {
			return nil, nil
		}

		return resolver, err
	}

	resolveCommit := func(ctx context.Context, repositoryResolver resolverstubs.RepositoryResolver, commit string) (resolverstubs.GitCommitResolver, error) {
		commitID, err := gitserverClient.ResolveRevision(ctx, api.RepoName(repositoryResolver.Name()), commit, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
		if err != nil {
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
				return nil, nil
			}
			return nil, err
		}

		commitResolver := NewGitCommitResolver(gitserverClient, repositoryResolver, commitID, commit)
		return commitResolver, nil
	}

	resolvePath := func(commitResolver resolverstubs.GitCommitResolver, path string, isDir bool) *cachedGitTreeEntryResolver {
		return &cachedGitTreeEntryResolver{NewGitTreeEntryResolver(commitResolver, path, isDir, gitserverClient)}
	}

	resolveRepositoryCached := func(ctx context.Context, repoID api.RepoID) (*cachedRepositoryResolver, error) {
		repositoryResolver, err := resolveRepo(ctx, repoID)
		if err != nil || repositoryResolver == nil {
			return nil, err
		}

		resolveCommitCached := func(ctx context.Context, commit string) (*cachedCommitResolver, error) {
			commitResolver, err := resolveCommit(ctx, repositoryResolver, commit)
			if err != nil || commitResolver == nil {
				return nil, err
			}

			return &cachedCommitResolver{
				commitResolver: commitResolver,
				dirCache: dataloader.NewDoubleLockedCache(dataloader.NewMultiFactoryFromFactoryFunc(func(ctx context.Context, path string) (*cachedGitTreeEntryResolver, error) {
					return resolvePath(commitResolver, path, true), nil
				})),
				pathCache: dataloader.NewDoubleLockedCache(dataloader.NewMultiFactoryFromFactoryFunc(func(ctx context.Context, path string) (*cachedGitTreeEntryResolver, error) {
					return resolvePath(commitResolver, path, false), nil
				})),
			}, nil
		}

		return &cachedRepositoryResolver{
			repositoryResolver: repositoryResolver,
			commitCache:        dataloader.NewDoubleLockedCache(dataloader.NewMultiFactoryFromFallibleFactoryFunc(resolveCommitCached)),
		}, nil
	}

	return &CachedLocationResolver{
		repositoryCache: dataloader.NewDoubleLockedCache(dataloader.NewMultiFactoryFromFallibleFactoryFunc(resolveRepositoryCached)),
	}
}

// Repository resolves (once) the given repository. May return nil if the repository is not available.
func (r *CachedLocationResolver) Repository(ctx context.Context, id api.RepoID) (resolverstubs.RepositoryResolver, error) {
	repositoryWrapper, ok, err := r.repositoryCache.GetOrLoad(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	return repositoryWrapper.repositoryResolver, nil
}

// Commit resolves (once) the given repository and commit. May return nil if the repository or commit is unknown.
func (r *CachedLocationResolver) Commit(ctx context.Context, id api.RepoID, commit string) (resolverstubs.GitCommitResolver, error) {
	repositoryWrapper, ok, err := r.repositoryCache.GetOrLoad(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	commitWrapper, ok, err := repositoryWrapper.commitCache.GetOrLoad(ctx, commit)
	if err != nil || !ok {
		return nil, err
	}

	return commitWrapper.commitResolver, nil
}

// Path resolves (once) the given repository, commit, and path. May return nil if the repository, commit, or path is unknown.
func (r *CachedLocationResolver) Path(ctx context.Context, id api.RepoID, commit, path string, isDir bool) (resolverstubs.GitTreeEntryResolver, error) {
	repositoryWrapper, ok, err := r.repositoryCache.GetOrLoad(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	commitWrapper, ok, err := repositoryWrapper.commitCache.GetOrLoad(ctx, commit)
	if err != nil || !ok {
		return nil, err
	}

	cache := commitWrapper.pathCache
	if isDir {
		cache = commitWrapper.dirCache
	}

	resolver, _, err := cache.GetOrLoad(ctx, path)
	return resolver.treeEntryResolver, err
}
