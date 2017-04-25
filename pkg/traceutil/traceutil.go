package traceutil

import opentracing "github.com/opentracing/opentracing-go"

// SpanURL returns the URL to the tracing UI for the given span. The span must be non-nil.
var SpanURL = func(span opentracing.Span) string {
	return "#tracer-not-enabled"
}
