package subscriptions_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/databasetest"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/tables"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
)

func TestSubscriptionsStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := databasetest.NewTestDB(t, "enterprise-portal", "SubscriptionsStore", tables.All()...)

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
		subscriptions.UpsertSubscriptionOptions{InstanceDomain: "s1.sourcegraph.com"},
	)
	require.NoError(t, err)
	s2, err := s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{InstanceDomain: "s2.sourcegraph.com"},
	)
	require.NoError(t, err)
	_, err = s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{InstanceDomain: "s3.sourcegraph.com"},
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
			InstanceDomains: []string{s1.InstanceDomain, s2.InstanceDomain}},
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
	// Create initial test record.
	s1, err := s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{InstanceDomain: "s1.sourcegraph.com"},
	)
	require.NoError(t, err)

	got, err := s.Get(ctx, s1.ID)
	require.NoError(t, err)
	assert.Equal(t, s1.ID, got.ID)
	assert.Equal(t, s1.InstanceDomain, got.InstanceDomain)

	t.Run("noop", func(t *testing.T) {
		got, err = s.Upsert(ctx, s1.ID, subscriptions.UpsertSubscriptionOptions{})
		require.NoError(t, err)
		assert.Equal(t, s1.InstanceDomain, got.InstanceDomain)
	})

	t.Run("update", func(t *testing.T) {
		got, err = s.Upsert(ctx, s1.ID, subscriptions.UpsertSubscriptionOptions{InstanceDomain: "s1-new.sourcegraph.com"})
		require.NoError(t, err)
		assert.Equal(t, "s1-new.sourcegraph.com", got.InstanceDomain)
	})

	t.Run("force update", func(t *testing.T) {
		got, err = s.Upsert(ctx, s1.ID, subscriptions.UpsertSubscriptionOptions{ForceUpdate: true})
		require.NoError(t, err)
		assert.Empty(t, got.InstanceDomain)
	})
}

func SubscriptionsStoreGet(t *testing.T, ctx context.Context, s *subscriptions.Store) {
	// Create initial test record.
	s1, err := s.Upsert(
		ctx,
		uuid.New().String(),
		subscriptions.UpsertSubscriptionOptions{InstanceDomain: "s1.sourcegraph.com"},
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
