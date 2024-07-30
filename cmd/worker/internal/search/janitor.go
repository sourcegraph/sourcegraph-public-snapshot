package search

import (
	"context"
	"io"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/conc"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newJanitorJob(observationCtx *observation.Context, db database.DB, svc *service.Service) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		return runJanitor(ctx, db, svc)
	})

	operation := observationCtx.Operation(observation.Op{
		Name: "search.jobs.janitor",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"search_jobs_janitor",
			metrics.WithCountHelp("Total number of search_jobs_janitor executions"),
		),
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("search_jobs_janitor"),
		goroutine.WithDescription("refresh analytics cache"),
		goroutine.WithInterval(8*time.Hour),
		goroutine.WithOperation(operation),
	)
}

func runJanitor(ctx context.Context, db database.DB, svc *service.Service) error {
	jobs, err := listSearchJobs(ctx, db)
	if err != nil {
		return err
	}

	var errs error
	for _, job := range jobs {
		// Use the initiator as the actor for this operation
		ctx = actor.WithActor(ctx, actor.FromUser(job.Initiator))

		aggStatus, err := getAggregateStatus(ctx, svc, job.ID)
		if err != nil {
			return err
		}
		if aggStatus.IsTerminal() {
			err = db.WithTransact(ctx, func(tx database.DB) error {
				if err := updateSearchJobStatus(ctx, tx, job.ID, aggStatus); err != nil {
					return err
				}

				if err := uploadLogsToBlobstore(ctx, svc, job.ID); err != nil {
					return err
				}

				if err := deleteRepoJobs(ctx, tx, job.ID); err != nil {
					return err
				}

				if err := setJobAsAggregated(ctx, tx, job.ID); err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				errs = errors.Append(errs, err)
				// best effort cleanup
				_ = svc.DeleteJobLogs(ctx, job.ID)
				continue
			}
		}
	}

	return errs
}

func setJobAsAggregated(ctx context.Context, db database.DB, searchJobID int64) error {
	q := sqlf.Sprintf("UPDATE exhaustive_search_jobs SET is_aggregated = true WHERE id = %s", searchJobID)
	_, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

type job struct {
	ID        int64
	Initiator int32
}

// listSearchJobs returns a list of search jobs that haven't been aggregated
// yet.
func listSearchJobs(ctx context.Context, db database.DB) ([]job, error) {
	q := sqlf.Sprintf("SELECT id, initiator_id FROM exhaustive_search_jobs WHERE is_aggregated = false")
	rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []job
	for rows.Next() {
		var j job
		if err := rows.Scan(
			&j.ID,
			&j.Initiator,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func getAggregateStatus(ctx context.Context, service *service.Service, searchJobID int64) (types.JobState, error) {
	job, err := service.GetSearchJob(ctx, searchJobID)
	if err != nil {
		return "", err
	}

	return job.AggState, nil
}

func updateSearchJobStatus(ctx context.Context, db database.DB, searchJobID int64, status types.JobState) error {
	q := sqlf.Sprintf("UPDATE exhaustive_search_jobs SET state = %s WHERE id = %s", status.String(), searchJobID)
	_, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

func uploadLogsToBlobstore(ctx context.Context, svc *service.Service, searchJobID int64) error {
	csvWriterTo, err := svc.GetSearchJobLogsWriterTo(ctx, searchJobID)
	if err != nil {
		return err
	}

	r, w := io.Pipe()

	var g conc.WaitGroup
	defer g.Wait()

	g.Go(func() {
		_, err := csvWriterTo.WriteTo(w)
		w.CloseWithError(err)
	})

	_, err = svc.UploadJobLogs(ctx, searchJobID, r)

	return err
}

func deleteRepoJobs(ctx context.Context, db database.DB, searchJobID int64) error {
	q := sqlf.Sprintf("DELETE FROM exhaustive_search_repo_jobs WHERE search_job_id = %s", searchJobID)
	_, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}
