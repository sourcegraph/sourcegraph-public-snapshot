package graph

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/openfga/openfga/pkg/tuple"
)

type CycleDetectionCheckResolver struct {
	delegate CheckResolver
}

var _ CheckResolver = (*CycleDetectionCheckResolver)(nil)

// Close implements CheckResolver.
func (*CycleDetectionCheckResolver) Close() {}

func NewCycleDetectionCheckResolver() *CycleDetectionCheckResolver {
	c := &CycleDetectionCheckResolver{}
	c.delegate = c

	return c
}

// ResolveCheck implements CheckResolver.
func (c *CycleDetectionCheckResolver) ResolveCheck(
	ctx context.Context,
	req *ResolveCheckRequest,
) (*ResolveCheckResponse, error) {
	span := trace.SpanFromContext(ctx)

	key := tuple.TupleKeyToString(req.GetTupleKey())

	if req.VisitedPaths == nil {
		req.VisitedPaths = map[string]struct{}{}
	}

	_, cycleDetected := req.VisitedPaths[key]
	span.SetAttributes(attribute.Bool("cycle_detected", cycleDetected))
	if cycleDetected {
		return &ResolveCheckResponse{
			Allowed: false,
			ResolutionMetadata: &ResolveCheckResponseMetadata{
				CycleDetected: true,
			},
		}, nil
	}

	req.VisitedPaths[key] = struct{}{}

	return c.delegate.ResolveCheck(ctx, &ResolveCheckRequest{
		StoreID:              req.GetStoreID(),
		AuthorizationModelID: req.GetAuthorizationModelID(),
		TupleKey:             req.GetTupleKey(),
		ContextualTuples:     req.GetContextualTuples(),
		RequestMetadata:      req.GetRequestMetadata(),
		VisitedPaths:         req.VisitedPaths,
		Context:              req.GetContext(),
	})
}

func (c *CycleDetectionCheckResolver) SetDelegate(delegate CheckResolver) {
	c.delegate = delegate
}

func (c *CycleDetectionCheckResolver) GetDelegate() CheckResolver {
	return c.delegate
}
