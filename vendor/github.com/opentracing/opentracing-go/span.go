package opentracing

import "time"

// SpanContext represents Span state that must propagate to descendant Spans and across process
// boundaries (e.g., a <trace_id, span_id, sampled> tuple).
type SpanContext interface {
	// ForeachBaggageItem grants access to all baggage items stored in the
	// SpanContext.
	// The handler function will be called for each baggage key/value pair.
	// The ordering of items is not guaranteed.
	//
	// The bool return value indicates if the handler wants to continue iterating
	// through the rest of the baggage items; for example if the handler is trying to
	// find some baggage item by pattern matching the name, it can return false
	// as soon as the item is found to stop further iterations.
	ForeachBaggageItem(handler func(k, v string) bool)
}

// Span represents an active, un-finished span in the OpenTracing system.
//
// Spans are created by the Tracer interface.
type Span interface {
	// Sets the end timestamp and finalizes Span state.
	//
	// With the exception of calls to Context() (which are always allowed),
	// Finish() must be the last call made to any span instance, and to do
	// otherwise leads to undefined behavior.
	Finish()
	// FinishWithOptions is like Finish() but with explicit control over
	// timestamps and log data.
	FinishWithOptions(opts FinishOptions)

	// Context() yields the SpanContext for this Span. Note that the return
	// value of Context() is still valid after a call to Span.Finish(), as is
	// a call to Span.Context() after a call to Span.Finish().
	Context() SpanContext

	// Sets or changes the operation name.
	SetOperationName(operationName string) Span

	// Adds a tag to the span.
	//
	// If there is a pre-existing tag set for `key`, it is overwritten.
	//
	// Tag values can be numeric types, strings, or bools. The behavior of
	// other tag value types is undefined at the OpenTracing level. If a
	// tracing system does not know how to handle a particular value type, it
	// may ignore the tag, but shall not panic.
	SetTag(key string, value interface{}) Span

	// LogEvent() is equivalent to
	//
	//   Log(LogData{Event: event})
	//
	LogEvent(event string)

	// LogEventWithPayload() is equivalent to
	//
	//   Log(LogData{Event: event, Payload: payload0})
	//
	LogEventWithPayload(event string, payload interface{})

	// Log() records `data` to this Span.
	//
	// See LogData for semantic details.
	Log(data LogData)

	// SetBaggageItem sets a key:value pair on this Span and its SpanContext
	// that also propagates to descendants of this Span.
	//
	// SetBaggageItem() enables powerful functionality given a full-stack
	// opentracing integration (e.g., arbitrary application data from a mobile
	// app can make it, transparently, all the way into the depths of a storage
	// system), and with it some powerful costs: use this feature with care.
	//
	// IMPORTANT NOTE #1: SetBaggageItem() will only propagate baggage items to
	// *future* causal descendants of the associated Span.
	//
	// IMPORTANT NOTE #2: Use this thoughtfully and with care. Every key and
	// value is copied into every local *and remote* child of the associated
	// Span, and that can add up to a lot of network and cpu overhead.
	//
	// Returns a reference to this Span for chaining.
	SetBaggageItem(restrictedKey, value string) Span

	// Gets the value for a baggage item given its key. Returns the empty string
	// if the value isn't found in this Span.
	BaggageItem(restrictedKey string) string

	// Provides access to the Tracer that created this Span.
	Tracer() Tracer
}

// LogData is data associated with a Span. Every LogData instance should
// specify at least one of Event and/or Payload.
type LogData struct {
	// The timestamp of the log record; if set to the default value (the unix
	// epoch), implementations should use time.Now() implicitly.
	Timestamp time.Time

	// Event (if non-empty) should be the stable name of some notable moment in
	// the lifetime of a Span. For instance, a Span representing a browser page
	// load might add an Event for each of the Performance.timing moments
	// here: https://developer.mozilla.org/en-US/docs/Web/API/PerformanceTiming
	//
	// While it is not a formal requirement, Event strings will be most useful
	// if they are *not* unique; rather, tracing systems should be able to use
	// them to understand how two similar Spans relate from an internal timing
	// perspective.
	Event string

	// Payload is a free-form potentially structured object which Tracer
	// implementations may retain and record all, none, or part of.
	//
	// If included, `Payload` should be restricted to data derived from the
	// instrumented application; in particular, it should not be used to pass
	// semantic flags to a Log() implementation.
	//
	// For example, an RPC system could log the wire contents in both
	// directions, or a SQL library could log the query (with or without
	// parameter bindings); tracing implementations may truncate or otherwise
	// record only a snippet of these payloads (or may strip out PII, etc,
	// etc).
	Payload interface{}
}

// FinishOptions allows Span.FinishWithOptions callers to override the finish
// timestamp and provide log data via a bulk interface.
type FinishOptions struct {
	// FinishTime overrides the Span's finish time, or implicitly becomes
	// time.Now() if FinishTime.IsZero().
	//
	// FinishTime must resolve to a timestamp that's >= the Span's StartTime
	// (per StartSpanOptions).
	FinishTime time.Time

	// BulkLogData allows the caller to specify the contents of many Log()
	// calls with a single slice. May be nil.
	//
	// None of the LogData.Timestamp values may be .IsZero() (i.e., they must
	// be set explicitly). Also, they must be >= the Span's start timestamp and
	// <= the FinishTime (or time.Now() if FinishTime.IsZero()). Otherwise the
	// behavior of FinishWithOptions() is undefined.
	//
	// If specified, the caller hands off ownership of BulkLogData at
	// FinishWithOptions() invocation time.
	BulkLogData []LogData
}
