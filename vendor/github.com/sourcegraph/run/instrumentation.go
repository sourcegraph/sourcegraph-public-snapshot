package run

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	contextKeyShouldTrace contextKey = "shouldTrace"
	contextKeyShouldLog   contextKey = "shouldLog"
)

// ExecutedCommand represents a command that has been started.
type ExecutedCommand struct {
	Args    []string
	Dir     string
	Environ []string
}

// LogFunc can be used to generate a log entry for the executed command.
type LogFunc func(ExecutedCommand)

// LogCommands enables logging on all commands executed by sourcegraph/run within this
// context. Set to nil to disable (default), or use run.DefaultTraceAttributes for some
// recommended defaults.
//
// Note that arguments and environments may contain sensitive information.
//
// If you use loggers carrying contexts, e.g. via sourcegraph/log, it is recommended that
// you only enable this within relevant scopes.
func LogCommands(ctx context.Context, log LogFunc) context.Context {
	return context.WithValue(ctx, contextKeyShouldLog, log)
}

// getLogger turns a log func if logging is enabled, otherwise returns nil.
func getLogger(ctx context.Context) LogFunc {
	v, _ := ctx.Value(contextKeyShouldLog).(LogFunc)
	return v
}

// TraceAttributesFunc can be used to generate attributes to attach to a span for the
// executed command.
type TraceAttributesFunc func(ExecutedCommand) []attribute.KeyValue

var _ TraceAttributesFunc = DefaultTraceAttributes

// DefaultTraceAttributes adds Args and Dir as attributes. Note that Args may contain
// sensitive data.
func DefaultTraceAttributes(e ExecutedCommand) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.StringSlice("Args", e.Args),
		attribute.String("Dir", e.Dir),
	}
}

// TraceCommands toggles OpenTelemetry tracing on all usages of sourcegraph/run within
// this context. Set to nil to disable (default).
//
// Note that arguments and environments may contain sensitive information.
func TraceCommands(ctx context.Context, attrs TraceAttributesFunc) context.Context {
	return context.WithValue(ctx, contextKeyShouldTrace, attrs)
}

// getTracer returns a tracer if tracing is enabled, otherwise returns a no-op tracer.
func getTracer(ctx context.Context) (trace.Tracer, TraceAttributesFunc) {
	v, _ := ctx.Value(contextKeyShouldTrace).(TraceAttributesFunc)
	if v != nil {
		return otel.GetTracerProvider().Tracer("sourcegraph/run"), v
	}
	// Return no-ops.
	return trace.NewNoopTracerProvider().Tracer("sourcegraph/run"),
		func(ExecutedCommand) []attribute.KeyValue { return nil }
}
