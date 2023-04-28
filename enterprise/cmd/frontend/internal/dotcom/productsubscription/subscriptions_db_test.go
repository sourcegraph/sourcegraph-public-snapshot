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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	t.Run("no account number", func(t *testing.T) {
		u, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
		require.NoError(t, err)

		sub, err := dbSubscriptions{db: db}.Create(ctx, u.ID, u.Username)
		require.NoError(t, err)

		got, err := dbSubscriptions{db: db}.GetByID(ctx, sub)
		require.NoError(t, err)
		assert.Equal(t, sub, got.ID)
		assert.Equal(t, u.ID, got.UserID)

		require.NotNil(t, got.AccountNumber)
		assert.Empty(t, *got.AccountNumber)
	})

	u, err := db.Users().Create(ctx, database.NewUser{Username: "u-11223344"})
	require.NoError(t, err)

	sub, err := dbSubscriptions{db: db}.Create(ctx, u.ID, u.Username)
	require.NoError(t, err)

	got, err := dbSubscriptions{db: db}.GetByID(ctx, sub)
	require.NoError(t, err)
	assert.Equal(t, sub, got.ID)
	assert.Equal(t, u.ID, got.UserID)
	assert.Nil(t, got.BillingSubscriptionID)

	require.NotNil(t, got.AccountNumber)
	assert.Equal(t, "11223344", *got.AccountNumber)

	ts, err := dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{UserID: u.ID})
	require.NoError(t, err)
	assert.Len(t, ts, 1)

	ts, err = dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{UserID: 123 /* invalid */})
	require.NoError(t, err)
	assert.Len(t, ts, 0)
}

func TestProductSubscriptions_List(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	u1, err := db.Users().Create(ctx, database.NewUser{Username: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	u2, err := db.Users().Create(ctx, database.NewUser{Username: "u2"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = dbSubscriptions{db: db}.Create(ctx, u1.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = dbSubscriptions{db: db}.Create(ctx, u1.ID, "")
	if err != nil {
		t.Fatal(err)
	}

	{
		// List all product subscriptions.
		ts, err := dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d product subscriptions, want %d", len(ts), want)
		}
		count, err := dbSubscriptions{db: db}.Count(ctx, dbSubscriptionsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List u1's product subscriptions.
		ts, err := dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{UserID: u1.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d product subscriptions, want %d", len(ts), want)
		}
	}

	{
		// List u2's product subscriptions.
		ts, err := dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{UserID: u2.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := 0; len(ts) != want {
			t.Errorf("got %d product subscriptions, want %d", len(ts), want)
		}
	}
}

func TestProductSubscriptions_Update(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	u, err := db.Users().Create(ctx, database.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	sub0, err := dbSubscriptions{db: db}.Create(ctx, u.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := (dbSubscriptions{db: db}).GetByID(ctx, sub0); err != nil {
		t.Fatal(err)
	} else if got.BillingSubscriptionID != nil {
		t.Errorf("got %q, want nil", *got.BillingSubscriptionID)
	}

	t.Run("billingSubscriptionID", func(t *testing.T) {
		// Set non-null value.
		if err := (dbSubscriptions{db: db}).Update(ctx, sub0, dbSubscriptionUpdate{
			billingSubscriptionID: &sql.NullString{
				String: "x",
				Valid:  true,
			},
		}); err != nil {
			t.Fatal(err)
		}
		if got, err := (dbSubscriptions{db: db}).GetByID(ctx, sub0); err != nil {
			t.Fatal(err)
		} else {
			autogold.Expect(valast.Addr("x").(*string)).Equal(t, got.BillingSubscriptionID)
		}

		// Update no fields.
		if err := (dbSubscriptions{db: db}).Update(ctx, sub0, dbSubscriptionUpdate{billingSubscriptionID: nil}); err != nil {
			t.Fatal(err)
		}
		if got, err := (dbSubscriptions{db: db}).GetByID(ctx, sub0); err != nil {
			t.Fatal(err)
		} else {
			autogold.Expect(valast.Addr("x").(*string)).Equal(t, got.BillingSubscriptionID)
		}

		// Set null value.
		if err := (dbSubscriptions{db: db}).Update(ctx, sub0, dbSubscriptionUpdate{
			billingSubscriptionID: &sql.NullString{Valid: false},
		}); err != nil {
			t.Fatal(err)
		}
		if got, err := (dbSubscriptions{db: db}).GetByID(ctx, sub0); err != nil {
			t.Fatal(err)
		} else {
			autogold.Expect((*string)(nil)).Equal(t, got.BillingSubscriptionID)
		}
	})

	t.Run("llmProxyAccess", func(t *testing.T) {
		// Set non-null values.
		if err := (dbSubscriptions{db: db}).Update(ctx, sub0, dbSubscriptionUpdate{
			llmProxyAccess: &graphqlbackend.UpdateLLMProxyAccessInput{
				Enabled:                  pointify(true),
				RateLimit:                pointify[int32](12),
				RateLimitIntervalSeconds: pointify(int32(time.Hour.Seconds())),
			},
		}); err != nil {
			t.Fatal(err)
		}
		if got, err := (dbSubscriptions{db: db}).GetByID(ctx, sub0); err != nil {
			t.Fatal(err)
		} else {
			autogold.Expect(dbLLMProxyAccess{
				Enabled: true, RateLimit: valast.Addr(int32(12)).(*int32),
				RateIntervalSeconds: valast.Addr(int32(3600)).(*int32),
			}).Equal(t, got.LLMProxyAccess)
		}

		// Set to zero/null value.
		if err := (dbSubscriptions{db: db}).Update(ctx, sub0, dbSubscriptionUpdate{
			llmProxyAccess: &graphqlbackend.UpdateLLMProxyAccessInput{
				Enabled:                  pointify(false),
				RateLimit:                pointify(int32(0)),
				RateLimitIntervalSeconds: pointify[int32](0),
			},
		}); err != nil {
			t.Fatal(err)
		}
		if got, err := (dbSubscriptions{db: db}).GetByID(ctx, sub0); err != nil {
			t.Fatal(err)
		} else {
			autogold.Expect(dbLLMProxyAccess{}).Equal(t, got.LLMProxyAccess)
		}
	})
}
