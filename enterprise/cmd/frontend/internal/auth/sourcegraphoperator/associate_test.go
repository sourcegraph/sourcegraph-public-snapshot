package sourcegraphoperator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	osssourcegraphoperator "github.com/sourcegraph/sourcegraph/internal/auth/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAddSourcegraphOperatorExternalAccountBinding(t *testing.T) {
	assert.NotNil(t, osssourcegraphoperator.AddSourcegraphOperatorExternalAccount)
}

func TestAddSourcegraphOperatorExternalAccount(t *testing.T) {
	ctx := context.Background()
	serviceID := (fakeSoapProvider{}).ConfigID().ID

	for _, tc := range []struct {
		name string
		// db, user, and other setup
		setup func(t *testing.T) (userID int32, db database.DB)
		// accountDetails parameter
		accountDetails *accountDetailsBody
		// validate result of AddSourcegraphOperatorExternalAccount
		expectErr autogold.Value
		// assert state of the DB (optional)
		assert func(t *testing.T, uid int32, db database.DB)
	}{
		{
			name: "provider does not exist",
			setup: func(t *testing.T) (int32, database.DB) {
				return 42, database.NewMockDB()
			},
			expectErr: autogold.Expect("provider does not exist"),
		},
		{
			name: "incorrect details for SOAP provider",
			setup: func(t *testing.T) (int32, database.DB) {
				providers.MockProviders = []providers.Provider{fakeSoapProvider{}}
				t.Cleanup(func() { providers.MockProviders = nil })

				return 42, database.NewMockDB()
			},
			accountDetails: &accountDetailsBody{
				ClientID:  "foobar",
				AccountID: "bob",
				ExternalAccountData: ExternalAccountData{
					ServiceAccount: true,
				},
			},
			expectErr: autogold.Expect(`unknown client ID "foobar"`),
		},
		{
			name: "new user associate",
			setup: func(t *testing.T) (int32, database.DB) {
				if testing.Short() {
					t.Skip()
				}

				providers.MockProviders = []providers.Provider{fakeSoapProvider{}}
				t.Cleanup(func() { providers.MockProviders = nil })

				logger := logtest.NoOp(t)
				db := database.NewDB(logger, dbtest.NewDB(logger, t))
				u, err := db.Users().Create(
					ctx,
					database.NewUser{
						Username: "logan",
					},
				)
				require.NoError(t, err)
				return u.ID, db
			},
			accountDetails: &accountDetailsBody{
				ClientID:  "soap_client",
				AccountID: "bob",
				ExternalAccountData: ExternalAccountData{
					ServiceAccount: true,
				},
			},
			expectErr: autogold.Expect(nil),
			assert: func(t *testing.T, uid int32, db database.DB) {
				accts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
					UserID: uid,
				})
				require.NoError(t, err)
				require.Len(t, accts, 1)
				assert.Equal(t, auth.SourcegraphOperatorProviderType, accts[0].ServiceType)
				assert.Equal(t, "bob", accts[0].AccountID)
				assert.Equal(t, "soap_client", accts[0].ClientID)
				assert.Equal(t, serviceID, accts[0].ServiceID)

				data, err := GetAccountData(ctx, accts[0].AccountData)
				require.NoError(t, err)
				assert.True(t, data.ServiceAccount)
			},
		},
		{
			name: "double associate is not allowed (prevents escalation)",
			setup: func(t *testing.T) (int32, database.DB) {
				if testing.Short() {
					t.Skip()
				}

				providers.MockProviders = []providers.Provider{fakeSoapProvider{}}
				t.Cleanup(func() { providers.MockProviders = nil })

				logger := logtest.NoOp(t)
				db := database.NewDB(logger, dbtest.NewDB(logger, t))
				u, err := db.Users().Create(
					ctx,
					database.NewUser{
						Username: "bib",
					},
				)
				require.NoError(t, err)
				err = db.UserExternalAccounts().AssociateUserAndSave(ctx, u.ID, extsvc.AccountSpec{
					ServiceType: auth.SourcegraphOperatorProviderType,
					ServiceID:   serviceID,
					ClientID:    "soap_client",
					AccountID:   "bib",
				}, extsvc.AccountData{}) // not a service account initially
				require.NoError(t, err)
				return u.ID, db
			},
			accountDetails: &accountDetailsBody{
				ClientID:  "soap_client",
				AccountID: "bob", // trying to change account ID
				ExternalAccountData: ExternalAccountData{
					ServiceAccount: true, // trying to promote themselves to service account
				},
			},
			expectErr: autogold.Expect("user already has an associated SOAP account"),
			assert: func(t *testing.T, uid int32, db database.DB) {
				accts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
					UserID: uid,
				})
				require.NoError(t, err)
				require.Len(t, accts, 1)
				assert.Equal(t, auth.SourcegraphOperatorProviderType, accts[0].ServiceType)
				assert.Equal(t, "bib", accts[0].AccountID) // the original account
				assert.Equal(t, "soap_client", accts[0].ClientID)
				assert.Equal(t, serviceID, accts[0].ServiceID)

				data, err := GetAccountData(ctx, accts[0].AccountData)
				require.NoError(t, err)
				assert.False(t, data.ServiceAccount) // still not a service account
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uid, db := tc.setup(t)
			details, err := json.Marshal(tc.accountDetails)
			require.NoError(t, err)
			err = addSourcegraphOperatorExternalAccount(ctx, db, uid, serviceID, string(details))
			if err != nil {
				tc.expectErr.Equal(t, err.Error())
			} else {
				tc.expectErr.Equal(t, nil)
			}
			if tc.assert != nil {
				tc.assert(t, uid, db)
			}
		})
	}
}

type fakeSoapProvider struct{}

// ConfigID returns the identifier for this provider's config in the auth.providers site
// configuration array.
//
// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
// and anonymous clients.
func (p fakeSoapProvider) ConfigID() providers.ConfigID {
	return providers.ConfigID{
		Type: auth.SourcegraphOperatorProviderType,
		ID:   "soap",
	}
}

// Config is the entry in the site configuration "auth.providers" array that this provider
// represents.
//
// 🚨 SECURITY: This value contains secret information that must not be shown to
// non-site-admins.
func (p fakeSoapProvider) Config() schema.AuthProviders { return schema.AuthProviders{} }

// CachedInfo returns cached information about the provider.
func (p fakeSoapProvider) CachedInfo() *providers.Info {
	return &providers.Info{
		ClientID: "soap_client",
	}
}

// Refresh refreshes the provider's information with an external service, if any.
func (p fakeSoapProvider) Refresh(ctx context.Context) error { return nil }

// Provides basic external account from this auth provider
func (p fakeSoapProvider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	return nil, nil
}
