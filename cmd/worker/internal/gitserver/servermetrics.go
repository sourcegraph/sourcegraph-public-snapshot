package gitserver

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
)

func RegisterPrometheusCollectors(logger log.Logger) error {
	db, err := workerdb.Init()
	if err != nil {
		return err
	}

	c := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_gitserver_repo_last_error_total",
		Help: "Number of repositories whose last_error column is not empty.",
	}, func() float64 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var count int64
		err := db.QueryRowContext(ctx, `SELECT SUM(failed_fetch) FROM gitserver_repos_statistics`).Scan(&count)
		if err != nil {
			logger.Error("failed to count repository errors", log.Error(err))
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
		err := db.QueryRowContext(ctx, `SELECT SUM(total) FROM repo_statistics`).Scan(&count)
		if err != nil {
			logger.Error("failed to count repositories", log.Error(err))
			return 0
		}
		return float64(count)
	})
	prometheus.MustRegister(c)

	return nil
}
