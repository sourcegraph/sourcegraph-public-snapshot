package database

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// An ObservedDatabase wraps another Database with error logging, Prometheus metrics, and tracing.
type ObservedDatabase struct {
	database                    Database
	filename                    string
	existsOperation             *observation.Operation
	definitionsOperation        *observation.Operation
	referencesOperation         *observation.Operation
	hoverOperation              *observation.Operation
	monikersByPositionOperation *observation.Operation
	monikerResultsOperation     *observation.Operation
	packageInformationOperation *observation.Operation
}

var _ Database = &ObservedDatabase{}

// singletonMetrics ensures that the operation metrics required by ObservedDatabase are
// constructed only once as there may be many databases instantiated by a single replica
// of precise-code-intel-bundle-manager.
var singletonMetrics = &metrics.SingletonOperationMetrics{}

// NewObservedDatabase wraps the given Database with error logging, Prometheus metrics, and tracing.
func NewObserved(database Database, filename string, observationContext *observation.Context) Database {
	metrics := singletonMetrics.Get(func() *metrics.OperationMetrics {
		return metrics.NewOperationMetrics(
			observationContext.Registerer,
			"bundle_database",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of results returned"),
		)
	})

	return &ObservedDatabase{
		database: database,
		filename: filename,
		existsOperation: observationContext.Operation(observation.Op{
			Name:         "Database.Exists",
			MetricLabels: []string{"exists"},
			Metrics:      metrics,
		}),
		definitionsOperation: observationContext.Operation(observation.Op{
			Name:         "Database.Definitions",
			MetricLabels: []string{"definitions"},
			Metrics:      metrics,
		}),
		referencesOperation: observationContext.Operation(observation.Op{
			Name:         "Database.References",
			MetricLabels: []string{"references"},
			Metrics:      metrics,
		}),
		hoverOperation: observationContext.Operation(observation.Op{
			Name:         "Database.Hover",
			MetricLabels: []string{"hover"},
			Metrics:      metrics,
		}),
		monikersByPositionOperation: observationContext.Operation(observation.Op{
			Name:         "Database.MonikersByPosition",
			MetricLabels: []string{"monikers_by_position"},
			Metrics:      metrics,
		}),
		monikerResultsOperation: observationContext.Operation(observation.Op{
			Name:         "Database.MonikerResults",
			MetricLabels: []string{"moniker_results"},
			Metrics:      metrics,
		}),
		packageInformationOperation: observationContext.Operation(observation.Op{
			Name:         "Database.PackageInformation",
			MetricLabels: []string{"package_information"},
			Metrics:      metrics,
		}),
	}
}

// Close calls into the inner Database.
func (db *ObservedDatabase) Close() error {
	return db.database.Close()
}

// Exists calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Exists(ctx context.Context, path string) (_ bool, err error) {
	ctx, endObservation := db.existsOperation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("filename", db.filename),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})
	return db.database.Exists(ctx, path)
}

// Definitions calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Definitions(ctx context.Context, path string, line, character int) (definitions []Location, err error) {
	ctx, endObservation := db.definitionsOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("filename", db.filename),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer func() { endObservation(float64(len(definitions)), observation.Args{}) }()
	return db.database.Definitions(ctx, path, line, character)
}

// References calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) References(ctx context.Context, path string, line, character int) (references []Location, err error) {
	ctx, endObservation := db.referencesOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("filename", db.filename),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer func() { endObservation(float64(len(references)), observation.Args{}) }()
	return db.database.References(ctx, path, line, character)
}

// Hover calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Hover(ctx context.Context, path string, line, character int) (_ string, _ Range, _ bool, err error) {
	ctx, endObservation := db.hoverOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("filename", db.filename),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation(1, observation.Args{})
	return db.database.Hover(ctx, path, line, character)
}

// MonikersByPosition calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) MonikersByPosition(ctx context.Context, path string, line, character int) (monikers [][]types.MonikerData, err error) {
	ctx, endObservation := db.monikersByPositionOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("filename", db.filename),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer func() {
		count := 0
		for _, group := range monikers {
			count += len(group)
		}

		endObservation(float64(count), observation.Args{})
	}()

	return db.database.MonikersByPosition(ctx, path, line, character)
}

// MonikerResults calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) MonikerResults(ctx context.Context, tableName, scheme, identifier string, skip, take int) (locations []Location, _ int, err error) {
	ctx, endObservation := db.monikerResultsOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("filename", db.filename),
			log.String("tableName", tableName),
			log.String("scheme", scheme),
			log.String("identifier", identifier),
		},
	})
	defer func() { endObservation(float64(len(locations)), observation.Args{}) }()
	return db.database.MonikerResults(ctx, tableName, scheme, identifier, skip, take)
}

// PackageInformation calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) PackageInformation(ctx context.Context, path string, packageInformationID types.ID) (_ types.PackageInformationData, _ bool, err error) {
	ctx, endObservation := db.packageInformationOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("filename", db.filename),
			log.String("path", path),
			log.String("packageInformationId", string(packageInformationID)),
		},
	})
	defer endObservation(1, observation.Args{})
	return db.database.PackageInformation(ctx, path, packageInformationID)
}
