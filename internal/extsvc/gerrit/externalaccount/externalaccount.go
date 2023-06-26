package externalaccount

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func AddGerritExternalAccount(ctx context.Context, db database.DB, userID int32, serviceID string, accountDetails string) (err error) {
	var accountCredentials gerrit.AccountCredentials
	err = json.Unmarshal([]byte(accountDetails), &accountCredentials)
	if err != nil {
		return err
	}

	serviceURL, err := url.Parse(serviceID)
	if err != nil {
		return err
	}
	serviceURL = extsvc.NormalizeBaseURL(serviceURL)

	gerritAccount, err := gerrit.VerifyAccount(ctx, serviceURL, &accountCredentials)
	if err != nil {
		return err
	}

	accountSpec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGerrit,
		ServiceID:   serviceID,
		AccountID:   strconv.Itoa(int(gerritAccount.ID)),
	}

	accountData := extsvc.AccountData{}
	if err = gerrit.SetExternalAccountData(&accountData, gerritAccount, &accountCredentials); err != nil {
		return err
	}

	if err = db.UserExternalAccounts().AssociateUserAndSave(ctx, userID, accountSpec, accountData); err != nil {
		return err
	}

	logger := log.Scoped("AddGerritExternalAccount", "Add Gerrit External Account to existing user")
	// Schedule a permission sync, since this is a new account for the user
	permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
		UserIDs:           []int32{userID},
		Reason:            database.ReasonExternalAccountAdded,
		TriggeredByUserID: userID,
	})
	return nil
}
