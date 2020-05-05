package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// resolveRepository returns a repository resolver for the given name.
func resolveRepository(ctx context.Context, repoID api.RepoID) (*graphqlbackend.RepositoryResolver, error) {
	repo, err := backend.Repos.Get(ctx, repoID)
	if err != nil {
		return nil, err
	}

	return graphqlbackend.NewRepositoryResolver(repo), nil
}

// resolveCommit returns the GitCommitResolver for the given repository and commit. If the
// commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func resolveCommit(ctx context.Context, repositoryResolver *graphqlbackend.RepositoryResolver, commit string) (*graphqlbackend.GitCommitResolver, error) {
	return resolveCommitFrom(ctx, repositoryResolver, commit)
}

// resolveCommitFrom returns the GitCommitResolver for the given repository resolver and commit.
// If the commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func resolveCommitFrom(ctx context.Context, repositoryResolver *graphqlbackend.RepositoryResolver, commit string) (*graphqlbackend.GitCommitResolver, error) {
	gitserverRepo, err := backend.CachedGitRepo(ctx, repositoryResolver.Type())
	if err != nil {
		return nil, err
	}

	commitID, err := git.ResolveRevision(ctx, *gitserverRepo, nil, commit, &git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		if gitserver.IsRevisionNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return repositoryResolver.CommitFromID(ctx, &graphqlbackend.RepositoryCommitArgs{Rev: commit}, commitID)
}

// resolvePath returns the GitTreeResolver for the given repository, commit, and path. If the
// commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func resolvePath(ctx context.Context, repoID api.RepoID, commit, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	repositoryResolver, err := resolveRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	commitResolver, err := resolveCommit(ctx, repositoryResolver, commit)
	if err != nil {
		return nil, err
	}

	return resolvePathFrom(ctx, commitResolver, path)
}

// resolvePath returns the GitTreeResolver for the given commit resolver, and path. If the
// commit resolver is nil, a nil resolver is returned. Any other error is returned unmodified.
func resolvePathFrom(ctx context.Context, commitResolver *graphqlbackend.GitCommitResolver, path string) (*graphqlbackend.GitTreeEntryResolver, error) {
	if commitResolver == nil {
		return nil, nil
	}

	return graphqlbackend.NewGitTreeEntryResolver(commitResolver, graphqlbackend.CreateFileInfo(path, true)), nil
}
