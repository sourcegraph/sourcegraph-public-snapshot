package db

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// An ObservedDB wraps another DB with error logging, Prometheus metrics, and tracing.
type ObservedDB struct {
	db                                 DB
	savepointOperation                 *observation.Operation
	rollbackToSavepointOperation       *observation.Operation
	doneOperation                      *observation.Operation
	getUploadByIDOperation             *observation.Operation
	getUploadsByRepoOperation          *observation.Operation
	queueSizeOperation                 *observation.Operation
	insertUploadOperation              *observation.Operation
	addUploadPartOperation             *observation.Operation
	markQueuedOperation                *observation.Operation
	markCompleteOperation              *observation.Operation
	markErroredOperation               *observation.Operation
	dequeueOperation                   *observation.Operation
	getStatesOperation                 *observation.Operation
	deleteUploadByIDOperation          *observation.Operation
	resetStalledOperation              *observation.Operation
	getDumpIDsOperation                *observation.Operation
	getDumpByIDOperation               *observation.Operation
	findClosestDumpsOperation          *observation.Operation
	deleteOldestDumpOperation          *observation.Operation
	updateDumpsVisibleFromTipOperation *observation.Operation
	deleteOverlappingDumpsOperation    *observation.Operation
	getPackageOperation                *observation.Operation
	updatePackagesOperation            *observation.Operation
	sameRepoPagerOperation             *observation.Operation
	updatePackageReferencesOperation   *observation.Operation
	packageReferencePagerOperation     *observation.Operation
	hasCommitOperation                 *observation.Operation
	updateCommitsOperation             *observation.Operation
	indexableRepositoriesOperation     *observation.Operation
	updateIndexableRepositoryOperation *observation.Operation
	getIndexByIDOperation              *observation.Operation
	indexQueueSizeOperation            *observation.Operation
	isQueuedOperation                  *observation.Operation
	insertIndexOperation               *observation.Operation
	markIndexCompleteOperation         *observation.Operation
	markIndexErroredOperation          *observation.Operation
	dequeueIndexOperation              *observation.Operation
	repoUsageStatisticsOperation       *observation.Operation
	repoNameOperation                  *observation.Operation
}

var _ DB = &ObservedDB{}

// NewObservedDB wraps the given DB with error logging, Prometheus metrics, and tracing.
func NewObserved(db DB, observationContext *observation.Context) DB {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"code_intel_db",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of results returned"),
	)

	return &ObservedDB{
		db: db,
		savepointOperation: observationContext.Operation(observation.Op{
			Name:         "DB.Savepoint",
			MetricLabels: []string{"savepoint"},
			Metrics:      metrics,
		}),
		rollbackToSavepointOperation: observationContext.Operation(observation.Op{
			Name:         "DB.RollbackToSavepoint",
			MetricLabels: []string{"rollback_to_savepoint"},
			Metrics:      metrics,
		}),
		doneOperation: observationContext.Operation(observation.Op{
			Name:         "DB.Done",
			MetricLabels: []string{"done"},
			Metrics:      metrics,
		}),
		getUploadByIDOperation: observationContext.Operation(observation.Op{
			Name:         "DB.GetUploadByID",
			MetricLabels: []string{"get_upload_by_id"},
			Metrics:      metrics,
		}),
		getUploadsByRepoOperation: observationContext.Operation(observation.Op{
			Name:         "DB.GetUploadsByRepo",
			MetricLabels: []string{"get_uploads_by_repo"},
			Metrics:      metrics,
		}),
		queueSizeOperation: observationContext.Operation(observation.Op{
			Name:         "DB.QueueSize",
			MetricLabels: []string{"queue_size"},
			Metrics:      metrics,
		}),
		insertUploadOperation: observationContext.Operation(observation.Op{
			Name:         "DB.InsertUpload",
			MetricLabels: []string{"insert_upload"},
			Metrics:      metrics,
		}),
		addUploadPartOperation: observationContext.Operation(observation.Op{
			Name:         "DB.AddUploadPart",
			MetricLabels: []string{"add_upload_part"},
			Metrics:      metrics,
		}),
		markQueuedOperation: observationContext.Operation(observation.Op{
			Name:         "DB.MarkQueued",
			MetricLabels: []string{"mark_queued"},
			Metrics:      metrics,
		}),
		markCompleteOperation: observationContext.Operation(observation.Op{
			Name:         "DB.MarkComplete",
			MetricLabels: []string{"mark_complete"},
			Metrics:      metrics,
		}),
		markErroredOperation: observationContext.Operation(observation.Op{
			Name:         "DB.MarkErrored",
			MetricLabels: []string{"mark_errored"},
			Metrics:      metrics,
		}),
		dequeueOperation: observationContext.Operation(observation.Op{
			Name:         "DB.Dequeue",
			MetricLabels: []string{"dequeue"},
			Metrics:      metrics,
		}),
		getStatesOperation: observationContext.Operation(observation.Op{
			Name:         "DB.GetStates",
			MetricLabels: []string{"get_states"},
			Metrics:      metrics,
		}),
		deleteUploadByIDOperation: observationContext.Operation(observation.Op{
			Name:         "DB.DeleteUploadByID",
			MetricLabels: []string{"delete_upload_by_id"},
			Metrics:      metrics,
		}),
		resetStalledOperation: observationContext.Operation(observation.Op{
			Name:         "DB.ResetStalled",
			MetricLabels: []string{"reset_stalled"},
			Metrics:      metrics,
		}),
		getDumpIDsOperation: observationContext.Operation(observation.Op{
			Name:         "DB.GetDumpIDs",
			MetricLabels: []string{"get_dump_ids"},
			Metrics:      metrics,
		}),
		getDumpByIDOperation: observationContext.Operation(observation.Op{
			Name:         "DB.GetDumpByID",
			MetricLabels: []string{"get_dump_by_id"},
			Metrics:      metrics,
		}),
		findClosestDumpsOperation: observationContext.Operation(observation.Op{
			Name:         "DB.FindClosestDumps",
			MetricLabels: []string{"find_closest_dumps"},
			Metrics:      metrics,
		}),
		deleteOldestDumpOperation: observationContext.Operation(observation.Op{
			Name:         "DB.DeleteOldestDump",
			MetricLabels: []string{"delete_oldest_dump"},
			Metrics:      metrics,
		}),
		updateDumpsVisibleFromTipOperation: observationContext.Operation(observation.Op{
			Name:         "DB.UpdateDumpsVisibleFromTip",
			MetricLabels: []string{"update_dumps_visible_from_tip"},
			Metrics:      metrics,
		}),
		deleteOverlappingDumpsOperation: observationContext.Operation(observation.Op{
			Name:         "DB.DeleteOverlappingDumps",
			MetricLabels: []string{"delete_overlapping_dumps"},
			Metrics:      metrics,
		}),
		getPackageOperation: observationContext.Operation(observation.Op{
			Name:         "DB.GetPackage",
			MetricLabels: []string{"get_package"},
			Metrics:      metrics,
		}),
		updatePackagesOperation: observationContext.Operation(observation.Op{
			Name:         "DB.UpdatePackages",
			MetricLabels: []string{"update_packages"},
			Metrics:      metrics,
		}),
		sameRepoPagerOperation: observationContext.Operation(observation.Op{
			Name:         "DB.SameRepoPager",
			MetricLabels: []string{"same_repo_pager"},
			Metrics:      metrics,
		}),
		updatePackageReferencesOperation: observationContext.Operation(observation.Op{
			Name:         "DB.UpdatePackageReferences",
			MetricLabels: []string{"update_package_references"},
			Metrics:      metrics,
		}),
		packageReferencePagerOperation: observationContext.Operation(observation.Op{
			Name:         "DB.PackageReferencePager",
			MetricLabels: []string{"package_reference_pager"},
			Metrics:      metrics,
		}),
		hasCommitOperation: observationContext.Operation(observation.Op{
			Name:         "DB.HasCommit",
			MetricLabels: []string{"has_commit"},
			Metrics:      metrics,
		}),
		updateCommitsOperation: observationContext.Operation(observation.Op{
			Name:         "DB.UpdateCommits",
			MetricLabels: []string{"update_commits"},
			Metrics:      metrics,
		}),
		indexableRepositoriesOperation: observationContext.Operation(observation.Op{
			Name:         "DB.IndexableRepositories",
			MetricLabels: []string{"indexable_repositories"},
			Metrics:      metrics,
		}),
		updateIndexableRepositoryOperation: observationContext.Operation(observation.Op{
			Name:         "DB.UpdateIndexableRepository",
			MetricLabels: []string{"update_indexable_repository"},
			Metrics:      metrics,
		}),
		getIndexByIDOperation: observationContext.Operation(observation.Op{
			Name:         "DB.GetIndexByID",
			MetricLabels: []string{"get_index_by_id"},
			Metrics:      metrics,
		}),
		indexQueueSizeOperation: observationContext.Operation(observation.Op{
			Name:         "DB.IndexQueueSize",
			MetricLabels: []string{"index_queue_size"},
			Metrics:      metrics,
		}),
		isQueuedOperation: observationContext.Operation(observation.Op{
			Name:         "DB.IsQueued",
			MetricLabels: []string{"is_queued"},
			Metrics:      metrics,
		}),
		insertIndexOperation: observationContext.Operation(observation.Op{
			Name:         "DB.InsertIndex",
			MetricLabels: []string{"insert_index"},
			Metrics:      metrics,
		}),
		markIndexCompleteOperation: observationContext.Operation(observation.Op{
			Name:         "DB.MarkIndexComplete",
			MetricLabels: []string{"mark_index_complete"},
			Metrics:      metrics,
		}),
		markIndexErroredOperation: observationContext.Operation(observation.Op{
			Name:         "DB.MarkIndexErrored",
			MetricLabels: []string{"mark_index_errored"},
			Metrics:      metrics,
		}),
		dequeueIndexOperation: observationContext.Operation(observation.Op{
			Name:         "DB.DequeueIndex",
			MetricLabels: []string{"dequeue_index"},
			Metrics:      metrics,
		}),
		repoUsageStatisticsOperation: observationContext.Operation(observation.Op{
			Name:         "DB.RepoUsageStatistics",
			MetricLabels: []string{"repo_usage_statistics"},
			Metrics:      metrics,
		}),
		repoNameOperation: observationContext.Operation(observation.Op{
			Name:         "DB.RepoName",
			MetricLabels: []string{"repo_name"},
			Metrics:      metrics,
		}),
	}
}

// wrap the given database with the same observed operations as the receiver database.
func (db *ObservedDB) wrap(other DB) DB {
	if other == nil {
		return nil
	}

	return &ObservedDB{
		db:                                 other,
		savepointOperation:                 db.savepointOperation,
		rollbackToSavepointOperation:       db.rollbackToSavepointOperation,
		doneOperation:                      db.doneOperation,
		getUploadByIDOperation:             db.getUploadByIDOperation,
		getUploadsByRepoOperation:          db.getUploadsByRepoOperation,
		queueSizeOperation:                 db.queueSizeOperation,
		insertUploadOperation:              db.insertUploadOperation,
		addUploadPartOperation:             db.addUploadPartOperation,
		markQueuedOperation:                db.markQueuedOperation,
		markCompleteOperation:              db.markCompleteOperation,
		markErroredOperation:               db.markErroredOperation,
		dequeueOperation:                   db.dequeueOperation,
		getStatesOperation:                 db.getStatesOperation,
		deleteUploadByIDOperation:          db.deleteUploadByIDOperation,
		resetStalledOperation:              db.resetStalledOperation,
		getDumpIDsOperation:                db.getDumpIDsOperation,
		getDumpByIDOperation:               db.getDumpByIDOperation,
		findClosestDumpsOperation:          db.findClosestDumpsOperation,
		deleteOldestDumpOperation:          db.deleteOldestDumpOperation,
		updateDumpsVisibleFromTipOperation: db.updateDumpsVisibleFromTipOperation,
		deleteOverlappingDumpsOperation:    db.deleteOverlappingDumpsOperation,
		getPackageOperation:                db.getPackageOperation,
		updatePackagesOperation:            db.updatePackagesOperation,
		sameRepoPagerOperation:             db.sameRepoPagerOperation,
		updatePackageReferencesOperation:   db.updatePackageReferencesOperation,
		packageReferencePagerOperation:     db.packageReferencePagerOperation,
		hasCommitOperation:                 db.hasCommitOperation,
		updateCommitsOperation:             db.updateCommitsOperation,
		indexableRepositoriesOperation:     db.indexableRepositoriesOperation,
		updateIndexableRepositoryOperation: db.updateIndexableRepositoryOperation,
		getIndexByIDOperation:              db.getIndexByIDOperation,
		indexQueueSizeOperation:            db.indexQueueSizeOperation,
		isQueuedOperation:                  db.isQueuedOperation,
		insertIndexOperation:               db.insertIndexOperation,
		markIndexCompleteOperation:         db.markIndexCompleteOperation,
		markIndexErroredOperation:          db.markIndexErroredOperation,
		dequeueIndexOperation:              db.dequeueIndexOperation,
		repoUsageStatisticsOperation:       db.repoUsageStatisticsOperation,
		repoNameOperation:                  db.repoNameOperation,
	}
}

// Transact calls into the inner DB and wraps the resulting value in an ObservedDB.
func (db *ObservedDB) Transact(ctx context.Context) (DB, error) {
	tx, err := db.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return db.wrap(tx), nil
}

// Savepoint calls into the inner DB and registers the observed results.
func (db *ObservedDB) Savepoint(ctx context.Context) (_ string, err error) {
	ctx, endObservation := db.savepointOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.Savepoint(ctx)
}

// RollbackToSavepoint calls into the inner DB and registers the observed results.
func (db *ObservedDB) RollbackToSavepoint(ctx context.Context, name string) (err error) {
	ctx, endObservation := db.rollbackToSavepointOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.RollbackToSavepoint(ctx, name)
}

// Done calls into the inner DB and registers the observed results.
func (db *ObservedDB) Done(e error) error {
	var observedErr error = nil
	_, endObservation := db.doneOperation.With(context.Background(), &observedErr, observation.Args{})
	defer endObservation(1, observation.Args{})

	err := db.db.Done(e)
	if err != e {
		// Only observe the error if it's a commit/rollback failure
		observedErr = err
	}
	return err
}

// GetUploadByID calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetUploadByID(ctx context.Context, id int) (_ Upload, _ bool, err error) {
	ctx, endObservation := db.getUploadByIDOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.GetUploadByID(ctx, id)
}

// GetUploadsByRepo calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetUploadsByRepo(ctx context.Context, repositoryID int, state, term string, visibleAtTip bool, limit, offset int) (uploads []Upload, _ int, err error) {
	ctx, endObservation := db.getUploadsByRepoOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(uploads)), observation.Args{}) }()
	return db.db.GetUploadsByRepo(ctx, repositoryID, state, term, visibleAtTip, limit, offset)
}

// QueueSize  calls into the inner DB and registers the observed results.
func (db *ObservedDB) QueueSize(ctx context.Context) (_ int, err error) {
	ctx, endObservation := db.queueSizeOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.QueueSize(ctx)
}

// InsertUpload calls into the inner DB and registers the observed result.
func (db *ObservedDB) InsertUpload(ctx context.Context, upload Upload) (_ int, err error) {
	ctx, endObservation := db.insertUploadOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.InsertUpload(ctx, upload)
}

// AddUploadPart calls into the inner DB and registers the observed result.
func (db *ObservedDB) AddUploadPart(ctx context.Context, uploadID, partIndex int) (err error) {
	ctx, endObservation := db.addUploadPartOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.AddUploadPart(ctx, uploadID, partIndex)
}

// MarkQueued calls into the inner DB and registers the observed result.
func (db *ObservedDB) MarkQueued(ctx context.Context, uploadID int) (err error) {
	ctx, endObservation := db.markQueuedOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.MarkQueued(ctx, uploadID)
}

// MarkComplete calls into the inner DB and registers the observed results.
func (db *ObservedDB) MarkComplete(ctx context.Context, id int) (err error) {
	ctx, endObservation := db.markCompleteOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.MarkComplete(ctx, id)
}

// MarkErrored calls into the inner DB and registers the observed results.
func (db *ObservedDB) MarkErrored(ctx context.Context, id int, failureSummary, failureStacktrace string) (err error) {
	ctx, endObservation := db.markErroredOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.MarkErrored(ctx, id, failureSummary, failureStacktrace)
}

// Dequeue calls into the inner DB and registers the observed results.
func (db *ObservedDB) Dequeue(ctx context.Context) (_ Upload, _ DB, _ bool, err error) {
	ctx, endObservation := db.dequeueOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	upload, tx, ok, err := db.db.Dequeue(ctx)
	return upload, db.wrap(tx), ok, err
}

// GetStates calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetStates(ctx context.Context, ids []int) (states map[int]string, err error) {
	ctx, endObservation := db.getStatesOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(states)), observation.Args{}) }()
	return db.db.GetStates(ctx, ids)
}

// DeleteUploadByID calls into the inner DB and registers the observed results.
func (db *ObservedDB) DeleteUploadByID(ctx context.Context, id int, getTipCommit GetTipCommitFn) (_ bool, err error) {
	ctx, endObservation := db.deleteUploadByIDOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.DeleteUploadByID(ctx, id, getTipCommit)
}

// ResetStalled calls into the inner DB and registers the observed results.
func (db *ObservedDB) ResetStalled(ctx context.Context, now time.Time) (ids []int, err error) {
	ctx, endObservation := db.resetStalledOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(ids)), observation.Args{}) }()
	return db.db.ResetStalled(ctx, now)
}

// GetDumpIDs calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetDumpIDs(ctx context.Context) (_ []int, err error) {
	ctx, endObservation := db.getDumpIDsOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.GetDumpIDs(ctx)
}

// GetDumpByID calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetDumpByID(ctx context.Context, id int) (_ Dump, _ bool, err error) {
	ctx, endObservation := db.getDumpByIDOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.GetDumpByID(ctx, id)
}

// FindClosestDumps calls into the inner DB and registers the observed results.
func (db *ObservedDB) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) (dumps []Dump, err error) {
	ctx, endObservation := db.findClosestDumpsOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(dumps)), observation.Args{}) }()
	return db.db.FindClosestDumps(ctx, repositoryID, commit, file)
}

// DeleteOldestDump calls into the inner DB and registers the observed results.
func (db *ObservedDB) DeleteOldestDump(ctx context.Context) (_ int, _ bool, err error) {
	ctx, endObservation := db.deleteOldestDumpOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.DeleteOldestDump(ctx)
}

// UpdateDumpsVisibleFromTip calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdateDumpsVisibleFromTip(ctx context.Context, repositoryID int, tipCommit string) (err error) {
	ctx, endObservation := db.updateDumpsVisibleFromTipOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.UpdateDumpsVisibleFromTip(ctx, repositoryID, tipCommit)
}

// DeleteOverlappingDumps calls into the inner DB and registers the observed results.
func (db *ObservedDB) DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) (err error) {
	ctx, endObservation := db.deleteOverlappingDumpsOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.DeleteOverlappingDumps(ctx, repositoryID, commit, root, indexer)
}

// GetPackage calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetPackage(ctx context.Context, scheme, name, version string) (_ Dump, _ bool, err error) {
	ctx, endObservation := db.getPackageOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.GetPackage(ctx, scheme, name, version)
}

// UpdatePackages calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdatePackages(ctx context.Context, packages []types.Package) (err error) {
	ctx, endObservation := db.updatePackagesOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.UpdatePackages(ctx, packages)
}

// SameRepoPager calls into the inner DB and registers the observed results.
func (db *ObservedDB) SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (_ int, _ ReferencePager, err error) {
	ctx, endObservation := db.sameRepoPagerOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.SameRepoPager(ctx, repositoryID, commit, scheme, name, version, limit)
}

// UpdatePackageReferences calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdatePackageReferences(ctx context.Context, packageReferences []types.PackageReference) (err error) {
	ctx, endObservation := db.updatePackageReferencesOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.UpdatePackageReferences(ctx, packageReferences)
}

// PackageReferencePager calls into the inner DB and registers the observed results.
func (db *ObservedDB) PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (_ int, _ ReferencePager, err error) {
	ctx, endObservation := db.packageReferencePagerOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.PackageReferencePager(ctx, scheme, name, version, repositoryID, limit)
}

// HasCommit calls into the inner DB and registers the observed results.
func (db *ObservedDB) HasCommit(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, endObservation := db.hasCommitOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.HasCommit(ctx, repositoryID, commit)
}

// UpdateCommits calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdateCommits(ctx context.Context, repositoryID int, commits map[string][]string) (err error) {
	ctx, endObservation := db.updateCommitsOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.UpdateCommits(ctx, repositoryID, commits)
}

// IndexableRepositories calls into the inner DB and registers the observed results.
func (db *ObservedDB) IndexableRepositories(ctx context.Context, opts IndexableRepositoryQueryOptions) (_ []IndexableRepository, err error) {
	ctx, endObservation := db.indexableRepositoriesOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.IndexableRepositories(ctx, opts)
}

// UpdateIndexableRepository calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdateIndexableRepository(ctx context.Context, indexableRepository UpdateableIndexableRepository) (err error) {
	ctx, endObservation := db.updateIndexableRepositoryOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.UpdateIndexableRepository(ctx, indexableRepository)
}

// GetIndexByID calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetIndexByID(ctx context.Context, id int) (_ Index, _ bool, err error) {
	ctx, endObservation := db.getIndexByIDOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.GetIndexByID(ctx, id)
}

// IndexableRepositories calls into the inner DB and registers the observed results.
func (db *ObservedDB) IndexQueueSize(ctx context.Context) (_ int, err error) {
	ctx, endObservation := db.indexQueueSizeOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.IndexQueueSize(ctx)
}

// IsQueued calls into the inner DB and registers the observed results.
func (db *ObservedDB) IsQueued(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, endObservation := db.isQueuedOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.IsQueued(ctx, repositoryID, commit)
}

// InsertIndex calls into the inner DB and registers the observed results.
func (db *ObservedDB) InsertIndex(ctx context.Context, index Index) (_ int, err error) {
	ctx, endObservation := db.insertIndexOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.InsertIndex(ctx, index)
}

// MarkIndexComplete calls into the inner DB and registers the observed results.
func (db *ObservedDB) MarkIndexComplete(ctx context.Context, id int) (err error) {
	ctx, endObservation := db.markIndexCompleteOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.MarkIndexComplete(ctx, id)
}

// MarkIndexErrored calls into the inner DB and registers the observed results.
func (db *ObservedDB) MarkIndexErrored(ctx context.Context, id int, failureSummary, failureStacktrace string) (err error) {
	ctx, endObservation := db.markIndexErroredOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.MarkIndexErrored(ctx, id, failureSummary, failureStacktrace)
}

// DequeueIndex calls into the inner DB and registers the observed results.
func (db *ObservedDB) DequeueIndex(ctx context.Context) (_ Index, _ DB, _ bool, err error) {
	ctx, endObservation := db.dequeueIndexOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.DequeueIndex(ctx)
}

// RepoUsageStatistics calls into the inner DB and registers the observed results.
func (db *ObservedDB) RepoUsageStatistics(ctx context.Context) (_ []RepoUsageStatistics, err error) {
	ctx, endObservation := db.repoUsageStatisticsOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.RepoUsageStatistics(ctx)
}

// RepoName calls into the inner DB and registers the observed results.
func (db *ObservedDB) RepoName(ctx context.Context, repositoryID int) (_ string, err error) {
	ctx, endObservation := db.repoNameOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return db.db.RepoName(ctx, repositoryID)
}
