package repos_test

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSyncWorkerPlumbing(t *testing.T) {
	t.Parallel()
	store := getTestRepoStore(t)

	ctx := context.Background()
	testSvc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "TestService",
		Config:      extsvc.NewUnencryptedConfig(basicGitHubConfig),
	}

	// Create external service
	err := store.ExternalServiceStore().Upsert(ctx, testSvc)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Test service created, ID: %d", testSvc.ID)

	// Add item to queue
	q := sqlf.Sprintf(`insert into external_service_sync_jobs (external_service_id) values (%s);`, testSvc.ID)
	result, err := store.Handle().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	if rowsAffected != 1 {
		t.Fatalf("Expected 1 row to be affected, got %d", rowsAffected)
	}

	jobChan := make(chan *repos.SyncJob)

	h := &fakeRepoSyncHandler{
		jobChan: jobChan,
	}
	worker, resetter, janitor := repos.NewSyncWorker(ctx, observation.TestContextTB(t), store.Handle(), h, repos.SyncWorkerOptions{
		NumHandlers:    1,
		WorkerInterval: 1 * time.Millisecond,
	})
	go worker.Start()
	go resetter.Start()
	go janitor.Start()

	// There is a race between the worker being stopped and the worker util
	// finalising the row which means that when running tests in verbose mode we'll
	// see "sql: transaction has already been committed or rolled back". These
	// errors can be ignored.
	t.Cleanup(func() {
		err := janitor.Stop(ctx)
		require.NoError(t, err)
		err = resetter.Stop(ctx)
		require.NoError(t, err)
		err = worker.Stop(ctx)
		require.NoError(t, err)
	})

	var job *repos.SyncJob
	select {
	case job = <-jobChan:
		t.Log("Job received")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout")
	}

	if job.ExternalServiceID != testSvc.ID {
		t.Fatalf("Expected %d, got %d", testSvc.ID, job.ExternalServiceID)
	}
}

type fakeRepoSyncHandler struct {
	jobChan chan *repos.SyncJob
}

func (h *fakeRepoSyncHandler) Handle(ctx context.Context, logger log.Logger, sj *repos.SyncJob) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case h.jobChan <- sj:
		return nil
	}
}
