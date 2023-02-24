package scim

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func getMockDB() *database.MockDB {
	users := []*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "external1"},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "First Middle Last"}, Emails: []string{"b@example.com"}, SCIMExternalID: ""},
		{User: types.User{ID: 3, Username: "user3", DisplayName: "First Last"}},
		{User: types.User{ID: 4, Username: "user4"}},
	}

	userStore := database.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	userStore.ListForSCIMFunc.SetDefaultHook(func(ctx context.Context, opt *database.UsersListOptions) ([]*types.UserForSCIM, error) {
		// Return the users with the given IDs
		if opt.UserIDs != nil {
			var filteredUsers []*types.UserForSCIM
			for _, id := range opt.UserIDs {
				for _, user := range users {
					if user.ID == id {
						filteredUsers = append(filteredUsers, user)
					}
				}
			}
			return applyLimitOffset(filteredUsers, opt.LimitOffset)
		}

		return applyLimitOffset(users, opt.LimitOffset)
	})
	userStore.CountFunc.SetDefaultReturn(4, nil)
	userStore.CreateFunc.SetDefaultHook(func(ctx context.Context, user database.NewUser) (*types.User, error) {
		return &types.User{ID: 5, Username: user.Username, DisplayName: user.DisplayName}, nil
	})

	// Create DB
	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(userStore)
	return db
}

func applyLimitOffset(users []*types.UserForSCIM, limitOffset *database.LimitOffset) ([]*types.UserForSCIM, error) {
	// Return all users
	if limitOffset == nil {
		return users, nil
	}

	// Return a slice of users based on the limit and offset
	start := limitOffset.Offset
	end := start + limitOffset.Limit
	if end > len(users) {
		end = len(users)
	}
	return users[start:end], nil
}
