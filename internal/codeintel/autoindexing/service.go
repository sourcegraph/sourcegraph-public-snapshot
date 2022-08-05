package autoindexing

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ service = (*Service)(nil)

type service interface {
	// Not in use yet.
	List(ctx context.Context, opts ListOpts) (jobs []IndexJob, err error)
	Get(ctx context.Context, id int) (job IndexJob, ok bool, err error)
	GetBatch(ctx context.Context, ids ...int) (jobs []IndexJob, err error)
	Delete(ctx context.Context, id int) (err error)
	Enqueue(ctx context.Context, jobs []IndexJob) (err error)
	Infer(ctx context.Context, repoID int) (jobs []IndexJob, err error)
	UpdateIndexingConfiguration(ctx context.Context, repoID int) (jobs []IndexJob, err error)

	// Commits
	GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (indexesDeleted int, err error)

	// Indexes
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
}

type Service struct {
	autoindexingStore store.Store
	dbStore           DBStore // TODO - roll into store
	gitserverClient   GitserverClient
	repoUpdater       RepoUpdaterClient
	inferenceService  inferenceService
	operations        *operations
	logger            log.Logger
}

func newService(
	autoindexingStore store.Store,
	dbStore DBStore,
	gitserverClient GitserverClient,
	repoUpdaterClient RepoUpdaterClient,
	inferenceService inferenceService,
	observationContext *observation.Context,
) *Service {
	return &Service{
		autoindexingStore: autoindexingStore,
		dbStore:           dbStore,
		gitserverClient:   gitserverClient,
		repoUpdater:       repoUpdaterClient,
		inferenceService:  inferenceService,
		operations:        newOperations(observationContext),
		logger:            observationContext.Logger,
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

func (s *Service) DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.deleteIndexesWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.autoindexingStore.DeleteIndexesWithoutRepository(ctx, now)
}

func (s *Service) GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, _, endObservation := s.operations.getStaleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.autoindexingStore.GetStaleSourcedCommits(ctx, minimumTimeSinceLastCheck, limit, now)
}

func (s *Service) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error) {
	ctx, _, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.autoindexingStore.UpdateSourcedCommits(ctx, repositoryID, commit, now)
}

func (s *Service) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (indexesDeleted int, err error) {
	ctx, _, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.autoindexingStore.DeleteSourcedCommits(ctx, repositoryID, commit, maximumCommitLag)
}
