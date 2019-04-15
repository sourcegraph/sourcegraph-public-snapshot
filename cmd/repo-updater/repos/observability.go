package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// ErrorLogger captures the method required for logging an error.
type ErrorLogger interface {
	Error(msg string, ctx ...interface{})
}

// ObservedSource returns a decorator that wraps a Source
// with error logging, Prometheus metrics and tracing.
func ObservedSource(l ErrorLogger, m *SourceMetrics) func(Source) Source {
	return func(s Source) Source {
		return &observedSource{
			Source:  s,
			log:     l,
			metrics: m,
		}
	}
}

// An observedSource wraps another Source with error logging,
// Prometheus metrics and tracing.
type observedSource struct {
	Source
	log     ErrorLogger
	metrics *SourceMetrics
}

// SourceMetrics encapsulates the Prometheus metrics of a Source.
type SourceMetrics struct {
	Duration *prometheus.HistogramVec
	Repos    *prometheus.CounterVec
	Errors   *prometheus.CounterVec
}

// NewSourceMetrics returns SourceMetrics that need to be registered
// in a Prometheus registry.
func NewSourceMetrics() *SourceMetrics {
	return &SourceMetrics{
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "repoupdater",
			Name:      "source_duration_seconds",
			Help:      "Time spent sourcing repos",
		}, []string{"kind"}),
		Repos: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "repoupdater",
			Name:      "source_repos_total",
			Help:      "Total number of sourced repositories",
		}, []string{"kind"}),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "repoupdater",
			Name:      "source_errors_total",
			Help:      "Total number of sourcing errors",
		}, []string{"kind"}),
	}
}

// ListRepos calls into the inner Source registers the observed results.
func (o *observedSource) ListRepos(ctx context.Context) (rs []*Repo, err error) {
	defer log(o.log, "source.list-repos", &err)
	if o.metrics != nil {
		defer func(began time.Time) {
			took := time.Since(began).Seconds()
			for _, kind := range o.Source.Kinds() {
				o.metrics.Duration.WithLabelValues(kind).Observe(took)
				o.metrics.Repos.WithLabelValues(kind).Add(float64(len(rs)))
				if err != nil {
					o.metrics.Errors.WithLabelValues(kind).Add(1)
				}
			}
		}(time.Now())
	}
	return o.Source.ListRepos(ctx)
}

// NewObservedStore wraps the given Store with error logging,
// Prometheus metrics and tracing.
func NewObservedStore(s Store, l ErrorLogger) *ObservedStore {
	return &ObservedStore{
		store: s,
		log:   l,
	}
}

// An ObservedStore wraps another Store with error logging,
// Prometheus metrics and tracing.
type ObservedStore struct {
	store Store
	log   ErrorLogger
}

// Transact calls into the inner Store Transact method and
// returns an observed TxStore.
func (o *ObservedStore) Transact(ctx context.Context) (TxStore, error) {
	txstore, err := o.store.(Transactor).Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "observed store")
	}
	return &ObservedStore{
		store: txstore,
		log:   o.log,
	}, nil
}

// Done calls into the inner Store Done method.
func (o *ObservedStore) Done(errs ...*error) {
	defer func() {
		for _, err := range errs {
			log(o.log, "txstore.done", err)
		}
	}()
	o.store.(TxStore).Done(errs...)
}

// ListExternalServices calls into the inner Store and registers the observed results.
func (o *ObservedStore) ListExternalServices(ctx context.Context, args StoreListExternalServicesArgs) (es []*ExternalService, err error) {
	defer log(o.log, "store.list-external-services", &err, "args", fmt.Sprintf("%+v", args))
	return o.store.ListExternalServices(ctx, args)
}

// UpsertExternalServices calls into the inner Store and registers the observed results.
func (o *ObservedStore) UpsertExternalServices(ctx context.Context, svcs ...*ExternalService) (err error) {
	defer log(o.log, "store.upsert-external-services", &err,
		"count", len(svcs),
		"names", ExternalServices(svcs).DisplayNames(),
	)
	return o.store.UpsertExternalServices(ctx, svcs...)
}

// ListRepos calls into the inner Store and registers the observed results.
func (o *ObservedStore) ListRepos(ctx context.Context, args StoreListReposArgs) (rs []*Repo, err error) {
	defer log(o.log, "store.list-external-services", &err, "args", fmt.Sprintf("%+v", args))
	return o.store.ListRepos(ctx, args)
}

// UpsertRepos calls into the inner Store and registers the observed results.
func (o *ObservedStore) UpsertRepos(ctx context.Context, repos ...*Repo) (err error) {
	defer log(o.log, "store.list-external-services", &err,
		"count", len(repos),
		"names", Repos(repos).Names(),
	)
	return o.store.UpsertRepos(ctx, repos...)
}

func log(lg ErrorLogger, msg string, err *error, ctx ...interface{}) {
	if err == nil || *err == nil {
		return
	}

	args := append(make([]interface{}, 0, 2+len(ctx)), "error", *err)
	args = append(args, ctx...)

	lg.Error(msg, args...)
}
