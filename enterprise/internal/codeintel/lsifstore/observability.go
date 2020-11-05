package lsifstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// An ObservedStore wraps another Store with error logging, Prometheus metrics, and tracing.
type ObservedStore struct {
	store                       Store
	existsOperation             *observation.Operation
	rangesOperation             *observation.Operation
	definitionsOperation        *observation.Operation
	referencesOperation         *observation.Operation
	hoverOperation              *observation.Operation
	diagnosticsOperation        *observation.Operation
	monikersByPositionOperation *observation.Operation
	monikerResultsOperation     *observation.Operation
	packageInformationOperation *observation.Operation
	clearOperation              *observation.Operation
	readMetaOperation           *observation.Operation
	pathsWithPrefixOperation    *observation.Operation
	readDocumentOperation       *observation.Operation
	readResultChunkOperation    *observation.Operation
	readDefinitionsOperation    *observation.Operation
	readReferencesOperation     *observation.Operation
	doneOperation               *observation.Operation
	writeMetaOperation          *observation.Operation
	writeDocumentsOperation     *observation.Operation
	writeResultChunksOperation  *observation.Operation
	writeDefinitionsOperation   *observation.Operation
	writeReferencesOperation    *observation.Operation
}

var _ Store = &ObservedStore{}

// singletonMetrics ensures that the operation metrics required by ObservedStore are
// constructed only once as there may be many stores instantiated by a single replica
// of precise-code-intel-bundle-manages.
var singletonMetrics = &metrics.SingletonOperationMetrics{}

// NewObservedStore wraps the given store with error logging, Prometheus metrics, and tracing.
func NewObserved(store Store, observationContext *observation.Context) Store {
	metrics := singletonMetrics.Get(func() *metrics.OperationMetrics {
		return metrics.NewOperationMetrics(
			observationContext.Registerer,
			"code_intel_codeintel_db_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of results returned"),
		)
	})

	return &ObservedStore{
		store: store,
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
		clearOperation: observationContext.Operation(observation.Op{
			Name:         "Store.Clear",
			MetricLabels: []string{"clear"},
			Metrics:      metrics,
		}),
		readMetaOperation: observationContext.Operation(observation.Op{
			Name:         "Store.ReadMeta",
			MetricLabels: []string{"read_meta"},
			Metrics:      metrics,
		}),
		pathsWithPrefixOperation: observationContext.Operation(observation.Op{
			Name:         "Store.PathsWithPrefix",
			MetricLabels: []string{"paths_with_prefix"},
			Metrics:      metrics,
		}),
		readDocumentOperation: observationContext.Operation(observation.Op{
			Name:         "Store.ReadDocument",
			MetricLabels: []string{"read_document"},
			Metrics:      metrics,
		}),
		readResultChunkOperation: observationContext.Operation(observation.Op{
			Name:         "Store.ReadResultChunk",
			MetricLabels: []string{"read_result-chunk"},
			Metrics:      metrics,
		}),
		readDefinitionsOperation: observationContext.Operation(observation.Op{
			Name:         "Store.ReadDefinitions",
			MetricLabels: []string{"read_definitions"},
			Metrics:      metrics,
		}),
		readReferencesOperation: observationContext.Operation(observation.Op{
			Name:         "Store.ReadReferences",
			MetricLabels: []string{"read_references"},
			Metrics:      metrics,
		}),
		doneOperation: observationContext.Operation(observation.Op{
			Name:         "Store.Done",
			MetricLabels: []string{"done"},
			Metrics:      metrics,
		}),
		writeMetaOperation: observationContext.Operation(observation.Op{
			Name:         "Store.WriteMeta",
			MetricLabels: []string{"write_meta"},
			Metrics:      metrics,
		}),
		writeDocumentsOperation: observationContext.Operation(observation.Op{
			Name:         "Store.WriteDocuments",
			MetricLabels: []string{"write_documents"},
			Metrics:      metrics,
		}),
		writeResultChunksOperation: observationContext.Operation(observation.Op{
			Name:         "Store.WriteResultChunks",
			MetricLabels: []string{"write_result_chunks"},
			Metrics:      metrics,
		}),
		writeDefinitionsOperation: observationContext.Operation(observation.Op{
			Name:         "Store.WriteDefinitions",
			MetricLabels: []string{"write_definitions"},
			Metrics:      metrics,
		}),
		writeReferencesOperation: observationContext.Operation(observation.Op{
			Name:         "Store.WriteReferences",
			MetricLabels: []string{"write_references"},
			Metrics:      metrics,
		}),
	}
}

// Clear calls into the inner Store and registers the observed results.
func (s *ObservedStore) Clear(ctx context.Context, bundleIDs ...int) (err error) {
	ctx, endObservation := s.clearOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.Clear(ctx, bundleIDs...)
}

// ReadMeta calls into the inner Store and registers the observed results.
func (s *ObservedStore) ReadMeta(ctx context.Context, bundleID int) (_ MetaData, err error) {
	ctx, endObservation := s.readMetaOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.ReadMeta(ctx, bundleID)
}

// PathsWithPrefix calls into the inner Store and registers the observed results.
func (s *ObservedStore) PathsWithPrefix(ctx context.Context, bundleID int, prefix string) (_ []string, err error) {
	ctx, endObservation := s.pathsWithPrefixOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.PathsWithPrefix(ctx, bundleID, prefix)
}

// ReadDocument calls into the inner Store and registers the observed results.
func (s *ObservedStore) ReadDocument(ctx context.Context, bundleID int, path string) (_ DocumentData, _ bool, err error) {
	ctx, endObservation := s.readDocumentOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.ReadDocument(ctx, bundleID, path)
}

// ReadResultChunk calls into the inner Store and registers the observed results.
func (s *ObservedStore) ReadResultChunk(ctx context.Context, bundleID int, id int) (_ ResultChunkData, _ bool, err error) {
	ctx, endObservation := s.readResultChunkOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.ReadResultChunk(ctx, bundleID, id)
}

// ReadDefinitions calls into the inner Store and registers the observed results.
func (s *ObservedStore) ReadDefinitions(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) (locations []Location, _ int, err error) {
	ctx, endObservation := s.readDefinitionsOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(locations)), observation.Args{}) }()
	return s.store.ReadDefinitions(ctx, bundleID, scheme, identifier, skip, take)
}

// ReadReferences calls into the inner Store and registers the observed results.
func (s *ObservedStore) ReadReferences(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) (locations []Location, _ int, err error) {
	ctx, endObservation := s.readReferencesOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(locations)), observation.Args{}) }()
	return s.store.ReadReferences(ctx, bundleID, scheme, identifier, skip, take)
}

// Transact calls into the inner Store and registers the observed result.
func (s *ObservedStore) Transact(ctx context.Context) (_ Store, err error) {
	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &ObservedStore{
		store:                      tx,
		readMetaOperation:          s.readMetaOperation,
		pathsWithPrefixOperation:   s.pathsWithPrefixOperation,
		readDocumentOperation:      s.readDocumentOperation,
		readResultChunkOperation:   s.readResultChunkOperation,
		readDefinitionsOperation:   s.readDefinitionsOperation,
		readReferencesOperation:    s.readReferencesOperation,
		doneOperation:              s.doneOperation,
		writeMetaOperation:         s.writeMetaOperation,
		writeDocumentsOperation:    s.writeDocumentsOperation,
		writeResultChunksOperation: s.writeResultChunksOperation,
		writeDefinitionsOperation:  s.writeDefinitionsOperation,
		writeReferencesOperation:   s.writeReferencesOperation,
	}, nil
}

// Done calls into the inner Store and registers the observed result.
func (s *ObservedStore) Done(e error) error {
	var observedErr error = nil
	_, endObservation := s.doneOperation.With(context.Background(), &observedErr, observation.Args{})
	defer endObservation(1, observation.Args{})

	err := s.store.Done(e)
	if err != e {
		// Only observe the error if it's a commit/rollback failure
		observedErr = err
	}
	return err
}

// WriteMeta calls into the inner Store and registers the observed result.
func (s *ObservedStore) WriteMeta(ctx context.Context, bundleID int, meta MetaData) (err error) {
	ctx, endObservation := s.writeMetaOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.WriteMeta(ctx, bundleID, meta)
}

// WriteDocuments calls into the inner Store and registers the observed result.
func (s *ObservedStore) WriteDocuments(ctx context.Context, bundleID int, documents chan KeyedDocumentData) (err error) {
	ctx, endObservation := s.writeDocumentsOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.WriteDocuments(ctx, bundleID, documents)
}

// WriteResultChunks calls into the inner Store and registers the observed result.
func (s *ObservedStore) WriteResultChunks(ctx context.Context, bundleID int, resultChunks chan IndexedResultChunkData) (err error) {
	ctx, endObservation := s.writeResultChunksOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.WriteResultChunks(ctx, bundleID, resultChunks)
}

// WriteDefinitions calls into the inner Store and registers the observed result.
func (s *ObservedStore) WriteDefinitions(ctx context.Context, bundleID int, monikerLocations chan MonikerLocations) (err error) {
	ctx, endObservation := s.writeDefinitionsOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.WriteDefinitions(ctx, bundleID, monikerLocations)
}

// WriteReferences calls into the inner Store and registers the observed result.
func (s *ObservedStore) WriteReferences(ctx context.Context, bundleID int, monikerLocations chan MonikerLocations) (err error) {
	ctx, endObservation := s.writeReferencesOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return s.store.WriteReferences(ctx, bundleID, monikerLocations)
}
