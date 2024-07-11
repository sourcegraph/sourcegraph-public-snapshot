package authz

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

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
)

const errorMsg = "Sorry, wrong number."
const allProvidersFailedMsg = "All providers failed to sync permissions."

func TestPermsSyncerWorker_Handle(t *testing.T) {
	ctx := context.Background()
	dummySyncer := &dummyPermsSyncer{}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	syncJobsStore := db.PermissionSyncJobs()

	t.Run("user sync request", func(t *testing.T) {
		worker := makePermsSyncerWorker(observation.TestContextTB(t), dummySyncer, syncTypeUser, syncJobsStore)
		_ = worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               99,
			UserID:           1234,
			InvalidateCaches: true,
			Priority:         database.HighPriorityPermissionsSync,
			NoPerms:          true,
		})

		wantRequest := combinedRequest{
			UserID:  1234,
			NoPerms: true,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: true,
			},
		}
		if diff := cmp.Diff(dummySyncer.request, wantRequest); diff != "" {
			t.Fatalf("wrong sync request: %s", diff)
		}
	})

	t.Run("repo sync request", func(t *testing.T) {
		worker := makePermsSyncerWorker(observation.TestContextTB(t), dummySyncer, syncTypeRepo, syncJobsStore)
		_ = worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               777,
			RepositoryID:     4567,
			InvalidateCaches: false,
			Priority:         database.LowPriorityPermissionsSync,
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
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating users and repos.
	userStore := db.Users()
	user1, err := userStore.Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := userStore.Create(ctx, database.NewUser{Username: "user2"})
	require.NoError(t, err)
	repoStore := db.Repos()
	err = repoStore.Create(ctx, &types.Repo{Name: "github.com/soucegraph/sourcegraph"}, &types.Repo{Name: "github.com/soucegraph/about"}, &types.Repo{Name: "github.com/soucegraph/hello"})
	require.NoError(t, err)

	// Creating a worker.
	observationCtx := observation.TestContextTB(t)
	dummySyncer := &dummySyncerWithErrors{
		repoIDErrors: map[api.RepoID]errorType{2: allProvidersFailed, 3: realError},
	}

	syncJobsStore := db.PermissionSyncJobs()
	workerStore := makeStore(observationCtx, db.Handle(), syncTypeRepo)
	worker := MakeTestWorker(ctx, observationCtx, workerStore, dummySyncer, syncTypeRepo, syncJobsStore)
	go worker.Start()
	t.Cleanup(func() {
		err := worker.Stop(ctx)
		require.NoError(t, err)
	})

	// Adding repo perms sync jobs.
	err = syncJobsStore.CreateRepoSyncJob(ctx, api.RepoID(1), database.PermissionSyncJobOpts{Reason: database.ReasonManualRepoSync, Priority: database.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	err = syncJobsStore.CreateRepoSyncJob(ctx, api.RepoID(2), database.PermissionSyncJobOpts{Reason: database.ReasonManualRepoSync, Priority: database.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	err = syncJobsStore.CreateRepoSyncJob(ctx, api.RepoID(3), database.PermissionSyncJobOpts{Reason: database.ReasonManualRepoSync, Priority: database.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding user perms sync job, which should not be processed by current worker!
	err = syncJobsStore.CreateUserSyncJob(ctx, user2.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonRepoNoPermissions, Priority: database.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
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
			// We don't check job with ID=4 because it is a user sync job which is not
			// processed by current worker.
			if job.ID != 4 && (job.State == database.PermissionsSyncJobStateQueued || job.State == database.PermissionsSyncJobStateProcessing) {
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
			if job.ID == 4 && remainingRounds > 0 {
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

		// Check that repo sync job was completed and results were saved.
		if jobID == 1 {
			require.Equal(t, database.PermissionsSyncJobStateCompleted, job.State)
			require.Nil(t, job.FailureMessage)
			require.Equal(t, 1, job.PermissionsAdded)
			require.Equal(t, 2, job.PermissionsRemoved)
			require.Equal(t, 5, job.PermissionsFound)
			require.False(t, job.IsPartialSuccess)
		}

		// Check that repo sync job has the failure message.
		if jobID == 2 {
			require.NotNil(t, job.FailureMessage)
			require.Equal(t, allProvidersFailedMsg, *job.FailureMessage)
			require.Equal(t, 1, job.NumFailures)
			require.Equal(t, 0, job.PermissionsAdded)
			require.Equal(t, 0, job.PermissionsRemoved)
			require.Equal(t, 0, job.PermissionsFound)
		}

		// Check that failed job has the failure message.
		if jobID == 3 {
			require.NotNil(t, job.FailureMessage)
			require.Equal(t, errorMsg, *job.FailureMessage)
			require.Equal(t, 1, job.NumFailures)
			require.Equal(t, 0, job.PermissionsAdded)
			require.Equal(t, 0, job.PermissionsRemoved)
			require.Equal(t, 0, job.PermissionsFound)
		}

		// Check that user sync job wasn't picked up by repo sync worker.
		if jobID == 4 {
			require.Equal(t, database.PermissionsSyncJobStateQueued, job.State)
			require.Nil(t, job.FailureMessage)
			require.Equal(t, 0, job.PermissionsAdded)
			require.Equal(t, 0, job.PermissionsRemoved)
			require.Equal(t, 0, job.PermissionsFound)
		}
	}
}

func TestPermsSyncerWorker_UserSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	// Creating users and repos.
	userStore := db.Users()
	user1, err := userStore.Create(ctx, database.NewUser{Username: "user1"})
	require.NoError(t, err)
	user2, err := userStore.Create(ctx, database.NewUser{Username: "user2"})
	require.NoError(t, err)
	user3, err := userStore.Create(ctx, database.NewUser{Username: "user3"})
	require.NoError(t, err)
	user4, err := userStore.Create(ctx, database.NewUser{Username: "user4"})
	require.NoError(t, err)
	repoStore := db.Repos()
	err = repoStore.Create(ctx, &types.Repo{Name: "github.com/soucegraph/sourcegraph"}, &types.Repo{Name: "github.com/soucegraph/about"})
	require.NoError(t, err)

	// Creating a worker.
	observationCtx := observation.TestContextTB(t)
	dummySyncer := &dummySyncerWithErrors{
		userIDErrors:      map[int32]errorType{2: allProvidersFailed, 3: realError},
		userIDNoProviders: map[int32]struct{}{4: {}},
	}

	syncJobsStore := db.PermissionSyncJobs()
	workerStore := makeStore(observationCtx, db.Handle(), syncTypeUser)
	worker := MakeTestWorker(ctx, observationCtx, workerStore, dummySyncer, syncTypeUser, syncJobsStore)
	go worker.Start()
	t.Cleanup(func() {
		err := worker.Stop(ctx)
		require.NoError(t, err)
	})

	// Adding user perms sync jobs.
	err = syncJobsStore.CreateUserSyncJob(ctx, user1.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonUserOutdatedPermissions, Priority: database.LowPriorityPermissionsSync})
	require.NoError(t, err)

	err = syncJobsStore.CreateUserSyncJob(ctx, user2.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonRepoNoPermissions, NoPerms: true, Priority: database.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	err = syncJobsStore.CreateUserSyncJob(ctx, user3.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonRepoNoPermissions, NoPerms: true, Priority: database.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding user perms sync job without perms providers synced.
	err = syncJobsStore.CreateUserSyncJob(ctx, user4.ID,
		database.PermissionSyncJobOpts{Reason: database.ReasonRepoNoPermissions, NoPerms: true, Priority: database.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding repo perms sync job, which should not be processed by current worker!
	err = syncJobsStore.CreateRepoSyncJob(ctx, api.RepoID(1), database.PermissionSyncJobOpts{Reason: database.ReasonManualRepoSync, Priority: database.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
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
			// We don't check job with ID=5 because it is a repo sync job which is not
			// processed by current worker.
			if job.ID != 5 && (job.State == database.PermissionsSyncJobStateQueued || job.State == database.PermissionsSyncJobStateProcessing) {
				// wait and retry
				time.Sleep(500 * time.Millisecond)
				continue loop
			}
		}

		// Adding additional 3 rounds of checks to make sure that we've waited enough
		// time to get a chance for repo sync job to be processed (by mistake).
		for _, job := range jobs {
			// We only check job with ID=5 because it is a repo sync job which should not
			// processed by current worker.
			if job.ID == 5 && remainingRounds > 0 {
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

		// Check that user sync job was completed and results were saved.
		if jobID == 1 {
			require.Equal(t, database.PermissionsSyncJobStateCompleted, job.State)
			require.Nil(t, job.FailureMessage)
			require.Equal(t, 1, job.PermissionsAdded)
			require.Equal(t, 2, job.PermissionsRemoved)
			require.Equal(t, 5, job.PermissionsFound)
			require.True(t, job.IsPartialSuccess)
		}

		// Check that failed job has the failure message.
		if jobID == 2 {
			require.NotNil(t, job.FailureMessage)
			require.Equal(t, allProvidersFailedMsg, *job.FailureMessage)
			require.Equal(t, 1, job.NumFailures)
			require.True(t, job.NoPerms)
			require.Equal(t, 0, job.PermissionsAdded)
			require.Equal(t, 0, job.PermissionsRemoved)
			require.Equal(t, 0, job.PermissionsFound)
		}

		// Check that failed job has the failure message.
		if jobID == 3 {
			require.NotNil(t, job.FailureMessage)
			require.Equal(t, errorMsg, *job.FailureMessage)
			require.Equal(t, 1, job.NumFailures)
			require.True(t, job.NoPerms)
			require.Equal(t, 0, job.PermissionsAdded)
			require.Equal(t, 0, job.PermissionsRemoved)
			require.Equal(t, 0, job.PermissionsFound)
		}

		// Check that user sync job was completed and results were saved even though
		// there weren't any perms providers.
		if jobID == 4 {
			require.Equal(t, database.PermissionsSyncJobStateCompleted, job.State)
			require.Nil(t, job.FailureMessage)
			require.Equal(t, 0, job.PermissionsAdded)
			require.Equal(t, 0, job.PermissionsRemoved)
			require.Equal(t, 0, job.PermissionsFound)
			require.False(t, job.IsPartialSuccess)
		}

		// Check that repo sync job wasn't picked up by user sync worker.
		if jobID == 5 {
			require.Equal(t, database.PermissionsSyncJobStateQueued, job.State)
			require.Nil(t, job.FailureMessage)
			require.Equal(t, 0, job.PermissionsAdded)
			require.Equal(t, 0, job.PermissionsRemoved)
			require.Equal(t, 0, job.PermissionsFound)
		}
	}
}

func TestPermsSyncerWorker_Store_Dequeue_Order(t *testing.T) {
	logger := logtest.Scoped(t)
	dbt := dbtest.NewDB(t)
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

	store := makeStore(observation.TestContextTB(t), db.Handle(), syncTypeRepo)
	jobIDs := make([]int, 0)
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

func MakeTestWorker(ctx context.Context, observationCtx *observation.Context, workerStore dbworkerstore.Store[*database.PermissionSyncJob], permsSyncer permsSyncer, typ syncType, jobsStore database.PermissionSyncJobStore) *workerutil.Worker[*database.PermissionSyncJob] {
	handler := makePermsSyncerWorker(observationCtx, permsSyncer, typ, jobsStore)
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
	sync.Mutex
	request combinedRequest
}

func (d *dummyPermsSyncer) syncRepoPerms(_ context.Context, repoID api.RepoID, noPerms bool, options authz.FetchPermsOptions) (*database.SetPermissionsResult, database.CodeHostStatusesSet, error) {
	d.Lock()
	defer d.Unlock()

	d.request = combinedRequest{
		RepoID:  repoID,
		NoPerms: noPerms,
		Options: options,
	}
	return &database.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}, database.CodeHostStatusesSet{}, nil
}
func (d *dummyPermsSyncer) syncUserPerms(_ context.Context, userID int32, noPerms bool, options authz.FetchPermsOptions) (*database.SetPermissionsResult, database.CodeHostStatusesSet, error) {
	d.Lock()
	defer d.Unlock()

	d.request = combinedRequest{
		UserID:  userID,
		NoPerms: noPerms,
		Options: options,
	}
	return &database.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}, database.CodeHostStatusesSet{}, nil
}

type errorType string

const (
	realError          errorType = "REAL_ERROR"
	allProvidersFailed errorType = "ALL_PROVIDERS_FAILED"
)

type dummySyncerWithErrors struct {
	sync.Mutex
	request           combinedRequest
	userIDErrors      map[int32]errorType
	repoIDErrors      map[api.RepoID]errorType
	userIDNoProviders map[int32]struct{}
}

func (d *dummySyncerWithErrors) syncRepoPerms(_ context.Context, repoID api.RepoID, noPerms bool, options authz.FetchPermsOptions) (*database.SetPermissionsResult, database.CodeHostStatusesSet, error) {
	d.Lock()
	defer d.Unlock()

	if errorTyp, ok := d.repoIDErrors[repoID]; ok && errorTyp == realError {
		return nil, nil, errors.New(errorMsg)
	}
	d.request = combinedRequest{
		RepoID:  repoID,
		NoPerms: noPerms,
		Options: options,
	}

	codeHostStates := database.CodeHostStatusesSet{{ProviderID: "id1", Status: database.CodeHostStatusSuccess}, {ProviderID: "id2", Status: database.CodeHostStatusSuccess}}
	result := database.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}
	if typ, ok := d.repoIDErrors[repoID]; ok && typ == allProvidersFailed {
		for idx := range codeHostStates {
			codeHostStates[idx].Status = database.CodeHostStatusError
		}
		result = database.SetPermissionsResult{}
	}

	return &result, codeHostStates, nil
}
func (d *dummySyncerWithErrors) syncUserPerms(_ context.Context, userID int32, noPerms bool, options authz.FetchPermsOptions) (*database.SetPermissionsResult, database.CodeHostStatusesSet, error) {
	d.Lock()
	defer d.Unlock()

	if errorTyp, ok := d.userIDErrors[userID]; ok && errorTyp == realError {
		return nil, nil, errors.New(errorMsg)
	}
	d.request = combinedRequest{
		UserID:  userID,
		NoPerms: noPerms,
		Options: options,
	}

	codeHostStates := database.CodeHostStatusesSet{{ProviderID: "id1", Status: database.CodeHostStatusError}, {ProviderID: "id2", Status: database.CodeHostStatusSuccess}}
	result := database.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}
	if typ, ok := d.userIDErrors[userID]; ok && typ == allProvidersFailed {
		for idx := range codeHostStates {
			codeHostStates[idx].Status = database.CodeHostStatusError
		}
		return &database.SetPermissionsResult{}, codeHostStates, nil
	}

	if _, ok := d.userIDNoProviders[userID]; ok {
		codeHostStates = database.CodeHostStatusesSet{}
		result = database.SetPermissionsResult{}
	}

	return &result, codeHostStates, nil
}
