package campaigns

import (
	"container/heap"
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/campaigns"

	"github.com/google/go-cmp/cmp"
)

func TestNextSync(t *testing.T) {
	clock := func() time.Time { return time.Date(2020, 01, 01, 01, 01, 01, 01, time.UTC) }
	tests := []struct {
		name string
		h    campaigns.ChangesetSyncData
		want time.Time
	}{
		{
			name: "No time passed",
			h: campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock(),
			},
			want: clock().Add(minSyncDelay),
		},
		{
			name: "Linear backoff",
			h: campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-1 * time.Hour),
			},
			want: clock().Add(1 * time.Hour),
		},
		{
			name: "Use max of ExternalUpdateAt and LatestEvent",
			h: campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-2 * time.Hour),
				LatestEvent:       clock().Add(-1 * time.Hour),
			},
			want: clock().Add(1 * time.Hour),
		},
		{
			name: "Diff max is capped",
			h: campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-2 * maxSyncDelay),
			},
			want: clock().Add(maxSyncDelay),
		},
		{
			name: "Diff min is capped",
			h: campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-1 * minSyncDelay / 2),
			},
			want: clock().Add(minSyncDelay),
		},
		{
			name: "Event arrives after sync",
			h: campaigns.ChangesetSyncData{
				UpdatedAt:         clock(),
				ExternalUpdatedAt: clock().Add(-1 * maxSyncDelay / 2),
				LatestEvent:       clock().Add(10 * time.Minute),
			},
			want: clock().Add(10 * time.Minute).Add(minSyncDelay),
		},
		{
			name: "Never synced",
			h:    campaigns.ChangesetSyncData{},
			want: clock(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextSync(clock, tt.h)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestChangesetPriorityQueue(t *testing.T) {
	assertOrder := func(t *testing.T, q *changesetPriorityQueue, expected []int64) {
		t.Helper()
		ids := make([]int64, len(q.items))
		for i := range ids {
			ids[i] = q.items[i].changesetID
		}
		if diff := cmp.Diff(expected, ids); diff != "" {
			t.Fatal(diff)
		}
	}

	now := time.Now()
	q := newChangesetPriorityQueue()

	items := []scheduledSync{
		{
			changesetID: 1,
			nextSync:    now,
			priority:    priorityNormal,
		},
		{
			changesetID: 2,
			nextSync:    now,
			priority:    priorityHigh,
		},
		{
			changesetID: 3,
			nextSync:    now.Add(-1 * time.Minute),
			priority:    priorityNormal,
		},
		{
			changesetID: 4,
			nextSync:    now.Add(-2 * time.Hour),
			priority:    priorityNormal,
		},
		{
			changesetID: 5,
			nextSync:    now.Add(1 * time.Hour),
			priority:    priorityNormal,
		},
	}

	for i := range items {
		q.Upsert(items[i])
	}

	assertOrder(t, q, []int64{2, 4, 3, 1, 5})

	// Set item to high priority
	q.Upsert(scheduledSync{
		changesetID: 4,
		nextSync:    now.Add(-2 * time.Hour),
		priority:    priorityHigh,
	})

	assertOrder(t, q, []int64{4, 2, 3, 1, 5})

	// Can't reduce priority of existing item
	q.Upsert(scheduledSync{
		changesetID: 4,
		nextSync:    now.Add(-2 * time.Hour),
		priority:    priorityNormal,
	})

	if q.Len() != len(items) {
		t.Fatalf("Expected %d, got %d", q.Len(), len(items))
	}

	assertOrder(t, q, []int64{4, 2, 3, 1, 5})

	for i := 0; i < len(items); i++ {
		peeked, ok := q.Peek()
		if !ok {
			t.Fatalf("Queue should not be empty")
		}
		item := heap.Pop(q).(scheduledSync)
		if peeked.changesetID != item.changesetID {
			t.Fatalf("Peeked and Popped item should have the same id")
		}
	}

	// Len() should be zero after all items popped
	if q.Len() != 0 {
		t.Fatalf("Expected %d, got %d", q.Len(), 0)
	}
}

func TestSyncerRun(t *testing.T) {
	t.Run("Sync due", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		now := time.Now()
		store := MockSyncStore{
			listChangesetSyncData: func(ctx context.Context) ([]campaigns.ChangesetSyncData, error) {
				return []campaigns.ChangesetSyncData{
					{
						ChangesetID:       1,
						UpdatedAt:         now.Add(-2 * maxSyncDelay),
						LatestEvent:       now.Add(-2 * maxSyncDelay),
						ExternalUpdatedAt: now.Add(-2 * maxSyncDelay),
					},
				}, nil
			},
		}
		syncFunc := func(ctx context.Context, ids int64) error {
			cancel()
			return nil
		}
		syncer := &ChangesetSyncer{
			Store:                   store,
			ComputeScheduleInterval: 10 * time.Minute,
			syncFunc:                syncFunc,
		}
		go syncer.Run(ctx)
		select {
		case <-ctx.Done():
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Sync should have been triggered")
		}
	})

	t.Run("Sync not due", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		now := time.Now()
		store := MockSyncStore{
			listChangesetSyncData: func(ctx context.Context) ([]campaigns.ChangesetSyncData, error) {
				return []campaigns.ChangesetSyncData{
					{
						ChangesetID:       1,
						UpdatedAt:         now,
						LatestEvent:       now,
						ExternalUpdatedAt: now,
					},
				}, nil
			},
		}
		var syncCalled bool
		syncFunc := func(ctx context.Context, ids int64) error {
			syncCalled = true
			return nil
		}
		syncer := &ChangesetSyncer{
			Store:                   store,
			ComputeScheduleInterval: 10 * time.Minute,
			syncFunc:                syncFunc,
		}
		syncer.Run(ctx)
		if syncCalled {
			t.Fatal("Sync should not have been triggered")
		}
	})

	t.Run("Priority added", func(t *testing.T) {
		// Empty schedule but then we add an item
		ctx, cancel := context.WithCancel(context.Background())
		store := MockSyncStore{
			listChangesetSyncData: func(ctx context.Context) ([]campaigns.ChangesetSyncData, error) {
				return []campaigns.ChangesetSyncData{}, nil
			},
		}
		syncFunc := func(ctx context.Context, ids int64) error {
			cancel()
			return nil
		}
		syncer := &ChangesetSyncer{
			Store:                   store,
			ComputeScheduleInterval: 10 * time.Minute,
			syncFunc:                syncFunc,
			priorityNotify:          make(chan []int64, 1),
		}
		syncer.priorityNotify <- []int64{1}
		go syncer.Run(ctx)
		select {
		case <-ctx.Done():
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Sync not called")
		}
	})

}

type MockSyncStore struct {
	listChangesetSyncData func(context.Context) ([]campaigns.ChangesetSyncData, error)
	getChangeset          func(context.Context, GetChangesetOpts) (*campaigns.Changeset, error)
	listChangesets        func(context.Context, ListChangesetsOpts) ([]*campaigns.Changeset, int64, error)
	transact              func(context.Context) (*Store, error)
}

func (m MockSyncStore) ListChangesetSyncData(ctx context.Context) ([]campaigns.ChangesetSyncData, error) {
	return m.listChangesetSyncData(ctx)
}

func (m MockSyncStore) GetChangeset(ctx context.Context, opts GetChangesetOpts) (*campaigns.Changeset, error) {
	return m.getChangeset(ctx, opts)
}

func (m MockSyncStore) ListChangesets(ctx context.Context, opts ListChangesetsOpts) ([]*campaigns.Changeset, int64, error) {
	return m.listChangesets(ctx, opts)
}

func (m MockSyncStore) Transact(ctx context.Context) (*Store, error) {
	return m.transact(ctx)
}
