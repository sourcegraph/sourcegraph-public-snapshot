package auth

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cloud"
	"github.com/sourcegraph/sourcegraph/internal/auth"
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
	//   1. logan, who has no external accounts
	//   2. morgan, who is an expired SOAP user but has more external accounts
	//   3. jordan, who is a SOAP user that has not expired
	//   4. riley, who is an expired SOAP user (will be cleaned up)
	//   5. cris, who has a non-SOAP external account
	//   6. cami, who is an expired SOAP user on the permanent accounts list
	// All the above except riley will be deleted.
	wantNotDeleted := []string{"logan", "morgan", "jordan", "cris", "cami"}

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
	_, err = db.Handle().ExecContext(ctx, `UPDATE users SET created_at = $1 WHERE id = $2`, time.Now().Add(-61*time.Minute), morgan.ID)
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

	_, err = db.UserExternalAccounts().CreateUserAndSave(
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
	_, err = db.Handle().ExecContext(ctx, `UPDATE users SET created_at = $1 WHERE id = $2`, time.Now().Add(-61*time.Minute), riley.ID)
	require.NoError(t, err)

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
	_, err = db.Handle().ExecContext(ctx, `UPDATE users SET created_at = $1 WHERE id = $2`, time.Now().Add(-61*time.Minute), cami.ID)
	require.NoError(t, err)

	t.Run("handle with cleanup", func(t *testing.T) {
		err = handler.Handle(ctx)
		require.NoError(t, err)

		users, err := db.Users().List(ctx, nil)
		require.NoError(t, err)

		got := make([]string, 0, len(users))
		for _, u := range users {
			got = append(got, u.Username)
		}
		assert.Equal(t, wantNotDeleted, got)
	})
}
