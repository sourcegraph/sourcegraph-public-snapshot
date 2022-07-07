package repos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func PrintRowsSync(repoStore repos.Store, ctx context.Context) {
	q := sqlf.Sprintf(`select * from external_service_sync_jobs;`)
	rows, err := repoStore.Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		fmt.Println("error printing rows")
	}
	var res []repos.SyncJob

	for rows.Next() {
		var dest repos.SyncJob
		rows.Scan(dest)
		res = append(res, dest)
	}

	fmt.Println("len:", len(res))
	for _, r := range res {
		fmt.Printf("r:%+v\n", r)
	}
}

func testSyncWorkerPlumbing(repoStore repos.Store) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		testSvc := &types.ExternalService{
			ID:          77,
			Kind:        extsvc.KindGitHub,
			DisplayName: "TestService",
			Config:      "{}",
		}
		PrintRowsSync(repoStore, ctx)

		// Create external service
		err := repoStore.ExternalServiceStore().Upsert(ctx, testSvc)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Test service created, ID: %d", testSvc.ID)

		fmt.Println("id:", testSvc.ID)
		// Add item to queue
		q1 := sqlf.Sprintf(`insert into external_service_sync_jobs (external_service_id) values (%s);`, testSvc.ID)
		result, err := repoStore.Handle().ExecContext(ctx, q1.Query(sqlf.PostgresBindVar), q1.Args()...)
		if err != nil {
			t.Fatal(err)
		}
		PrintRowsSync(repoStore, ctx)
		q2 := sqlf.Sprintf(`insert into external_service_sync_jobs (external_service_id) values (%s);`, 77)
		_, err = repoStore.Handle().ExecContext(ctx, q2.Query(sqlf.PostgresBindVar), q2.Args()...)
		if err != nil {
			t.Fatal(err)
		}
		PrintRowsSync(repoStore, ctx)

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
		worker, _ := repos.NewSyncWorker(ctx, repoStore.Handle(), h, repos.SyncWorkerOptions{
			NumHandlers:    1,
			WorkerInterval: 1 * time.Millisecond,
		})
		go worker.Start()
		// go resetter.Start()

		// There is a race between the worker being stopped and the worker util
		// finalising the row which means that when running tests in verbose mode we'll
		// see "sql: transaction has already been committed or rolled back". These
		// errors can be ignored.
		defer worker.Stop()
		// defer resetter.Stop()

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
}

type fakeRepoSyncHandler struct {
	jobChan chan *repos.SyncJob
}

func (h *fakeRepoSyncHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	fmt.Println("in the fake Handle")
	sj, ok := record.(*repos.SyncJob)
	fmt.Printf("sj:%+v\n", sj)
	if !ok {
		return errors.Errorf("expected repos.SyncJob, got %T", record)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case h.jobChan <- sj:
		fmt.Println("putting sync job in channel")
		return nil
	}
}
