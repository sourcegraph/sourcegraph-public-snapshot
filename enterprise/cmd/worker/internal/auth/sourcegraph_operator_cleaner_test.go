package auth

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/cloud"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func TestSourcegraphOperatorCleanHandler(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// NOTE: We cannot run this test with t.Parallel() because this mock mutates a
	// shared global state.
	cloud.MockSiteConfig(
		t,
		&cloud.SchemaSiteConfig{
			AuthProviders: &cloud.SchemaAuthProviders{
				SourcegraphOperator: &cloud.SchemaAuthProviderSourcegraphOperator{},
			},
		},
	)

	ctx := context.Background()
	logger := logtest.NoOp(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	handler := sourcegraphOperatorCleanHandler{
		db:                db,
		lifecycleDuration: 60 * time.Minute,
	}

	t.Run("handle with nothing to clean up", func(t *testing.T) {
		// Make sure it doesn't blow up if there is nothing to clean up
		err := handler.Handle(ctx)
		require.NoError(t, err)
	})

	// Create test users:
	//   1. logan, who has no external accounts and is a site admin (will not be changed)
	//      (like a customer site admin)
	//   2. morgan, who is an expired SOAP user but has more external accounts (will be demoted)
	//      (like a Sourcegraph teammate who used SOAP via Entitle, and has an external account)
	//   3. jordan, who is a SOAP user that has not expired (will not be changed)
	//   4. riley, who is an expired SOAP user with no external accounts (will be deleted)
	//   5. cris, who has a non-SOAP external account and is not a site admin (will not be changed)
	//   6. cami, who is an expired SOAP user and is a service account (will not be changed)
	//   7. dani, who has no external accounts and is not a site admin (will not be changed)
	// In all of the above, SOAP users are also made site admins.
	// The lists below indicate who will and will not be deleted or otherwise
	// modified.
	wantNotDeleted := []string{"logan", "morgan", "jordan", "cris", "cami", "dani"}
	wantAdmins := []string{"logan", "jordan", "cami"}
	wantNonSOAPUsers := []string{"logan", "morgan", "cris", "dani"}

	_, err := db.Users().Create(
		ctx,
		database.NewUser{
			Username: "logan",
		},
	)
	require.NoError(t, err)

	morgan, err := db.UserExternalAccounts().CreateUserAndSave(
		ctx,
		database.NewUser{
			Username: "morgan",
		},
		extsvc.AccountSpec{
			ServiceType: auth.SourcegraphOperatorProviderType,
			ServiceID:   "https://sourcegraph.com",
			ClientID:    "soap",
			AccountID:   "morgan",
		},
		extsvc.AccountData{},
	)
	require.NoError(t, err)
	_, err = db.Handle().ExecContext(ctx, `UPDATE user_external_accounts SET created_at = $1 WHERE user_id = $2`,
		time.Now().Add(-61*time.Minute), morgan.ID)
	require.NoError(t, err)
	err = db.UserExternalAccounts().AssociateUserAndSave(
		ctx,
		morgan.ID,
		extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
			ClientID:    "github",
			AccountID:   "morgan",
		},
		extsvc.AccountData{},
	)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, morgan.ID, true))

	jordan, err := db.UserExternalAccounts().CreateUserAndSave(
		ctx,
		database.NewUser{
			Username: "jordan",
		},
		extsvc.AccountSpec{
			ServiceType: auth.SourcegraphOperatorProviderType,
			ServiceID:   "https://sourcegraph.com",
			ClientID:    "soap",
			AccountID:   "jordan",
		},
		extsvc.AccountData{},
	)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, jordan.ID, true))

	riley, err := db.UserExternalAccounts().CreateUserAndSave(
		ctx,
		database.NewUser{
			Username: "riley",
		},
		extsvc.AccountSpec{
			ServiceType: auth.SourcegraphOperatorProviderType,
			ServiceID:   "https://sourcegraph.com",
			ClientID:    "soap",
			AccountID:   "riley",
		},
		extsvc.AccountData{},
	)
	require.NoError(t, err)
	_, err = db.Handle().ExecContext(ctx, `UPDATE user_external_accounts SET created_at = $1 WHERE user_id = $2`,
		time.Now().Add(-61*time.Minute), riley.ID)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, riley.ID, true))

	_, err = db.UserExternalAccounts().CreateUserAndSave(
		ctx,
		database.NewUser{
			Username: "cris",
		},
		extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
			ClientID:    "github",
			AccountID:   "cris",
		},
		extsvc.AccountData{},
	)
	require.NoError(t, err)

	accountData, err := sourcegraphoperator.MarshalAccountData(sourcegraphoperator.ExternalAccountData{
		ServiceAccount: true,
	})
	require.NoError(t, err)
	cami, err := db.UserExternalAccounts().CreateUserAndSave(
		ctx,
		database.NewUser{
			Username: "cami",
		},
		extsvc.AccountSpec{
			ServiceType: auth.SourcegraphOperatorProviderType,
			ServiceID:   "https://sourcegraph.com",
			ClientID:    "soap",
			AccountID:   "cami",
		},
		accountData,
	)
	require.NoError(t, err)
	_, err = db.Handle().ExecContext(ctx, `UPDATE user_external_accounts SET created_at = $1 WHERE user_id = $2`,
		time.Now().Add(-61*time.Minute), cami.ID)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, cami.ID, true))

	_, err = db.Users().Create(ctx, database.NewUser{
		Username:        "dani",
		Email:           "dani@example.com",
		EmailIsVerified: true,
	})
	require.NoError(t, err)

	t.Run("handle with cleanup", func(t *testing.T) {
		err = handler.Handle(ctx)
		require.NoError(t, err)

		users, err := db.Users().List(ctx, nil)
		require.NoError(t, err)

		got := make([]string, 0, len(users))
		gotAdmins := make([]string, 0, len(users))
		gotNonSOAPUsers := make([]string, 0, len(users))
		for _, u := range users {
			got = append(got, u.Username)
			if u.SiteAdmin {
				gotAdmins = append(gotAdmins, u.Username)
			}
			ext, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
				UserID:      u.ID,
				ServiceType: auth.SourcegraphOperatorProviderType,
			})
			require.NoError(t, err)
			if len(ext) == 0 {
				gotNonSOAPUsers = append(gotNonSOAPUsers, u.Username)
			}
		}

		slices.Sort(wantNotDeleted)
		slices.Sort(got)
		slices.Sort(wantAdmins)
		slices.Sort(gotAdmins)
		slices.Sort(wantNonSOAPUsers)
		slices.Sort(gotNonSOAPUsers)

		assert.Equal(t, wantNotDeleted, got, "want not deleted")
		assert.Equal(t, wantAdmins, gotAdmins, "want admins")
		assert.Equal(t, wantNonSOAPUsers, gotNonSOAPUsers, "want SOAP")
	})
}
