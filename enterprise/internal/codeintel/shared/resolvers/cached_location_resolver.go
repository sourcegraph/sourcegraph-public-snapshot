package sharedresolvers

import (
	"context"
	"io/fs"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
	repositoryCache *DoubleLockedCache[api.RepoID, cachedRepositoryResolver]
}

type cachedRepositoryResolver struct {
	repositoryResolver *RepositoryResolver
	commitCache        *DoubleLockedCache[string, cachedCommitResolver]
}

type cachedCommitResolver struct {
	commitResolver *GitCommitResolver
	dirCache       *DoubleLockedCache[string, *GitTreeEntryResolver]
	pathCache      *DoubleLockedCache[string, *GitTreeEntryResolver]
}

func (r *GitTreeEntryResolver) RecordID() string        { return r.stat.Name() }
func (r cachedRepositoryResolver) RecordID() api.RepoID { return r.repositoryResolver.repo.ID }
func (r cachedCommitResolver) RecordID() string         { return string(r.commitResolver.oid) }

func NewCachedLocationResolver(
	cloneURLToRepoName CloneURLToRepoNameFunc,
	repoStore database.RepoStore,
	gitserverClient gitserver.Client,
) *CachedLocationResolver {
	resolveRepo := func(ctx context.Context, repoID api.RepoID) (*RepositoryResolver, error) {
		repo, err := repoStore.Get(ctx, repoID)
		if err != nil {
			if errcode.IsNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		return NewRepositoryResolver(repo), nil
	}

	resolveCommit := func(ctx context.Context, repositoryResolver *RepositoryResolver, commit string) (*GitCommitResolver, error) {
		commitID, err := gitserverClient.ResolveRevision(ctx, repositoryResolver.repo.Name, commit, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
		if err != nil {
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
				return nil, nil
			}
			return nil, err
		}

		commitResolver := NewGitCommitResolver(repositoryResolver, commitID)
		commitResolver.inputRev = &commit
		return commitResolver, nil
	}

	resolvePath := func(commitResolver *GitCommitResolver, path string, isDir bool) *GitTreeEntryResolver {
		return NewGitTreeEntryResolver(cloneURLToRepoName, commitResolver, CreateFileInfo(path, isDir))
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
				dirCache: NewDoubleLockedCache(MultiFactoryFromFactoryFunc(func(ctx context.Context, path string) (*GitTreeEntryResolver, error) {
					return resolvePath(commitResolver, path, true), nil
				})),
				pathCache: NewDoubleLockedCache(MultiFactoryFromFactoryFunc(func(ctx context.Context, path string) (*GitTreeEntryResolver, error) {
					return resolvePath(commitResolver, path, false), nil
				})),
			}, nil
		}

		return &cachedRepositoryResolver{
			repositoryResolver: repositoryResolver,
			commitCache:        NewDoubleLockedCache(MultiFactoryFromFallibleFactoryFunc(resolveCommitCached)),
		}, nil
	}

	return &CachedLocationResolver{
		repositoryCache: NewDoubleLockedCache(MultiFactoryFromFallibleFactoryFunc(resolveRepositoryCached)),
	}
}

// Repository resolves (once) the given repository. May return nil if the repository is not available.
func (r *CachedLocationResolver) Repository(ctx context.Context, id api.RepoID) (*RepositoryResolver, error) {
	repositoryWrapper, ok, err := r.repositoryCache.GetOrLoad(ctx, id)
	if err != nil || !ok {
		return nil, err
	}

	return repositoryWrapper.repositoryResolver, nil
}

// Commit resolves (once) the given repository and commit. May return nil if the repository or commit is unknown.
func (r *CachedLocationResolver) Commit(ctx context.Context, id api.RepoID, commit string) (*GitCommitResolver, error) {
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
func (r *CachedLocationResolver) Path(ctx context.Context, id api.RepoID, commit, path string, isDir bool) (*GitTreeEntryResolver, error) {
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
	return resolver, err
}

type fileInfo struct {
	path  string
	size  int64
	isDir bool
}

func CreateFileInfo(path string, isDir bool) fs.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return f.size }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() any           { return any(nil) }
