package graphqlbackend

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalAccounts_DeleteExternalAccount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)

	t.Run("has github account", func(t *testing.T) {
		db := database.NewDB(logger, dbtest.NewDB(t))
		act := actor.Actor{UID: 1}
		ctx := actor.WithActor(context.Background(), &act)
		sr := newSchemaResolver(db, gitserver.NewTestClient(t))

		spec := extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "xd",
		}

		_, err := db.Users().CreateWithExternalAccount(ctx, database.NewUser{Username: "u"}, &extsvc.Account{AccountSpec: spec})
		require.NoError(t, err)

		graphqlArgs := struct {
			ExternalAccount graphql.ID
		}{
			ExternalAccount: graphql.ID(base64.URLEncoding.EncodeToString([]byte("ExternalAccount:1"))),
		}
		permssync.MockSchedulePermsSync = func(_ context.Context, _ log.Logger, _ database.DB, req permssync.ScheduleSyncOpts) {
			if req.Reason != database.ReasonExternalAccountDeleted {
				t.Errorf("got reason %s, want %s", req.Reason, database.ReasonExternalAccountDeleted)
			}
		}
		_, err = sr.DeleteExternalAccount(ctx, &graphqlArgs)
		require.NoError(t, err)

		accts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{UserID: 1})
		require.NoError(t, err)
		require.Equal(t, 0, len(accts))
	})
}

func TestExternalAccounts_AddExternalAccount(t *testing.T) {
	db := dbmocks.NewMockDB()

	users := dbmocks.NewMockUserStore()
	db.UsersFunc.SetDefaultReturn(users)
	extservices := dbmocks.NewMockExternalServiceStore()
	db.ExternalServicesFunc.SetDefaultReturn(extservices)
	userextaccts := dbmocks.NewMockUserExternalAccountsStore()
	db.UserExternalAccountsFunc.SetDefaultReturn(userextaccts)

	gerritURL := "https://gerrit.mycorp.com/"
	testCases := map[string]struct {
		user            *types.User
		serviceType     string
		serviceID       string
		accountDetails  string
		wantErr         bool
		wantErrContains string
	}{
		"unauthed returns err": {
			user:    nil,
			wantErr: true,
		},
		"non-gerrit returns err": {
			user:        &types.User{ID: 1},
			serviceType: extsvc.TypeGitHub,
			wantErr:     true,
		},
		"no gerrit connection for serviceID returns err": {
			user:        &types.User{ID: 1},
			serviceType: extsvc.TypeGerrit,
			serviceID:   "https://wrong.id.com",
			wantErr:     true,
		},
		"correct gerrit connection for serviceID returns no err": {
			user:           &types.User{ID: 1},
			serviceType:    extsvc.TypeGerrit,
			serviceID:      gerritURL,
			wantErr:        false,
			accountDetails: `{"username": "alice", "password": "test"}`,
		},
		// OSS packages cannot import enterprise packages, but when we build the entire
		// application this will be implemented.
		//
		// See cmd/frontend/internal/auth/sourcegraphoperator for more details
		// and additional test coverage on the functionality.
		"Sourcegraph operator unimplemented in OSS": {
			user:            &types.User{ID: 1, SiteAdmin: true},
			serviceType:     auth.SourcegraphOperatorProviderType,
			wantErr:         true,
			wantErrContains: "unimplemented in Sourcegraph OSS",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(tc.user, nil)

			gerritConfig := &schema.GerritConnection{
				Url: gerritURL,
			}
			gerritConf, err := json.Marshal(gerritConfig)
			require.NoError(t, err)
			extservices.ListFunc.SetDefaultReturn([]*types.ExternalService{
				{
					Kind:   extsvc.KindGerrit,
					Config: extsvc.NewUnencryptedConfig(string(gerritConf)),
				},
			}, nil)

			userextaccts.InsertFunc.SetDefaultHook(func(ctx context.Context, acct *extsvc.Account) (*extsvc.Account, error) {
				if acct.UserID != tc.user.ID {
					t.Errorf("got userID %d, want %d", acct.UserID, tc.user.ID)
				}
				if acct.AccountSpec.ServiceType != extsvc.TypeGerrit {
					t.Errorf("got service type %q, want %q", acct.AccountSpec.ServiceType, extsvc.TypeGerrit)
				}
				if acct.AccountSpec.ServiceID != gerritURL {
					t.Errorf("got service ID %q, want %q", acct.AccountSpec.ServiceID, "https://gerrit.example.com/")
				}
				if acct.AccountSpec.AccountID != "1234" {
					t.Errorf("got account ID %q, want %q", acct.AccountSpec.AccountID, "alice")
				}
				return nil, nil
			})
			confGet := func() *conf.Unified {
				return &conf.Unified{}
			}
			err = db.ExternalServices().Create(context.Background(), confGet, &types.ExternalService{
				Kind:   extsvc.KindGerrit,
				Config: extsvc.NewUnencryptedConfig(string(gerritConf)),
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tc.user != nil {
				act := actor.Actor{UID: tc.user.ID}
				ctx = actor.WithActor(ctx, &act)
			}

			sr := newSchemaResolver(db, gitserver.NewTestClient(t))

			args := struct {
				ServiceType    string
				ServiceID      string
				AccountDetails string
			}{
				ServiceType:    tc.serviceType,
				ServiceID:      tc.serviceID,
				AccountDetails: tc.accountDetails,
			}

			permssync.MockSchedulePermsSync = func(_ context.Context, _ log.Logger, _ database.DB, req permssync.ScheduleSyncOpts) {
				if req.UserIDs[0] != tc.user.ID {
					t.Errorf("got userID %d, want %d", req.UserIDs[0], tc.user.ID)
				}
				if req.Reason != database.ReasonExternalAccountAdded {
					t.Errorf("got reason %s, want %s", req.Reason, database.ReasonExternalAccountAdded)
				}
			}

			gerrit.MockVerifyAccount = func(_ context.Context, _ *url.URL, _ *gerrit.AccountCredentials) (*gerrit.Account, error) {
				return &gerrit.Account{
					ID:       1234,
					Username: "alice",
				}, nil
			}
			_, err = sr.AddExternalAccount(ctx, &args)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrContains != "" {
					assert.Contains(t, err.Error(), tc.wantErrContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
