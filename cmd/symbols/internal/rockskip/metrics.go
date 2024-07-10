package rockskip

import (
	"context"
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	searchRunning  prometheus.Gauge
	searchFailed   prometheus.Counter
	searchDuration prometheus.Histogram
	indexRunning   prometheus.Gauge
	indexFailed    prometheus.Counter
	indexDuration  prometheus.Histogram
	queueAge       prometheus.Histogram
}

func newMetrics(observationCtx *observation.Context, db *sql.DB) *metrics {
	scanCount := func(sql string) (float64, error) {
		row := db.QueryRowContext(context.Background(), sql)
		var count int64
		err := row.Scan(&count)
		if err != nil {
			return 0, err
		}
		return float64(count), nil
	}

	ns := "src_rockskip_service"

	indexedRepos := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: ns,
		Name:      "repos_indexed",
		Help:      "The number of repositories indexed by rockskip",
	}, func() float64 {
		count, err := scanCount(`SELECT COUNT(*) FROM rockskip_repos`)
		if err != nil {
			observationCtx.Logger.Error("failed to get number of index repos", log.Error(err))
			return 0
		}
		return count
	})
	observationCtx.Registerer.MustRegister(indexedRepos)

	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      name,
			Help:      help,
		})
		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	gauge := func(name, help string) prometheus.Gauge {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      name,
			Help:      help,
		})
		observationCtx.Registerer.MustRegister(gauge)
		return gauge
	}

	histogram := func(name, help string, buckets []float64) prometheus.Histogram {
		histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		})
		observationCtx.Registerer.MustRegister(histogram)
		return histogram
	}

	return &metrics{
		searchRunning: gauge("in_flight_search_requests", "Number of in-flight search requests"),
		searchFailed:  counter("search_request_errors", "Number of search requests that returned an error"),
		searchDuration: histogram(
			"search_request_duration_seconds",
			"Search request duration in seconds.",
			[]float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 20, 30},
		),
		indexRunning: gauge("in_flight_index_jobs", "Number of in-flight index jobs"),
		indexFailed:  counter("index_job_errors", "Number of index jobs that returned an error"),
		indexDuration: histogram(
			"index_job_duration_seconds",
			"Search request duration in seconds.",
			prometheus.ExponentialBuckets(0.1, 2, 22),
		),
		queueAge: histogram(
			"index_queue_age_seconds",
			"A histogram of the amount of time a popped index request spent sitting in the queue beforehand.",
			[]float64{
				60,     // 1m
				300,    // 5m
				1200,   // 20m
				2400,   // 40m
				3600,   // 1h
				10800,  // 3h
				18000,  // 5h
				36000,  // 10h
				43200,  // 12h
				54000,  // 15h
				72000,  // 20h
				86400,  // 24h
				108000, // 30h
				126000, // 35h
				172800, // 48h
			},
		),
	}
}
