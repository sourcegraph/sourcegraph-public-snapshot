package authz

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestPermsSyncerWorker_Handle(t *testing.T) {
	ctx := context.Background()

	dummySyncer := &dummyPermsSyncer{}

	worker := MakePermsSyncerWorker(&observation.TestContext, dummySyncer)

	t.Run("user sync request", func(t *testing.T) {
		_ = worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               99,
			UserID:           1234,
			InvalidateCaches: true,
			Priority:         database.HighPriorityPermissionSync,
		})

		wantRequest := combinedRequest{
			UserID:  1234,
			NoPerms: false,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: true,
			},
		}
		if diff := cmp.Diff(dummySyncer.request, wantRequest); diff != "" {
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

		wantRequest := combinedRequest{
			RepoID:  4567,
			NoPerms: false,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: false,
			},
		}
		if diff := cmp.Diff(dummySyncer.request, wantRequest); diff != "" {
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
		INSERT INTO users (id, username)
		VALUES (1, 'test_user_1')
	`); err != nil {
		t.Fatalf("unexpected error creating user: %s", err)
	}

	if _, err := dbt.ExecContext(context.Background(), `
		INSERT INTO repo (id, name)
		VALUES (1, 'test_repo_1')
	`); err != nil {
		t.Fatalf("unexpected error creating repo: %s", err)
	}

	if _, err := dbt.ExecContext(context.Background(), `
		INSERT INTO permission_sync_jobs (id, state, user_id, repository_id, priority, process_after, reason)
		VALUES
			(1, 'queued', 1, null, 0, null, 'test'),
			(2, 'queued', null, 1, 0, null, 'test'),
			(3, 'queued', 1, null, 5, null, 'test'),
			(4, 'queued', null, 1, 5, null, 'test'),
			(5, 'queued', 1, null, 10, null, 'test'),
			(6, 'queued', null, 1, 10, null, 'test'),
			(7, 'queued', 1, null, 10, NOW() - '1 minute'::interval, 'test'),
			(8, 'queued', null, 1, 10, NOW() - '2 minute'::interval, 'test'),
			(9, 'queued', 1, null, 5, NOW() - '1 minute'::interval, 'test'),
			(10, 'queued', null, 1, 5, NOW() - '2 minute'::interval, 'test'),
			(11, 'queued', 1, null, 0, NOW() - '1 minute'::interval, 'test'),
			(12, 'queued', null, 1, 0, NOW() - '2 minute'::interval, 'test'),
			(13, 'processing', 1, null, 10, null, 'test'),
			(14, 'completed', null, 1, 10, null, 'test'),
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

// combinedRequest is a test entity which contains properties of both user and
// repo perms sync requests.
type combinedRequest struct {
	RepoID  api.RepoID
	UserID  int32
	NoPerms bool
	Options authz.FetchPermsOptions
}

type dummyPermsSyncer struct {
	request combinedRequest
}

func (d *dummyPermsSyncer) syncRepoPerms(_ context.Context, repoID api.RepoID, noPerms bool, options authz.FetchPermsOptions) ([]syncjobs.ProviderStatus, error) {
	d.request = combinedRequest{
		RepoID:  repoID,
		NoPerms: noPerms,
		Options: options,
	}
	return []syncjobs.ProviderStatus{}, nil
}
func (d *dummyPermsSyncer) syncUserPerms(_ context.Context, userID int32, noPerms bool, options authz.FetchPermsOptions) ([]syncjobs.ProviderStatus, error) {
	d.request = combinedRequest{
		UserID:  userID,
		NoPerms: noPerms,
		Options: options,
	}
	return []syncjobs.ProviderStatus{}, nil
}
