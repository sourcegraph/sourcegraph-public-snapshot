package productsubscription

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestProductSubscriptionByAccessToken(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	r := ProductSubscriptionLicensingResolver{Logger: logger, DB: db}

	alice, err := db.Users().Create(ctx, database.NewUser{Username: "alice"})
	require.NoError(t, err)

	t.Run("failure case", func(t *testing.T) {
		_, err := r.ProductSubscriptionByAccessToken(
			actor.WithActor(ctx, &actor.Actor{UID: alice.ID}),
			&graphqlbackend.ProductSubscriptionByAccessTokenArgs{
				AccessToken: "404",
			},
		)
		_, got := err.(ErrProductSubscriptionNotFound)
		assert.True(t, got, "should have error type ErrProductSubscriptionNotFound")
	})
}
