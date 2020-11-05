package lsifstore

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client_types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Exists calls into the inner Database and registers the observed results.
func (db *ObservedStore) Exists(ctx context.Context, bundleID int, path string) (_ bool, err error) {
	ctx, endObservation := db.existsOperation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})
	return db.store.Exists(ctx, bundleID, path)
}

// Ranges calls into the inner Database and registers the observed results.
func (db *ObservedStore) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) (ranges []bundles.CodeIntelligenceRange, err error) {
	ctx, endObservation := db.rangesOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("startLine", startLine),
			log.Int("endLine", endLine),
		},
	})
	defer func() { endObservation(float64(len(ranges)), observation.Args{}) }()
	return db.store.Ranges(ctx, bundleID, path, startLine, endLine)
}

// Definitions calls into the inner Database and registers the observed results.
func (db *ObservedStore) Definitions(ctx context.Context, bundleID int, path string, line, character int) (definitions []bundles.Location, err error) {
	ctx, endObservation := db.definitionsOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer func() { endObservation(float64(len(definitions)), observation.Args{}) }()
	return db.store.Definitions(ctx, bundleID, path, line, character)
}

// References calls into the inner Database and registers the observed results.
func (db *ObservedStore) References(ctx context.Context, bundleID int, path string, line, character int) (references []bundles.Location, err error) {
	ctx, endObservation := db.referencesOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer func() { endObservation(float64(len(references)), observation.Args{}) }()
	return db.store.References(ctx, bundleID, path, line, character)
}

// Hover calls into the inner Database and registers the observed results.
func (db *ObservedStore) Hover(ctx context.Context, bundleID int, path string, line, character int) (_ string, _ bundles.Range, _ bool, err error) {
	ctx, endObservation := db.hoverOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.Int("line", line),
			log.Int("character", character),
		},
	})
	defer endObservation(1, observation.Args{})
	return db.store.Hover(ctx, bundleID, path, line, character)
}

// Diagnostics calls into the inner Database and registers the observed results.
func (db *ObservedStore) Diagnostics(ctx context.Context, bundleID int, prefix string, skip, take int) (diagnostics []bundles.Diagnostic, _ int, err error) {
	ctx, endObservation := db.hoverOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("prefix", prefix),
		},
	})
	defer func() { endObservation(float64(len(diagnostics)), observation.Args{}) }()
	return db.store.Diagnostics(ctx, bundleID, prefix, skip, take)
}

// MonikersByPosition calls into the inner Database and registers the observed results.
func (db *ObservedStore) MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (monikers [][]bundles.MonikerData, err error) {
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

	return db.store.MonikersByPosition(ctx, bundleID, path, line, character)
}

// MonikerResults calls into the inner Database and registers the observed results.
func (db *ObservedStore) MonikerResults(ctx context.Context, bundleID int, tableName, scheme, identifier string, skip, take int) (locations []bundles.Location, _ int, err error) {
	ctx, endObservation := db.monikerResultsOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("tableName", tableName),
			log.String("scheme", scheme),
			log.String("identifier", identifier),
		},
	})
	defer func() { endObservation(float64(len(locations)), observation.Args{}) }()
	return db.store.MonikerResults(ctx, bundleID, tableName, scheme, identifier, skip, take)
}

// PackageInformation calls into the inner Database and registers the observed results.
func (db *ObservedStore) PackageInformation(ctx context.Context, bundleID int, path string, packageInformationID string) (_ bundles.PackageInformationData, _ bool, err error) {
	ctx, endObservation := db.packageInformationOperation.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("bundleID", bundleID),
			log.String("path", path),
			log.String("packageInformationId", string(packageInformationID)),
		},
	})
	defer endObservation(1, observation.Args{})
	return db.store.PackageInformation(ctx, bundleID, path, packageInformationID)
}
