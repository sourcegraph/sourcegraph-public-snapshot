package autoindexing

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	autoindexingStore Store
	operations        *operations
}

func newService(autoindexingStore Store, observationContext *observation.Context) *Service {
	return &Service{
		autoindexingStore: autoindexingStore,
		operations:        newOperations(observationContext),
	}
}

type IndexJob struct {
	// TODO
}

type ListOpts struct {
	// TODO
}

func (s *Service) List(ctx context.Context, opts ListOpts) (jobs []IndexJob, err error) {
	ctx, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

func (s *Service) Get(ctx context.Context, id int) (job IndexJob, ok bool, err error) {
	ctx, endObservation := s.operations.get.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return IndexJob{}, false, nil
}

func (s *Service) GetBatch(ctx context.Context, ids ...int) (jobs []IndexJob, err error) {
	ctx, endObservation := s.operations.getBatch.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

func (s *Service) Delete(ctx context.Context, id int) (err error) {
	ctx, endObservation := s.operations.delete.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil
}

func (s *Service) Enqueue(ctx context.Context, jobs []IndexJob) (err error) {
	ctx, endObservation := s.operations.enqueue.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil
}

func (s *Service) Infer(ctx context.Context, repoID int) (jobs []IndexJob, err error) {
	ctx, endObservation := s.operations.infer.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

// TODO - CRUD on indexing configuration
