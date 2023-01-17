package policy

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const traceMetadataKey = "sg-should-trace"

type ShouldTracePropagator struct{}

func (ShouldTracePropagator) ExtractContext(ctx context.Context) metadata.MD {
	return metadata.Pairs(traceMetadataKey, strconv.FormatBool(ShouldTrace(ctx)))
}

func (ShouldTracePropagator) InjectContext(ctx context.Context, md metadata.MD) context.Context {
	vals := md.Get(traceMetadataKey)
	if len(vals) > 0 {
		shouldTrace, err := strconv.ParseBool(vals[0])
		if err != nil {
			// Ignore error, just returning the context
			return ctx
		}
		return WithShouldTrace(ctx, shouldTrace)
	}
	return ctx
}
