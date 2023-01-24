package externalaccount

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func AddGerritExternalAccount(ctx context.Context, db database.DB, userID int32, serviceID string, accountDetails string) error {
	var accountCredentials gerrit.AccountCredentials
	err := json.Unmarshal([]byte(accountDetails), &accountCredentials)
	if err != nil {
		return err
	}

	// Fetch external service matching ServiceID
	svcs, err := db.ExternalServices().List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{extsvc.KindGerrit},
	})
	if err != nil {
		return err
	}

	serviceURL, err := url.Parse(serviceID)
	if err != nil {
		return err
	}
	serviceURL = extsvc.NormalizeBaseURL(serviceURL)

	var gerritConn *types.GerritConnection
	for _, svc := range svcs {
		cfg, err := extsvc.ParseEncryptableConfig(ctx, svc.Kind, svc.Config)
		if err != nil {
			continue
		}
		if c, ok := cfg.(*schema.GerritConnection); ok {
			connURL, err := url.Parse(c.Url)
			if err != nil {
				continue
			}
			connURL = extsvc.NormalizeBaseURL(connURL)

			if connURL.String() != serviceURL.String() {
				continue
			}
			gerritConn = &types.GerritConnection{
				URN:              svc.URN(),
				GerritConnection: c,
			}
			break
		}
	}
	if gerritConn == nil {
		return errors.New("no gerrit connection found")
	}

	gerritAccount, err := gerrit.VerifyAccount(ctx, gerritConn, &accountCredentials)
	if err != nil {
		return err
	}

	accountSpec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGerrit,
		ServiceID:   serviceID,
		ClientID:    "",
		AccountID:   strconv.Itoa(int(gerritAccount.ID)),
	}

	accountData := extsvc.AccountData{}
	if err = gerrit.SetExternalAccountData(&accountData, gerritAccount, &accountCredentials); err != nil {
		return err
	}

	if err = db.UserExternalAccounts().AssociateUserAndSave(ctx, userID, accountSpec, accountData); err != nil {
		return err
	}
	return nil
}
