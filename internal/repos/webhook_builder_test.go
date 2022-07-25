package repos_test

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var updateWebhooks = flag.Bool("updateWebhooks", false, "update testdata for webhook build worker integration test")

func testWebhookBuilderPlumbing(store repos.Store) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		ghRepo := &types.Repo{
			ID:   1,
			Name: "github.com/susantoscott/hi-mom",
		}

		err := store.RepoStore().Create(ctx, ghRepo)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("GitHub repo created, name: %s", ghRepo.Name)

		q := sqlf.Sprintf(`insert into webhook_build_jobs (repo_id, repo_name, queued_at, extsvc_kind) values (%s, %s, %s, %s);`,
			ghRepo.ID, ghRepo.Name, timeutil.Now(), "GITHUB")
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

		jobChan := make(chan *repos.WebhookBuildJob)
		handler := &fakeWebhookBuildJobHandler{
			jobChan: jobChan,
		}

		worker, resetter := repos.NewWebhookBuildWorker(ctx, store.Handle(), handler, repos.WebhookBuildOptions{
			NumHandlers:    1,
			WorkerInterval: 1 * time.Millisecond,
		})
		go worker.Start()
		go resetter.Start()

		defer worker.Stop()
		defer resetter.Stop()

		var job *repos.WebhookBuildJob
		select {
		case job = <-jobChan:
			t.Log("Job received")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		if job.RepoName != string(ghRepo.Name) {
			t.Fatalf("Expected %s, got %s", ghRepo.Name, job.RepoName)
		}
	}
}

type fakeWebhookBuildJobHandler struct {
	jobChan chan *repos.WebhookBuildJob
}

func (h *fakeWebhookBuildJobHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	wbj, ok := record.(*repos.WebhookBuildJob)
	if !ok {
		return errors.Errorf("expected repos.WebhookBuildJob, got %T", record)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case h.jobChan <- wbj:
		return nil
	}
}
