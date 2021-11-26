package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type sourceLocation struct {
	repo   *gql.RepositoryResolver
	commit *gql.GitCommitResolver
	trees  []*gql.GitTreeEntryResolver

	repoName api.RepoName
	commitID api.CommitID
	paths    []string
}

func (r *componentResolver) sourceLocations(ctx context.Context) ([]*sourceLocation, error) {
	getSourceLocations := func(ctx context.Context) ([]*sourceLocation, error) {
		slocs := make([]*sourceLocation, len(r.component.SourceLocations))
		for i, sloc := range r.component.SourceLocations {
			// 🚨 SECURITY: database.Repos.Get uses the authzFilter under the hood and filters out
			// repositories that the user doesn't have access to.
			repo, err := r.db.Repos().GetByName(ctx, sloc.Repo)
			if err != nil {
				return nil, err
			}
			repoResolver := gql.NewRepositoryResolver(r.db, repo)

			commit, err := git.GetCommit(ctx, sloc.Repo, "HEAD", git.ResolveRevisionOptions{})
			if err != nil {
				return nil, err
			}
			commitResolver := gql.NewGitCommitResolver(r.db, repoResolver, commit.ID, commit)

			treeResolvers := make([]*gql.GitTreeEntryResolver, len(sloc.Paths))
			for i, path := range sloc.Paths {
				treeResolvers[i] = gql.NewGitTreeEntryResolver(r.db, commitResolver, gql.CreateFileInfo(path, true))
			}

			slocs[i] = &sourceLocation{
				repo:   repoResolver,
				commit: commitResolver,
				trees:  treeResolvers,

				repoName: repo.Name,
				commitID: commit.ID,
				paths:    sloc.Paths,
			}
		}
		return slocs, nil
	}

	r.sourceLocationsOnce.Do(func() {
		r.sourceLocationsCached, r.sourceLocationsErr = getSourceLocations(ctx)
	})
	return r.sourceLocationsCached, r.sourceLocationsErr
}

func (r *componentResolver) SourceLocations(ctx context.Context) ([]*gql.GitTreeEntryResolver, error) {
	slocs, err := r.sourceLocations(ctx)
	if err != nil {
		return nil, err
	}

	var treeResolvers []*gql.GitTreeEntryResolver
	for _, sloc := range slocs {
		treeResolvers = append(treeResolvers, sloc.trees...)
	}

	return treeResolvers, nil
}
