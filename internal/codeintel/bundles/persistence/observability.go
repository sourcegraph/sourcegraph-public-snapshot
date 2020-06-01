package persistence

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// An ObservedReader wraps another Reader with error logging, Prometheus metrics, and tracing.
type ObservedReader struct {
	reader                   Reader
	readMetaOperation        *observation.Operation
	readDocumentOperation    *observation.Operation
	readResultChunkOperation *observation.Operation
	readDefinitionsOperation *observation.Operation
	readReferencesOperation  *observation.Operation
}

var _ Reader = &ObservedReader{}

// singletonMetrics ensures that the operation metrics required by ObservedReader are
// constructed only once as there may be many readers instantiated by a single replica
// of precise-code-intel-bundle-manager.
var singletonMetrics = &metrics.SingletonOperationMetrics{}

// NewObservedReader wraps the given Reader with error logging, Prometheus metrics, and tracing.
func NewObserved(reader Reader, observationContext *observation.Context) Reader {
	metrics := singletonMetrics.Get(func() *metrics.OperationMetrics {
		return metrics.NewOperationMetrics(
			observationContext.Registerer,
			"bundle_reader",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of results returned"),
		)
	})

	return &ObservedReader{
		reader: reader,
		readMetaOperation: observationContext.Operation(observation.Op{
			Name:         "Reader.ReadMeta",
			MetricLabels: []string{"read_meta"},
			Metrics:      metrics,
		}),
		readDocumentOperation: observationContext.Operation(observation.Op{
			Name:         "Reader.ReadDocument",
			MetricLabels: []string{"read_document"},
			Metrics:      metrics,
		}),
		readResultChunkOperation: observationContext.Operation(observation.Op{
			Name:         "Reader.ReadResultChunk",
			MetricLabels: []string{"read_result-chunk"},
			Metrics:      metrics,
		}),
		readDefinitionsOperation: observationContext.Operation(observation.Op{
			Name:         "Reader.ReadDefinitions",
			MetricLabels: []string{"read_definitions"},
			Metrics:      metrics,
		}),
		readReferencesOperation: observationContext.Operation(observation.Op{
			Name:         "Reader.ReadReferences",
			MetricLabels: []string{"read_references"},
			Metrics:      metrics,
		}),
	}
}

// ReadMeta calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadMeta(ctx context.Context) (_ types.MetaData, err error) {
	ctx, endObservation := r.readMetaOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return r.reader.ReadMeta(ctx)
}

// ReadDocument calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadDocument(ctx context.Context, path string) (_ types.DocumentData, _ bool, err error) {
	ctx, endObservation := r.readDocumentOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return r.reader.ReadDocument(ctx, path)
}

// ReadResultChunk calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadResultChunk(ctx context.Context, id int) (_ types.ResultChunkData, _ bool, err error) {
	ctx, endObservation := r.readResultChunkOperation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	return r.reader.ReadResultChunk(ctx, id)
}

// ReadDefinitions calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) (locations []types.Location, _ int, err error) {
	ctx, endObservation := r.readDefinitionsOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(locations)), observation.Args{}) }()
	return r.reader.ReadDefinitions(ctx, scheme, identifier, skip, take)
}

// ReadReferences calls into the inner Reader and registers the observed results.
func (r *ObservedReader) ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) (locations []types.Location, _ int, err error) {
	ctx, endObservation := r.readReferencesOperation.With(ctx, &err, observation.Args{})
	defer func() { endObservation(float64(len(locations)), observation.Args{}) }()
	return r.reader.ReadReferences(ctx, scheme, identifier, skip, take)
}

func (r *ObservedReader) Close() error {
	return r.reader.Close()
}
