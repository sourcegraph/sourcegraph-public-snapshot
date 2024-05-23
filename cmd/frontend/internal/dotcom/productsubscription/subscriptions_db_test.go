package productsubscription

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestProductSubscriptions_Create(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	subscriptions := dbSubscriptions{db: db}

	t.Run("no account number", func(t *testing.T) {
		u, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
		require.NoError(t, err)

		sub, err := subscriptions.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)

		got, err := subscriptions.GetByID(ctx, sub)
		require.NoError(t, err)
		assert.Equal(t, sub, got.ID)
		assert.Equal(t, u.ID, got.UserID)

		require.NotNil(t, got.AccountNumber)
		assert.Empty(t, *got.AccountNumber)
	})

	u, err := db.Users().Create(ctx, database.NewUser{Username: "u-11223344"})
	require.NoError(t, err)

	sub, err := subscriptions.Create(ctx, u.ID, u.Username)
	require.NoError(t, err)

	got, err := subscriptions.GetByID(ctx, sub)
	require.NoError(t, err)
	assert.Equal(t, sub, got.ID)
	assert.Equal(t, u.ID, got.UserID)
	assert.Nil(t, got.BillingSubscriptionID)

	require.NotNil(t, got.AccountNumber)
	assert.Equal(t, "11223344", *got.AccountNumber)

	ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: u.ID})
	require.NoError(t, err)
	assert.Len(t, ts, 1)

	ts, err = subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: 123 /* invalid */})
	require.NoError(t, err)
	assert.Len(t, ts, 0)
}

func TestProductSubscriptions_List(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	u1, err := db.Users().Create(ctx, database.NewUser{Username: "u1"})
	require.NoError(t, err)

	u2, err := db.Users().Create(ctx, database.NewUser{Username: "u2"})
	require.NoError(t, err)

	subscriptions := dbSubscriptions{db: db}

	_, err = subscriptions.Create(ctx, u1.ID, "")
	require.NoError(t, err)
	_, err = subscriptions.Create(ctx, u1.ID, "")
	require.NoError(t, err)

	t.Run("List all product subscriptions", func(t *testing.T) {
		ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, len(ts))
		count, err := subscriptions.Count(ctx, dbSubscriptionsListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("List u1's product subscriptions", func(t *testing.T) {
		// List u1's product subscriptions.
		ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: u1.ID})
		require.NoError(t, err)
		assert.Equal(t, 2, len(ts))
	})

	t.Run("List u2's product subscriptions", func(t *testing.T) {
		ts, err := subscriptions.List(ctx, dbSubscriptionsListOptions{UserID: u2.ID})
		require.NoError(t, err)
		assert.Equal(t, 0, len(ts))
	})
}

func TestProductSubscriptions_Update(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	u, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	require.NoError(t, err)

	subscriptions := dbSubscriptions{db: db}

	sub0, err := subscriptions.Create(ctx, u.ID, "")
	require.NoError(t, err)
	got, err := subscriptions.GetByID(ctx, sub0)
	require.NoError(t, err)
	require.Nil(t, got.BillingSubscriptionID)

	t.Run("billingSubscriptionID", func(t *testing.T) {
		t.Run("set non-null value", func(t *testing.T) {
			err := subscriptions.Update(ctx, sub0, DBSubscriptionUpdate{
				BillingSubscriptionID: &sql.NullString{
					String: "x",
					Valid:  true,
				},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			autogold.Expect(valast.Addr("x").(*string)).Equal(t, got.BillingSubscriptionID)
		})

		t.Run("update no fields", func(t *testing.T) {
			err := subscriptions.Update(ctx, sub0, DBSubscriptionUpdate{})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			autogold.Expect(valast.Addr("x").(*string)).Equal(t, got.BillingSubscriptionID)
		})

		// Set null value.
		t.Run("set null value", func(t *testing.T) {
			err := subscriptions.Update(ctx, sub0, DBSubscriptionUpdate{
				BillingSubscriptionID: &sql.NullString{Valid: false},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			autogold.Expect((*string)(nil)).Equal(t, got.BillingSubscriptionID)
		})
	})

	t.Run("codyGatewayAccess", func(t *testing.T) {
		t.Run("set non-null values", func(t *testing.T) {
			err := subscriptions.Update(ctx, sub0, DBSubscriptionUpdate{
				CodyGatewayAccess: &graphqlbackend.UpdateCodyGatewayAccessInput{
					Enabled:                                 pointify(true),
					ChatCompletionsRateLimit:                pointify(graphqlbackend.BigInt(12)),
					ChatCompletionsRateLimitIntervalSeconds: pointify(int32(time.Hour.Seconds())),
					ChatCompletionsAllowedModels:            pointify([]string{"claude-v1"}),
					CodeCompletionsRateLimit:                pointify(graphqlbackend.BigInt(13)),
					CodeCompletionsRateLimitIntervalSeconds: pointify(int32(2 * time.Hour.Seconds())),
					CodeCompletionsAllowedModels:            pointify([]string{"claude-v2"}),
					EmbeddingsRateLimit:                     pointify(graphqlbackend.BigInt(14)),
					EmbeddingsRateLimitIntervalSeconds:      pointify(int32(3 * time.Hour.Seconds())),
					EmbeddingsAllowedModels:                 pointify([]string{"claude-v3"}),
				},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			autogold.Expect(dbCodyGatewayAccess{
				Enabled: true,
				ChatRateLimit: dbRateLimit{
					RateLimit:           valast.Addr(int64(12)).(*int64),
					RateIntervalSeconds: valast.Addr(int32(3600)).(*int32),
					AllowedModels:       []string{"claude-v1"},
				},
				CodeRateLimit: dbRateLimit{
					RateLimit:           valast.Addr(int64(13)).(*int64),
					RateIntervalSeconds: valast.Addr(int32(2 * 3600)).(*int32),
					AllowedModels:       []string{"claude-v2"},
				},
				EmbeddingsRateLimit: dbRateLimit{
					RateLimit:           valast.Addr(int64(14)).(*int64),
					RateIntervalSeconds: valast.Addr(int32(3 * 3600)).(*int32),
					AllowedModels:       []string{"claude-v3"},
				},
			}).Equal(t, got.CodyGatewayAccess)
		})

		t.Run("set to zero/null values", func(t *testing.T) {
			err := subscriptions.Update(ctx, sub0, DBSubscriptionUpdate{
				CodyGatewayAccess: &graphqlbackend.UpdateCodyGatewayAccessInput{
					Enabled:                                 pointify(false),
					ChatCompletionsRateLimit:                pointify(graphqlbackend.BigInt(0)),
					ChatCompletionsRateLimitIntervalSeconds: pointify(int32(0)),
					ChatCompletionsAllowedModels:            pointify([]string{}),
					CodeCompletionsRateLimit:                pointify(graphqlbackend.BigInt(0)),
					CodeCompletionsRateLimitIntervalSeconds: pointify(int32(0)),
					CodeCompletionsAllowedModels:            pointify([]string{}),
					EmbeddingsRateLimit:                     pointify(graphqlbackend.BigInt(0)),
					EmbeddingsRateLimitIntervalSeconds:      pointify(int32(0)),
					EmbeddingsAllowedModels:                 pointify([]string{}),
				},
			})
			require.NoError(t, err)
			got, err := subscriptions.GetByID(ctx, sub0)
			require.NoError(t, err)
			autogold.Expect(dbCodyGatewayAccess{}).Equal(t, got.CodyGatewayAccess)
		})
	})
}
