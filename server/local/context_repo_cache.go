package local

import (
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/store"

	"golang.org/x/net/context"
)

// TODO(sqs): The repo cache is a quick workaround to improve
// performance by avoiding the garbage creation and duplicative effort
// by opening the same repo many times. There are maybe better
// solutions, such as using a long-running child process that
// maintains state in memory and knows how to evict cache entries, or
// an RPC interface to git shell commands invoked on an external
// server, etc.

type repoCacheKey string

// withRepoCache returns a copy of ctx that reuses the existing repo
// intead of reopening it (and rereading the .git/ dir's files, for
// example). It can significantly speed up operations that need to
// repeatedly open the same repository (e.g., Deltas.ListFiles, which
// issues many calls to RepoTree.Get, each of which must open the
// repository).
//
// The repo cache is opt-in because it is not totally safe. If the
// repo is modified on disk after we open it, the in-memory repo will
// not reflect the changes. For example, if we perform an operation
// that modifies the repo, such as "git merge", then the cache will
// not necessarily reflect the operation. As a general guideline, only
// use the repo cache for short-lived operations (such as handling a
// single gRPC method call) and only for "read" operations.
func withCachedRepo(ctx context.Context, repo string, obj vcs.Repository) context.Context {
	return context.WithValue(ctx, repoCacheKey(repo), repoCacheEntry{repo, obj})
}

// cachedRepoFromContext returns the cached, already opened repo, if
// it has been previously cached (using withCachedRepo). If no such
// repo exists in this context's repo cache, nil is returned.
func cachedRepoFromContext(ctx context.Context, repo string) vcs.Repository {
	e, ok := ctx.Value(repoCacheKey(repo)).(repoCacheEntry)
	if !ok {
		return nil
	}
	return e.obj
}

type repoCacheEntry struct {
	repo string         // repo URI
	obj  vcs.Repository // opened VCS repository
}

// cachedRepoVCSOpen wraps RepoVCS.Open with the repo cache, reusing a
// cached repo instead of calling RepoVCS.Open.
//
// It does not store the repo in the repo cache. To store it, you must
// opt-in to that behavior by applying withCachedRepo to the context.
func cachedRepoVCSOpen(ctx context.Context, repo string) (vcs.Repository, error) {
	if vcsRepo := cachedRepoFromContext(ctx, repo); vcsRepo != nil {
		return vcsRepo, nil
	}
	return store.RepoVCSFromContext(ctx).Open(ctx, repo)
}
