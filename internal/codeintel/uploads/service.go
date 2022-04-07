package uploads

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	uploadsStore Store
	operations   *operations
}

func newService(uploadsStore Store, observationContext *observation.Context) *Service {
	return &Service{
		uploadsStore: uploadsStore,
		operations:   newOperations(observationContext),
	}
}

type Upload struct {
	// TODO
}

type ListOpts struct {
	// TODO
}

func (s *Service) List(ctx context.Context, opts ListOpts) (uploads []Upload, err error) {
	ctx, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

func (s *Service) Get(ctx context.Context, id int) (upload Upload, ok bool, err error) {
	ctx, endObservation := s.operations.get.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return Upload{}, false, nil
}

func (s *Service) GetBatch(ctx context.Context, ids ...int) (uploads []Upload, err error) {
	ctx, endObservation := s.operations.getBatch.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

type State struct {
	// TODO
}

func (s *Service) Enqueue(ctx context.Context, state State, reader io.Reader) (err error) {
	ctx, endObservation := s.operations.enqueue.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil
}

func (s *Service) Delete(ctx context.Context, id int) (err error) {
	ctx, endObservation := s.operations.delete.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil
}

func (s *Service) CommitsVisibleToUpload(ctx context.Context, id int) (commits []string, err error) {
	ctx, endObservation := s.operations.commitsVisibleTo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

func (s *Service) UploadsVisibleToCommit(ctx context.Context, commit string) (uploads []Upload, err error) {
	ctx, endObservation := s.operations.uploadsVisibleTo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}
