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

var _ graphqlbackend.UserConnectionResolver = &userConnectionResolver{}

// userConnectionResolver resolves a list of user from the roaring bitmap with pagination.
type userConnectionResolver struct {
	ids []int32 //sorted slice of IDs
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
func (r *userConnectionResolver) compute(ctx context.Context) ([]*types.User, *graphqlutil.PageInfo, error) {
	r.once.Do(func() {
		// Create the bitmap iterator and advance to the next value of r.after.
		var afterID api.RepoID
		if r.after != nil {
			afterID, r.err = graphqlbackend.UnmarshalRepositoryID(graphql.ID(*r.after))
			if r.err != nil {
				return
			}
		}

		userIDs := make([]int32, 0, r.first)
		idsSize := int32(len(r.ids))

		// Generate a slice of user IDs ranging from index after+1 to: after+first or until the end of the slice, whichever is less.
		if idsSize >= int32(afterID)+1 {
			count := int32(1)
			for _, id := range r.ids[afterID:] {
				if count > r.first {
					break
				}
				userIDs = append(userIDs, id)
				count++
			}
		}

		r.users, r.err = r.db.Users().List(ctx, &database.UsersListOptions{
			UserIDs: userIDs,
		})
		if r.err != nil {
			return
		}

		// No more user IDs to paginate through.
		if idsSize <= int32(afterID)+r.first {
			r.pageInfo = graphqlutil.HasNextPage(false)
		} else { // Additional user IDs to paginate through.
			endCursor := string(graphqlbackend.MarshalUserID(userIDs[len(userIDs)-1]))
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
