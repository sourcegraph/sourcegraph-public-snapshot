package events

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.GetTracerProvider().Tracer("cody-gateway/internal/events")

type instrumentedLogger struct {
	Scope string
	Logger
}

var _ Logger = &DelayedLogger{}

func (i *instrumentedLogger) LogEvent(spanCtx context.Context, event Event) error {
	_, span := tracer.Start(backgroundContextWithSpan(spanCtx), fmt.Sprintf("%s.LogEvent", i.Scope),
		trace.WithAttributes(
			attribute.String("source", event.Source),
			attribute.String("event.name", string(event.Name))))
	defer span.End()

	// Best-effort attempt to record event metadata
	if metadataJSON, err := json.Marshal(event.Metadata); err == nil {
		span.SetAttributes(attribute.String("event.metadata", string(metadataJSON)))
	}

	if err := i.Logger.LogEvent(spanCtx, event); err != nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.SetStatus(codes.Error, "failed to log event")
		return err
	}
	return nil
}

// backgroundContextWithSpan extracts the span from the context and creates a new
// context.Background() with the span attached. Using context.Background() is
// desireable in Logger implementations because we still want to log the event
// in the case of a request cancellation, but we want to retain the parent span.
func backgroundContextWithSpan(ctx context.Context) context.Context {
	return trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
}
