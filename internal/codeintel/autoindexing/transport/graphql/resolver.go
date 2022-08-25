package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type Resolver interface {
	// Indexes
	GetIndexByID(ctx context.Context, id int) (shared.Index, bool, error)    // simple dbstore
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]shared.Index, error) // simple dbstore
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	DeleteIndexByID(ctx context.Context, id int) error                                                                  // simple dbstore
	QueueAutoIndexJobsForRepo(ctx context.Context, repositoryID int, rev, configuration string) ([]shared.Index, error) // in the service QueueIndexes

	// Index Configuration
	GetIndexConfiguration(ctx context.Context, repositoryID int) ([]byte, bool, error)                                        // GetIndexConfigurationByRepositoryID
	InferedIndexConfiguration(ctx context.Context, repositoryID int, commit string) (*config.IndexConfiguration, bool, error) // in the service InferIndexConfiguration first return
	InferedIndexConfigurationHints(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error)       // in the service InferIndexConfiguration second return
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, configuration string) error                 // simple dbstore

	// Index Connection Factory
	IndexConnectionResolverFromFactory(opts shared.GetIndexesOptions) *IndexesResolver // for the resolver
}

type resolver struct {
	svc        *autoindexing.Service
	operations *operations
}

func New(svc *autoindexing.Service, observationContext *observation.Context) Resolver {
	return &resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *resolver) GetIndexByID(ctx context.Context, id int) (_ shared.Index, _ bool, err error) {
	ctx, _, endObservation := r.operations.getIndexByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.svc.GetIndexByID(ctx, id)
}

func (r *resolver) GetIndexesByIDs(ctx context.Context, ids ...int) (_ []shared.Index, err error) {
	ctx, _, endObservation := r.operations.getIndexesByIDs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.svc.GetIndexesByIDs(ctx, ids...)
}

func (r *resolver) GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error) {
	ctx, _, endObservation := r.operations.getRecentIndexesSummary.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.svc.GetRecentIndexesSummary(ctx, repositoryID)
}

func (r *resolver) GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := r.operations.getLastIndexScanForRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.svc.GetLastIndexScanForRepository(ctx, repositoryID)
}

func (r *resolver) DeleteIndexByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := r.operations.deleteIndexByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	_, err = r.svc.DeleteIndexByID(ctx, id)
	return err
}

func (r *resolver) QueueAutoIndexJobsForRepo(ctx context.Context, repositoryID int, rev, configuration string) (_ []shared.Index, err error) {
	ctx, _, endObservation := r.operations.queueAutoIndexJobsForRepo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.svc.QueueIndexes(ctx, repositoryID, rev, configuration, true, true)
}

func (r *resolver) GetIndexConfiguration(ctx context.Context, repositoryID int) (_ []byte, _ bool, err error) {
	ctx, _, endObservation := r.operations.getIndexConfiguration.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	configuration, exists, err := r.svc.GetIndexConfigurationByRepositoryID(ctx, repositoryID)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	return configuration.Data, true, nil
}

func (r *resolver) InferedIndexConfiguration(ctx context.Context, repositoryID int, commit string) (_ *config.IndexConfiguration, _ bool, err error) {
	ctx, _, endObservation := r.operations.inferedIndexConfiguration.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	maybeConfig, _, err := r.svc.InferIndexConfiguration(ctx, repositoryID, commit, true)
	if err != nil || maybeConfig == nil {
		return nil, false, err
	}

	return maybeConfig, true, nil
}

func (r *resolver) InferedIndexConfigurationHints(ctx context.Context, repositoryID int, commit string) (_ []config.IndexJobHint, err error) {
	ctx, _, endObservation := r.operations.inferedIndexConfigurationHints.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	_, hints, err := r.svc.InferIndexConfiguration(ctx, repositoryID, commit, true)
	if err != nil {
		return nil, err
	}

	return hints, nil
}

func (r *resolver) UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, configuration string) (err error) {
	ctx, _, endObservation := r.operations.updateIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if _, err := config.UnmarshalJSON([]byte(configuration)); err != nil {
		return err
	}

	return r.svc.UpdateIndexConfigurationByRepositoryID(ctx, repositoryID, []byte(configuration))
}

func (r *resolver) IndexConnectionResolverFromFactory(opts shared.GetIndexesOptions) *IndexesResolver {
	return NewIndexesResolver(r.svc, opts)
}
