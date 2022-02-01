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
//     metrics := metrics.NewREDMetrics(
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
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// Context carries context about where to send logs, trace spans, and register
// metrics. It should be created once on service startup, and passed around to
// any location that wants to use it for observing operations.
type Context struct {
	Logger       logging.ErrorLogger
	Tracer       *trace.Tracer
	Registerer   prometheus.Registerer
	HoneyDataset *honey.Dataset
	Sentry       *sentry.Hub
}

// TestContext is a behaviorless Context usable for unit tests.
var TestContext = Context{Registerer: metrics.TestRegisterer}

var TestOperation *Operation

type ErrorFilterBehaviour uint8

const (
	EmitForNone    ErrorFilterBehaviour = 0
	EmitForMetrics ErrorFilterBehaviour = 1 << iota
	EmitForLogs
	EmitForTraces
	EmitForHoney
	EmitForSentry

	EmitForDefault = EmitForMetrics | EmitForLogs | EmitForTraces
)

// Op configures an Operation instance.
type Op struct {
	// Metrics sets the RED metrics triplet used to monitor & track metrics for this operation.
	// This field is optional, with `nil` meaning no metrics will be tracked for this.
	Metrics *metrics.REDMetrics
	// Name configures the trace and error log names. This string should be of the
	// format {GroupName}.{OperationName}, where both sections are title cased
	// (e.g. Store.GetRepoByID).
	Name string
	// MetricLabelValues that apply for every invocation of this operation.
	MetricLabelValues []string
	// LogFields that apply for for every invocation of this operation.
	LogFields []log.Field
	// ErrorFilter returns true for any error that should be converted to nil
	// for the purposes of metrics and tracing. If this field is not set then
	// error values are unaltered.
	//
	// This is useful when, for example, a revision not found error is expected by
	// a process interfacing with gitserver. Such an error should not be treated as
	// an unexpected value in metrics and traces but should be handled higher up in
	// the stack.
	ErrorFilter func(err error) ErrorFilterBehaviour
}

// Operation combines the state of the parent context to create a new operation. This value
// should be owned and used by the code that performs the operation it represents.
func (c *Context) Operation(args Op) *Operation {
	return &Operation{
		context:      c,
		metrics:      args.Metrics,
		name:         args.Name,
		kebabName:    kebabCase(args.Name),
		metricLabels: args.MetricLabelValues,
		logFields:    args.LogFields,
		errorFilter:  args.ErrorFilter,
	}
}

// Operation represents an interesting section of code that can be invoked.
type Operation struct {
	context      *Context
	metrics      *metrics.REDMetrics
	errorFilter  func(err error) ErrorFilterBehaviour
	name         string
	kebabName    string
	metricLabels []string
	logFields    []log.Field
}

// TraceLogger is returned from WithAndLogger and can be used to add timestamped key and
// value pairs into a related opentracing span.
type TraceLogger interface {
	Log(fields ...log.Field)
	Tag(fields ...log.Field)
}

var TestTraceLogger = &traceLogger{}

type traceLogger struct {
	opName string
	event  honey.Event
	trace  *trace.Trace
}

func (t traceLogger) Log(fields ...log.Field) {
	if honey.Enabled() {
		for _, field := range fields {
			t.event.AddField(t.opName+"."+toSnakeCase(field.Key()), field.Value())
		}
	}
	if t.trace != nil {
		t.trace.LogFields(fields...)
	}
}

func (t traceLogger) Tag(fields ...log.Field) {
	if honey.Enabled() {
		for _, field := range fields {
			t.event.AddField(t.opName+"."+toSnakeCase(field.Key()), field.Value())
		}
	}
	if t.trace != nil {
		t.trace.TagFields(fields...)
	}
}

// FinishFunc is the shape of the function returned by With and should be invoked within
// a defer directly before the observed function returns or when a context is cancelled
// with OnCancel.
type FinishFunc func(count float64, args Args)

// OnCancel allows for ending an observation when a context is cancelled as opposed to the
// more common scenario of when the observed function returns through a defer. This can
// be used for continuing an observation beyond the lifetime of a function if that function
// returns more units of work that you want to observe as part of the original function.
func (f FinishFunc) OnCancel(ctx context.Context, count float64, args Args) {
	go func() {
		<-ctx.Done()
		f(count, args)
	}()
}

// ErrCollector represents multiple errors and additional log fields that arose from those errors.
type ErrCollector struct {
	multi       *multierror.Error
	extraFields []log.Field
}

func NewErrorCollector() *ErrCollector { return &ErrCollector{multi: &multierror.Error{}} }

func (e *ErrCollector) Collect(err *error, fields ...log.Field) {
	if err != nil && *err != nil {
		e.multi.Errors = append(e.multi.Errors, *err)
		e.extraFields = append(e.extraFields, fields...)
	}
}

func (e *ErrCollector) Error() string {
	return e.multi.Error()
}

// Args configures the observation behavior of an invocation of an operation.
type Args struct {
	// MetricLabelValues that apply only to this invocation of the operation.
	MetricLabelValues []string
	// LogFields that apply only to this invocation of the operation.
	LogFields []log.Field
}

// LogFieldMap returns a string-to-interface map containing the contents of this Arg value's
// log fields.
func (args Args) LogFieldMap() map[string]interface{} {
	fields := make(map[string]interface{}, len(args.LogFields))
	for _, field := range args.LogFields {
		fields[field.Key()] = field.Value()
	}

	return fields
}

// LogFieldPairs returns a slice of key, value, key, value, ... pairs containing the contents
// of this Arg value's log fields.
func (args Args) LogFieldPairs() []interface{} {
	pairs := make([]interface{}, 0, len(args.LogFields)*2)
	for _, field := range args.LogFields {
		pairs = append(pairs, field.Key(), field.Value())
	}

	return pairs
}

// WithErrors prepares the necessary timers, loggers, and metrics to observe the invocation of an
// operation. This method returns a modified context, an multi-error capturing type and a function to be deferred until the
// end of the operation. It can be used with FinishFunc.OnCancel to capture multiple async errors.
func (op *Operation) WithErrors(ctx context.Context, root *error, args Args) (context.Context, *ErrCollector, FinishFunc) {
	ctx, collector, _, endObservation := op.WithErrorsAndLogger(ctx, root, args)
	return ctx, collector, endObservation
}

// WithErrorsAndLogger prepares the necessary timers, loggers, and metrics to observe the invocation of an
// operation. This method returns a modified context, an multi-error capturing type, a function that will add a log field
// to the active trace, and a function to be deferred until the end of the operation. It can be used with
// FinishFunc.OnCancel to capture multiple async errors.
func (op *Operation) WithErrorsAndLogger(ctx context.Context, root *error, args Args) (context.Context, *ErrCollector, TraceLogger, FinishFunc) {
	errTracer := NewErrorCollector()
	err := error(errTracer)

	ctx, traceLogger, endObservation := op.WithAndLogger(ctx, &err, args)

	// to avoid recursion stack overflow, we need a new binding
	endFunc := endObservation

	if root != nil {
		endFunc = func(count float64, args Args) {
			if *root != nil {
				errTracer.multi.Errors = append(errTracer.multi.Errors, *root)
			}
			endObservation(count, args)
		}
	}
	return ctx, errTracer, traceLogger, endFunc
}

// With prepares the necessary timers, loggers, and metrics to observe the invocation of an
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
	tr, ctx := op.startTrace(ctx, args)

	event := honey.NoopEvent()
	snakecaseOpName := toSnakeCase(op.name)
	if op.context.HoneyDataset != nil {
		event = op.context.HoneyDataset.EventWithFields(map[string]interface{}{
			"operation":     snakecaseOpName,
			"meta.hostname": hostname.Get(),
			"meta.version":  version.Version(),
		})
	}

	logFields := traceLogger{
		opName: snakecaseOpName,
		event:  event,
		trace:  tr,
	}

	if mergedFields := mergeLogFields(op.logFields, args.LogFields); len(mergedFields) > 0 {
		logFields.Tag(mergedFields...)
	}

	if traceID := trace.ID(ctx); traceID != "" {
		event.AddField("traceID", traceID)
	}

	return ctx, logFields, func(count float64, finishArgs Args) {
		since := time.Since(start)
		elapsed := since.Seconds()
		elapsedMs := since.Milliseconds()
		defaultFinishFields := []log.Field{log.Float64("count", count), log.Float64("elapsed", elapsed)}
		logFields := mergeLogFields(defaultFinishFields, finishArgs.LogFields, args.LogFields)
		metricLabels := mergeLabels(op.metricLabels, args.MetricLabelValues, finishArgs.MetricLabelValues)

		if multi := new(ErrCollector); err != nil && errors.As(*err, &multi) {
			if len(multi.multi.Errors) == 0 {
				err = nil
			}
			logFields = append(logFields, multi.extraFields...)
		}

		var (
			logErr     = op.applyErrorFilter(err, EmitForLogs)
			metricsErr = op.applyErrorFilter(err, EmitForMetrics)
			traceErr   = op.applyErrorFilter(err, EmitForTraces)
			honeyErr   = op.applyErrorFilter(err, EmitForHoney)
			sentryErr  = op.applyErrorFilter(err, EmitForSentry)
		)
		op.emitErrorLogs(logErr, logFields)
		op.emitHoneyEvent(honeyErr, snakecaseOpName, event, finishArgs.LogFields, elapsedMs) // op. and args.LogFields already added at start
		op.emitMetrics(metricsErr, count, elapsed, metricLabels)
		op.finishTrace(traceErr, tr, logFields)
		op.emitSentryError(sentryErr, logFields)
	}
}

// startTrace creates a new Trace object and returns the wrapped context. This returns
// an unmodified context and a nil startTrace if no tracer was supplied on the observation context.
func (op *Operation) startTrace(ctx context.Context, args Args) (*trace.Trace, context.Context) {
	if op.context.Tracer == nil {
		return nil, ctx
	}

	tr, ctx := op.context.Tracer.New(ctx, op.kebabName, "")
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

func (op *Operation) emitHoneyEvent(err *error, opName string, event honey.Event, logFields []log.Field, duration int64) {
	if err != nil && *err != nil {
		event.AddField("error", (*err).Error())
	}

	event.AddField("duration_ms", duration)

	for _, field := range logFields {
		event.AddField(opName+"."+toSnakeCase(field.Key()), field.Value())
	}

	event.Send()
}

// emitSentryError will send errors to Sentry.
func (op *Operation) emitSentryError(err *error, logFields []log.Field) {
	if err == nil || *err == nil {
		return
	}

	if op.context.Sentry == nil {
		return
	}

	logs := make(map[string]string)
	for _, field := range logFields {
		logs[field.Key()] = fmt.Sprintf("%v", field.Value())
	}

	logs["operation"] = op.name

	op.context.Sentry.CaptureError(*err, logs)
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
func (op *Operation) applyErrorFilter(err *error, behaviour ErrorFilterBehaviour) *error {
	if op.errorFilter != nil && err != nil && op.errorFilter(*err)&behaviour == 0 {
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
