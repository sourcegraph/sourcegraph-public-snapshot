package sourcegraphoperator

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type accountDetailsBody struct {
	ClientID  string `json:"clientID"`
	AccountID string `json:"accountID"`

	ExternalAccountData
}

// AddSourcegraphOperatorExternalAccount links the given user with a Sourcegraph Operator
// provider, if and only if it already exists. The provider can only be added through
// Enterprise Sourcegraph Cloud config, so this essentially no-ops outside of Cloud.
//
// 🚨 SECURITY: Some important things to note:
//   - The caller must check that the user is a site administrator.
//   - Being a SOAP user does not grant any extra privilege over being a site admin.
//   - The operation will fail if the user is already a SOAP user, which prevents escalating
//     time-bound accounts to permanent service accounts.
//   - Both the client ID and the service ID must match the SOAP configuration exactly.
func AddSourcegraphOperatorExternalAccount(ctx context.Context, db database.DB, userID int32, serviceID string, accountDetails string) error {
	p := providers.GetProviderByConfigID(providers.ConfigID{
		Type: auth.SourcegraphOperatorProviderType,
		ID:   serviceID,
	})
	if p == nil {
		return errors.New("provider does not exist")
	}

	if accountDetails == "" {
		return errors.New("invalid account details")
	}
	var details accountDetailsBody
	if err := json.Unmarshal([]byte(accountDetails), &details); err != nil {
		return errors.Wrap(err, "invalid account details")
	}

	// Simple validation
	if details.ClientID != p.CachedInfo().ClientID {
		return errors.Newf("unknown client ID %q", details.ClientID)
	}
	soapAccounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		UserID: userID,
		// Provider
		ServiceType: auth.SourcegraphOperatorProviderType,
		ServiceID:   serviceID,
	})
	if err != nil {
		return err
	}
	if len(soapAccounts) > 0 {
		return errors.New("user already has an associated SOAP account")
	}

	// Create an association
	accountData, err := MarshalAccountData(details.ExternalAccountData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal account data")
	}
	if err := db.UserExternalAccounts().AssociateUserAndSave(ctx, userID, extsvc.AccountSpec{
		ServiceType: auth.SourcegraphOperatorProviderType,
		ServiceID:   serviceID,
		ClientID:    details.ClientID,

		AccountID: details.AccountID,
	}, accountData); err != nil {
		return errors.Wrap(err, "failed to associate user with SOAP provider")
	}

	return nil
}
