package subscriptions_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/databasetest"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/tables"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestSubscriptionsStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := databasetest.NewTestDB(t, "enterprise-portal", t.Name(), tables.All()...)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, s *subscriptions.Store)
	}{
		{"List", SubscriptionsStoreList},
		{"Upsert", SubscriptionsStoreUpsert},
		{"Get", SubscriptionsStoreGet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			databasetest.ClearTablesAfterTest(t, db, tables.All()...)
			tc.test(t, ctx, subscriptions.NewStore(db))
		})
		if t.Failed() {
			break
		}
	}
}

func SubscriptionsStoreList(t *testing.T, ctx context.Context, s *subscriptions.Store) {
	// Create test records.
	s1, err := s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{
			InstanceDomain: pointers.Ptr(database.NewNullString("s1.sourcegraph.com")),
		},
	)
	require.NoError(t, err)
	s2, err := s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{
			InstanceDomain: pointers.Ptr(database.NewNullString("s2.sourcegraph.com")),
		},
	)
	require.NoError(t, err)
	_, err = s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{
			InstanceDomain: pointers.Ptr(database.NewNullString("s3.sourcegraph.com")),
		},
	)
	require.NoError(t, err)

	t.Run("list by IDs", func(t *testing.T) {
		ss, err := s.List(ctx, subscriptions.ListEnterpriseSubscriptionsOptions{IDs: []string{s1.ID, s2.ID}})
		require.NoError(t, err)
		require.Len(t, ss, 2)
		assert.Equal(t, s1.ID, ss[0].ID)
		assert.Equal(t, s2.ID, ss[1].ID)

		t.Run("no match", func(t *testing.T) {
			ss, err = s.List(ctx, subscriptions.ListEnterpriseSubscriptionsOptions{
				IDs: []string{uuid.Must(uuid.NewV7()).String()},
			})
			require.NoError(t, err)
			require.Len(t, ss, 0)
		})
	})

	t.Run("list by instance domains", func(t *testing.T) {
		ss, err := s.List(ctx, subscriptions.ListEnterpriseSubscriptionsOptions{
			InstanceDomains: []string{*s1.InstanceDomain, *s2.InstanceDomain}},
		)
		require.NoError(t, err)
		require.Len(t, ss, 2)
		assert.Equal(t, s1.ID, ss[0].ID)
		assert.Equal(t, s2.ID, ss[1].ID)

		t.Run("no match", func(t *testing.T) {
			ss, err = s.List(ctx, subscriptions.ListEnterpriseSubscriptionsOptions{
				InstanceDomains: []string{"1234"},
			})
			require.NoError(t, err)
			require.Len(t, ss, 0)
		})
	})

	t.Run("list with page size", func(t *testing.T) {
		ss, err := s.List(
			ctx,
			subscriptions.ListEnterpriseSubscriptionsOptions{
				IDs:      []string{s1.ID, s2.ID}, // Two matching but only of them will be returned.
				PageSize: 1,
			},
		)
		require.NoError(t, err)
		assert.Len(t, ss, 1)
	})
}

func SubscriptionsStoreUpsert(t *testing.T, ctx context.Context, s *subscriptions.Store) {
	// Create initial test record. The currentSubscription should be reassigned
	// throughout various test cases to represent the current state of the test
	// record, as the subtests are run in sequence.
	currentSubscription, err := s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{
			InstanceDomain: pointers.Ptr(database.NewNullString("s1.sourcegraph.com")),
		},
	)
	require.NoError(t, err)

	got, err := s.Get(ctx, currentSubscription.ID)
	require.NoError(t, err)
	assert.Equal(t, currentSubscription.ID, got.ID)
	assert.Equal(t, *currentSubscription.InstanceDomain, *got.InstanceDomain)
	assert.Empty(t, got.DisplayName)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
	assert.Nil(t, got.ArchivedAt) // not archived yet

	t.Run("noop", func(t *testing.T) {
		t.Cleanup(func() { currentSubscription = got })

		got, err = s.Upsert(ctx, currentSubscription.ID, subscriptions.UpsertSubscriptionOptions{})
		require.NoError(t, err)
		assert.Equal(t,
			pointers.DerefZero(currentSubscription.InstanceDomain),
			pointers.DerefZero(got.InstanceDomain))
	})

	t.Run("update only domain", func(t *testing.T) {
		t.Cleanup(func() { currentSubscription = got })

		got, err = s.Upsert(ctx, currentSubscription.ID, subscriptions.UpsertSubscriptionOptions{
			InstanceDomain: pointers.Ptr(database.NewNullString("s1-new.sourcegraph.com")),
		})
		require.NoError(t, err)
		assert.Equal(t, "s1-new.sourcegraph.com", pointers.DerefZero(got.InstanceDomain))
		assert.Equal(t, currentSubscription.DisplayName, got.DisplayName)
	})

	t.Run("update only display name", func(t *testing.T) {
		t.Cleanup(func() { currentSubscription = got })

		got, err = s.Upsert(ctx, currentSubscription.ID, subscriptions.UpsertSubscriptionOptions{
			DisplayName: pointers.Ptr(database.NewNullString("My New Display Name")),
		})
		require.NoError(t, err)
		assert.Equal(t, *currentSubscription.InstanceDomain, *got.InstanceDomain)
		assert.Equal(t, "My New Display Name", pointers.DerefZero(got.DisplayName))
	})

	t.Run("update only created at", func(t *testing.T) {
		t.Cleanup(func() { currentSubscription = got })

		yesterday := time.Now().Add(-24 * time.Hour)
		got, err = s.Upsert(ctx, currentSubscription.ID, subscriptions.UpsertSubscriptionOptions{
			CreatedAt: yesterday,
		})
		require.NoError(t, err)
		assert.Equal(t,
			pointers.DerefZero(currentSubscription.InstanceDomain),
			pointers.DerefZero(got.InstanceDomain))
		assert.Equal(t, currentSubscription.DisplayName, got.DisplayName)
		// Round times to allow for some precision drift in CI
		assert.Equal(t, yesterday.Round(time.Second).UTC(), got.CreatedAt.GetTime().Round(time.Second))
	})

	t.Run("update only archived at", func(t *testing.T) {
		t.Cleanup(func() { currentSubscription = got })

		yesterday := time.Now().Add(-24 * time.Hour)
		got, err = s.Upsert(ctx, currentSubscription.ID, subscriptions.UpsertSubscriptionOptions{
			ArchivedAt: pointers.Ptr(yesterday),
		})
		require.NoError(t, err)
		assert.Equal(t, *currentSubscription.InstanceDomain, *got.InstanceDomain)
		assert.Equal(t, *currentSubscription.DisplayName, *got.DisplayName)
		assert.Equal(t, currentSubscription.CreatedAt, got.CreatedAt)
		// Round times to allow for some precision drift in CI
		assert.Equal(t, yesterday.Round(time.Second).UTC(), got.ArchivedAt.GetTime().Round(time.Second))
	})

	t.Run("force update to zero values", func(t *testing.T) {
		t.Cleanup(func() { currentSubscription = got })

		got, err = s.Upsert(ctx, currentSubscription.ID, subscriptions.UpsertSubscriptionOptions{
			ForceUpdate: true,
		})
		require.NoError(t, err)
		assert.Empty(t, got.InstanceDomain)
		assert.Empty(t, got.DisplayName)
		assert.Nil(t, got.ArchivedAt)

		// Some fields cannot be updated in a force-update.
		assert.Equal(t, currentSubscription.ID, got.ID)
		assert.Equal(t, currentSubscription.CreatedAt, got.CreatedAt)
	})
}

func SubscriptionsStoreGet(t *testing.T, ctx context.Context, s *subscriptions.Store) {
	// Create initial test record.
	s1, err := s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{
			InstanceDomain: pointers.Ptr(database.NewNullString("s1.sourcegraph.com")),
		},
	)
	require.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		_, err := s.Get(ctx, uuid.New().String())
		assert.Equal(t, pgx.ErrNoRows, err)
	})

	t.Run("found", func(t *testing.T) {
		got, err := s.Get(ctx, s1.ID)
		require.NoError(t, err)
		assert.Equal(t, s1.ID, got.ID)
		assert.Equal(t, s1.InstanceDomain, got.InstanceDomain)
	})
}
