package repos

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ObservedSource returns a decorator that wraps a Source
// with error logging, Prometheus metrics and tracing.
func ObservedSource(l log.Logger, m SourceMetrics) func(Source) Source {
	return func(s Source) Source {
		return &observedSource{
			Source:  s,
			metrics: m,
			logger:  l,
		}
	}
}

// An observedSource wraps another Source with error logging,
// Prometheus metrics and tracing.
type observedSource struct {
	Source
	metrics SourceMetrics
	logger  log.Logger
}

// SourceMetrics encapsulates the Prometheus metrics of a Source.
type SourceMetrics struct {
	ListRepos *metrics.REDMetrics
	GetRepo   *metrics.REDMetrics
}

// MustRegister registers all metrics in SourceMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (sm SourceMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(sm.ListRepos.Count)
	r.MustRegister(sm.ListRepos.Duration)
	r.MustRegister(sm.ListRepos.Errors)
	r.MustRegister(sm.GetRepo.Count)
	r.MustRegister(sm.GetRepo.Duration)
	r.MustRegister(sm.GetRepo.Errors)
}

// NewSourceMetrics returns SourceMetrics that need to be registered
// in a Prometheus registry.
func NewSourceMetrics() SourceMetrics {
	return SourceMetrics{
		ListRepos: &metrics.REDMetrics{
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
		GetRepo: &metrics.REDMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_source_get_repo_duration_seconds",
				Help: "Time spent calling GetRepo",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_source_get_repo_total",
				Help: "Total number of GetRepo calls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_source_get_repo_errors_total",
				Help: "Total number of GetRepo errors",
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
		if err != nil {
			o.logger.Error("source.list-repos", log.Error(err))
		}
	}(time.Now())

	uncounted := make(chan SourceResult)
	go func() {
		o.Source.ListRepos(ctx, uncounted)
		close(uncounted)
	}()

	var errs error
	for res := range uncounted {
		results <- res
		if res.Err != nil {
			errs = errors.Append(errs, res.Err)
		}
		count++
	}
	if errs != nil {
		err = errs
	}
}

// GetRepo calls into the inner Source and registers the observed results.
func (o *observedSource) GetRepo(ctx context.Context, path string) (sourced *types.Repo, err error) {
	rg, ok := o.Source.(RepoGetter)
	if !ok {
		return nil, errors.New("RepoGetter not implemented")
	}

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		o.metrics.GetRepo.Observe(secs, 1, &err)
		if err != nil {
			o.logger.Error("source.get-repo", log.Error(err))
		}
	}(time.Now())

	return rg.GetRepo(ctx, path)
}
