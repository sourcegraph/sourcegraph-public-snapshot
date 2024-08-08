package policy

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const shouldTraceMetadataKey = "sg-should-trace"

// ShouldTracePropagator implements (internal/grpc).Propagator so that the
// ShouldTrace key can be propagated across gRPC API calls.
type ShouldTracePropagator struct{}

func (ShouldTracePropagator) FromContext(ctx context.Context) metadata.MD {
	return metadata.Pairs(shouldTraceMetadataKey, strconv.FormatBool(ShouldTrace(ctx)))
}

func (ShouldTracePropagator) InjectContext(ctx context.Context, md metadata.MD) (context.Context, error) {
	vals := md.Get(shouldTraceMetadataKey)
	if len(vals) > 0 {
		shouldTrace, err := strconv.ParseBool(vals[0])
		if err != nil {
			// Ignore error, just returning the context
			return ctx, nil
		}
		return WithShouldTrace(ctx, shouldTrace), nil
	}
	return ctx, nil
}
