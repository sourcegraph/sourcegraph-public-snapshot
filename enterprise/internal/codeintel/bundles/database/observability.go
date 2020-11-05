package database

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client_types"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// An ObservedDatabase wraps another Database with error logging, Prometheus metrics, and tracing.
type ObservedDatabase struct {
	database                    Database
	existsOperation             *observation.Operation
	rangesOperation             *observation.Operation
	definitionsOperation        *observation.Operation
	referencesOperation         *observation.Operation
	hoverOperation              *observation.Operation
	diagnosticsOperation        *observation.Operation
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
func NewObserved(database Database, observationContext *observation.Context) Database {
	metrics := singletonMetrics.Get(func() *metrics.OperationMetrics {
		return metrics.NewOperationMetrics(
			observationContext.Registerer,
			"code_intel_bundle_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of results returned"),
		)
	})

	return &ObservedDatabase{
		database: database,
		existsOperation: observationContext.Operation(observation.Op{
			Name:         "Database.Exists",
			MetricLabels: []string{"exists"},
			Metrics:      metrics,
		}),
		rangesOperation: observationContext.Operation(observation.Op{
			Name:         "Database.Ranges",
			MetricLabels: []string{"ranges"},
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
		diagnosticsOperation: observationContext.Operation(observation.Op{
			Name:         "Database.Diagnostics",
			MetricLabels: []string{"diagnostics"},
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

// Exists calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Exists(ctx context.Context, bundleID int, path string) (_ bool, err error) {
	ctx, endObservation := db.existsOperation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})
	return db.database.Exists(ctx, bundleID, path)
}

// Ranges calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) (ranges []bundles.CodeIntelligenceRange, err error) {
	ctx, endObservation := db.rangesOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("startLine", startLine),
			log.Int("endLine", endLine),
		},
	})
	defer func() { endObservation(float64(len(ranges)), observation.Args{}) }()
	return db.database.Ranges(ctx, bundleID, path, startLine, endLine)
}

// Definitions calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Definitions(ctx context.Context, bundleID int, path string, line, character int) (definitions []bundles.Location, err error) {
	ctx, endObservation := db.definitionsOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer func() { endObservation(float64(len(definitions)), observation.Args{}) }()
	return db.database.Definitions(ctx, bundleID, path, line, character)
}

// References calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) References(ctx context.Context, bundleID int, path string, line, character int) (references []bundles.Location, err error) {
	ctx, endObservation := db.referencesOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer func() { endObservation(float64(len(references)), observation.Args{}) }()
	return db.database.References(ctx, bundleID, path, line, character)
}

// Hover calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Hover(ctx context.Context, bundleID int, path string, line, character int) (_ string, _ bundles.Range, _ bool, err error) {
	ctx, endObservation := db.hoverOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation(1, observation.Args{})
	return db.database.Hover(ctx, bundleID, path, line, character)
}

// Diagnostics calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) Diagnostics(ctx context.Context, bundleID int, prefix string, skip, take int) (diagnostics []bundles.Diagnostic, _ int, err error) {
	ctx, endObservation := db.hoverOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("prefix", prefix),
		},
	})
	defer func() { endObservation(float64(len(diagnostics)), observation.Args{}) }()
	return db.database.Diagnostics(ctx, bundleID, prefix, skip, take)
}

// MonikersByPosition calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (monikers [][]bundles.MonikerData, err error) {
	ctx, endObservation := db.monikersByPositionOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
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

	return db.database.MonikersByPosition(ctx, bundleID, path, line, character)
}

// MonikerResults calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) MonikerResults(ctx context.Context, bundleID int, tableName, scheme, identifier string, skip, take int) (locations []bundles.Location, _ int, err error) {
	ctx, endObservation := db.monikerResultsOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("tableName", tableName),
			log.String("scheme", scheme),
			log.String("identifier", identifier),
		},
	})
	defer func() { endObservation(float64(len(locations)), observation.Args{}) }()
	return db.database.MonikerResults(ctx, bundleID, tableName, scheme, identifier, skip, take)
}

// PackageInformation calls into the inner Database and registers the observed results.
func (db *ObservedDatabase) PackageInformation(ctx context.Context, bundleID int, path string, packageInformationID string) (_ bundles.PackageInformationData, _ bool, err error) {
	ctx, endObservation := db.packageInformationOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.String("packageInformationId", string(packageInformationID)),
		},
	})
	defer endObservation(1, observation.Args{})
	return db.database.PackageInformation(ctx, bundleID, path, packageInformationID)
}
