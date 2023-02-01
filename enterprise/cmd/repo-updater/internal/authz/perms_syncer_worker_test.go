package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

const errorMsg = "Sorry, wrong number."

func TestPermsSyncerWorker_Handle(t *testing.T) {
	ctx := context.Background()
	dummySyncer := &dummyPermsSyncer{}

	t.Run("user sync request", func(t *testing.T) {
		worker := MakePermsSyncerWorker(&observation.TestContext, dummySyncer, SyncTypeUser)
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
		worker := MakePermsSyncerWorker(&observation.TestContext, dummySyncer, SyncTypeRepo)
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

func TestPermsSyncerWorker_RepoSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating users and repos.
	userStore := db.Users()
	user1, err := userStore.Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := userStore.Create(ctx, database.NewUser{Username: "user2"})
	require.NoError(t, err)
	repoStore := db.Repos()
	err = repoStore.Create(ctx, &types.Repo{Name: "github.com/soucegraph/sourcegraph"}, &types.Repo{Name: "github.com/soucegraph/about"})
	require.NoError(t, err)

	// Creating a worker.
	observationCtx := &observation.TestContext
	dummySyncer := &dummySyncerWithErrors{
		repoIDErrors: map[api.RepoID]struct{}{2: {}},
	}

	workerStore := MakeStore(observationCtx, db.Handle(), SyncTypeRepo)
	worker := MakeTestWorker(ctx, observationCtx, workerStore, dummySyncer, SyncTypeRepo)
	go worker.Start()
	t.Cleanup(worker.Stop)

	// Adding repo perms sync jobs.
	syncJobsStore := db.PermissionSyncJobs()
	err = syncJobsStore.CreateRepoSyncJob(ctx, api.RepoID(1), database.PermissionSyncJobOpts{Reason: database.ReasonManualRepoSync, Priority: database.MediumPriorityPermissionSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	err = syncJobsStore.CreateRepoSyncJob(ctx, api.RepoID(2), database.PermissionSyncJobOpts{Reason: database.ReasonManualRepoSync, Priority: database.MediumPriorityPermissionSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding user perms sync job, which should not be processed by current worker!
	err = syncJobsStore.CreateUserSyncJob(ctx, user2.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonRepoNoPermissions, Priority: database.HighPriorityPermissionSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Wait for all jobs to be processed.
	timeout := time.After(60 * time.Second)
	remainingRounds := 3
loop:
	for {
		jobs, err := syncJobsStore.List(ctx, database.ListPermissionSyncJobOpts{})
		if err != nil {
			t.Fatal(err)
		}
		for _, job := range jobs {
			// We don't check job with ID=3 because it is a user sync job which is not
			// processed by current worker.
			if job.ID != 3 && (job.State == "queued" || job.State == "processing") {
				// wait and retry
				time.Sleep(500 * time.Millisecond)
				continue loop
			}
		}

		// Adding additional 3 rounds of checks to make sure that we've waited enough
		// time to get a chance for user sync job to be processed (by mistake).
		for _, job := range jobs {
			// We only check job with ID=3 because it is a user sync job which should not
			// processed by current worker.
			if job.ID == 3 && remainingRounds > 0 {
				// wait and retry
				time.Sleep(500 * time.Millisecond)
				remainingRounds = remainingRounds - 1
				continue loop
			}
		}

		select {
		case <-timeout:
			t.Fatal("Perms sync jobs are not processing or processing takes too much time.")
		default:
			break loop
		}
	}

	jobs, err := syncJobsStore.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	for _, job := range jobs {
		jobID := job.ID

		// Check that repo IDs are correctly assigned.
		if job.RepositoryID > 0 {
			require.Equal(t, jobID, job.RepositoryID)
		}

		// Check that failed job has the failure message.
		if jobID == 2 {
			require.NotNil(t, job.FailureMessage)
			require.Equal(t, errorMsg, *job.FailureMessage)
			require.Equal(t, 1, job.NumFailures)
		}

		// Check that user sync job wasn't picked up by repo sync worker.
		if jobID == 3 {
			require.Equal(t, "queued", job.State)
			require.Nil(t, job.FailureMessage)
		}
	}
}

func TestPermsSyncerWorker_UserSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// Creating users and repos.
	userStore := db.Users()
	user1, err := userStore.Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := userStore.Create(ctx, database.NewUser{Username: "user2"})
	require.NoError(t, err)
	repoStore := db.Repos()
	err = repoStore.Create(ctx, &types.Repo{Name: "github.com/soucegraph/sourcegraph"}, &types.Repo{Name: "github.com/soucegraph/about"})
	require.NoError(t, err)

	// Creating a worker.
	observationCtx := &observation.TestContext
	dummySyncer := &dummySyncerWithErrors{
		userIDErrors: map[int32]struct{}{2: {}},
	}

	workerStore := MakeStore(observationCtx, db.Handle(), SyncTypeUser)
	worker := MakeTestWorker(ctx, observationCtx, workerStore, dummySyncer, SyncTypeUser)
	go worker.Start()
	t.Cleanup(worker.Stop)

	// Adding user perms sync jobs.
	syncJobsStore := db.PermissionSyncJobs()
	err = syncJobsStore.CreateUserSyncJob(ctx, user1.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonUserOutdatedPermissions, Priority: database.LowPriorityPermissionSync})
	require.NoError(t, err)

	err = syncJobsStore.CreateUserSyncJob(ctx, user2.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonRepoNoPermissions, Priority: database.HighPriorityPermissionSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding repo perms sync job, which should not be processed by current worker!
	err = syncJobsStore.CreateRepoSyncJob(ctx, api.RepoID(1), database.PermissionSyncJobOpts{Reason: database.ReasonManualRepoSync, Priority: database.MediumPriorityPermissionSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Wait for all jobs to be processed.
	timeout := time.After(60 * time.Second)
	remainingRounds := 3
loop:
	for {
		jobs, err := syncJobsStore.List(ctx, database.ListPermissionSyncJobOpts{})
		if err != nil {
			t.Fatal(err)
		}
		for _, job := range jobs {
			// We don't check job with ID=3 because it is a repo sync job which is not
			// processed by current worker.
			if job.ID != 3 && (job.State == "queued" || job.State == "processing") {
				// wait and retry
				time.Sleep(500 * time.Millisecond)
				continue loop
			}
		}

		// Adding additional 3 rounds of checks to make sure that we've waited enough
		// time to get a chance for repo sync job to be processed (by mistake).
		for _, job := range jobs {
			// We only check job with ID=3 because it is a repo sync job which should not
			// processed by current worker.
			if job.ID == 3 && remainingRounds > 0 {
				// wait and retry
				time.Sleep(500 * time.Millisecond)
				remainingRounds = remainingRounds - 1
				continue loop
			}
		}

		select {
		case <-timeout:
			t.Fatal("Perms sync jobs are not processing or processing takes too much time.")
		default:
			break loop
		}
	}

	jobs, err := syncJobsStore.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	for _, job := range jobs {
		jobID := job.ID

		// Check that user IDs are correctly assigned.
		if job.UserID > 0 {
			require.Equal(t, jobID, job.UserID)
		}

		// Check that failed job has the failure message.
		if jobID == 2 {
			require.NotNil(t, job.FailureMessage)
			require.Equal(t, errorMsg, *job.FailureMessage)
			require.Equal(t, 1, job.NumFailures)
		}

		// Check that repo sync job wasn't picked up by user sync worker.
		if jobID == 3 {
			require.Equal(t, "queued", job.State)
			require.Nil(t, job.FailureMessage)
		}
	}
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

	store := MakeStore(&observation.TestContext, db.Handle(), SyncTypeRepo)
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

func MakeTestWorker(ctx context.Context, observationCtx *observation.Context, workerStore dbworkerstore.Store[*database.PermissionSyncJob], permsSyncer permsSyncer, typ syncType) *workerutil.Worker[*database.PermissionSyncJob] {
	handler := MakePermsSyncerWorker(observationCtx, permsSyncer, typ)
	return dbworker.NewWorker[*database.PermissionSyncJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "permission_sync_job_worker",
		Interval:          time.Second,
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "permission_sync_job_worker"),
		NumHandlers:       4,
	})
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

type dummySyncerWithErrors struct {
	request      combinedRequest
	userIDErrors map[int32]struct{}
	repoIDErrors map[api.RepoID]struct{}
}

func (d *dummySyncerWithErrors) syncRepoPerms(_ context.Context, repoID api.RepoID, noPerms bool, options authz.FetchPermsOptions) ([]syncjobs.ProviderStatus, error) {
	if _, ok := d.repoIDErrors[repoID]; ok {
		return nil, errors.New(errorMsg)
	}
	d.request = combinedRequest{
		RepoID:  repoID,
		NoPerms: noPerms,
		Options: options,
	}
	return []syncjobs.ProviderStatus{}, nil
}
func (d *dummySyncerWithErrors) syncUserPerms(_ context.Context, userID int32, noPerms bool, options authz.FetchPermsOptions) ([]syncjobs.ProviderStatus, error) {
	if _, ok := d.userIDErrors[userID]; ok {
		return nil, errors.New(errorMsg)
	}
	d.request = combinedRequest{
		UserID:  userID,
		NoPerms: noPerms,
		Options: options,
	}
	return []syncjobs.ProviderStatus{}, nil
}
