// Package observation provides a unified way to wrap an operation with logging, tracing, and metrics.
//
// High-level ideas:
//
//     - Each service creates an observation Context that carries a root logger, tracer,
//       and a metrics registerer as its context.
//
//     - An observation Context can create an observation Operation which represents a
//       section of code that can be invoked many times. An observation Operation is
//       configured with state that applies to all invocation of the code.
//
//     - An observation Operation can wrap a an invocation of a section of code by calling its
//       With method. This prepares a trace and some state to be reconciled after the invocation
//       has completed. The With method returns a function that, when deferred, will emit metrics,
//       additional logs, and finalize the trace span.
//
// Sample usage:
//
//     observationContext := observation.NewContex(
//         log15.Root(),
//         &trace.Tracer{Tracer: opentracing.GlobalTracer()},
//         prometheus.DefaultRegisterer,
//     )
//
//     metrics := metrics.NewOperationMetrics(
//         observationContext.Registerer,
//         "thing",
//         metrics.WithLabels("op"),
//     )
//
//     operation := observationContext.Operation(observation.Op{
//         Name:         "Thing.SomeOperation",
//         MetricLabels: []string{"some_operation"},
//         Metrics:      metrics,
//     })
//
//     function SomeOperation(ctx context.Context) (err error) {
//         // logs and metrics may be available before or after the operation, so they
//         // can be supplied either at the start of the operation, or after in the
//         // defer of endObservation.
//
//         ctx, endObservation := operation.With(ctx, &err, observation.Args{ /* logs and metrics */ })
//         defer func() { endObservation(1, observation.Args{ /* additional logs and metrics */ }) }()
//
//         // ...
//     }
//
// Log fields and metric labels can be supplied at construction of an Operation, at invocation
// of an operation (the With function), or after the invocation completes but before the observation
// has terminated (the endObservation function). Log fields and metric labels are concatenated
// together in the order they are attached to an operation.
package observation

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Context carries context about where to send logs, trace spans, and register
// metrics. It should be created once on service startup, and passed around to
// any location that wants to use it for observing operations.
type Context struct {
	Logger     logging.ErrorLogger
	Tracer     *trace.Tracer
	Registerer prometheus.Registerer
}

// TestContext is a behaviorless Context usable for unit tests.
var TestContext = Context{Registerer: metrics.TestRegisterer}

// Op configures an Operation instance.
type Op struct {
	Metrics *metrics.OperationMetrics
	// Name configures the trace and error log names. This string should be of the
	// format {GroupName}.{OperationName}, where both sections are title cased
	// (e.g. Store.GetRepoByID).
	Name string
	// MetricLabels that apply for every invocation of this operation.
	MetricLabels []string
	// LogFields that apply for for every invocation of this operation.
	LogFields []log.Field
	// ErrorFilter returns false for any error that should be converted to nil
	// for the purposes of metrics and tracing. If this field is not set then
	// error values are unaltered.
	//
	// This is useful when, for example, a revision not found error is expected by
	// a process interfacing with gitserver. Such an error should not be treated as
	// an unexpected value in metrics and traces but should be handled higher up in
	// the stack.
	ErrorFilter func(err error) bool
}

// Operation combines the state of the parent context to create a new operation. This value
// should be owned and used by the code that performs the operation it represents.
func (c *Context) Operation(args Op) *Operation {
	return &Operation{
		context:      c,
		metrics:      args.Metrics,
		name:         args.Name,
		kebabName:    kebabCase(args.Name),
		metricLabels: args.MetricLabels,
		logFields:    args.LogFields,
		errorFilter:  args.ErrorFilter,
	}
}

// Operation represents an interesting section of code that can be invoked.
type Operation struct {
	context      *Context
	metrics      *metrics.OperationMetrics
	name         string
	kebabName    string
	metricLabels []string
	logFields    []log.Field
	errorFilter  func(err error) bool
}

// TraceLogger is returned from WithAndLogger and can be used to add timestamped key and
// value pairs into a related opentracing span.
type TraceLogger func(fields ...log.Field)

// FinishFunc is the shape of the function returned by With and should be invoked within
// a defer directly before the observed function returns.
type FinishFunc func(count float64, args Args)

// Args configures the observation behavior of an invocation of an operation.
type Args struct {
	// MetricLabels that apply only to this invocation of the operation.
	MetricLabels []string
	// LogFields that apply only to this invocation of the operation.
	LogFields []log.Field
}

// With prepares the necessary timers, loggers, and metrics to observe the invocation  of an
// operation. This method returns a modified context and a function to be deferred until the
// end of the operation.
func (op *Operation) With(ctx context.Context, err *error, args Args) (context.Context, FinishFunc) {
	ctx, _, endObservation := op.WithAndLogger(ctx, err, args)
	return ctx, endObservation
}

// WithAndLogger prepares the necessary timers, loggers, and metrics to observe the invocation
// of an operation. This method returns a modified context, a function that will add a log field
// to the active trace, and a function to be deferred until the end of the operation.
func (op *Operation) WithAndLogger(ctx context.Context, err *error, args Args) (context.Context, TraceLogger, FinishFunc) {
	start := time.Now()
	tr, ctx := op.trace(ctx, args)

	var logFields TraceLogger
	if tr != nil {
		logFields = tr.LogFields
	} else {
		logFields = func(fields ...log.Field) {}
	}

	return ctx, logFields, func(count float64, finishArgs Args) {
		elapsed := time.Since(start).Seconds()
		defaultFinishFields := []log.Field{log.Float64("count", count), log.Float64("elapsed", elapsed)}
		logFields := mergeLogFields(defaultFinishFields, finishArgs.LogFields)
		metricLabels := mergeLabels(op.metricLabels, args.MetricLabels, finishArgs.MetricLabels)

		err = op.applyErrorFilter(err)
		op.emitErrorLogs(err, logFields)
		op.emitMetrics(err, count, elapsed, metricLabels)
		op.finishTrace(err, tr, logFields)
	}
}

// trace creates a new Trace object and returns the wrapped context. If any log fields are
// attached to the operation or to the args to With, they are emitted immediately. This returns
// an unmodified context and a nil trace if no tracer was supplied on the observation context.
func (op *Operation) trace(ctx context.Context, args Args) (*trace.Trace, context.Context) {
	if op.context.Tracer == nil {
		return nil, ctx
	}

	tr, ctx := op.context.Tracer.New(ctx, op.kebabName, "")
	tr.LogFields(mergeLogFields(op.logFields, args.LogFields)...)
	return tr, ctx
}

// emitErrorLogs will log as message if the operation has failed. This log contains the error
// as well as all of the log fields attached ot the operation, the args to With, and the args
// to the finish function. This does nothing if the no logger was supplied on the observation
// context.
func (op *Operation) emitErrorLogs(err *error, logFields []log.Field) {
	if op.context.Logger == nil {
		return
	}

	var kvs []interface{}
	for _, field := range logFields {
		kvs = append(kvs, field.Key(), field.Value())
	}

	logging.Log(op.context.Logger, op.name, err, kvs...)
}

// emitMetrics will emit observe the duration, operation/result, and error counter metrics
// for this operation. This does nothing if no metric was supplied to the observation.
func (op *Operation) emitMetrics(err *error, count, elapsed float64, labels []string) {
	if op.metrics == nil {
		return
	}

	op.metrics.Observe(elapsed, count, err, labels...)
}

// finishTrace will set the error value, log additional fields supplied after the operation's
// execution, and finalize the trace span. This does nothing if no trace was constructed at
// the start of the operation.
func (op *Operation) finishTrace(err *error, tr *trace.Trace, logFields []log.Field) {
	if tr == nil {
		return
	}

	if err != nil {
		tr.SetError(*err)
	}

	tr.LogFields(logFields...)
	tr.Finish()
}

// applyErrorFilter returns nil if the given error does not pass the registered error filter.
// The original value is returned otherwise.
func (op *Operation) applyErrorFilter(err *error) *error {
	if op.errorFilter != nil && err != nil && op.errorFilter(*err) {
		return nil
	}

	return err
}

// mergeLabels flattens slices of slices of strings.
func mergeLabels(groups ...[]string) []string {
	size := 0
	for _, group := range groups {
		size += len(group)
	}

	labels := make([]string, 0, size)
	for _, group := range groups {
		labels = append(labels, group...)
	}

	return labels
}

// mergeLogFields flattens slices of slices of log fields.
func mergeLogFields(groups ...[]log.Field) []log.Field {
	size := 0
	for _, group := range groups {
		size += len(group)
	}

	logFields := make([]log.Field, 0, size)
	for _, group := range groups {
		logFields = append(logFields, group...)
	}

	return logFields
}
