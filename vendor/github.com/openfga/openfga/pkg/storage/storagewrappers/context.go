package storagewrappers

import (
	"context"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"go.opentelemetry.io/otel/trace"

	"github.com/openfga/openfga/pkg/storage"
)

// ContextTracerWrapper is a wrapper for a datastore that introduces a new context to the underlying datastore methods.
// Its purpose is to prevent the closure of the underlying database connection in case the original context is cancelled,
// such as when a client cancels the context. This ensures that ongoing queries are allowed to complete even if the
// original context is cancelled, helping to avoid unnecessary database connection churn.
type ContextTracerWrapper struct {
	storage.OpenFGADatastore
}

var _ storage.OpenFGADatastore = (*ContextTracerWrapper)(nil)

// NewContextWrapper creates a new instance of [ContextTracerWrapper], wrapping the specified datastore. It is crucial
// for [ContextTracerWrapper] to be the first wrapper around the datastore for traces to function correctly.
func NewContextWrapper(inner storage.OpenFGADatastore) *ContextTracerWrapper {
	return &ContextTracerWrapper{inner}
}

// queryContext generates a new context that is independent of the provided
// context and its timeout with the exception of the trace context.
func queryContext(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	return trace.ContextWithSpan(context.Background(), span)
}

// Close ensures proper cleanup and closure of resources associated with the OpenFGADatastore.
func (c *ContextTracerWrapper) Close() {
	c.OpenFGADatastore.Close()
}

// Read see [storage.RelationshipTupleReader.ReadUserTuple].
func (c *ContextTracerWrapper) Read(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (storage.TupleIterator, error) {
	queryCtx := queryContext(ctx)

	return c.OpenFGADatastore.Read(queryCtx, store, tupleKey)
}

// ReadPage see [storage.RelationshipTupleReader.ReadPage].
func (c *ContextTracerWrapper) ReadPage(ctx context.Context, store string, tupleKey *openfgav1.TupleKey, opts storage.PaginationOptions) ([]*openfgav1.Tuple, []byte, error) {
	queryCtx := queryContext(ctx)

	return c.OpenFGADatastore.ReadPage(queryCtx, store, tupleKey, opts)
}

// ReadUserTuple see [storage.RelationshipTupleReader].ReadUserTuple.
func (c *ContextTracerWrapper) ReadUserTuple(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (*openfgav1.Tuple, error) {
	queryCtx := queryContext(ctx)

	return c.OpenFGADatastore.ReadUserTuple(queryCtx, store, tupleKey)
}

// ReadUsersetTuples see [storage.RelationshipTupleReader].ReadUsersetTuples.
func (c *ContextTracerWrapper) ReadUsersetTuples(ctx context.Context, store string, filter storage.ReadUsersetTuplesFilter) (storage.TupleIterator, error) {
	queryCtx := queryContext(ctx)

	return c.OpenFGADatastore.ReadUsersetTuples(queryCtx, store, filter)
}

// ReadStartingWithUser see [storage.RelationshipTupleReader].ReadStartingWithUser.
func (c *ContextTracerWrapper) ReadStartingWithUser(ctx context.Context, store string, opts storage.ReadStartingWithUserFilter) (storage.TupleIterator, error) {
	queryCtx := queryContext(ctx)

	return c.OpenFGADatastore.ReadStartingWithUser(queryCtx, store, opts)
}
