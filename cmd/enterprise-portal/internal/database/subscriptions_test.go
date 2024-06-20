package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionsStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := newSubscriptionsStore(newTestDB(t, "enterprise-portal", "SubscriptionsStore", allTables...))

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, s *SubscriptionsStore)
	}{
		{"List", SubscriptionsStoreList},
		{"Upsert", SubscriptionsStoreUpsert},
		{"Get", SubscriptionsStoreGet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := clearTables(t, db.db, allTables...)
				require.NoError(t, err)
			})
			tc.test(t, ctx, db)
		})
		if t.Failed() {
			break
		}
	}
}

func SubscriptionsStoreList(t *testing.T, ctx context.Context, s *SubscriptionsStore) {
	// Create test records.
	s1, err := s.Upsert(
		ctx,
		uuid.New().String(),
		UpsertSubscriptionOptions{InstanceDomain: "s1.sourcegraph.com"},
	)
	require.NoError(t, err)
	s2, err := s.Upsert(
		ctx,
		uuid.New().String(),
		UpsertSubscriptionOptions{InstanceDomain: "s2.sourcegraph.com"},
	)
	require.NoError(t, err)
	_, err = s.Upsert(
		ctx,
		uuid.New().String(),
		UpsertSubscriptionOptions{InstanceDomain: "s3.sourcegraph.com"},
	)
	require.NoError(t, err)

	t.Run("list by IDs", func(t *testing.T) {
		subscriptions, err := s.List(ctx, ListEnterpriseSubscriptionsOptions{IDs: []string{s1.ID, s2.ID}})
		require.NoError(t, err)
		require.Len(t, subscriptions, 2)
		assert.Equal(t, s1.ID, subscriptions[0].ID)
		assert.Equal(t, s2.ID, subscriptions[1].ID)

		t.Run("no match", func(t *testing.T) {
			subscriptions, err = s.List(ctx, ListEnterpriseSubscriptionsOptions{IDs: []string{"1234"}})
			require.NoError(t, err)
			require.Len(t, subscriptions, 0)
		})
	})

	t.Run("list by instance domains", func(t *testing.T) {
		subscriptions, err := s.List(ctx, ListEnterpriseSubscriptionsOptions{
			InstanceDomains: []string{s1.InstanceDomain, s2.InstanceDomain}},
		)
		require.NoError(t, err)
		require.Len(t, subscriptions, 2)
		assert.Equal(t, s1.ID, subscriptions[0].ID)
		assert.Equal(t, s2.ID, subscriptions[1].ID)

		t.Run("no match", func(t *testing.T) {
			subscriptions, err = s.List(ctx, ListEnterpriseSubscriptionsOptions{InstanceDomains: []string{"1234"}})
			require.NoError(t, err)
			require.Len(t, subscriptions, 0)
		})
	})

	t.Run("list with page size", func(t *testing.T) {
		subscriptions, err := s.List(
			ctx,
			ListEnterpriseSubscriptionsOptions{
				IDs:      []string{s1.ID, s2.ID}, // Two matching but only of them will be returned.
				PageSize: 1,
			},
		)
		require.NoError(t, err)
		assert.Len(t, subscriptions, 1)
	})
}

func SubscriptionsStoreUpsert(t *testing.T, ctx context.Context, s *SubscriptionsStore) {
	// Create initial test record.
	s1, err := s.Upsert(
		ctx,
		uuid.New().String(),
		UpsertSubscriptionOptions{InstanceDomain: "s1.sourcegraph.com"},
	)
	require.NoError(t, err)

	got, err := s.Get(ctx, s1.ID)
	require.NoError(t, err)
	assert.Equal(t, s1.ID, got.ID)
	assert.Equal(t, s1.InstanceDomain, got.InstanceDomain)

	t.Run("noop", func(t *testing.T) {
		got, err = s.Upsert(ctx, s1.ID, UpsertSubscriptionOptions{})
		require.NoError(t, err)
		assert.Equal(t, s1.InstanceDomain, got.InstanceDomain)
	})

	t.Run("update", func(t *testing.T) {
		got, err = s.Upsert(ctx, s1.ID, UpsertSubscriptionOptions{InstanceDomain: "s1-new.sourcegraph.com"})
		require.NoError(t, err)
		assert.Equal(t, "s1-new.sourcegraph.com", got.InstanceDomain)
	})

	t.Run("force update", func(t *testing.T) {
		got, err = s.Upsert(ctx, s1.ID, UpsertSubscriptionOptions{ForceUpdate: true})
		require.NoError(t, err)
		assert.Empty(t, got.InstanceDomain)
	})
}

func SubscriptionsStoreGet(t *testing.T, ctx context.Context, s *SubscriptionsStore) {
	// Create initial test record.
	s1, err := s.Upsert(
		ctx,
		uuid.New().String(),
		UpsertSubscriptionOptions{InstanceDomain: "s1.sourcegraph.com"},
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
