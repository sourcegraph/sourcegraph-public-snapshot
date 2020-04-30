package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// ErrorLogger captures the method required for logging an error.
type ErrorLogger interface {
	Error(msg string, ctx ...interface{})
}

// OperationMetrics contains three common metrics for any operation.
type OperationMetrics struct {
	Duration *prometheus.HistogramVec // How long did it take?
	Count    *prometheus.CounterVec   // How many things were processed?
	Errors   *prometheus.CounterVec   // How many errors occurred?
}

// Observe registers an observation of a single operation.
func (m *OperationMetrics) Observe(secs, count float64, err error, lvals ...string) {
	if m == nil {
		return
	}

	m.Duration.WithLabelValues(lvals...).Observe(secs)
	m.Count.WithLabelValues(lvals...).Add(count)
	if err != nil {
		m.Errors.WithLabelValues(lvals...).Add(1)
	}
}

// MustRegister registers all metrics in OperationMetrics in the given prometheus.Registerer.
// It panics in case of failure.
func (m *OperationMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(m.Duration)
	r.MustRegister(m.Count)
	r.MustRegister(m.Errors)
}

// DatabaseMetrics encapsulates the Prometheus metrics of a Database.
type DatabaseMetrics struct {
	GetUploadByID             *OperationMetrics
	GetUploadsByRepo          *OperationMetrics
	Enqueue                   *OperationMetrics
	Dequeue                   *OperationMetrics
	GetStates                 *OperationMetrics
	DeleteUploadByID          *OperationMetrics
	ResetStalled              *OperationMetrics
	GetDumpByID               *OperationMetrics
	FindClosestDumps          *OperationMetrics
	DeleteOldestDump          *OperationMetrics
	UpdateDumpsVisibleFromTip *OperationMetrics
	DeleteOverlappingDumps    *OperationMetrics
	GetPackage                *OperationMetrics
	UpdatePackages            *OperationMetrics
	SameRepoPager             *OperationMetrics
	UpdatePackageReferences   *OperationMetrics
	PackageReferencePager     *OperationMetrics
	UpdateCommits             *OperationMetrics
	RepoName                  *OperationMetrics
	PageFromOffset            *OperationMetrics
}

// NewDatabaseMetrics returns DatabaseMetrics that need to be registered in a Prometheus registry.
func NewDatabaseMetrics(subsystem string) DatabaseMetrics {
	return DatabaseMetrics{
		GetUploadByID: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_upload_by_id_duration_seconds",
				Help:      "Time spent performing get upload by id queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_upload_by_id_total",
				Help:      "Total number of get upload by id queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_upload_by_id_errors_total",
				Help:      "Total number of errors when performing get upload by id queries",
			}, []string{}),
		},
		GetUploadsByRepo: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_upload_by_repo_duration_seconds",
				Help:      "Time spent performing get upload by repo queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_upload_by_repo_total",
				Help:      "Total number of get upload by repo queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_upload_by_repo_errors_total",
				Help:      "Total number of errors when performing get upload by repo queries",
			}, []string{}),
		},
		Enqueue: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_enqueue_duration_seconds",
				Help:      "Time spent performing enqueue queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_enqueue_total",
				Help:      "Total number of enqueue queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_enqueue_errors_total",
				Help:      "Total number of errors when performing enqueue queries",
			}, []string{}),
		},
		Dequeue: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_dequeue_duration_seconds",
				Help:      "Time spent performing dequeue queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_dequeue_total",
				Help:      "Total number of dequeue queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_dequeue_errors_total",
				Help:      "Total number of errors when performing dequeue queries",
			}, []string{}),
		},
		GetStates: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_states_duration_seconds",
				Help:      "Time spent performing get states queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_states_total",
				Help:      "Total number of get states queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_states_errors_total",
				Help:      "Total number of errors when performing get states queries",
			}, []string{}),
		},
		DeleteUploadByID: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_upload_by_id_duration_seconds",
				Help:      "Time spent performing delete upload by id queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_upload_by_id_total",
				Help:      "Total number of delete upload by id queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_upload_by_id_errors_total",
				Help:      "Total number of errors when performing delete upload by id queries",
			}, []string{}),
		},
		ResetStalled: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_reset_stalled_duration_seconds",
				Help:      "Time spent performing reset stalled queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_reset_stalled_total",
				Help:      "Total number of reset stalled queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_reset_stalled_errors_total",
				Help:      "Total number of errors when performing reset stalled queries",
			}, []string{}),
		},
		GetDumpByID: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_dump_by_id_duration_seconds",
				Help:      "Time spent performing get dump by id queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_dump_by_id_total",
				Help:      "Total number of get dump by id queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_dump_by_id_errors_total",
				Help:      "Total number of errors when performing get dump by id queries",
			}, []string{}),
		},
		FindClosestDumps: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_find_closest_dumps_duration_seconds",
				Help:      "Time spent performing find closest dumps queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_find_closest_dumps_total",
				Help:      "Total number of find closest dumps queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_find_closest_dumps_errors_total",
				Help:      "Total number of errors when performing find closest dumps queries",
			}, []string{}),
		},
		DeleteOldestDump: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_oldest_dump_duration_seconds",
				Help:      "Time spent performing delete oldest dump queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_oldest_dump_total",
				Help:      "Total number of delete oldest dump queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_oldest_dump_errors_total",
				Help:      "Total number of errors when performing delete oldest dump queries",
			}, []string{}),
		},
		UpdateDumpsVisibleFromTip: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_dumps_visible_from_tip_duration_seconds",
				Help:      "Time spent performing update dumps visible from tip queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_dumps_visible_from_tip_total",
				Help:      "Total number of update dumps visible from tip queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_dumps_visible_from_tip_errors_total",
				Help:      "Total number of errors when performing update dumps visible from tip queries",
			}, []string{}),
		},
		DeleteOverlappingDumps: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_overlapping_dumps_duration_seconds",
				Help:      "Time spent performing delete overlapping dumps queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_overlapping_dumps_total",
				Help:      "Total number of delete overlapping dumps queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_delete_overlapping_dumps_errors_total",
				Help:      "Total number of errors when performing delete overlapping dumps queries",
			}, []string{}),
		},
		GetPackage: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_package_duration_seconds",
				Help:      "Time spent performing get package queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_package_total",
				Help:      "Total number of get package queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_get_package_errors_total",
				Help:      "Total number of errors when performing get package queries",
			}, []string{}),
		},
		UpdatePackages: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_packages_duration_seconds",
				Help:      "Time spent performing update packages queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_packages_total",
				Help:      "Total number of update packages queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_packages_errors_total",
				Help:      "Total number of errors when performing update packages queries",
			}, []string{}),
		},
		SameRepoPager: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_same_repo_pager_duration_seconds",
				Help:      "Time spent performing same repo pager queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_same_repo_pager_total",
				Help:      "Total number of same repo pager queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_same_repo_pager_errors_total",
				Help:      "Total number of errors when performing same repo pager queries",
			}, []string{}),
		},
		UpdatePackageReferences: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_package_references_duration_seconds",
				Help:      "Time spent performing update package references queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_package_references_total",
				Help:      "Total number of update package references queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_package_references_errors_total",
				Help:      "Total number of errors when performing update package references queries",
			}, []string{}),
		},
		PackageReferencePager: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_package_reference_pager_duration_seconds",
				Help:      "Time spent performing package reference pager queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_package_reference_pager_total",
				Help:      "Total number of package reference pager queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_package_reference_pager_errors_total",
				Help:      "Total number of errors when performing package reference pager queries",
			}, []string{}),
		},
		UpdateCommits: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_commits_duration_seconds",
				Help:      "Time spent performing update commits queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_commits_total",
				Help:      "Total number of update commits queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_update_commits_errors_total",
				Help:      "Total number of errors when performing update commits queries",
			}, []string{}),
		},
		RepoName: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_repo_name_duration_seconds",
				Help:      "Time spent performing repo name queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_repo_name_total",
				Help:      "Total number of repo name queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_repo_name_errors_total",
				Help:      "Total number of errors when performing repo name queries",
			}, []string{}),
		},
		PageFromOffset: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_page_from_offset_duration_seconds",
				Help:      "Time spent performing page from offset queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_page_from_offset_total",
				Help:      "Total number of page from offset queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: subsystem,
				Name:      "db_page_from_offset_errors_total",
				Help:      "Total number of errors when performing page from offset queries",
			}, []string{}),
		},
	}
}

// An ObservedDatabase wraps another Database with error logging, Prometheus metrics, and tracing.
type ObservedDatabase struct {
	database DB
	logger   ErrorLogger
	metrics  DatabaseMetrics
	tracer   trace.Tracer
}

var _ DB = &ObservedDatabase{}

// NewObservedDatabase wraps the given DB with error logging, Prometheus metrics, and tracing.
func NewObservedDatabase(database DB, logger ErrorLogger, metrics DatabaseMetrics, tracer trace.Tracer) DB {
	return &ObservedDatabase{
		database: database,
		logger:   logger,
		metrics:  metrics,
		tracer:   tracer,
	}
}

func (db *ObservedDatabase) GetUploadByID(ctx context.Context, id int) (_ Upload, _ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.GetUploadByID", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.GetUploadByID.Observe(secs, 1, err)
		log(db.logger, "db.get-upload-by-id", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.GetUploadByID(ctx, id)
}

func (db *ObservedDatabase) GetUploadsByRepo(ctx context.Context, repositoryID int, state, term string, visibleAtTip bool, limit, offset int) (_ []Upload, _ int, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.GetUploadsByRepo", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.GetUploadsByRepo.Observe(secs, 1, err)
		log(db.logger, "db.get-uploads-by-repo", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.GetUploadsByRepo(ctx, repositoryID, state, term, visibleAtTip, limit, offset)
}

func (db *ObservedDatabase) Enqueue(ctx context.Context, commit, root, tracingContext string, repositoryID int, indexerName string) (_ int, _ TxCloser, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.Enqueue", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.Enqueue.Observe(secs, 1, err)
		log(db.logger, "db.enqueue", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.Enqueue(ctx, commit, root, tracingContext, repositoryID, indexerName)
}

func (db *ObservedDatabase) Dequeue(ctx context.Context) (_ Upload, _ JobHandle, _ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.Dequeue", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.Dequeue.Observe(secs, 1, err)
		log(db.logger, "db.dequeue", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	// TODO - observe the job handle
	return db.database.Dequeue(ctx)
}

func (db *ObservedDatabase) GetStates(ctx context.Context, ids []int) (_ map[int]string, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.GetStates", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.GetStates.Observe(secs, 1, err)
		log(db.logger, "db.get-states", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.GetStates(ctx, ids)
}

func (db *ObservedDatabase) DeleteUploadByID(ctx context.Context, id int, getTipCommit func(repositoryID int) (string, error)) (_ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.DeleteUploadByID", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.DeleteUploadByID.Observe(secs, 1, err)
		log(db.logger, "db.delete-upload-by-id", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.DeleteUploadByID(ctx, id, getTipCommit)
}

func (db *ObservedDatabase) ResetStalled(ctx context.Context, now time.Time) (_ []int, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.ResetStalled", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.ResetStalled.Observe(secs, 1, err)
		log(db.logger, "db.reset-stalled", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.ResetStalled(ctx, now)
}

func (db *ObservedDatabase) GetDumpByID(ctx context.Context, id int) (_ Dump, _ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.GetDumpByID", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.GetDumpByID.Observe(secs, 1, err)
		log(db.logger, "db.get-dump-by-id", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.GetDumpByID(ctx, id)
}

func (db *ObservedDatabase) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) (_ []Dump, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.FindClosestDumps", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.FindClosestDumps.Observe(secs, 1, err)
		log(db.logger, "db.find-closest-dumps", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.FindClosestDumps(ctx, repositoryID, commit, file)
}

func (db *ObservedDatabase) DeleteOldestDump(ctx context.Context) (_ int, _ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.DeleteOldestDump", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.DeleteOldestDump.Observe(secs, 1, err)
		log(db.logger, "db.delete-oldest-dump", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.DeleteOldestDump(ctx)
}

func (db *ObservedDatabase) UpdateDumpsVisibleFromTip(ctx context.Context, tx *sql.Tx, repositoryID int, tipCommit string) (err error) {
	tr, ctx := db.tracer.New(ctx, "DB.UpdateDumpsVisibleFromTip", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.UpdateDumpsVisibleFromTip.Observe(secs, 1, err)
		log(db.logger, "db.update-dumps-visible-from-tip", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.UpdateDumpsVisibleFromTip(ctx, tx, repositoryID, tipCommit)
}

func (db *ObservedDatabase) DeleteOverlappingDumps(ctx context.Context, tx *sql.Tx, repositoryID int, commit, root, indexer string) (err error) {
	tr, ctx := db.tracer.New(ctx, "DB.DeleteOverlappingDumps", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.DeleteOverlappingDumps.Observe(secs, 1, err)
		log(db.logger, "db.delete-overlapping-dumps", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.DeleteOverlappingDumps(ctx, tx, repositoryID, commit, root, indexer)
}

func (db *ObservedDatabase) GetPackage(ctx context.Context, scheme, name, version string) (_ Dump, _ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.GetPackage", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.GetPackage.Observe(secs, 1, err)
		log(db.logger, "db.get-package", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.GetPackage(ctx, scheme, name, version)
}

func (db *ObservedDatabase) UpdatePackages(ctx context.Context, tx *sql.Tx, packages []types.Package) (err error) {
	tr, ctx := db.tracer.New(ctx, "DB.UpdatePackages", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.UpdatePackages.Observe(secs, 1, err)
		log(db.logger, "db.update-packages", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.UpdatePackages(ctx, tx, packages)
}

func (db *ObservedDatabase) SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (_ int, _ ReferencePager, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.SameRepoPager", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.SameRepoPager.Observe(secs, 1, err)
		log(db.logger, "db.same-repo-pager", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	totalCount, pager, err := db.database.SameRepoPager(ctx, repositoryID, commit, scheme, name, version, limit)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, NewObservedReferencePager(pager, db.logger, db.metrics, db.tracer), nil
}

func (db *ObservedDatabase) UpdatePackageReferences(ctx context.Context, tx *sql.Tx, packageReferences []types.PackageReference) (err error) {
	tr, ctx := db.tracer.New(ctx, "DB.UpdatePackageReferences", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.UpdatePackageReferences.Observe(secs, 1, err)
		log(db.logger, "db.update-package-references", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.UpdatePackageReferences(ctx, tx, packageReferences)
}

func (db *ObservedDatabase) PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (_ int, _ ReferencePager, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.PackageReferencePager", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.PackageReferencePager.Observe(secs, 1, err)
		log(db.logger, "db.package-reference-pager", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	totalCount, pager, err := db.database.PackageReferencePager(ctx, scheme, name, version, repositoryID, limit)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, NewObservedReferencePager(pager, db.logger, db.metrics, db.tracer), nil
}

func (db *ObservedDatabase) UpdateCommits(ctx context.Context, tx *sql.Tx, repositoryID int, commits map[string][]string) (err error) {
	tr, ctx := db.tracer.New(ctx, "DB.UpdateCommits", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.UpdateCommits.Observe(secs, 1, err)
		log(db.logger, "db.update-commits", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.UpdateCommits(ctx, tx, repositoryID, commits)
}

func (db *ObservedDatabase) RepoName(ctx context.Context, repositoryID int) (_ string, err error) {
	tr, ctx := db.tracer.New(ctx, "DB.RepoName", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.RepoName.Observe(secs, 1, err)
		log(db.logger, "db.repo-name", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.RepoName(ctx, repositoryID)
}

// An ObservedReferencePager wraps another ReferencePager with error logging, Prometheus metrics, and tracing.
type ObservedReferencePager struct {
	pager   ReferencePager
	logger  ErrorLogger
	metrics DatabaseMetrics
	tracer  trace.Tracer
}

var _ ReferencePager = &ObservedReferencePager{}

// NewObservedReferencePager wraps the given ReferencePager with error logging, Prometheus metrics, and tracing.
func NewObservedReferencePager(pager ReferencePager, logger ErrorLogger, metrics DatabaseMetrics, tracer trace.Tracer) ReferencePager {
	return &ObservedReferencePager{
		pager:   pager,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

func (p *ObservedReferencePager) CloseTx(err error) error {
	return p.pager.CloseTx(err)
}

func (p *ObservedReferencePager) PageFromOffset(ctx context.Context, offset int) (_ []types.PackageReference, err error) {
	tr, ctx := p.tracer.New(ctx, "ReferencePager.PageFromOffset", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		p.metrics.PageFromOffset.Observe(secs, 1, err)
		log(p.logger, "reference-pager.page-from-offset", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return p.pager.PageFromOffset(ctx, offset)
}

func log(lg ErrorLogger, msg string, err error, ctx ...interface{}) {
	if err == nil {
		return
	}

	lg.Error(msg, append(append(make([]interface{}, 0, len(ctx)+2), "error", err), ctx...)...)
}
