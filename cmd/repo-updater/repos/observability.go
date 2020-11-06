package repos

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

// ObservedSource returns a decorator that wraps a Source
// with error logging, Prometheus metrics and tracing.
func ObservedSource(l logging.ErrorLogger, m SourceMetrics) func(Source) Source {
	return func(s Source) Source {
		return &observedSource{
			Source:  s,
			metrics: m,
			log:     l,
		}
	}
}

// An observedSource wraps another Source with error logging,
// Prometheus metrics and tracing.
type observedSource struct {
	Source
	metrics SourceMetrics
	log     logging.ErrorLogger
}

// SourceMetrics encapsulates the Prometheus metrics of a Source.
type SourceMetrics struct {
	ListRepos *metrics.OperationMetrics
}

// MustRegister registers all metrics in SourceMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (sm SourceMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(sm.ListRepos.Count)
	r.MustRegister(sm.ListRepos.Duration)
	r.MustRegister(sm.ListRepos.Errors)
}

// NewSourceMetrics returns SourceMetrics that need to be registered
// in a Prometheus registry.
func NewSourceMetrics() SourceMetrics {
	return SourceMetrics{
		ListRepos: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_source_duration_seconds",
				Help: "Time spent sourcing repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_source_repos_total",
				Help: "Total number of sourced repositories",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_source_errors_total",
				Help: "Total number of sourcing errors",
			}, []string{}),
		},
	}
}

// ListRepos calls into the inner Source registers the observed results.
func (o *observedSource) ListRepos(ctx context.Context, results chan SourceResult) {
	var (
		err   error
		count float64
	)

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		o.metrics.ListRepos.Observe(secs, count, &err)
		logging.Log(o.log, "source.list-repos", &err)
	}(time.Now())

	uncounted := make(chan SourceResult)
	go func() {
		o.Source.ListRepos(ctx, uncounted)
		close(uncounted)
	}()

	var errs *multierror.Error
	for res := range uncounted {
		results <- res
		if res.Err != nil {
			errs = multierror.Append(errs, res.Err)
		}
		count++
	}
	if errs != nil {
		err = errs.ErrorOrNil()
	}
}
