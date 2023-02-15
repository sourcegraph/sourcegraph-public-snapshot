package gitserver

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metricsJob struct{}

func NewMetricsJob() job.Job {
	return &metricsJob{}
}

func (j *metricsJob) Description() string {
	return ""
}

func (j *metricsJob) Config() []env.Config {
	return nil
}

func (j *metricsJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	c := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_gitserver_repo_last_error_total",
		Help: "Number of repositories whose last_error column is not empty.",
	}, func() float64 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var count int64
		err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(failed_fetch), 0) FROM gitserver_repos_statistics`).Scan(&count)
		if err != nil {
			observationCtx.Logger.Error("failed to count repository errors", log.Error(err))
			return 0
		}
		return float64(count)
	})
	prometheus.MustRegister(c)

	c = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_gitserver_repo_count",
		Help: "Number of repos.",
	}, func() float64 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var count int64
		err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(total), 0) FROM repo_statistics`).Scan(&count)
		if err != nil {
			observationCtx.Logger.Error("failed to count repositories", log.Error(err))
			return 0
		}
		return float64(count)
	})
	prometheus.MustRegister(c)

	return nil, nil
}
