package database

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// DatabaseMetrics encapsulates the Prometheus metrics of a Database.
type DatabaseMetrics struct {
	Exists             *metrics.OperationMetrics
	Definitions        *metrics.OperationMetrics
	References         *metrics.OperationMetrics
	Hover              *metrics.OperationMetrics
	MonikersByPosition *metrics.OperationMetrics
	MonikerResults     *metrics.OperationMetrics
	PackageInformation *metrics.OperationMetrics
}

// MustRegister registers all metrics in DatabaseMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (dm DatabaseMetrics) MustRegister(r prometheus.Registerer) {
	for _, om := range []*metrics.OperationMetrics{
		dm.Exists,
		dm.Definitions,
		dm.References,
		dm.Hover,
		dm.MonikersByPosition,
		dm.MonikerResults,
		dm.PackageInformation,
	} {
		om.MustRegister(prometheus.DefaultRegisterer)
	}
}

// NewDatabaseMetrics returns DatabaseMetrics that need to be registered in a Prometheus registry.
func NewDatabaseMetrics(subsystem string) DatabaseMetrics {
	return DatabaseMetrics{
		Exists:             metrics.NewOperationMetrics(subsystem, "database", "exists"),
		Definitions:        metrics.NewOperationMetrics(subsystem, "database", "definitions", metrics.WithCountHelp("The total number of definitions returned")),
		References:         metrics.NewOperationMetrics(subsystem, "database", "references", metrics.WithCountHelp("The total number of references returned")),
		Hover:              metrics.NewOperationMetrics(subsystem, "database", "hover"),
		MonikersByPosition: metrics.NewOperationMetrics(subsystem, "database", "monikers_by_position", metrics.WithCountHelp("The total number of monikers returned")),
		MonikerResults:     metrics.NewOperationMetrics(subsystem, "database", "moniker_results", metrics.WithCountHelp("The total number of locations returned")),
		PackageInformation: metrics.NewOperationMetrics(subsystem, "database", "package_information"),
	}
}

// An ObservedDatabase wraps another Database with error logging, Prometheus metrics, and tracing.
type ObservedDatabase struct {
	database Database
	logger   logging.ErrorLogger
	metrics  DatabaseMetrics
	tracer   trace.Tracer
}

var _ Database = &ObservedDatabase{}

// NewObservedDatabase wraps the given Database with error logging, Prometheus metrics, and tracing.
func NewObservedDatabase(database Database, logger logging.ErrorLogger, metrics DatabaseMetrics, tracer trace.Tracer) Database {
	return &ObservedDatabase{
		database: database,
		logger:   logger,
		metrics:  metrics,
		tracer:   tracer,
	}
}

// Close calls into the inner Database.
func (db *ObservedDatabase) Close() error {
	// TODO - trace
	return db.database.Close()
}

// Exists calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Exists(ctx context.Context, path string) (_ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.Exists, "Database.Exists", "database.exists")
	defer endObservation(1)

	return db.database.Exists(ctx, path)
}

// Definitions calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Definitions(ctx context.Context, path string, line, character int) (definitions []Location, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.Definitions, "Database.Definitions", "database.definitions")
	defer func() {
		endObservation(float64(len(definitions)))
	}()

	return db.database.Definitions(ctx, path, line, character)
}

// References calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) References(ctx context.Context, path string, line, character int) (references []Location, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.References, "Database.References", "database.references")
	defer func() {
		endObservation(float64(len(references)))
	}()

	return db.database.References(ctx, path, line, character)
}

// Hover calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Hover(ctx context.Context, path string, line, character int) (_ string, _ Range, _ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.Hover, "Database.Hover", "database.hover")
	defer endObservation(1)

	return db.database.Hover(ctx, path, line, character)
}

// MonikersByPosition calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) MonikersByPosition(ctx context.Context, path string, line, character int) (monikers [][]types.MonikerData, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.MonikersByPosition, "Database.MonikersByPosition", "database.monikers-by-position")
	defer func() {
		count := 0
		for _, group := range monikers {
			count += len(group)
		}

		endObservation(float64(count))
	}()

	return db.database.MonikersByPosition(ctx, path, line, character)
}

// MonikerResults calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) MonikerResults(ctx context.Context, tableName, scheme, identifier string, skip, take int) (locations []Location, _ int, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.MonikerResults, "Database.MonikerResults", "database.moniker-results")
	defer func() {
		endObservation(float64(len(locations)))
	}()

	return db.database.MonikerResults(ctx, tableName, scheme, identifier, skip, take)
}

// PackageInformation calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) PackageInformation(ctx context.Context, path string, packageInformationID types.ID) (_ types.PackageInformationData, _ bool, err error) {
	ctx, endObservation := db.prepObservation(ctx, &err, db.metrics.PackageInformation, "Database.PackageInformation", "database.package-information")
	defer endObservation(1)

	return db.database.PackageInformation(ctx, path, packageInformationID)
}

func (db *ObservedDatabase) prepObservation(
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
