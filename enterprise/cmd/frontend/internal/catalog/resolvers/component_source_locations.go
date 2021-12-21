package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *componentResolver) sourceLocations(ctx context.Context) ([]*componentSourceLocationResolver, error) {
	getSourceLocations := func(ctx context.Context) ([]*componentSourceLocationResolver, error) {
		slocs := make([]*componentSourceLocationResolver, 0, len(r.component.SourceLocations))
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
			for j, path := range sloc.Paths {
				treeResolver := gql.NewGitTreeEntryResolver(r.db, commitResolver, gql.CreateFileInfo(path, true))
				slocs = append(slocs, &componentSourceLocationResolver{
					repo:   repoResolver,
					commit: commitResolver,
					tree:   treeResolver,

					repoName:  repo.Name,
					commitID:  commit.ID,
					path:      path,
					isPrimary: i == 0 && j == 0,
				})
			}
		}
		return slocs, nil
	}

	r.sourceLocationsOnce.Do(func() {
		r.sourceLocationsCached, r.sourceLocationsErr = getSourceLocations(ctx)
	})
	return r.sourceLocationsCached, r.sourceLocationsErr
}

func (r *componentResolver) SourceLocations(ctx context.Context) ([]gql.ComponentSourceLocationResolver, error) {
	slocs, err := r.sourceLocations(ctx)
	if err != nil {
		return nil, err
	}

	rs := make([]gql.ComponentSourceLocationResolver, len(slocs))
	for i, sloc := range slocs {
		rs[i] = sloc
	}

	return rs, nil
}

func groupSourceLocationsByRepo(slocs []*componentSourceLocationResolver) map[*componentSourceLocationResolver][]string {
	byRepoName := map[api.RepoName][]*componentSourceLocationResolver{}
	for _, sloc := range slocs {
		byRepoName[sloc.repoName] = append(byRepoName[sloc.repoName], sloc)
	}

	grouped := make(map[*componentSourceLocationResolver][]string, len(byRepoName))
	for _, repoSlocs := range byRepoName {
		var paths []string
		for _, sloc := range repoSlocs {
			paths = append(paths, sloc.path)
		}
		grouped[repoSlocs[0]] = paths
	}
	return grouped
}

type componentSourceLocationResolver struct {
	repo   *gql.RepositoryResolver
	commit *gql.GitCommitResolver
	tree   *gql.GitTreeEntryResolver

	repoName  api.RepoName
	commitID  api.CommitID
	path      string
	isPrimary bool
}

func (r *componentSourceLocationResolver) ID() graphql.ID {
	return relay.MarshalID("ComponentSourceLocation", []interface{}{r.repoName, r.path})
}

func (r *componentSourceLocationResolver) RepositoryName() string { return string(r.repoName) }

func (r *componentSourceLocationResolver) Repository() (*gql.RepositoryResolver, error) {
	return r.repo, nil
}

func (r *componentSourceLocationResolver) Path() *string {
	if r.path == "" {
		return nil
	}
	return &r.path
}

func (r *componentSourceLocationResolver) IsEntireRepository() bool { return r.path == "" }

func (r *componentSourceLocationResolver) TreeEntry() (*gql.GitTreeEntryResolver, error) {
	return r.tree, nil
}

func (r *componentSourceLocationResolver) IsPrimary() bool { return r.isPrimary }
