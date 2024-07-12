package codyaccess_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/codyaccess"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/databasetest"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/tables"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/utctime"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func mockSubscriptionDisplayName(idx int) string {
	return fmt.Sprintf("Subscription %d", idx)
}

func TestCodyGatewayStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := databasetest.NewTestDB(t, "enterprise-portal", t.Name(), tables.All()...)

	// Create a test subscription as a parent.
	subscriptionIDs := make([]string, 10)
	for idx := range subscriptionIDs {
		subscriptionID := uuid.NewString()
		_, err := subscriptions.NewStore(db).Upsert(ctx, subscriptionID, subscriptions.UpsertSubscriptionOptions{
			DisplayName:    pointers.Ptr(database.NewNullString(mockSubscriptionDisplayName(idx))),
			InstanceDomain: pointers.Ptr(database.NewNullString(fmt.Sprintf("s%d.sourcegraph.com", idx))),
		})
		require.NoError(t, err)
		subscriptionIDs[idx] = subscriptionID

		now := time.Now()
		// We create 3 licenses for each:
		// - 1 expired
		// - 1 revoked
		// - 2 active
		for _, opt := range []subscriptions.CreateLicenseOpts{
			{
				Message:    "expired",
				Time:       pointers.Ptr(utctime.FromTime(now.Add(-10 * time.Hour))),
				ExpireTime: utctime.FromTime(now.Add(-1 * time.Hour)),
			},
			{
				Message:    "revoked",
				Time:       pointers.Ptr(utctime.FromTime(now.Add(-5 * time.Hour))),
				ExpireTime: utctime.FromTime(now.Add(-1 * time.Hour)),
			},
			{
				Message:    "activeOld",
				Time:       pointers.Ptr(utctime.FromTime(now.Add(-4 * time.Hour))),
				ExpireTime: utctime.FromTime(now.Add(10 * time.Hour)),
			},
			{
				Message:    "activeLicense",
				Time:       pointers.Ptr(utctime.FromTime(now.Add(-1 * time.Hour))),
				ExpireTime: utctime.FromTime(now.Add(10 * time.Hour)),
			},
		} {
			l, err := subscriptions.NewLicensesStore(db).CreateLicenseKey(ctx,
				subscriptionID,
				&subscriptions.LicenseKey{
					Info: license.Info{
						// Set properties that are easy to assert on later
						Tags: []string{
							strconv.Itoa(idx),
							opt.Message,
						},
						UserCount: uint(idx),
						CreatedAt: opt.Time.AsTime(),
						ExpiresAt: opt.ExpireTime.AsTime(),
					},
					SignedKey: subscriptionID, // stub value
				},
				opt)
			require.NoError(t, err)
			if opt.Message == "revoked" {
				_, err := subscriptions.NewLicensesStore(db).Revoke(ctx, l.ID, subscriptions.RevokeLicenseOpts{
					Message: t.Name(),
				})
				require.NoError(t, err)
			}
		}

	}
	databasetest.ClearTablesAfterTest(t, db, tables.All()...)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, subscriptionIDs []string, s *codyaccess.CodyGatewayStore)
	}{
		{"ListAndGet", CodyGatewayStoreListAndGet},
		{"Upsert", CodyGatewayStoreUpsert},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Only clear the CodyGatewayAccess table between tests.
			databasetest.ClearTablesAfterTest(t, db, &codyaccess.TableCodyGatewayAccess{})

			tc.test(t, ctx, subscriptionIDs, codyaccess.NewCodyGatewayStore(db))
		})
		if t.Failed() {
			break
		}
	}

	t.Run("archived subscriptions", func(t *testing.T) {
		t.Run("already archived", func(t *testing.T) {
			subscriptionID := uuid.NewString()
			_, err := subscriptions.NewStore(db).Upsert(ctx, subscriptionID, subscriptions.UpsertSubscriptionOptions{
				DisplayName:    pointers.Ptr(database.NewNullString("Archived subscription")),
				InstanceDomain: pointers.Ptr(database.NewNullString("archived.sourcegraph.com")),
				ArchivedAt:     pointers.Ptr(time.Now()),
			})
			require.NoError(t, err)

			_, err = codyaccess.NewCodyGatewayStore(db).Upsert(ctx, subscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
				Enabled: true,
			})
			assert.ErrorIs(t, err, codyaccess.ErrSubscriptionDoesNotExist)
		})

		t.Run("set then archive", func(t *testing.T) {
			subscriptionID := uuid.NewString()
			_, err := subscriptions.NewStore(db).Upsert(ctx, subscriptionID, subscriptions.UpsertSubscriptionOptions{
				DisplayName:    pointers.Ptr(database.NewNullString("Soon-to-be-archived subscription")),
				InstanceDomain: pointers.Ptr(database.NewNullString("not-yet-archived.sourcegraph.com")),
			})
			require.NoError(t, err)

			_, err = codyaccess.NewCodyGatewayStore(db).Upsert(ctx, subscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
				Enabled: true,
			})
			require.NoError(t, err)

			_, err = subscriptions.NewStore(db).Upsert(ctx, subscriptionID, subscriptions.UpsertSubscriptionOptions{
				ArchivedAt: pointers.Ptr(time.Now()),
			})
			require.NoError(t, err)

			_, err = codyaccess.NewCodyGatewayStore(db).Get(ctx, subscriptionID)
			assert.ErrorIs(t, err, codyaccess.ErrSubscriptionDoesNotExist)
		})
	})
}

func CodyGatewayStoreListAndGet(t *testing.T, ctx context.Context, subscriptionIDs []string, s *codyaccess.CodyGatewayStore) {
	// Create explicit access for all but the last subscription. The last
	// subscription still be able to get zero-value Cody Access.
	for idx, sub := range subscriptionIDs[:len(subscriptionIDs)-1] {
		_, err := s.Upsert(ctx, sub, codyaccess.UpsertCodyGatewayAccessOptions{
			Enabled:                                 idx%2 == 0, // even
			ChatCompletionsRateLimit:                pointers.Ptr(int64(idx)),
			ChatCompletionsRateLimitIntervalSeconds: pointers.Ptr(int(idx)),
			CodeCompletionsRateLimit:                pointers.Ptr(int64(idx)),
			CodeCompletionsRateLimitIntervalSeconds: pointers.Ptr(int(idx)),
			EmbeddingsRateLimit:                     pointers.Ptr(int64(idx)),
			EmbeddingsRateLimitIntervalSeconds:      pointers.Ptr(int(idx)),
		})
		require.NoError(t, err)
	}
	assertAccess := func(idx int, got *codyaccess.CodyGatewayAccessWithSubscriptionDetails) {
		assert.Equal(t, subscriptionIDs[idx], got.SubscriptionID)
		assert.Equal(t, mockSubscriptionDisplayName(idx), got.DisplayName)

		// Last subscription has no explicit access, only has default values.
		if idx == len(subscriptionIDs)-1 {
			assert.Equal(t,
				codyaccess.CodyGatewayAccess{
					SubscriptionID: subscriptionIDs[idx],
					Enabled:        false,
				},
				got.CodyGatewayAccess)
			return
		}

		assert.Equal(t, idx%2 == 0, got.Enabled)
		assert.Equal(t, int64(idx), pointers.DerefZero(got.ChatCompletionsRateLimit))
		assert.Equal(t, idx, pointers.DerefZero(got.ChatCompletionsRateLimitIntervalSeconds))
		assert.Equal(t, int64(idx), pointers.DerefZero(got.CodeCompletionsRateLimit))
		assert.Equal(t, idx, pointers.DerefZero(got.CodeCompletionsRateLimitIntervalSeconds))
		assert.Equal(t, int64(idx), pointers.DerefZero(got.EmbeddingsRateLimit))
		assert.Equal(t, idx, pointers.DerefZero(got.EmbeddingsRateLimitIntervalSeconds))

		assert.Equal(t, []string{strconv.Itoa(idx), "activeLicense"}, got.ActiveLicenseInfo.Tags)
		assert.Equal(t, uint(idx), got.ActiveLicenseInfo.UserCount)
		assert.Len(t, got.LicenseKeyHashes, 2) // 2 valid licenses
		for _, hash := range got.LicenseKeyHashes {
			assert.NotEmpty(t, hash)
		}
	}

	t.Run("List", func(t *testing.T) {
		accs, err := s.List(ctx)
		require.NoError(t, err)
		require.Len(t, accs, len(subscriptionIDs))

		for idx, got := range accs {
			assertAccess(idx, got)
		}
	})

	t.Run("Get", func(t *testing.T) {
		for idx, sub := range subscriptionIDs {
			t.Run(fmt.Sprintf("idx=%d", idx), func(t *testing.T) {
				got, err := s.Get(ctx, sub)
				require.NoError(t, err)

				assertAccess(idx, got)
			})
		}

		t.Run("ErrSubscriptionDoesNotExist", func(t *testing.T) {
			_, err := s.Get(ctx, uuid.NewString())
			assert.ErrorIs(t, err, codyaccess.ErrSubscriptionDoesNotExist)
		})
	})
}

func CodyGatewayStoreUpsert(t *testing.T, ctx context.Context, subscriptionIDs []string, s *codyaccess.CodyGatewayStore) {
	// Create initial test record. The currentAccess should be reassigned
	// throughout various test cases to represent the current state of the test
	// record, as the subtests are run in sequence.
	currentAccess, err := s.Upsert(
		ctx,
		subscriptionIDs[0],
		codyaccess.UpsertCodyGatewayAccessOptions{
			Enabled: false,
		},
	)
	require.NoError(t, err)

	got, err := s.Get(ctx, currentAccess.SubscriptionID)
	require.NoError(t, err)
	assert.False(t, got.Enabled)
	assert.Equal(t, currentAccess.SubscriptionID, got.SubscriptionID)
	assert.Equal(t, mockSubscriptionDisplayName(0), got.DisplayName)
	assert.Nil(t, got.ChatCompletionsRateLimit)
	assert.Nil(t, got.ChatCompletionsRateLimitIntervalSeconds)
	assert.Nil(t, got.CodeCompletionsRateLimit)
	assert.Nil(t, got.CodeCompletionsRateLimitIntervalSeconds)
	assert.Nil(t, got.EmbeddingsRateLimit)
	assert.Nil(t, got.EmbeddingsRateLimitIntervalSeconds)

	t.Run("noop", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{})
		require.NoError(t, err)
		assert.Equal(t, currentAccess, got)
	})

	t.Run("subscription does not exist", func(t *testing.T) {
		subscriptionID := uuid.NewString()
		_, err = s.Upsert(ctx, subscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			Enabled: true,
		})
		assert.ErrorIs(t, err, codyaccess.ErrSubscriptionDoesNotExist)
	})

	t.Run("update only chat completions", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			ChatCompletionsRateLimit: pointers.Ptr(int64(1234)),
		})
		require.NoError(t, err)
		assert.Equal(t, currentAccess.Enabled, got.Enabled)
		assert.Equal(t, currentAccess.DisplayName, got.DisplayName)
		assert.Equal(t, currentAccess.CodeCompletionsRateLimit, got.CodeCompletionsRateLimit)
		assert.EqualValues(t, 1234, *got.ChatCompletionsRateLimit)
	})

	t.Run("update only enabled", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			Enabled: true,
		})
		require.NoError(t, err)
		assert.True(t, got.Enabled)
		assert.Equal(t, currentAccess.DisplayName, got.DisplayName)
		assert.Equal(t, currentAccess.CodeCompletionsRateLimit, got.CodeCompletionsRateLimit)
		assert.EqualValues(t, 1234, *got.ChatCompletionsRateLimit)
	})

	t.Run("force update to zero values", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			ForceUpdate: true,
		})
		require.NoError(t, err)

		// Only fields that cannot be changed are not reset to zero values
		assert.Equal(t, codyaccess.CodyGatewayAccessWithSubscriptionDetails{
			CodyGatewayAccess: codyaccess.CodyGatewayAccess{
				SubscriptionID: currentAccess.SubscriptionID,
			},
			DisplayName:       currentAccess.DisplayName,
			ActiveLicenseInfo: currentAccess.ActiveLicenseInfo,
			LicenseKeyHashes:  currentAccess.LicenseKeyHashes,
		}, *got)
	})
}
