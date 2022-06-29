package uploads

import (
	"context"
	"io"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

type Upload = shared.Upload

type ListOpts struct {
	Limit int
}

func (s *Service) List(ctx context.Context, opts ListOpts) (uploads []Upload, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.uploadsStore.List(ctx, store.ListOpts(opts))
}

func (s *Service) DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.uploadsStore.DeleteUploadsWithoutRepository(ctx, now)
}

func (s *Service) Get(ctx context.Context, id int) (upload Upload, ok bool, err error) {
	ctx, _, endObservation := s.operations.get.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return Upload{}, false, errors.Newf("unimplemented: uploads.Get")
}

func (s *Service) GetBatch(ctx context.Context, ids ...int) (uploads []Upload, err error) {
	ctx, _, endObservation := s.operations.getBatch.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return nil, errors.Newf("unimplemented: uploads.GetBatch")
}

type UploadState struct{}

func (s *Service) Enqueue(ctx context.Context, state UploadState, reader io.Reader) (err error) {
	ctx, _, endObservation := s.operations.enqueue.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return errors.Newf("unimplemented: uploads.Enqueue")
}

func (s *Service) Delete(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.delete.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return errors.Newf("unimplemented: uploads.Delete")
}

func (s *Service) CommitsVisibleToUpload(ctx context.Context, id int) (commits []string, err error) {
	ctx, _, endObservation := s.operations.commitsVisibleTo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return nil, errors.Newf("unimplemented: uploads.CommitsVisibleToUpload")
}

func (s *Service) UploadsVisibleToCommit(ctx context.Context, commit string) (uploads []Upload, err error) {
	ctx, _, endObservation := s.operations.uploadsVisibleTo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return nil, errors.Newf("unimplemented: uploads.UploadsVisibleToCommit")
}

func (s *Service) StaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, _, endObservation := s.operations.staleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.uploadsStore.StaleSourcedCommits(ctx, minimumTimeSinceLastCheck, limit, now)
}

func (s *Service) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error) {
	ctx, _, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.uploadsStore.UpdateSourcedCommits(ctx, repositoryID, commit, now)
}

func (s *Service) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error) {
	ctx, _, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.uploadsStore.DeleteSourcedCommits(ctx, repositoryID, commit, maximumCommitLag, now)
}
