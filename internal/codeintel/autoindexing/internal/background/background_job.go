package background

import (
	"context"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BackgroundJob interface {
	WorkerutilStore() dbworkerstore.Store
	DependencySyncStore() dbworkerstore.Store
	DependencyIndexingStore() dbworkerstore.Store
}

type backgroundJob struct {
	uploadSvc shared.UploadService
	depsSvc   DependenciesService

	repoUpdater shared.RepoUpdaterClient

	workerutilStore         dbworkerstore.Store
	dependencySyncStore     dbworkerstore.Store
	dependencyIndexingStore dbworkerstore.Store

	operations *operations
}

func New(
	db database.DB,
	uploadSvc shared.UploadService,
	repoUpdater shared.RepoUpdaterClient,
	observationContext *observation.Context,
) BackgroundJob {
	workerutilStore := dbworkerstore.NewWithMetrics(db.Handle(), indexWorkerStoreOptions, observationContext)
	dependencySyncStore := dbworkerstore.NewWithMetrics(db.Handle(), dependencySyncingJobWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.NewWithMetrics(db.Handle(), dependencyIndexingJobWorkerStoreOptions, observationContext)

	return &backgroundJob{
		uploadSvc:   uploadSvc,
		repoUpdater: repoUpdater,

		workerutilStore:         workerutilStore,
		dependencySyncStore:     dependencySyncStore,
		dependencyIndexingStore: dependencyIndexingStore,

		operations: newOperations(observationContext),
	}
}

func (b backgroundJob) WorkerutilStore() dbworkerstore.Store     { return b.workerutilStore }
func (b backgroundJob) DependencySyncStore() dbworkerstore.Store { return b.dependencySyncStore }
func (b backgroundJob) DependencyIndexingStore() dbworkerstore.Store {
	return b.dependencyIndexingStore
}

// QueueIndexesForPackage enqueues index jobs for a dependency of a recently-processed precise code
// intelligence index.
func (b backgroundJob) queueIndexesForPackage(ctx context.Context, pkg precise.Package) (err error) {
	ctx, trace, endObservation := b.operations.queueIndexForPackage.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.String("scheme", pkg.Scheme),
			otlog.String("name", pkg.Name),
			otlog.String("version", pkg.Version),
		},
	})
	defer endObservation(1, observation.Args{})

	repoName, revision, ok := InferRepositoryAndRevision(pkg)
	if !ok {
		return nil
	}
	trace.Log(otlog.String("repoName", string(repoName)))
	trace.Log(otlog.String("revision", revision))

	resp, err := s.repoUpdater.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "repoUpdater.EnqueueRepoUpdate")
	}

	commit, err := b.gitserverClient.ResolveRevision(ctx, int(resp.ID), revision)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "gitserverClient.ResolveRevision")
	}

	_, err = b.autoindexingSvc.queueIndexForRepositoryAndCommit(ctx, int(resp.ID), string(commit), "", false, false, nil) // trace)
	return err
}
