package scim

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func getMockDB() *database.MockDB {
	users := []*types.UserForSCIM{
		{User: types.User{ID: 1, Username: "user1", DisplayName: "First Last"}, Emails: []string{"a@example.com"}, SCIMExternalID: "id1"},
		{User: types.User{ID: 2, Username: "user2", DisplayName: "First Middle Last"}, Emails: []string{"b@example.com"}, SCIMExternalID: "id2"},
		{User: types.User{ID: 3, Username: "user3", DisplayName: "First Last"}, SCIMExternalID: "id3"},
		{User: types.User{ID: 4, Username: "user4"}, SCIMAccountData: "{\"externalUsername\":\"user4@company.com\"}", SCIMExternalID: "id4"},
	}

	userStore := database.NewMockUserStore()
	userStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		for _, user := range users {
			if user.ID == id {
				return &user.User, nil
			}
		}
		return nil, database.NewUserNotFoundErr()
	})
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
	userStore.GetByUsernameFunc.SetDefaultHook(func(ctx context.Context, username string) (*types.User, error) {
		for _, user := range users {
			if user.Username == username {
				return &user.User, nil
			}
		}
		return nil, database.NewUserNotFoundErr()
	})
	userStore.UpdateFunc.SetDefaultHook(func(ctx context.Context, userID int32, update database.UserUpdate) error {
		for _, u := range users {
			if u.ID == userID {
				if update.Username != "" {
					u.Username = update.Username
				}
				if update.DisplayName != nil {
					u.DisplayName = *update.DisplayName
				}

				return nil
			}
		}
		return database.NewUserNotFoundErr()
	})
	userStore.TransactFunc.SetDefaultHook(func(ctx context.Context) (database.UserStore, error) {
		return userStore, nil
	})

	userExternalAccountsStore := database.NewMockUserExternalAccountsStore()
	userExternalAccountsStore.LookupUserAndSaveFunc.SetDefaultHook(func(ctx context.Context, spec extsvc.AccountSpec, data extsvc.AccountData) (int32, error) {
		for _, user := range users {
			if user.SCIMExternalID == spec.AccountID {
				decrypted, err := data.Data.Decrypt(ctx)
				if err != nil {
					return 0, err
				}
				user.SCIMExternalID = decrypted.(map[string]interface{})["externalUsername"].(string)
				return user.ID, nil
			}
		}
		return 0, nil
	})

	// Create DB
	db := database.NewMockDB()
	db.WithTransactFunc.SetDefaultHook(func(ctx context.Context, tx func(database.DB) error) error {
		return tx(db)
	})
	db.UsersFunc.SetDefaultReturn(userStore)
	db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccountsStore)
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
