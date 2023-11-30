package scim

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var verifiedDate = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

// getMockDB returns a mock database that contains the given users.
// Note: IDs of users must be ascending.
func getMockDB(users []*types.UserForSCIM, usersEmails map[int32][]*database.UserEmail) *dbmocks.MockDB {
	userStore := dbmocks.NewMockUserStore()
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
	userStore.CountForSCIMFunc.SetDefaultReturn(len(users), nil)
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
	userStore.HardDeleteFunc.SetDefaultHook(func(ctx context.Context, userID int32) error {
		// Delete the user
		for i, u := range users {
			if u.ID == userID {
				// Delete the user
				users = append(users[:i], users[i+1:]...)
				// Delete the user's emails
				delete(usersEmails, userID)
				return nil
			}
		}

		return database.NewUserNotFoundErr()
	})
	userStore.RecoverUsersListFunc.SetDefaultHook(func(ctx context.Context, userIds []int32) ([]int32, error) {
		updated := []int32{}
		for _, id := range userIds {
			for _, user := range users {
				if user.ID == id {
					user.Active = true
					updated = append(updated, id)
				}
			}
		}
		return updated, nil
	})

	userStore.DeleteFunc.SetDefaultHook(func(ctx context.Context, id int32) error {
		for _, user := range users {
			if user.ID == id {
				user.Active = false
				return nil
			}
		}
		return errors.New("user not found")
	})
	userStore.CreateWithExternalAccountFunc.SetDefaultHook(func(ctx context.Context, newUser database.NewUser, acct *extsvc.Account) (*types.User, error) {
		nextID := 1
		if len(users) > 0 {
			nextID = int(users[len(users)-1].ID) + 1
		}
		userToCreate := types.UserForSCIM{User: types.User{ID: int32(nextID), Username: newUser.Username, DisplayName: newUser.DisplayName}}
		users = append(users, &userToCreate)
		return &userToCreate.User, nil
	})

	userExternalAccountsStore := dbmocks.NewMockUserExternalAccountsStore()
	userExternalAccountsStore.UpsertSCIMDataFunc.SetDefaultHook(func(ctx context.Context, userID int32, accountID string, data extsvc.AccountData) (err error) {
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

	userEmailsStore := dbmocks.NewMockUserEmailsStore()
	userEmailsStore.AddFunc.SetDefaultHook(func(ctx context.Context, userID int32, email string, verificationCode *string) error {
		usersEmails[userID] = append(usersEmails[userID], &database.UserEmail{UserID: userID, Email: email, VerificationCode: verificationCode})
		return nil
	})

	userEmailsStore.RemoveFunc.SetDefaultHook(func(ctx context.Context, userID int32, email string) error {
		var err error
		remove := func(currentEmails []*database.UserEmail, toRemove string) ([]*database.UserEmail, error) {
			for i, email := range currentEmails {
				if email.Email == toRemove {
					if email.Primary {
						return currentEmails, errors.New("can't delete primary email")
					}
					return append(currentEmails[:i], currentEmails[i+1:]...), nil
				}
			}
			return currentEmails, err
		}
		usersEmails[userID], err = remove(usersEmails[userID], email)
		return err
	})

	userEmailsStore.SetVerifiedFunc.SetDefaultHook(func(ctx context.Context, userID int32, email string, verified bool) error {
		for _, savedEmail := range usersEmails[userID] {
			if savedEmail.Email == email {
				savedEmail.VerifiedAt = &verifiedDate
			}
		}
		return nil
	})

	userEmailsStore.SetPrimaryEmailFunc.SetDefaultHook(func(ctx context.Context, userID int32, email string) error {
		for _, savedEmail := range usersEmails[userID] {
			savedEmail.Primary = strings.EqualFold(savedEmail.Email, email)
		}
		return nil
	})

	userEmailsStore.ListByUserFunc.SetDefaultHook(func(ctx context.Context, opts database.UserEmailsListOptions) ([]*database.UserEmail, error) {
		toReturn := make([]*database.UserEmail, 0)
		for _, email := range usersEmails[opts.UserID] {
			if !opts.OnlyVerified {
				toReturn = append(toReturn, email)
				continue
			}
			if email.VerifiedAt != nil {
				toReturn = append(toReturn, email)
			}
		}
		return toReturn, nil
	})
	userEmailsStore.GetVerifiedEmailsFunc.SetDefaultHook(func(ctx context.Context, emails ...string) ([]*database.UserEmail, error) {
		toReturn := make([]*database.UserEmail, 0)
		for _, email := range emails {
			for _, userEmails := range usersEmails {
				for _, userEmail := range userEmails {
					if userEmail.Email == email && userEmail.VerifiedAt != nil {
						toReturn = append(toReturn, userEmail)
					}
				}
			}
		}
		return toReturn, nil
	})

	authzStore := dbmocks.NewMockAuthzStore()
	authzStore.RevokeUserPermissionsListFunc.SetDefaultReturn(nil)

	// Create DB
	db := dbmocks.NewMockDB()
	db.WithTransactFunc.SetDefaultHook(func(ctx context.Context, tx func(database.DB) error) error {
		return tx(db)
	})
	db.UsersFunc.SetDefaultReturn(userStore)
	db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccountsStore)
	db.UserEmailsFunc.SetDefaultReturn(userEmailsStore)
	db.AuthzFunc.SetDefaultReturn(authzStore)
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
