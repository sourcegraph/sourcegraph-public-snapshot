package codyaccess_test

import (
	"context"
	"database/sql"
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
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/utctime"
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
			DisplayName:    database.NewNullString(mockSubscriptionDisplayName(idx)),
			InstanceDomain: database.NewNullString(fmt.Sprintf("s%d.sourcegraph.com", idx)),
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
				&subscriptions.DataLicenseKey{
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
				DisplayName:    database.NewNullString("Archived subscription"),
				InstanceDomain: database.NewNullString("archived.sourcegraph.com"),
				ArchivedAt:     pointers.Ptr(utctime.Now()),
			})
			require.NoError(t, err)

			_, err = codyaccess.NewCodyGatewayStore(db).Upsert(ctx, subscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
				Enabled: pointers.Ptr(true),
			})
			assert.ErrorIs(t, err, codyaccess.ErrSubscriptionNotFound)
		})

		t.Run("set then archive", func(t *testing.T) {
			subscriptionID := uuid.NewString()
			_, err := subscriptions.NewStore(db).Upsert(ctx, subscriptionID, subscriptions.UpsertSubscriptionOptions{
				DisplayName:    database.NewNullString("Soon-to-be-archived subscription"),
				InstanceDomain: database.NewNullString("not-yet-archived.sourcegraph.com"),
			})
			require.NoError(t, err)

			_, err = codyaccess.NewCodyGatewayStore(db).Upsert(ctx, subscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
				Enabled: pointers.Ptr(true),
			})
			require.NoError(t, err)

			_, err = subscriptions.NewStore(db).Upsert(ctx, subscriptionID, subscriptions.UpsertSubscriptionOptions{
				ArchivedAt: pointers.Ptr(utctime.Now()),
			})
			require.NoError(t, err)

			_, err = codyaccess.NewCodyGatewayStore(db).Get(ctx, codyaccess.GetCodyGatewayAccessOptions{
				SubscriptionID: subscriptionID,
			})
			assert.ErrorIs(t, err, codyaccess.ErrSubscriptionNotFound)
		})
	})
}

func CodyGatewayStoreListAndGet(t *testing.T, ctx context.Context, subscriptionIDs []string, s *codyaccess.CodyGatewayStore) {
	// Create explicit access for all but the last subscription. The last
	// subscription still be able to get zero-value Cody Access.
	for idx, sub := range subscriptionIDs[:len(subscriptionIDs)-1] {
		_, err := s.Upsert(ctx, sub, codyaccess.UpsertCodyGatewayAccessOptions{
			Enabled:                                 pointers.Ptr(idx%2 == 0), // even
			ChatCompletionsRateLimit:                database.NewNullInt64(idx),
			ChatCompletionsRateLimitIntervalSeconds: database.NewNullInt32(idx),
			CodeCompletionsRateLimit:                database.NewNullInt64(idx),
			CodeCompletionsRateLimitIntervalSeconds: database.NewNullInt32(idx),
			EmbeddingsRateLimit:                     database.NewNullInt64(idx),
			EmbeddingsRateLimitIntervalSeconds:      database.NewNullInt32(idx),
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
		assert.Equal(t, int64(idx), got.ChatCompletionsRateLimit.Int64)
		assert.Equal(t, int32(idx), got.ChatCompletionsRateLimitIntervalSeconds.Int32)
		assert.Equal(t, int64(idx), got.CodeCompletionsRateLimit.Int64)
		assert.Equal(t, int32(idx), got.CodeCompletionsRateLimitIntervalSeconds.Int32)
		assert.Equal(t, int64(idx), got.EmbeddingsRateLimit.Int64)
		assert.Equal(t, int32(idx), got.EmbeddingsRateLimitIntervalSeconds.Int32)

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
				got, err := s.Get(ctx, codyaccess.GetCodyGatewayAccessOptions{
					SubscriptionID: sub,
				})
				require.NoError(t, err)

				assertAccess(idx, got)

				// Reverse lookup by license key hash
				for _, hash := range got.LicenseKeyHashes {
					got2, err := s.Get(ctx, codyaccess.GetCodyGatewayAccessOptions{
						LicenseKeyHash: hash,
					})
					require.NoError(t, err)
					assert.Len(t, got2.LicenseKeyHashes, 2) // 2 valid licenses
					assert.Equal(t, got, got2)
				}
			})
		}

		t.Run("ErrSubscriptionDoesNotExist", func(t *testing.T) {
			_, err := s.Get(ctx, codyaccess.GetCodyGatewayAccessOptions{
				SubscriptionID: uuid.NewString(),
			})
			assert.Error(t, err)
			assert.ErrorIs(t, err, codyaccess.ErrSubscriptionNotFound)

			_, err = s.Get(ctx, codyaccess.GetCodyGatewayAccessOptions{
				LicenseKeyHash: []byte(uuid.NewString()),
			})
			assert.Error(t, err)
			assert.ErrorIs(t, err, codyaccess.ErrSubscriptionNotFound)
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
			Enabled: pointers.Ptr(false),
		},
	)
	require.NoError(t, err)

	got, err := s.Get(ctx, codyaccess.GetCodyGatewayAccessOptions{
		SubscriptionID: currentAccess.SubscriptionID,
	})
	require.NoError(t, err)
	assert.False(t, got.Enabled)
	assert.Equal(t, currentAccess.SubscriptionID, got.SubscriptionID)
	assert.Equal(t, mockSubscriptionDisplayName(0), got.DisplayName)
	assert.False(t, got.ChatCompletionsRateLimit.Valid)
	assert.False(t, got.ChatCompletionsRateLimitIntervalSeconds.Valid)
	assert.False(t, got.CodeCompletionsRateLimit.Valid)
	assert.False(t, got.CodeCompletionsRateLimitIntervalSeconds.Valid)
	assert.False(t, got.EmbeddingsRateLimit.Valid)
	assert.False(t, got.EmbeddingsRateLimitIntervalSeconds.Valid)

	t.Run("noop", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{})
		require.NoError(t, err)
		assert.Equal(t, currentAccess, got)
	})

	t.Run("subscription does not exist", func(t *testing.T) {
		subscriptionID := uuid.NewString()
		_, err = s.Upsert(ctx, subscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			Enabled: pointers.Ptr(false),
		})
		assert.ErrorIs(t, err, codyaccess.ErrSubscriptionNotFound)
	})

	t.Run("update only enabled", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			Enabled: pointers.Ptr(true),
		})
		require.NoError(t, err)
		assert.True(t, got.Enabled)
		assert.Equal(t, currentAccess.DisplayName, got.DisplayName)
		assert.Equal(t, currentAccess.CodeCompletionsRateLimit, got.CodeCompletionsRateLimit)
		assert.EqualValues(t, currentAccess.ChatCompletionsRateLimit, got.ChatCompletionsRateLimit)
	})

	t.Run("update only chat completions", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			ChatCompletionsRateLimit: database.NewNullInt64(1234),
		})
		require.NoError(t, err)
		assert.Equal(t, true, got.Enabled)
		assert.Equal(t, currentAccess.DisplayName, got.DisplayName)
		assert.Equal(t, currentAccess.CodeCompletionsRateLimit, got.CodeCompletionsRateLimit)
		assert.EqualValues(t, 1234, got.ChatCompletionsRateLimit.Int64)
	})

	t.Run("remove chat completion overrides", func(t *testing.T) {
		t.Cleanup(func() { currentAccess = got })

		got, err = s.Upsert(ctx, currentAccess.SubscriptionID, codyaccess.UpsertCodyGatewayAccessOptions{
			ChatCompletionsRateLimit:                &sql.NullInt64{},
			ChatCompletionsRateLimitIntervalSeconds: &sql.NullInt32{},
		})
		require.NoError(t, err)
		assert.Equal(t, true, got.Enabled)
		assert.Equal(t, currentAccess.DisplayName, got.DisplayName)

		// unchanged
		assert.Equal(t, currentAccess.CodeCompletionsRateLimit, got.CodeCompletionsRateLimit)

		// changed
		assert.False(t, got.ChatCompletionsRateLimit.Valid)
		assert.False(t, got.ChatCompletionsRateLimitIntervalSeconds.Valid)
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
