package scim

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Delete removes the resource with corresponding ID.
func (h *UserResourceHandler) Delete(r *http.Request, id string) error {
	// Find user
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.Wrap(err, "parse user ID")
	}
	user, err := findUser(r.Context(), h.db, idInt)
	if err != nil {
		return err
	}

	// If we found no user, we report “all clear” to match the spec
	if user.Username == "" {
		return nil
	}

	// Save username, verified emails, and external accounts to be used for revoking user permissions after deletion
	revokeUserPermissionsArgsList, err := getRevokeUserPermissionArgs(r.Context(), user, h.db)
	if err != nil {
		return err
	}

	// Delete user and revoke user permissions
	err = h.db.WithTransact(r.Context(), func(tx database.DB) error {
		if err := tx.Users().HardDelete(r.Context(), int32(idInt)); err != nil {
			return err
		}

		// NOTE: Practically, we don't reuse the ID for any new users, and the situation of left-over pending permissions
		// is possible but highly unlikely. Therefore, there is no need to roll back user deletion even if this step failed.
		// This call is purely for the purpose of cleanup.
		return tx.Authz().RevokeUserPermissionsList(r.Context(), []*database.RevokeUserPermissionsArgs{revokeUserPermissionsArgsList})
	})
	if err != nil {
		return errors.Wrap(err, "delete user")
	}

	return nil
}

// findUser finds the user with the given ID. If the user does not exist, it returns an empty user.
func findUser(ctx context.Context, db database.DB, id int) (types.UserForSCIM, error) {
	users, err := db.Users().ListForSCIM(ctx, &database.UsersListOptions{
		UserIDs: []int32{int32(id)},
	})
	if err != nil {
		return types.UserForSCIM{}, errors.Wrap(err, "list users by IDs")
	}
	if len(users) == 0 {
		return types.UserForSCIM{}, nil
	}
	if users[0].SCIMAccountData == "" {
		return types.UserForSCIM{}, errors.New("cannot delete user because it doesn't seem to be SCIM-controlled")
	}
	user := *users[0]
	return user, nil
}

// getRevokeUserPermissionArgs returns a list of arguments for revoking user permissions.
func getRevokeUserPermissionArgs(ctx context.Context, user types.UserForSCIM, db database.DB) (*database.RevokeUserPermissionsArgs, error) {
	// Collect external accounts
	var accounts []*extsvc.Accounts
	extAccounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{UserID: user.ID})
	if err != nil {
		return nil, errors.Wrap(err, "list external accounts")
	}
	for _, acct := range extAccounts {
		accounts = append(accounts, &extsvc.Accounts{
			ServiceType: acct.ServiceType,
			ServiceID:   acct.ServiceID,
			AccountIDs:  []string{acct.AccountID},
		})
	}

	// Add Sourcegraph account
	accounts = append(accounts, &extsvc.Accounts{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountIDs:  append(user.Emails, user.Username),
	})

	return &database.RevokeUserPermissionsArgs{
		UserID:   user.ID,
		Accounts: accounts,
	}, nil
}
