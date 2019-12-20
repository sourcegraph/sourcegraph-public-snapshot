package resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type treeResolver struct {
	lsifDump *lsif.LSIFDump
}

func newTreeResolver(lsifDump *lsif.LSIFDump) *treeResolver {
	return &treeResolver{lsifDump: lsifDump}
}

// resolveRepository returns a repository resolver for the given name.
func (r *treeResolver) resolveRepository(ctx context.Context, repoName string) (*graphqlbackend.RepositoryResolver, error) {
	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		return nil, err
	}

	return graphqlbackend.NewRepositoryResolver(repo), nil
}

// resolveCommit returns the GitCommitResolver for the given repository and commit. If the
// commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func (r *treeResolver) resolveCommit(ctx context.Context, repoName, commit string) (*graphqlbackend.GitCommitResolver, error) {
	repositoryResolver, err := r.resolveRepository(ctx, repoName)
	if err != nil {
		return nil, err
	}

	return r.resolveCommitFrom(ctx, repositoryResolver, commit)
}

// resolveCommitFrom returns the GitCommitResolver for the given repository resolver and commit.
// If the commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func (r *treeResolver) resolveCommitFrom(ctx context.Context, repositoryResolver *graphqlbackend.RepositoryResolver, commit string) (*graphqlbackend.GitCommitResolver, error) {
	commitResolver, err := repositoryResolver.Commit(ctx, &graphqlbackend.RepositoryCommitArgs{Rev: commit})
	if err != nil {
		return nil, err
	}

	if commitResolver == nil && r.lsifDump != nil {
		fmt.Printf("GOTTA REMOVE THIS GUY: %#v\n", r.lsifDump)
	}

	return commitResolver, nil
}

// resolvePath returns the GitTreeResolver for the given repository, commit, and path. If the
// commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func (r *treeResolver) resolvePath(ctx context.Context, repoName, commit, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	commitResolver, err := r.resolveCommit(ctx, repoName, commit)
	if err != nil {
		return nil, err
	}

	return r.resolvePathFrom(ctx, commitResolver, path)
}

// resolvePath returns the GitTreeResolver for the given commit resolver, and path. If the
// commit resolver is nil, a nil resolver is returned. Any other error is returned unmodified.
func (r *treeResolver) resolvePathFrom(ctx context.Context, commitResolver *graphqlbackend.GitCommitResolver, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	if commitResolver == nil {
		return nil, nil
	}

	return graphqlbackend.NewGitTreeEntryResolver(commitResolver, graphqlbackend.CreateFileInfo(path, true)), nil
}
