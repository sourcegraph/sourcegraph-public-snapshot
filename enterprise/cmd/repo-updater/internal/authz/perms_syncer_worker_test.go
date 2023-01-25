package authz

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func TestPermsSyncerWorker_Handle(t *testing.T) {
	ctx := context.Background()

	dummySyncer := &dummyPermsSyncer{}

	worker := MakePermsSyncerWorker(ctx, &observation.TestContext, dummySyncer)

	t.Run("user sync request", func(t *testing.T) {
		_ = worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               99,
			UserID:           1234,
			InvalidateCaches: true,
			Priority:         database.HighPriorityPermissionSync,
		})

		wantMeta := &requestMeta{
			ID:       1234,
			Priority: priorityHigh,
			Type:     requestTypeUser,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: true,
			},
		}

		if diff := cmp.Diff(dummySyncer.request.requestMeta, wantMeta); diff != "" {
			t.Fatalf("wrong sync request: %s", diff)
		}
	})

	t.Run("repo sync request", func(t *testing.T) {
		_ = worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               777,
			RepositoryID:     4567,
			InvalidateCaches: false,
			Priority:         database.LowPriorityPermissionSync,
		})

		wantMeta := &requestMeta{
			ID:       4567,
			Priority: priorityLow,
			Type:     requestTypeRepo,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: false,
			},
		}

		if diff := cmp.Diff(dummySyncer.request.requestMeta, wantMeta); diff != "" {
			t.Fatalf("wrong sync request: %s", diff)
		}
	})
}

func TestPermsSyncerWorker_Store_Dequeue_Order(t *testing.T) {
	logger := logtest.Scoped(t)
	dbt := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, dbt)

	if _, err := dbt.ExecContext(context.Background(), `DELETE FROM permission_sync_jobs;`); err != nil {
		t.Fatalf("unexpected error deleting records: %s", err)
	}

	if _, err := dbt.ExecContext(context.Background(), `
		INSERT INTO permission_sync_jobs (id, state, user_id, repository_id, priority, process_after, reason)
		VALUES
			(1, 'queued', 1, null, 0, null, 'test'),
			(2, 'queued', 0, null, 0, null, 'test'),
			(3, 'queued', 1, null, 5, null, 'test'),
			(4, 'queued', 0, null, 5, null, 'test'),
			(5, 'queued', 1, null, 10, null, 'test'),
			(6, 'queued', 0, null, 10, null, 'test'),
			(7, 'queued', 1, null, 10, NOW() - '1 minute'::interval, 'test'),
			(8, 'queued', 0, null, 10, NOW() - '2 minute'::interval, 'test'),
			(9, 'queued', 1, null, 5, NOW() - '1 minute'::interval, 'test'),
			(10, 'queued', 0, null, 5, NOW() - '2 minute'::interval, 'test'),
			(11, 'queued', 1, null, 0, NOW() - '1 minute'::interval, 'test'),
			(12, 'queued', 0, null, 0, NOW() - '2 minute'::interval, 'test'),
			(13, 'processing', 1, null, 10, null, 'test'),
			(14, 'completed', 0, null, 10, null, 'test'),
			(15, 'cancelled', 1, null, 10, null, 'test'),
			(16, 'queued', 1, null, 10, NOW() + '2 minute'::interval, 'test')
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	store := MakeStore(&observation.TestContext, db.Handle())
	jobIDs := []int{}
	wantJobIDs := []int{5, 6, 8, 7, 3, 4, 10, 9, 1, 2, 12, 11, 0, 0, 0, 0}
	var dequeueErr error
	for range wantJobIDs {
		record, _, err := store.Dequeue(context.Background(), "test", nil)
		if err == nil {
			if record == nil {
				jobIDs = append(jobIDs, 0)
			} else {
				jobIDs = append(jobIDs, record.ID)
			}
		} else {
			dequeueErr = err
		}
	}

	if dequeueErr != nil {
		t.Fatalf("dequeue operation failed: %s", dequeueErr)
	}

	if diff := cmp.Diff(jobIDs, wantJobIDs); diff != "" {
		t.Fatalf("jobs dequeued in wrong order: %s", diff)
	}
}

type dummyPermsSyncer struct {
	request *syncRequest
}

func (d *dummyPermsSyncer) syncPerms(_ context.Context, _ map[requestType]group.ContextGroup, request *syncRequest) {
	d.request = request
}
