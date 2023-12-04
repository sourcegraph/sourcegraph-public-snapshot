package enqueuer

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/jobselector"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexEnqueuer struct {
	store           store.Store
	repoStore       database.RepoStore
	gitserverClient gitserver.Client
	operations      *operations
	jobSelector     *jobselector.JobSelector
}

func NewIndexEnqueuer(
	observationCtx *observation.Context,
	store store.Store,
	repoStore database.RepoStore,
	gitserverClient gitserver.Client,
	jobSelector *jobselector.JobSelector,
) *IndexEnqueuer {
	return &IndexEnqueuer{
		store:           store,
		repoStore:       repoStore,
		gitserverClient: gitserverClient,
		operations:      newOperations(observationCtx),
		jobSelector:     jobSelector,
	}
}

// QueueIndexes enqueues a set of index jobs for the following repository and commit. If a non-empty
// configuration is given, it will be used to determine the set of jobs to enqueue. Otherwise, it will
// the configuration will be determined based on the regular index scheduling rules: first read any
// in-repo configuration (e.g., sourcegraph.yaml), then look for any existing in-database configuration,
// finally falling back to the automatically inferred configuration based on the repo contents at the
// target commit.
//
// If the force flag is false, then the presence of an upload or index record for this given repository and commit
// will cause this method to no-op. Note that this is NOT a guarantee that there will never be any duplicate records
// when the flag is false.
func (s *IndexEnqueuer) QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []uploadsshared.Index, err error) {
	ctx, trace, endObservation := s.operations.queueIndex.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := s.repoStore.Get(ctx, api.RepoID(repositoryID))
	if err != nil {
		return nil, err
	}

	commitID, err := s.gitserverClient.ResolveRevision(ctx, repo.Name, rev, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.ResolveRevision")
	}
	commit := string(commitID)
	trace.AddEvent("ResolveRevision", attribute.String("commit", commit))

	return s.queueIndexForRepositoryAndCommit(ctx, repositoryID, commit, configuration, force, bypassLimit)
}

// QueueIndexesForPackage enqueues index jobs for a dependency of a recently-processed precise code
// intelligence index.
func (s *IndexEnqueuer) QueueIndexesForPackage(ctx context.Context, pkg dependencies.MinimialVersionedPackageRepo) (err error) {
	ctx, trace, endObservation := s.operations.queueIndexForPackage.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("scheme", pkg.Scheme),
		attribute.String("name", string(pkg.Name)),
		attribute.String("version", pkg.Version),
	}})
	defer endObservation(1, observation.Args{})

	repoName, revision, ok := inference.InferRepositoryAndRevision(pkg)
	if !ok {
		return nil
	}
	trace.AddEvent("InferRepositoryAndRevision",
		attribute.String("repoName", string(repoName)),
		attribute.String("revision", revision))

	repo, err := s.repoStore.GetByName(ctx, repoName)
	if err != nil {
		return errors.Wrap(err, "store.Repos.GetByName")
	}
	repoID := int(repo.ID)

	commit, err := s.gitserverClient.ResolveRevision(ctx, repoName, revision, gitserver.ResolveRevisionOptions{})
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "gitserverClient.ResolveRevision")
	}

	_, err = s.queueIndexForRepositoryAndCommit(ctx, repoID, string(commit), "", false, false)
	return err
}

// queueIndexForRepositoryAndCommit determines a set of index jobs to enqueue for the given repository and commit.
//
// If the force flag is false, then the presence of an upload or index record for this given repository and commit
// will cause this method to no-op. Note that this is NOT a guarantee that there will never be any duplicate records
// when the flag is false.
func (s *IndexEnqueuer) queueIndexForRepositoryAndCommit(ctx context.Context, repositoryID int, commit, configuration string, force, bypassLimit bool) ([]uploadsshared.Index, error) {
	if !force {
		isQueued, err := s.store.IsQueued(ctx, repositoryID, commit)
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.IsQueued")
		}
		if isQueued {
			return nil, nil
		}
	}

	indexes, err := s.jobSelector.GetIndexRecords(ctx, repositoryID, commit, configuration, bypassLimit)
	if err != nil {
		return nil, err
	}
	if len(indexes) == 0 {
		return nil, nil
	}

	indexesToInsert := indexes
	if !force {
		indexesToInsert = []uploadsshared.Index{}
		for _, index := range indexes {
			isQueued, err := s.store.IsQueuedRootIndexer(ctx, repositoryID, commit, index.Root, index.Indexer)
			if err != nil {
				return nil, errors.Wrap(err, "dbstore.IsQueuedRootIndexer")
			}
			if !isQueued {
				indexesToInsert = append(indexesToInsert, index)
			}
		}
	}

	return s.store.InsertIndexes(ctx, indexesToInsert)
}
