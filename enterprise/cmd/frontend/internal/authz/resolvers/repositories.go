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
	ids []int32

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
func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*types.Repo, *graphqlutil.PageInfo, error) {
	r.once.Do(func() {
		// Create the bitmap iterator and advance to the next value of r.after.
		var afterID api.RepoID
		if r.after != nil {
			afterID, r.err = graphqlbackend.UnmarshalRepositoryID(graphql.ID(*r.after))
			if r.err != nil {
				return
			}
		}
		repoIDs := make([]api.RepoID, 0, r.first)
		idsSize := int32(len(r.ids))

		// Generate a slice of repo IDs ranging from index after+1 to: after+first or until the end of the slice, whichever is less.
		if idsSize >= int32(afterID)+1 {
			count := int32(1)
			for _, id := range r.ids[afterID:] {
				if count > r.first {
					break
				}
				repoIDs = append(repoIDs, api.RepoID(id))
				count++
			}
		}

		r.repos, r.err = r.db.Repos().GetByIDs(ctx, repoIDs...)
		if r.err != nil {
			return
		}

		// No more user IDs to paginate through.
		if idsSize <= int32(afterID)+r.first {
			r.pageInfo = graphqlutil.HasNextPage(false)
		} else { // Additional user IDs to paginate through.
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
