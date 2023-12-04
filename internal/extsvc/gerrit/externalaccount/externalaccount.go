package externalaccount

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
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

	if _, err = db.UserExternalAccounts().Upsert(ctx,
		&extsvc.Account{
			UserID:      userID,
			AccountSpec: accountSpec,
			AccountData: accountData,
		}); err != nil {
		return err
	}

	return nil
}
