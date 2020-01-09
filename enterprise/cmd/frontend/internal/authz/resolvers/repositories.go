package resolvers

import (
	"context"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

var _ graphqlbackend.RepositoryConnectionResolver = &repositoryConnectionResolver{}

type repositoryConnectionResolver struct {
	ids *roaring.Bitmap

	first int32
	after *string

	// cache results because they are used by multiple fields
	once        sync.Once
	repos       []*types.Repo
	endCursor   string
	hasNextPage bool
	err         error
}

// 🚨 SECURITY: It is the caller's responsibility to ensure the current authenticated user
// is the site admin because this method computes data from all available information in
// the database.
func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*types.Repo, error) {
	r.once.Do(func() {
		// Create the bitmap iterator and advance to the next value of r.after.
		var afterID api.RepoID
		if r.after != nil {
			afterID, r.err = graphqlbackend.UnmarshalRepositoryID(graphql.ID(*r.after))
			if r.err != nil {
				return
			}
		}
		iter := r.ids.Iterator()
		iter.AdvanceIfNeeded(uint32(afterID) + 1) // Plus 1 since r.after should be excluded.

		repoIDs := make([]api.RepoID, 0, r.first)
		for iter.HasNext() {
			repoIDs = append(repoIDs, api.RepoID(iter.Next()))
			if len(repoIDs) >= int(r.first) {
				break
			}
		}

		r.repos, r.err = db.Repos.GetByIDs(ctx, repoIDs...)
		if r.err != nil {
			return
		}

		r.endCursor = string(graphqlbackend.MarshalRepositoryID(repoIDs[len(repoIDs)-1]))
		r.hasNextPage = iter.HasNext()
	})
	return r.repos, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*graphqlbackend.RepositoryResolver, len(repos))
	for i := range repos {
		resolvers[i] = graphqlbackend.NewRepositoryResolver(repos[i])
	}
	return resolvers, nil
}

func (r *repositoryConnectionResolver) TotalCount(ctx context.Context, args *graphqlbackend.TotalCountArgs) (countptr *int32, err error) {
	count := int32(r.ids.GetCardinality())
	return &count, nil
}

func (r *repositoryConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.hasNextPage {
		return graphqlutil.NextPageCursor(r.endCursor), nil
	}
	return graphqlutil.HasNextPage(false), nil
}
