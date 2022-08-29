// Package observation provides a unified way to wrap an operation with logging, tracing, and metrics.
//
// To learn more, refer to "How to add observability": https://docs.sourcegraph.com/dev/how-to/add_observability
package observation

import (
	"context"
	"os"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// enableTraceLog toggles whether TraceLogger.Log events should be logged at info level,
// which is useful in environments like Datadog that don't support OpenTrace/OpenTelemetry
// trace log events.
var enableTraceLog = os.Getenv("SRC_TRACE_LOG") == "true"

// Context carries context about where to send logs, trace spans, and register
// metrics. It should be created once on service startup, and passed around to
// any location that wants to use it for observing operations.
type Context struct {
	Logger       log.Logger
	Tracer       *trace.Tracer
	Registerer   prometheus.Registerer
	HoneyDataset *honey.Dataset
}

// TestContext is a behaviorless Context usable for unit tests.
var TestContext = Context{Logger: log.Scoped("TestContext", ""), Registerer: metrics.TestRegisterer}

type ErrorFilterBehaviour uint8

const (
	EmitForNone    ErrorFilterBehaviour = 0
	EmitForMetrics ErrorFilterBehaviour = 1 << iota
	EmitForLogs
	EmitForTraces
	EmitForHoney

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
	// Description is a simple description for this Op.
	Description string
	// MetricLabelValues that apply for every invocation of this operation.
	MetricLabelValues []string
	// LogFields that apply for every invocation of this operation.
	LogFields []otlog.Field
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
	var logger log.Logger
	if c.Logger != nil {
		// Create a child logger, if a parent is provided.
		logger = c.Logger.Scoped(args.Name, args.Description)
	} else {
		// Create a new logger.
		logger = log.Scoped(args.Name, args.Description)
	}
	return &Operation{
		context:      c,
		metrics:      args.Metrics,
		name:         args.Name,
		kebabName:    kebabCase(args.Name),
		metricLabels: args.MetricLabelValues,
		logFields:    args.LogFields,
		errorFilter:  args.ErrorFilter,

		Logger: logger.With(toLogFields(args.LogFields)...),
	}
}

// Operation represents an interesting section of code that can be invoked. It has an
// embedded Logger that can be used directly.
type Operation struct {
	context      *Context
	metrics      *metrics.REDMetrics
	errorFilter  func(err error) ErrorFilterBehaviour
	name         string
	kebabName    string
	metricLabels []string
	logFields    []otlog.Field

	// Logger is a logger scoped to this operation. Must not be nil.
	log.Logger
}

// TraceLogger is returned from With and can be used to add timestamped key and
// value pairs into a related opentracing span. It has an embedded Logger that can be used
// directly to log messages in the context of a trace.
type TraceLogger interface {
	// Log logs and event with fields to the opentracing.Span as well as the nettrace.Trace,
	// and also logs an 'trace.event' log entry at INFO level with the fields, including
	// any existing tags and parent observation context.
	Log(fields ...otlog.Field)

	// Tag adds fields to the opentracing.Span as tags as well as as logs to the nettrace.Trace.
	//
	// Tag will add fields to the underlying logger.
	Tag(fields ...otlog.Field)

	// Logger is a logger scoped to this trace.
	log.Logger
}

// TestTraceLogger creates an empty TraceLogger that can be used for testing. The logger
// should be 'logtest.Scoped(t)'.
func TestTraceLogger(logger log.Logger) TraceLogger {
	return &traceLogger{Logger: logger}
}

type traceLogger struct {
	opName string
	event  honey.Event
	trace  *trace.Trace

	log.Logger
}

// initWithTags adds tags to everything except the underlying Logger, which should
// already have init fields due to being spawned from a parent Logger.
func (t *traceLogger) initWithTags(fields ...otlog.Field) {
	if honey.Enabled() {
		for _, field := range fields {
			t.event.AddField(t.opName+"."+toSnakeCase(field.Key()), field.Value())
		}
	}
	if t.trace != nil {
		t.trace.TagFields(fields...)
	}
}

func (t *traceLogger) Log(fields ...otlog.Field) {
	if honey.Enabled() {
		for _, field := range fields {
			t.event.AddField(t.opName+"."+toSnakeCase(field.Key()), field.Value())
		}
	}
	if t.trace != nil {
		t.trace.LogFields(fields...)
		if enableTraceLog {
			t.Logger.
				AddCallerSkip(1). // Log() -> Logger
				Info("trace.log", toLogFields(fields)...)
		}
	}
}

func (t *traceLogger) Tag(fields ...otlog.Field) {
	if honey.Enabled() {
		for _, field := range fields {
			t.event.AddField(t.opName+"."+toSnakeCase(field.Key()), field.Value())
		}
	}
	if t.trace != nil {
		t.trace.TagFields(fields...)
	}
	t.Logger = t.Logger.With(toLogFields(fields)...)
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
	errs        error
	extraFields []otlog.Field
}

func NewErrorCollector() *ErrCollector { return &ErrCollector{errs: nil} }

func (e *ErrCollector) Collect(err *error, fields ...otlog.Field) {
	if err != nil && *err != nil {
		e.errs = errors.Append(e.errs, *err)
		e.extraFields = append(e.extraFields, fields...)
	}
}

func (e *ErrCollector) Error() string {
	if e.errs == nil {
		return ""
	}
	return e.errs.Error()
}

// Args configures the observation behavior of an invocation of an operation.
type Args struct {
	// MetricLabelValues that apply only to this invocation of the operation.
	MetricLabelValues []string
	// LogFields that apply only to this invocation of the operation.
	LogFields []otlog.Field
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

	ctx, traceLogger, endObservation := op.With(ctx, &err, args)

	// to avoid recursion stack overflow, we need a new binding
	endFunc := endObservation

	if root != nil {
		endFunc = func(count float64, args Args) {
			if *root != nil {
				errTracer.errs = errors.Append(errTracer.errs, *root)
			}
			endObservation(count, args)
		}
	}
	return ctx, errTracer, traceLogger, endFunc
}

// With prepares the necessary timers, loggers, and metrics to observe the invocation
// of an operation. This method returns a modified context, a function that will add a log field
// to the active trace, and a function to be deferred until the end of the operation.
func (op *Operation) With(ctx context.Context, err *error, args Args) (context.Context, TraceLogger, FinishFunc) {
	parentTraceContext := trace.Context(ctx)
	start := time.Now()
	tr, ctx := op.startTrace(ctx)

	event := honey.NoopEvent()
	snakecaseOpName := toSnakeCase(op.name)
	if op.context.HoneyDataset != nil {
		event = op.context.HoneyDataset.EventWithFields(map[string]any{
			"operation":     snakecaseOpName,
			"meta.hostname": hostname.Get(),
			"meta.version":  version.Version(),
		})
	}

	logger := op.Logger.With(toLogFields(args.LogFields)...)

	if traceContext := trace.Context(ctx); traceContext.TraceID != "" {
		event.AddField("trace.trace_id", traceContext.TraceID)
		event.AddField("trace.span_id", traceContext.SpanID)
		if parentTraceContext.SpanID != "" {
			event.AddField("trace.parent_id", parentTraceContext.SpanID)
		}
		logger = logger.WithTrace(traceContext)
	}

	trLogger := &traceLogger{
		opName: snakecaseOpName,
		event:  event,
		trace:  tr,
		Logger: logger,
	}

	if mergedFields := mergeLogFields(op.logFields, args.LogFields); len(mergedFields) > 0 {
		trLogger.initWithTags(mergedFields...)
	}

	return ctx, trLogger, func(count float64, finishArgs Args) {
		since := time.Since(start)
		elapsed := since.Seconds()
		elapsedMs := since.Milliseconds()
		defaultFinishFields := []otlog.Field{otlog.Float64("count", count), otlog.Float64("elapsed", elapsed)}
		finishLogFields := mergeLogFields(defaultFinishFields, finishArgs.LogFields)

		logFields := mergeLogFields(defaultFinishFields, finishLogFields)
		metricLabels := mergeLabels(op.metricLabels, args.MetricLabelValues, finishArgs.MetricLabelValues)

		if multi := new(ErrCollector); err != nil && errors.As(*err, &multi) {
			if multi.errs == nil {
				err = nil
			}
			logFields = append(logFields, multi.extraFields...)
		}

		var (
			logErr     = op.applyErrorFilter(err, EmitForLogs)
			metricsErr = op.applyErrorFilter(err, EmitForMetrics)
			traceErr   = op.applyErrorFilter(err, EmitForTraces)
			honeyErr   = op.applyErrorFilter(err, EmitForHoney)
		)

		// already has all the other log fields
		op.emitErrorLogs(trLogger, logErr, finishLogFields)

		// op. and args.LogFields already added at start
		op.emitHoneyEvent(honeyErr, snakecaseOpName, event, finishArgs.LogFields, elapsedMs)

		op.emitMetrics(metricsErr, count, elapsed, metricLabels)
		op.finishTrace(traceErr, tr, logFields)
	}
}

// startTrace creates a new Trace object and returns the wrapped context. This returns
// an unmodified context and a nil startTrace if no tracer was supplied on the observation context.
func (op *Operation) startTrace(ctx context.Context) (*trace.Trace, context.Context) {
	if op.context.Tracer == nil {
		return nil, ctx
	}

	tr, ctx := op.context.Tracer.New(ctx, op.kebabName, "")
	return tr, ctx
}

// emitErrorLogs will log as message if the operation has failed. This log contains the error
// as well as all of the log fields attached ot the operation, the args to With, and the args
// to the finish function.
func (op *Operation) emitErrorLogs(trLogger TraceLogger, err *error, logFields []otlog.Field) {
	if err == nil || *err == nil {
		return
	}
	fields := append(toLogFields(logFields), log.Error(*err))

	trLogger.
		AddCallerSkip(2). // callback() -> emitErrorLogs() -> Logger
		Error("operation.error", fields...)
}

func (op *Operation) emitHoneyEvent(err *error, opName string, event honey.Event, logFields []otlog.Field, duration int64) {
	if err != nil && *err != nil {
		event.AddField("error", (*err).Error())
	}

	event.AddField("duration_ms", duration)

	for _, field := range logFields {
		event.AddField(opName+"."+toSnakeCase(field.Key()), field.Value())
	}

	event.Send()
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
func (op *Operation) finishTrace(err *error, tr *trace.Trace, logFields []otlog.Field) {
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
func mergeLogFields(groups ...[]otlog.Field) []otlog.Field {
	size := 0
	for _, group := range groups {
		size += len(group)
	}

	logFields := make([]otlog.Field, 0, size)
	for _, group := range groups {
		logFields = append(logFields, group...)
	}

	return logFields
}
