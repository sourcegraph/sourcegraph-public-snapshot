package resolvers

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type LocationsQueryOptions struct {
	Operation string
	RepoName  string
	Commit    graphqlbackend.GitObjectID
	Path      string
	Line      int32
	Character int32
	DumpID    int64
	Limit     *int32
	NextURL   *string
}

type locationConnectionResolver struct {
	locations []*lsif.LSIFLocation
	nextURL   string
}

var _ graphqlbackend.LocationConnectionResolver = &locationConnectionResolver{}

func (r *locationConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LocationResolver, error) {
	gitTreeResolvers, err := resolveGitTrees(ctx, partitionPathsByCommitByRepository(r.locations))
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LocationResolver
	for _, location := range r.locations {
		l = append(l, graphqlbackend.NewLocationResolver(
			gitTreeResolvers[location.Repository][location.Commit][location.Path],
			&location.Range,
		))
	}

	return l, nil
}

func (r *locationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if r.nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(r.nextURL))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

type pathsByCommitByRepositoryMap map[string]pathsByCommitMap
type pathsByCommitMap map[string]pathsSet
type pathsSet map[string]struct{}

// partitionPathsByCommitByRepository partitions locations as returned by a definitions or references
// query from the LSIF server into a nested map of the form `{repository} -> {commit} -> {set of paths}`.
// The set of paths are encoded as a map from the path name to an empty struct.
func partitionPathsByCommitByRepository(locations []*lsif.LSIFLocation) pathsByCommitByRepositoryMap {
	pathsByCommitByRepository := pathsByCommitByRepositoryMap{}
	for _, location := range locations {
		if _, ok := pathsByCommitByRepository[location.Repository]; !ok {
			pathsByCommitByRepository[location.Repository] = pathsByCommitMap{}
		}

		if _, ok := pathsByCommitByRepository[location.Repository][location.Commit]; !ok {
			pathsByCommitByRepository[location.Repository][location.Commit] = pathsSet{}
		}

		pathsByCommitByRepository[location.Repository][location.Commit][location.Path] = struct{}{}
	}

	return pathsByCommitByRepository
}

type resolversByPathByCommitByRepostioryMap map[string]resolversByPathByCommitMap
type resolversByPathByCommitMap map[string]resolversByPathMap
type resolversByPathMap map[string]*graphqlbackend.GitTreeEntryResolver

// resolveGitTrees takes the map produced by `resolveGitTrees` and returns a symmetric map, where the
// empty structs in the path set are replaced by git tree resolvers. This ensures that each repository,
// commit, and path are each resolved only once.
//
// This method resolves repositories. See `resolveGitCommitsForRepositories` and `resolveGitTreeForPaths`
// (which are called from here) for the resolution of commits and git trees, respectively.
func resolveGitTrees(ctx context.Context, pathsByCommitByRepository pathsByCommitByRepositoryMap) (resolversByPathByCommitByRepostioryMap, error) {
	resolversByRepositories := resolversByPathByCommitByRepostioryMap{}

	for repoName, commits := range pathsByCommitByRepository {
		repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
		if err != nil {
			return nil, err
		}
		repositoryResolver := graphqlbackend.NewRepositoryResolver(repo)

		resolvers, err := resolveGitCommitsForRepositories(ctx, repositoryResolver, commits)
		if err != nil {
			return nil, err
		}
		resolversByRepositories[repoName] = resolvers
	}

	return resolversByRepositories, nil
}

func resolveGitCommitsForRepositories(ctx context.Context, repositoryResolver *graphqlbackend.RepositoryResolver, pathsByCommit pathsByCommitMap) (resolversByPathByCommitMap, error) {
	resolversByCommit := resolversByPathByCommitMap{}

	for commit, paths := range pathsByCommit {
		commitResolver, err := repositoryResolver.Commit(ctx, &graphqlbackend.RepositoryCommitArgs{Rev: commit})
		if err != nil {
			return nil, err
		}

		resolvers, err := resolveGitTreeForPaths(ctx, commitResolver, paths)
		if err != nil {
			return nil, err
		}
		resolversByCommit[commit] = resolvers
	}

	return resolversByCommit, nil
}

func resolveGitTreeForPaths(ctx context.Context, commitResolver *graphqlbackend.GitCommitResolver, paths pathsSet) (resolversByPathMap, error) {
	resolversByPath := resolversByPathMap{}

	for path := range paths {
		gitTreeResolver, err := commitResolver.Blob(ctx, &struct{ Path string }{Path: path})
		if err != nil {
			return nil, err
		}
		resolversByPath[path] = gitTreeResolver
	}

	return resolversByPath, nil
}
