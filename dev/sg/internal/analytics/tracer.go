package analytics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// spanCategoryKey denotes the type of a span, e.g. "root" or "action"
const spanCategoryKey attribute.Key = "sg.span_category"

// StartSpan starts an OpenTelemetry span from context. Example:
//
//	ctx, span := analytics.StartSpan(ctx, spanName,
//		trace.WithAttributes(...)
//	defer span.End()
//	// ... do your things
//
// Span provides convenience functions for setting the status of the span.
func StartSpan(ctx context.Context, spanName string, category string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	opts = append(opts, trace.WithAttributes(spanCategoryKey.String(category)))
	ctx, s := otel.GetTracerProvider().Tracer("dev/sg/analytics").Start(ctx, spanName, opts...)
	return ctx, &Span{s}
}

// Span wraps an OpenTelemetry span with convenience functions.
type Span struct{ trace.Span }

// Error records and error in span.
func (s *Span) RecordError(kind string, err error, options ...trace.EventOption) {
	s.Failed(kind)
	s.Span.RecordError(err)
}

// Succeeded records a success in span.
func (s *Span) Succeeded() {
	// description is only kept if error, so we add an event
	s.Span.AddEvent("success")
	s.Span.SetStatus(codes.Ok, "success")
}

// Failed records a failure.
func (s *Span) Failed(reason ...string) {
	v := "failed"
	if len(reason) > 0 {
		v = reason[0]
	}
	s.Span.AddEvent(v)
	s.Span.SetStatus(codes.Error, v)
}

// Cancelled records a cancellation.
func (s *Span) Cancelled() {
	// description is only kept if error, so we add an event
	s.Span.AddEvent("cancelled")
	s.Span.SetStatus(codes.Ok, "cancelled")
}

// Skipped records a skipped task.
func (s *Span) Skipped(reason ...string) {
	v := "skipped"
	if len(reason) > 0 {
		v = reason[0]
	}
	// description is only kept if error, so we add an event
	s.Span.AddEvent(v)
	s.Span.SetStatus(codes.Ok, v)
}

// NoOpSpan is a safe-to-use, no-op span.
func NoOpSpan() *Span {
	_, s := trace.NewNoopTracerProvider().Tracer("").Start(context.Background(), "")
	return &Span{s}
}
