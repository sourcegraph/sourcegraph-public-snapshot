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
	_, span := tracer.Start(spanCtx, fmt.Sprintf("%s.LogEvent", i.Scope),
		trace.WithAttributes(
			attribute.String("source", event.Source),
			attribute.String("event.name", string(event.Name))))
	defer span.End()

	// Best-effort attempt to record event metadata
	if metadataJSON, err := json.Marshal(event.Metadata); err == nil {
		span.SetAttributes(attribute.String("event.metadata", string(metadataJSON)))
	}

	if err := i.Logger.LogEvent(spanCtx, event); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}
