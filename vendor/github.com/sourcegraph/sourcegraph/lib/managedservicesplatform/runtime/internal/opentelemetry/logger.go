package opentelemetry

import (
	"context"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/trace"
)

// TracedLogger returns a logger with trace context if available.
func TracedLogger(ctx context.Context, logger log.Logger) log.Logger {
	if otelSpan := trace.SpanContextFromContext(ctx); otelSpan.IsValid() {
		return logger.WithTrace(log.TraceContext{
			TraceID: otelSpan.TraceID().String(),
			SpanID:  otelSpan.SpanID().String(),
		})
	}
	return logger
}
