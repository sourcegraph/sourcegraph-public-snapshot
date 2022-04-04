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
		var afterID api.RepoID
		afterIDIdx := 0
		skipSearch := false
		if r.after != nil {
			afterID, r.err = graphqlbackend.UnmarshalRepositoryID(graphql.ID(*r.after))
			if r.err != nil {
				return
			}

			// Find the index of afterID in the ids slice, if afterID exists, and we can't find it in the loop, don't bother paginating
			skipSearch = true
			for idx, id := range r.ids {
				if id == int32(afterID) {
					// Only paginate if the index of the afterID isn't the last id of the slice.
					if afterIDIdx < len(r.ids)-1 {
						afterIDIdx = idx + 1 // set the start index to the next element
						skipSearch = false
					}
					break
				}
			}
		}

		userIDs := make([]int32, 0, r.first)
		idsSize := int32(len(r.ids))

		if !skipSearch {
			// Generate a slice of user IDs ranging from index afterIDIdx+1 to: afterIDIdx+first or until the end of the slice, whichever is less.
			count := int32(1)
			for _, id := range r.ids[afterIDIdx:] {
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
		if idsSize <= int32(afterIDIdx)+r.first {
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
