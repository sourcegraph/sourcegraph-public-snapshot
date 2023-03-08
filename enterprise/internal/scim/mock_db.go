package scim

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// getMockDB returns a mock database that contains the given users.
// Note: IDs of users must be ascending.
func getMockDB(users []*types.UserForSCIM) *database.MockDB {
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
	userStore.DeleteFunc.SetDefaultHook(func(ctx context.Context, userID int32) error {
		for i, u := range users {
			if u.ID == userID {
				users = append(users[:i], users[i+1:]...)
				return nil
			}
		}
		return database.NewUserNotFoundErr()
	})

	userExternalAccountsStore := database.NewMockUserExternalAccountsStore()
	userExternalAccountsStore.CreateUserAndSaveFunc.SetDefaultHook(func(ctx context.Context, newUser database.NewUser, spec extsvc.AccountSpec, data extsvc.AccountData) (*types.User, error) {
		nextID := 1
		if len(users) > 0 {
			nextID = int(users[len(users)-1].ID) + 1
		}
		userToCreate := types.UserForSCIM{User: types.User{ID: int32(nextID), Username: newUser.Username, DisplayName: newUser.DisplayName}}
		users = append(users, &userToCreate)
		return &userToCreate.User, nil
	})
	userExternalAccountsStore.UpdateOrInsertSCIMDataFunc.SetDefaultHook(func(ctx context.Context, userID int32, accountID string, data extsvc.AccountData) (err error) {
		for _, user := range users {
			if user.ID == userID {
				var decrypted interface{}
				decrypted, err = data.Data.Decrypt(ctx)
				if err != nil {
					return
				}

				var serialized []byte
				serialized, err = json.Marshal(decrypted)
				if err != nil {
					return
				}
				user.SCIMExternalID = accountID
				user.SCIMAccountData = string(serialized)
				break
			}
		}
		return
	})

	userEmailsStore := database.NewMockUserEmailsStore()
	userEmailsStore.AddFunc.SetDefaultHook(func(ctx context.Context, userID int32, email string, s2 *string) error {
		for _, user := range users {
			if user.ID == userID {
				user.Emails = append(user.Emails, email)
			}
		}
		return nil
	})

	userEmailsStore.SetVerifiedFunc.SetDefaultHook(func(ctx context.Context, i int32, s string, b bool) error {
		return nil
	})

	// Create DB
	db := database.NewMockDB()
	db.WithTransactFunc.SetDefaultHook(func(ctx context.Context, tx func(database.DB) error) error {
		return tx(db)
	})
	db.UsersFunc.SetDefaultReturn(userStore)
	db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccountsStore)
	db.UserEmailsFunc.SetDefaultReturn(userEmailsStore)
	return db
}

// applyLimitOffset returns a slice of users based on the limit and offset
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
