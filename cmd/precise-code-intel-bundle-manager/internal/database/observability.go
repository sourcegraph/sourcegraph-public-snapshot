package database

import (
	"context"
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
	Exists             *OperationMetrics
	Definitions        *OperationMetrics
	References         *OperationMetrics
	Hover              *OperationMetrics
	MonikersByPosition *OperationMetrics
	MonikerResults     *OperationMetrics
	PackageInformation *OperationMetrics
}

// NewDatabaseMetrics returns DatabaseMetrics that need to be registered in a Prometheus registry.
func NewDatabaseMetrics() DatabaseMetrics {
	return DatabaseMetrics{
		Exists: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_exists",
				Help:      "Time spent on exists queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_exists",
				Help:      "Total number of exists queries performed",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_exists",
				Help:      "Total number of errors when performing exists queries",
			}, []string{}),
		},
		Definitions: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_definitions",
				Help:      "Time spent on definitions queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_definitions",
				Help:      "Total number of definitions returned",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_definitions",
				Help:      "Total number of errors when performing definitions queries",
			}, []string{}),
		},
		References: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_references",
				Help:      "Time spent on references queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_references",
				Help:      "Total number of references returned",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_references",
				Help:      "Total number of errors when performing references queries",
			}, []string{}),
		},
		Hover: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_hover",
				Help:      "Time spent on hover queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_hover",
				Help:      "Total number of hover results returned",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_hover",
				Help:      "Total number of errors when performing hover queries",
			}, []string{}),
		},
		MonikersByPosition: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_monikers_by_position",
				Help:      "Time spent on moniker queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_monikers_by_position",
				Help:      "Total number of monikers returned",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_monikers_by_position",
				Help:      "Total number of errors when performing moniker queries",
			}, []string{}),
		},
		MonikerResults: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_moniker_results",
				Help:      "Time spent on moniker result queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_moniker_results",
				Help:      "Total number of moniker results returned",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_moniker_results",
				Help:      "Total number of errors when performing moniker result queries",
			}, []string{}),
		},
		PackageInformation: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_package_information",
				Help:      "Time spent on package information queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_package_information",
				Help:      "Total number of package informations returned",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-bundle-manager",
				Name:      "database_package_information",
				Help:      "Total number of errors when performing package information queries",
			}, []string{}),
		},
	}
}

// An ObservedDatabase wraps another Database with error logging, Prometheus metrics, and tracing.
type ObservedDatabase struct {
	database Database
	logger   ErrorLogger
	metrics  DatabaseMetrics
	tracer   trace.Tracer
}

var _ Database = &ObservedDatabase{}

// NewObservedDatabase wraps the given Database with error logging, Prometheus metrics, and tracing.
func NewObservedDatabase(database Database, logger ErrorLogger, metrics DatabaseMetrics, tracer trace.Tracer) Database {
	return &ObservedDatabase{
		database: database,
		logger:   logger,
		metrics:  metrics,
		tracer:   tracer,
	}
}

func (db *ObservedDatabase) Exists(ctx context.Context, path string) (_ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "Database.Exists", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.Exists.Observe(secs, 1, err)
		log(db.logger, "database.exists", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.Exists(ctx, path)
}

func (db *ObservedDatabase) Definitions(ctx context.Context, path string, line, character int) (definitions []Location, err error) {
	tr, ctx := db.tracer.New(ctx, "Database.Definitions", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.Definitions.Observe(secs, float64(len(definitions)), err)
		log(db.logger, "database.definitions", err, "count", len(definitions))
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.Definitions(ctx, path, line, character)
}

func (db *ObservedDatabase) References(ctx context.Context, path string, line, character int) (references []Location, err error) {
	tr, ctx := db.tracer.New(ctx, "Database.References", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.References.Observe(secs, float64(len(references)), err)
		log(db.logger, "database.references", err, "count", len(references))
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.References(ctx, path, line, character)
}

func (db *ObservedDatabase) Hover(ctx context.Context, path string, line, character int) (_ string, _ Range, _ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "Database.Hover", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.Hover.Observe(secs, 1, err)
		log(db.logger, "database.hover", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.Hover(ctx, path, line, character)
}

func (db *ObservedDatabase) MonikersByPosition(ctx context.Context, path string, line, character int) (monikers [][]types.MonikerData, err error) {
	tr, ctx := db.tracer.New(ctx, "Database.MonikersByPosition", "")
	defer func(began time.Time) {
		count := 0
		for _, monikerGroup := range monikers {
			count += len(monikerGroup)
		}

		secs := time.Since(began).Seconds()
		db.metrics.MonikersByPosition.Observe(secs, float64(count), err)
		log(db.logger, "database.monikers-by-position", err, "count", count)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.MonikersByPosition(ctx, path, line, character)
}

func (db *ObservedDatabase) MonikerResults(ctx context.Context, tableName, scheme, identifier string, skip, take int) (locations []Location, _ int, err error) {
	tr, ctx := db.tracer.New(ctx, "Database.MonikerResults", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.MonikerResults.Observe(secs, float64(len(locations)), err)
		log(db.logger, "database.moniker-results", err, "count", len(locations))
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.MonikerResults(ctx, tableName, scheme, identifier, skip, take)
}

func (db *ObservedDatabase) PackageInformation(ctx context.Context, path string, packageInformationID types.ID) (_ types.PackageInformationData, _ bool, err error) {
	tr, ctx := db.tracer.New(ctx, "Database.PackageInformation", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		db.metrics.PackageInformation.Observe(secs, 1, err)
		log(db.logger, "database.package-information", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return db.database.PackageInformation(ctx, path, packageInformationID)
}

func (db *ObservedDatabase) Close() error {
	return db.database.Close()
}

func log(lg ErrorLogger, msg string, err error, ctx ...interface{}) {
	if err == nil {
		return
	}

	lg.Error(msg, append(append(make([]interface{}, 0, len(ctx)+2), "error", err), ctx...)...)
}
