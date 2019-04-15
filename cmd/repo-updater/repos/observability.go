package repos

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// ErrorLogger captures the method required for logging an error.
type ErrorLogger interface {
	Error(msg string, ctx ...interface{})
}

// ObservedSource returns a decorator that wraps a Source
// with error logging, Prometheus metrics and tracing.
func ObservedSource(l ErrorLogger) func(Source) Source {
	return func(s Source) Source {
		return &observedSource{source: s, log: l}
	}
}

// An observedSource wraps another Source with error logging,
// Prometheus metrics and tracing.
type observedSource struct {
	source Source
	log    ErrorLogger
}

// ListRepos calls into the inner Source registers the observed results.
func (o *observedSource) ListRepos(ctx context.Context) (rs []*Repo, err error) {
	defer log(o.log, "source.list-repos", &err)
	return o.source.ListRepos(ctx)
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
