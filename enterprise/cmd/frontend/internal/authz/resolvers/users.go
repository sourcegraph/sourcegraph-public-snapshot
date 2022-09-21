package resolvers

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var _ graphqlbackend.UserConnectionResolver = &userConnectionResolver{}

// userConnectionResolver resolves a list of user from the roaring bitmap with pagination.
type userConnectionResolver struct {
	ids []int32 // Sorted slice in ascending order of user IDs.
	db  database.DB

	first int32
	after *string

	// cache results because they are used by multiple fields
	once     sync.Once
	users    []*types.User
	pageInfo *graphqlutil.PageInfo
	err      error
}

// 🚨 SECURITY: It is the caller's responsibility to ensure the current authenticated user
// is the site admin because this method computes data from all available information in
// the database.
// This function takes returns a pagination of the user IDs
//
//	r.ids - the full slice of sorted user IDs
//	r.after - (optional) the user ID to start the paging after (does not include the after ID itself)
//	r.first - the # of user IDs to return
func (r *userConnectionResolver) compute(ctx context.Context) ([]*types.User, *graphqlutil.PageInfo, error) {
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
			if len(idSubset) == 0 {
				// r.after is set, but there are no elements larger than it, so return empty slice.
				r.users = []*types.User{}
				r.pageInfo = graphqlutil.HasNextPage(false)
				return
			}
		}
		// If we have more ids than we need, trim them
		if int32(len(idSubset)) > r.first {
			idSubset = idSubset[:r.first]
		}

		r.users, r.err = r.db.Users().List(ctx, &database.UsersListOptions{
			UserIDs: idSubset,
		})
		if r.err != nil {
			return
		}

		// No more user IDs to paginate through.
		if idSubset[len(idSubset)-1] == r.ids[len(r.ids)-1] {
			r.pageInfo = graphqlutil.HasNextPage(false)
		} else { // Additional user IDs to paginate through.
			endCursor := string(graphqlbackend.MarshalUserID(idSubset[len(idSubset)-1]))
			r.pageInfo = graphqlutil.NextPageCursor(endCursor)
		}
	})
	return r.users, r.pageInfo, r.err
}

func (r *userConnectionResolver) Nodes(ctx context.Context) ([]*graphqlbackend.UserResolver, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	users, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*graphqlbackend.UserResolver, len(users))
	for i := range users {
		resolvers[i] = graphqlbackend.NewUserResolver(r.db, users[i])
	}
	return resolvers, nil
}

func (r *userConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return -1, err
	}

	return int32(len(r.ids)), nil
}

func (r *userConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	_, pageInfo, err := r.compute(ctx)
	return pageInfo, err
}
