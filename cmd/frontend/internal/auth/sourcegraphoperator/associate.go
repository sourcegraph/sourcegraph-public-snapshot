package sourcegraphoperator

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type accountDetailsBody struct {
	ClientID  string `json:"clientID"`
	AccountID string `json:"accountID"`

	sourcegraphoperator.ExternalAccountData
}

// AddSourcegraphOperatorExternalAccount links the given user with a Sourcegraph Operator
// provider, if and only if it already exists. The provider can only be added through
// Enterprise Sourcegraph Cloud config, so this essentially no-ops outside of Cloud.
//
// It implements internal/auth/sourcegraphoperator.AddSourcegraphOperatorExternalAccount
//
// 🚨 SECURITY: Some important things to note:
//   - Being a SOAP user does not grant any extra privilege over being a site admin.
//   - The operation will fail if the user is already a SOAP user, which prevents escalating
//     time-bound accounts to permanent service accounts.
//   - Both the client ID and the service ID must match the SOAP configuration exactly.
func AddSourcegraphOperatorExternalAccount(ctx context.Context, db database.DB, userID int32, serviceID string, accountDetails string) error {
	// 🚨 SECURITY: Caller must be a site admin.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return err
	}

	p := providers.GetProviderByConfigID(providers.ConfigID{
		Type: auth.SourcegraphOperatorProviderType,
		ID:   serviceID,
	})
	if p == nil {
		return errors.New("provider does not exist")
	}

	if accountDetails == "" {
		return errors.New("account details are required")
	}
	var details accountDetailsBody
	if err := json.Unmarshal([]byte(accountDetails), &details); err != nil {
		return errors.Wrap(err, "invalid account details")
	}

	// Additionally check client ID matches - service ID was already checked in the
	// initial GetProviderByConfigID call
	if details.ClientID != p.CachedInfo().ClientID {
		return errors.Newf("unknown client ID %q", details.ClientID)
	}

	// Run account count verification and association in a single transaction, to ensure
	// we have no funny business with accounts being created in the time between the two.
	return db.WithTransact(ctx, func(db database.DB) error {
		// Make sure this user has no other SOAP accounts.
		numSOAPAccounts, err := db.UserExternalAccounts().Count(ctx, database.ExternalAccountsListOptions{
			UserID: userID,
			// For provider matching, we explicitly do not provider the service ID - there
			// should only be one SOAP registered.
			ServiceType: auth.SourcegraphOperatorProviderType,
		})
		if err != nil {
			return errors.Wrap(err, "failed to check for an existing Sourcegraph Operator accounts")
		}
		if numSOAPAccounts > 0 {
			return errors.New("user already has an associated Sourcegraph Operator account")
		}

		// Create an association
		accountData, err := sourcegraphoperator.MarshalAccountData(details.ExternalAccountData)
		if err != nil {
			return errors.Wrap(err, "failed to marshal account data")
		}
		if _, err := db.UserExternalAccounts().Upsert(ctx,
			&extsvc.Account{
				UserID: userID,
				AccountSpec: extsvc.AccountSpec{
					ServiceType: auth.SourcegraphOperatorProviderType,
					ServiceID:   serviceID,
					ClientID:    details.ClientID,

					AccountID: details.AccountID,
				},
				AccountData: accountData,
			}); err != nil {
			return errors.Wrap(err, "failed to associate user with Sourcegraph Operator provider")
		}
		return nil
	})
}
