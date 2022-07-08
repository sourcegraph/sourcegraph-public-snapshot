package repos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	webhookapi "github.com/sourcegraph/sourcegraph/internal/repos/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func testSyncWebhookWorker(db database.DB) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		testRepo := &types.Repo{
			ID:       33,
			Name:     "github.com/susantoscott/Task-Tracker",
			Metadata: &github.Repository{},
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi-mom-12345",
				ServiceID:   "https://github.com/",
				ServiceType: extsvc.TypeGitHub,
			},
		}
		err := db.Repos().Create(ctx, testRepo)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("testRepo:%+v\n", testRepo)

		q1 := sqlf.Sprintf(`insert into create_webhook_jobs (repo_id, repo_name) values (%d, %s);`, testRepo.ID, testRepo.Name)
		result, err := db.Handle().ExecContext(ctx, q1.Query(sqlf.PostgresBindVar), q1.Args()...)
		if err != nil {
			t.Fatal(err)
		}
		PrintRows(db, ctx)
		rowsAffected, err := result.RowsAffected()
		if rowsAffected != 1 {
			t.Fatalf("Expected 1 row to be affected, got %d", rowsAffected)
		}

		jobChan := make(chan *repos.WhBuildJob)

		h := &fakeWebhookCreationHandler{
			jobChan: jobChan,
		}
		worker, _ := repos.NewWhBuildWorker(ctx, db.Handle(), h, repos.WhBuildOptions{
			NumHandlers:    1,
			WorkerInterval: 1 * time.Millisecond,
		})
		go worker.Start()
		// go resetter.Start()

		defer worker.Stop()
		// defer resetter.Stop()

		var job *repos.WhBuildJob
		select {
		case job = <-jobChan:
			fmt.Println("Job received")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}
		fmt.Printf("job:%+v\n", job)

	}
}

func PrintRows(db database.DB, ctx context.Context) {
	q := sqlf.Sprintf(`select * from create_webhook_jobs;`)
	rows, err := db.Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		fmt.Println("error printing rows")
	}
	var jobs []repos.WhBuildJob

	for rows.Next() {
		var job repos.WhBuildJob
		var executionLogs *[]any
		rows.Scan(
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			&executionLogs,
			&job.RepoID,
			&job.RepoName,
			&job.QueuedAt,
		)
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("err printing:", err)
	}
}

type fakeWebhookCreationHandler struct {
	jobChan chan *repos.WhBuildJob
}

func (h *fakeWebhookCreationHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	fmt.Println("in the fake Handlerr")
	cwj, ok := record.(*repos.WhBuildJob)
	if !ok {
		return errors.Errorf("expected repos.WhBuildJob, got %T", record)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case h.jobChan <- cwj:
		fmt.Println("putting creation job in channel")
		return webhookapi.CreateSyncWebhook(cwj.RepoName, "secret", "token")
	}
}
