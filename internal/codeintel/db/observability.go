package db

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// DBMetrics encapsulates the Prometheus metrics of a DB.
type DBMetrics struct {
	GetUploadByID             *metrics.OperationMetrics
	GetUploadsByRepo          *metrics.OperationMetrics
	Enqueue                   *metrics.OperationMetrics
	Dequeue                   *metrics.OperationMetrics
	GetStates                 *metrics.OperationMetrics
	DeleteUploadByID          *metrics.OperationMetrics
	ResetStalled              *metrics.OperationMetrics
	GetDumpByID               *metrics.OperationMetrics
	FindClosestDumps          *metrics.OperationMetrics
	DeleteOldestDump          *metrics.OperationMetrics
	UpdateDumpsVisibleFromTip *metrics.OperationMetrics
	DeleteOverlappingDumps    *metrics.OperationMetrics
	GetPackage                *metrics.OperationMetrics
	UpdatePackages            *metrics.OperationMetrics
	SameRepoPager             *metrics.OperationMetrics
	UpdatePackageReferences   *metrics.OperationMetrics
	PackageReferencePager     *metrics.OperationMetrics
	UpdateCommits             *metrics.OperationMetrics
	RepoName                  *metrics.OperationMetrics
}

// MustRegister registers all metrics in DBMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (dbm DBMetrics) MustRegister(r prometheus.Registerer) {
	for _, om := range []*metrics.OperationMetrics{
		dbm.GetUploadByID,
		dbm.GetUploadsByRepo,
		dbm.Enqueue,
		dbm.Dequeue,
		dbm.GetStates,
		dbm.DeleteUploadByID,
		dbm.ResetStalled,
		dbm.GetDumpByID,
		dbm.FindClosestDumps,
		dbm.DeleteOldestDump,
		dbm.UpdateDumpsVisibleFromTip,
		dbm.DeleteOverlappingDumps,
		dbm.GetPackage,
		dbm.UpdatePackages,
		dbm.SameRepoPager,
		dbm.UpdatePackageReferences,
		dbm.PackageReferencePager,
		dbm.UpdateCommits,
		dbm.RepoName,
	} {
		om.MustRegister(prometheus.DefaultRegisterer)
	}
}

// NewDBMetrics returns DBMetrics that need to be registered in a Prometheus registry.
func NewDBMetrics(subsystem string) DBMetrics {
	return DBMetrics{
		GetUploadByID:             metrics.NewOperationMetrics(subsystem, "db", "get_upload_by_id"),
		GetUploadsByRepo:          metrics.NewOperationMetrics(subsystem, "db", "get_uploads_by_repo"), // TODO
		Enqueue:                   metrics.NewOperationMetrics(subsystem, "db", "enqueue"),
		Dequeue:                   metrics.NewOperationMetrics(subsystem, "db", "dequeue"),
		GetStates:                 metrics.NewOperationMetrics(subsystem, "db", "get_states"), // TODO
		DeleteUploadByID:          metrics.NewOperationMetrics(subsystem, "db", "delete_upload_by_id"),
		ResetStalled:              metrics.NewOperationMetrics(subsystem, "db", "reset_stalled"), // TODO
		GetDumpByID:               metrics.NewOperationMetrics(subsystem, "db", "get_dump_by_id"),
		FindClosestDumps:          metrics.NewOperationMetrics(subsystem, "db", "find_closest_dumps"), // TODO
		DeleteOldestDump:          metrics.NewOperationMetrics(subsystem, "db", "delete_oldest_dump"),
		UpdateDumpsVisibleFromTip: metrics.NewOperationMetrics(subsystem, "db", "update_dumps_visible_from_tip"),
		DeleteOverlappingDumps:    metrics.NewOperationMetrics(subsystem, "db", "delete_overlapping_dumps"),
		GetPackage:                metrics.NewOperationMetrics(subsystem, "db", "get_package"),
		UpdatePackages:            metrics.NewOperationMetrics(subsystem, "db", "update_packages"),
		SameRepoPager:             metrics.NewOperationMetrics(subsystem, "db", "same_repo_pager"),
		UpdatePackageReferences:   metrics.NewOperationMetrics(subsystem, "db", "update_package_references"),
		PackageReferencePager:     metrics.NewOperationMetrics(subsystem, "db", "package_reference_pager"),
		UpdateCommits:             metrics.NewOperationMetrics(subsystem, "db", "update_commits"),
		RepoName:                  metrics.NewOperationMetrics(subsystem, "db", "repo_name"),
	}
}

// An ObservedDB wraps another DB with error logging, Prometheus metrics, and tracing.
type ObservedDB struct {
	db      DB
	logger  logging.ErrorLogger
	metrics DBMetrics
	tracer  trace.Tracer
}

var _ DB = &ObservedDB{}

// NewObservedDB wraps the given DB with error logging, Prometheus metrics, and tracing.
func NewObservedDB(db DB, logger logging.ErrorLogger, metrics DBMetrics, tracer trace.Tracer) DB {
	return &ObservedDB{
		db:      db,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// Transact calls into the inner DB.
func (db *ObservedDB) Transact(ctx context.Context) (DB, error) {
	return db.db.Transact(ctx)
}

// Done calls into the inner DB.
func (db *ObservedDB) Done(err error) error {
	return db.db.Done(err)
}

// GetUploadByID calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetUploadByID(ctx context.Context, id int) (_ Upload, _ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.GetUploadByID, "DB.GetUploadByID", "db.get-upload-by-id")
	defer endObservation(1)

	return db.db.GetUploadByID(ctx, id)
}

// GetUploadsByRepo calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetUploadsByRepo(ctx context.Context, repositoryID int, state, term string, visibleAtTip bool, limit, offset int) (uploads []Upload, _ int, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.GetUploadsByRepo, "DB.GetUploadsByRepo", "db.get-uploads-by-repo")
	defer func() {
		endObservation(float64(len(uploads)))
	}()

	return db.db.GetUploadsByRepo(ctx, repositoryID, state, term, visibleAtTip, limit, offset)
}

// Enqueue calls into the inner DB and registers the observed results.
func (db *ObservedDB) Enqueue(ctx context.Context, commit, root, tracingContext string, repositoryID int, indexerName string) (_ int, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.Enqueue, "DB.Enqueue", "db.enqueue")
	defer endObservation(1)

	return db.db.Enqueue(ctx, commit, root, tracingContext, repositoryID, indexerName)
}

// Dequeue calls into the inner DB and registers the observed results.
func (db *ObservedDB) Dequeue(ctx context.Context) (_ Upload, _ JobHandle, _ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.Dequeue, "DB.Dequeue", "db.dequeue")
	defer endObservation(1)

	return db.db.Dequeue(ctx)
}

// GetStates calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetStates(ctx context.Context, ids []int) (states map[int]string, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.GetStates, "DB.GetStates", "db.get-states")
	defer func() {
		endObservation(float64(len(states)))
	}()

	return db.db.GetStates(ctx, ids)
}

// DeleteUploadByID calls into the inner DB and registers the observed results.
func (db *ObservedDB) DeleteUploadByID(ctx context.Context, id int, getTipCommit GetTipCommitFn) (_ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.DeleteUploadByID, "DB.DeleteUploadByID", "db.delete-upload-by-id")
	defer endObservation(1)

	return db.db.DeleteUploadByID(ctx, id, getTipCommit)
}

// ResetStalled calls into the inner DB and registers the observed results.
func (db *ObservedDB) ResetStalled(ctx context.Context, now time.Time) (ids []int, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.ResetStalled, "DB.ResetStalled", "db.reset-stalled")
	defer func() {
		endObservation(float64(len(ids)))
	}()

	return db.db.ResetStalled(ctx, now)
}

// GetDumpByID calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetDumpByID(ctx context.Context, id int) (_ Dump, _ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.GetDumpByID, "DB.GetDumpByID", "db.get-dump-by-id")
	defer endObservation(1)

	return db.db.GetDumpByID(ctx, id)
}

// FindClosestDumps calls into the inner DB and registers the observed results.
func (db *ObservedDB) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) (dumps []Dump, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.FindClosestDumps, "DB.FindClosestDumps", "db.find-closest-dumps")
	defer func() {
		endObservation(float64(len(dumps)))
	}()

	return db.db.FindClosestDumps(ctx, repositoryID, commit, file)
}

// DeleteOldestDump calls into the inner DB and registers the observed results.
func (db *ObservedDB) DeleteOldestDump(ctx context.Context) (_ int, _ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.DeleteOldestDump, "DB.DeleteOldestDump", "db.delete-oldest-dump")
	defer endObservation(1)

	return db.db.DeleteOldestDump(ctx)
}

// UpdateDumpsVisibleFromTip calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdateDumpsVisibleFromTip(ctx context.Context, repositoryID int, tipCommit string) (err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.UpdateDumpsVisibleFromTip, "DB.UpdateDumpsVisibleFromTip", "db.update-dumps-visible-from-tip")
	defer endObservation(1)

	return db.db.UpdateDumpsVisibleFromTip(ctx, repositoryID, tipCommit)
}

// DeleteOverlappingDumps calls into the inner DB and registers the observed results.
func (db *ObservedDB) DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) (err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.DeleteOverlappingDumps, "DB.DeleteOverlappingDumps", "db.delete-overlapping-dumps")
	defer endObservation(1)

	return db.db.DeleteOverlappingDumps(ctx, repositoryID, commit, root, indexer)
}

// GetPackage calls into the inner DB and registers the observed results.
func (db *ObservedDB) GetPackage(ctx context.Context, scheme, name, version string) (_ Dump, _ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.GetPackage, "DB.GetPackage", "db.get-package")
	defer endObservation(1)

	return db.db.GetPackage(ctx, scheme, name, version)
}

// UpdatePackages calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdatePackages(ctx context.Context, packages []types.Package) (err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.UpdatePackages, "DB.UpdatePackages", "db.update-packages")
	defer endObservation(1)

	return db.db.UpdatePackages(ctx, packages)
}

// SameRepoPager calls into the inner DB and registers the observed results.
func (db *ObservedDB) SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (_ int, _ ReferencePager, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.SameRepoPager, "DB.SameRepoPager", "db.same-repo-pager")
	defer endObservation(1)

	return db.db.SameRepoPager(ctx, repositoryID, commit, scheme, name, version, limit)
}

// UpdatePackageReferences calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdatePackageReferences(ctx context.Context, packageReferences []types.PackageReference) (err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.UpdatePackageReferences, "DB.UpdatePackageReferences", "db.update-package-references")
	defer endObservation(1)

	return db.db.UpdatePackageReferences(ctx, packageReferences)
}

// PackageReferencePager calls into the inner DB and registers the observed results.
func (db *ObservedDB) PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (_ int, _ ReferencePager, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.PackageReferencePager, "DB.PackageReferencePager", "db.pacakge-reference-pager")
	defer endObservation(1)

	return db.db.PackageReferencePager(ctx, scheme, name, version, repositoryID, limit)
}

// UpdateCommits calls into the inner DB and registers the observed results.
func (db *ObservedDB) UpdateCommits(ctx context.Context, repositoryID int, commits map[string][]string) (err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.UpdateCommits, "DB.UpdateCommits", "db.update-commits")
	defer endObservation(1)

	return db.db.UpdateCommits(ctx, repositoryID, commits)
}

// RepoName calls into the inner DB and registers the observed results.
func (db *ObservedDB) RepoName(ctx context.Context, repositoryID int) (_ string, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.RepoName, "DB.RepoName", "db.repo-name")
	defer endObservation(1)

	return db.db.RepoName(ctx, repositoryID)
}

func (db *ObservedDB) prepObservation(
	ctx context.Context,
	err *error,
	metrics *metrics.OperationMetrics,
	traceName string,
	logName string,
	preFields ...log.Field,
) (context.Context, observation.FinishFn) {
	return observation.With(ctx, observation.Args{
		Logger:    db.logger,
		Metrics:   metrics,
		Tracer:    &db.tracer,
		Err:       err,
		TraceName: traceName,
		LogName:   logName,
		LogFields: preFields,
	})
}
