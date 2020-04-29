package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type JanitorMetrics struct {
	StalledJobs prometheus.Counter
	Errors      prometheus.Counter
}

func NewJanitorMetrics() JanitorMetrics {
	return JanitorMetrics{
		StalledJobs: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-api-server",
			Name:      "janitor_stalled_jobs",
			Help:      "Total number of reset stalled jobs",
		}),
		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-api-server",
			Name:      "janitor_errors",
			Help:      "Total number of errors when running the janitor",
		}),
	}
}

type Janitor struct {
	db              db.DB
	janitorInterval time.Duration
	metrics         JanitorMetrics
}

type JanitorOpts struct {
	DB              db.DB
	JanitorInterval time.Duration
	Metrics         JanitorMetrics
}

func NewJanitor(opts JanitorOpts) *Janitor {
	return &Janitor{
		db:              opts.DB,
		janitorInterval: opts.JanitorInterval,
		metrics:         opts.Metrics,
	}
}

func (j *Janitor) Start() {
	for {
		if err := j.step(); err != nil {
			log15.Error("Failed to run janitor process", "error", err)
		}

		time.Sleep(j.janitorInterval)
	}
}

// Run performs a best-effort cleanup. See the following methods for more specifics.
//   - resetStalled
func (j *Janitor) step() error {
	cleanupFns := []func() error{
		j.resetStalled,
	}

	for _, fn := range cleanupFns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// resetStalled moves all uploads that have been in the PROCESSING state for a while back
// to QUEUED. For each updated upload record, the conversion process that was responsible
// for handling the upload did not hold a row lock, indicating that it has died.
func (j *Janitor) resetStalled() error {
	ids, err := j.db.ResetStalled(context.Background(), time.Now())
	if err != nil {
		j.metrics.Errors.Inc()
		return err
	}

	for _, id := range ids {
		log15.Debug("Reset stalled upload", "uploadID", id)
	}

	j.metrics.StalledJobs.Add(float64(len(ids)))
	return nil
}
