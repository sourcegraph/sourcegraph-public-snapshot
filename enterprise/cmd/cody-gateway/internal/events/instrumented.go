package events

import (
	"context"
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
			attribute.String("name", string(event.Name))))
	defer span.End()

	if err := i.Logger.LogEvent(spanCtx, event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to log event")
		return err
	}
	return nil
}

func backgroundContextWithSpan(ctx context.Context) context.Context {
	// NOTE: Using context.Background() because we still want to log the event in the
	// case of a request cancellation, we only want the parent span.
	ctx = trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
	return ctx
}
