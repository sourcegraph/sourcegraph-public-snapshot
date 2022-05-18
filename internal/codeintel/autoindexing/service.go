package autoindexing

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	autoindexingStore store.Store
	operations        *operations
}

func newService(autoindexingStore store.Store, observationContext *observation.Context) *Service {
	return &Service{
		autoindexingStore: autoindexingStore,
		operations:        newOperations(observationContext),
	}
}

type IndexJob = shared.IndexJob

type ListOpts struct {
	Limit int
}

func (s *Service) List(ctx context.Context, opts ListOpts) (jobs []IndexJob, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.autoindexingStore.List(ctx, store.ListOpts(opts))
}

func (s *Service) Get(ctx context.Context, id int) (job IndexJob, ok bool, err error) {
	ctx, _, endObservation := s.operations.get.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_ = ctx
	return IndexJob{}, false, errors.Newf("unimplemented: autoindexing.Get")
}

func (s *Service) GetBatch(ctx context.Context, ids ...int) (jobs []IndexJob, err error) {
	ctx, _, endObservation := s.operations.getBatch.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_ = ctx
	return nil, errors.Newf("unimplemented: autoindexing.GetBatch")
}

func (s *Service) Delete(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.delete.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_ = ctx
	return errors.Newf("unimplemented: autoindexing.Delete")
}

func (s *Service) Enqueue(ctx context.Context, jobs []IndexJob) (err error) {
	ctx, _, endObservation := s.operations.enqueue.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_ = ctx
	return errors.Newf("unimplemented: autoindexing.Enqueue")
}

func (s *Service) Infer(ctx context.Context, repoID int) (jobs []IndexJob, err error) {
	ctx, _, endObservation := s.operations.infer.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_ = ctx
	return nil, errors.Newf("unimplemented: autoindexing.Infer")
}

func (s *Service) UpdateIndexingConfiguration(ctx context.Context, repoID int) (jobs []IndexJob, err error) {
	ctx, _, endObservation := s.operations.updateIndexingConfiguration.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33377
	_ = ctx
	return nil, errors.Newf("unimplemented: autoindexing.UpdateIndexingConfiguration")
}
