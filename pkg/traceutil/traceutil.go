package traceutil

import (
	"context"
	"strings"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

// SpanURL returns the URL to the tracing UI for the given span. The span must be non-nil.
var SpanURL = func(span opentracing.Span) string {
	return "#tracer-not-enabled"
}

const traceNameKey = "traceName"

func TraceName(ctx context.Context, name string) (string, context.Context) {
	val := ctx.Value(traceNameKey)
	var names []string
	if val, ok := val.([]string); ok {
		names = val
	}
	names = append(names, name)
	return strings.Join(names, " > "), context.WithValue(ctx, traceNameKey, names)
}

func init() {
	// Ignore warnings from the tracer about SetTag calls with unrecognized value types. The
	// github.com/lightstep/lightstep-tracer-go package calls fmt.Sprintf("%#v", ...) on them, which is fine.
	defaultHandler := lightstep.NewEventLogOneError()
	lightstep.SetGlobalEventHandler(func(e lightstep.Event) {
		if _, ok := e.(lightstep.EventUnsupportedValue); ok {
			// ignore
		} else {
			defaultHandler(e)
		}
	})
}
