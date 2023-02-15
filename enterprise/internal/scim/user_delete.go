package scim

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Delete removes the resource with corresponding ID.
func (h *UserResourceHandler) Delete(r *http.Request, id string) error {
	logger := h.observationCtx.Logger.Scoped("DeleteUsers", "SCIM delete user").With(log.String("user", id))

	// Find user
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.Wrap(err, "parse user ID")
	}
	users, err := h.db.Users().ListForSCIM(r.Context(), &database.UsersListOptions{
		UserIDs: []int32{int32(idInt)},
	})
	if err != nil {
		return errors.Wrap(err, "list users by IDs")
	}
	if len(users) == 0 {
		logger.Info("requested users to delete do not exist")
		return nil
	} else if len(users) > 1 {
		logger.Error("requested users to delete have duplicate IDs—that should not happen")
		return errors.New("requested users to delete have duplicate IDs")
	} else {
		logger.Debug("attempting to delete requested users")
	}
	if users[0].SCIMExternalID != "" {
		return errors.New("cannot delete user because it has no SCIM external ID set")
	}

	// Collect username, verified email addresses, and external accounts to be used for revoking user permissions.
	revokeUserPermissionsArgsList, err := getRevokeUserPermissionArgs(r.Context(), users, h.db)
	if err != nil {
		return err
	}

	// Delete user
	if err := h.db.Users().HardDelete(r.Context(), int32(idInt)); err != nil {
		return err
	}

	// NOTE: Practically, we don't reuse the ID for any new users, and the situation of left-over pending permissions
	// is possible but highly unlikely. Therefore, there is no need to roll back user deletion even if this step failed.
	// This call is purely for the purpose of cleanup.
	if err := h.db.Authz().RevokeUserPermissionsList(r.Context(), revokeUserPermissionsArgsList); err != nil {
		return err
	}

	return nil
}

// getRevokeUserPermissionArgs returns a list of arguments for revoking user permissions.
func getRevokeUserPermissionArgs(ctx context.Context, users []*types.UserForSCIM, db database.DB) ([]*database.RevokeUserPermissionsArgs, error) {
	accountsList := make([][]*extsvc.Accounts, len(users))
	var revokeUserPermissionsArgsList []*database.RevokeUserPermissionsArgs
	for index, user := range users {
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

		verifiedEmails, err := db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		})
		if err != nil {
			return nil, errors.Wrap(err, "list verified emails")
		}
		emailStrings := make([]string, len(verifiedEmails))
		for i := range verifiedEmails {
			emailStrings[i] = verifiedEmails[i].Email
		}
		accounts = append(accounts, &extsvc.Accounts{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			AccountIDs:  append(emailStrings, user.Username),
		})

		accountsList[index] = accounts

		revokeUserPermissionsArgsList = append(revokeUserPermissionsArgsList, &database.RevokeUserPermissionsArgs{
			UserID:   user.ID,
			Accounts: accounts,
		})
	}
	return revokeUserPermissionsArgsList, nil
}
