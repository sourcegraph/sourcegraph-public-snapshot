package graphqlbackend

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestExternalAccounts_DeleteExternalAccount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)

	t.Run("has github account", func(t *testing.T) {
		db := database.NewDB(logger, dbtest.NewDB(logger, t))
		act := actor.Actor{UID: 1}
		ctx := actor.WithActor(context.Background(), &act)
		sr := newSchemaResolver(db, gitserver.NewClient(db))

		spec := extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "xd",
		}

		_, err := db.UserExternalAccounts().CreateUserAndSave(ctx, database.NewUser{Username: "u"}, spec, extsvc.AccountData{})
		require.NoError(t, err)

		graphqlArgs := struct {
			ExternalAccount graphql.ID
		}{
			ExternalAccount: graphql.ID(base64.URLEncoding.EncodeToString([]byte("ExternalAccount:1"))),
		}
		_, err = sr.DeleteExternalAccount(ctx, &graphqlArgs)
		require.NoError(t, err)

		accts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{UserID: 1})
		require.NoError(t, err)
		require.Equal(t, 0, len(accts))
	})
}
