package resolvers

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var _ graphqlbackend.RepositoryConnectionResolver = &repositoryConnectionResolver{}

// repositoryConnectionResolver resolves a list of repositories from the roaring bitmap with pagination.
type repositoryConnectionResolver struct {
	db  database.DB
	ids []int32 // Sorted slice in ascending order of repo IDs.

	first int32
	after *string

	// cache results because they are used by multiple fields
	once     sync.Once
	repos    []*types.Repo
	pageInfo *graphqlutil.PageInfo
	err      error
}

// 🚨 SECURITY: It is the caller's responsibility to ensure the current authenticated user
// is the site admin because this method computes data from all available information in
// the database.
// This function takes returns a pagination of the repo IDs
//
//	r.ids - the full slice of sorted repo IDs
//	r.after - (optional) the repo ID to start the paging after (does not include the after ID itself)
//	r.first - the # of repo IDs to return
func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*types.Repo, *graphqlutil.PageInfo, error) {
	r.once.Do(func() {
		var idSubset []int32
		if r.after == nil {
			idSubset = r.ids
		} else {
			afterID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(*r.after))
			if err != nil {
				r.err = err
				return
			}
			for idx, id := range r.ids {
				if id == int32(afterID) {
					if idx < len(r.ids)-1 {
						idSubset = r.ids[idx+1:]
					}
					break
				} else if id > int32(afterID) {
					if idx < len(r.ids)-1 {
						idSubset = r.ids[idx:]
					}
					break
				}
			}
		}
		// No IDs to find, return early
		if len(idSubset) == 0 {
			r.repos = []*types.Repo{}
			r.pageInfo = graphqlutil.HasNextPage(false)
			return
		}
		// If we have more ids than we need, trim them
		if int32(len(idSubset)) > r.first {
			idSubset = idSubset[:r.first]
		}

		repoIDs := make([]api.RepoID, len(idSubset))
		for i := range idSubset {
			repoIDs[i] = api.RepoID(idSubset[i])
		}

		// TODO(asdine): GetByIDs now returns the complete repo information rather that only a subset.
		// Ensure this doesn't have an impact on performance and switch to using ListMinimalRepos if needed.
		r.repos, r.err = r.db.Repos().GetByIDs(ctx, repoIDs...)
		if r.err != nil {
			return
		}

		// The last id in this page is the last id in r.ids, no more pages
		if int32(repoIDs[len(repoIDs)-1]) == r.ids[len(r.ids)-1] {
			r.pageInfo = graphqlutil.HasNextPage(false)
		} else { // Additional repo IDs to paginate through.
			endCursor := string(graphqlbackend.MarshalRepositoryID(repoIDs[len(repoIDs)-1]))
			r.pageInfo = graphqlutil.NextPageCursor(endCursor)
		}
	})
	return r.repos, r.pageInfo, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repos, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*graphqlbackend.RepositoryResolver, len(repos))
	for i := range repos {
		resolvers[i] = graphqlbackend.NewRepositoryResolver(r.db, repos[i])
	}
	return resolvers, nil
}

func (r *repositoryConnectionResolver) TotalCount(ctx context.Context, args *graphqlbackend.TotalCountArgs) (*int32, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	count := int32(len(r.ids))
	return &count, nil
}

func (r *repositoryConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	_, pageInfo, err := r.compute(ctx)
	return pageInfo, err
}
