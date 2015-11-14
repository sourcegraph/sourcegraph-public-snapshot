// Package instrument is an instrumentation library for Traceguide. It
// is used to record information about the duration of key request
// handlers and other functions. It also can be used to record logs
// and other information.
//
// By default, a "no-op" implementation of Runtime is set as the
// default Runtime. This implementation has minimal overhead and
// dependencies. It logs to stdout.
//
// To record Span and log information persistently to the Traceguide
// service, use the "client" implementation of Runtime, found in a
// sub-package. A minimal integration looks like this:
//
//	import (
//		"gopkg.in/resonancelabs/go-pub.v0/instrument"
//		"gopkg.in/resonancelabs/go-pub.v0/instrument/client"
//	)
//
// And in a main() function:
//
// 	instrument.SetDefaultRuntime(client.NewRuntime(&client.Options{
//		AccessToken: "<your-access-token-here>",
//		ServiceHost: "api.traceguide.io",
//	}))
//
// See Runtime and ActiveSpan for how to instrument code or use one or
// more drop-in replacements for common packages that have already
// been instrumented. These include:
//
//   instrument/logwrapper
//   instrument/glogwrapper
//   instrument/sqlwrapper
//   instrument/httpwrapper
//
package instrument

const (
	SpanAttributeDeprecatedName = "deprecated_name"
)

// An ActiveSpan represents a time interval during which a user
// interaction or some other computation occurs. For example, it may
// be the time during which a page loads or an RPC is invoked. While
// active, logs and other data may be associated with the span. Once
// finished, ActiveSpans may be recorded by Traceguide and used in
// computing statistics about user interactions.
//
// See Runtime.StartSpan() and Runtime.RunInSpan().
type ActiveSpan interface {
	Logger

	// SetOperation records the operation [name] for the Span. The operation
	// should contain no spaces and use slashes to demarcate software component
	// boundaries. For example:
	//
	//   /node/express/servlet (for an Express.js library span)
	//   /instagram/mobile/post/load_image (for the Instagram mobile app span)
	//
	SetOperation(operation string) ActiveSpan

	// DEPRECATED. Use SetOperation.
	SetName(name string) ActiveSpan

	// AddTraceJoinId adds a TraceJoinId to the span.
	//
	// Spans with multiple TraceJoinIds are used as join-points for
	// traces that have matches for either of their join keys.
	//
	// `val` is immediately converted to a string using fmt.Sprint().
	//
	AddTraceJoinId(key string, val interface{}) ActiveSpan
	// TODO: if we release this to the world, add way more explanation here, or
	// a link to a doc/diagram.

	// SetEndUserId is shorthand for
	//	AddTraceJoinId(instrument.TraceJoinKeyEndUserId, id).
	SetEndUserId(id interface{}) ActiveSpan
	// TODO maybe SetUserId? Or DisplayName? think about better names here.

	// If an attribute value for `key` already exists, it is overwritten.
	//
	// See the SpanAttribute* constants, though it's permissible to add
	// attributes that aren't in that list.
	AddAttribute(key, val string) ActiveSpan

	// Explicitly sets a parent span for this ActiveSpan.  This span
	// will inherit all TraceJoinIds of the parent span.
	//
	// Returns the ActiveSpan object itself.
	//
	SetParent(parentSpan ActiveSpan) ActiveSpan

	// Sets the parent span of this ActiveSpan based on the last span
	// set on the stack (e.g. via a call to RunInSpan() with the OnStack
	// option).
	//
	// Returns an error if there is no ActiveSpan on the stack.
	//
	SetParentFromStack() error

	// MergeTraceJoinIdsFromStack takes any TraceJoinIds from Spans on
	// the stack and merges them into the current Span.  It returns an
	// error if there are no Spans on the stack.
	//
	// For example, suppose a Span is created here:
	//
	//	instrument.RunInSpan(func(span instrument.ActiveSpan) error {
	//		span.SetEndUserId(user)
	//		maybeDoWork() // possibly calls doWork below
	//
	// And that `maybeDoWork` invokes `doWork` below in some cases.
	//
	//	func doWork() {
	//		span := instrument.StartSpan()
	//		err = span.MergeTraceJoinIdsFromStack()
	//
	// This use of MergeTraceJoinIdsFromStack ensures that the EndUserId
	// (and other TraceJoinIds) are merged into the doWork Span,
	// effectively joining these Spans into a single Trace.
	//
	MergeTraceJoinIdsFromStack() error

	// Finish ends the active Span. All other ActiveSpan methods are
	// undefined after calling Finish(). It should only be called for Spans
	// created with StartSpan().
	Finish()

	// TraceJoinIds returns any identifiers currently associated with
	// this Span.
	TraceJoinIds() map[string]string

	// Guid returns the unique identifier of this Span. GUIDs are
	// immutable and assigned automatically.
	Guid() SpanGuid
}

// SpanOptions can be provided to RunInSpan when a Span is created.
type SpanOption int

const (
	// Use goroutine-local storage to register this Span. This incurs
	// additional overhead but enables the use of
	// AddTraceJoinIdToSpansInStack and MergeTraceJoinIdsFromStack.
	OnStack SpanOption = iota
)

// RuntimeGuid uniquely identifies a Runtime. Guids are generated
// automatically by the instrumentation library.
type RuntimeGuid string

// A Runtime represents a dynamic instance of a server, client, or
// some other component of a distributed software system. A Runtime
// may have many concurrent ActiveSpans and may be simultaneously
// handling requests from many different users (though some Runtimes
// may also be user-specific). A Runtime may share some resources
// between these request handlers, including caches and connection
// pools. Typically a server process has exactly one Runtime.
type Runtime interface {
	// RunInSpan runs f in a span, returning the result of f. The Span
	// will automatically be finished when f returns. It is f's
	// responsibility to call SetOperation and add any desired TraceJoinIds
	// (including EndUserId).
	//
	// If f returns an error then that error will automatically be
	// logged.
	//
	// If OnStack is provided as an option, the Span will be registered
	// using goroutine-local storage. The Span can then be used with
	// AddTraceJoinIdToSpansInStack and MergeTraceJoinIdsFromStack.
	//
	// Example:
	//	var rval *Results
	// 	err := instrument.RunInSpan(func(s instrument.ActiveSpan) error {
	// 	  s.SetOperation("api/users/search")
	// 	  s.AddEndUserId(context.GetActiveUser().Username())
	// 	  rval = backend.SearchUsers(context, args...)
	//	  return nil
	// 	})
	//	return rval, err
	//
	RunInSpan(f func(s ActiveSpan) error, options ...SpanOption) error

	// StartSpan starts a span. The Span must be terminated by calling
	// ActiveSpan.Finish()
	//
	// Example:
	// 	func HandleRequest(...) {
	// 	  defer instrument.StartSpan().SetOperation("api/request").Finish()
	//	  handleRequest(...)
	// 	}
	//
	StartSpan() ActiveSpan

	// AddTraceJoinIdToSpansInStack looks (using goroutine-local
	// storage) for any extant ActiveSpans that were created by the
	// current goroutine using the OnStack option. If found, the given
	// TraceJoinId is added to each such ActiveSpan. This effectively
	// joins these Spans into single trace. It returns an error if there
	// are no Spans on the stack.
	//
	// See RunInSpan().
	AddTraceJoinIdToSpansInStack(key string, val interface{}) error

	Logger

	// RecordTraceJoin is used to identify 2 or more Join Ids
	// corresponding to a single end user.
	//
	// The arguments are interpreted as a sequence of key-value pairs:
	//
	//    join_key_0, join_val_0, join_key_1, join_val_1, ...,
	//      join_key_n, join_val_n
	//
	// The implementation does nothing (rather than panic/crash) if
	// there are fewer than 4 arguments or the number of arguments is
	// not even. All even arguments must be strings, and all odd
	// arguments are converted to strings.
	//
	// Prefer calling AddTraceJoinId() if an ActiveSpan is available.
	RecordTraceJoin(joinIds ...interface{})

	// MergeAttributes adds the given key-values to the Runtime
	// attributes, overwriting any values previously associated with
	// these keys.
	MergeAttributes(attrs map[string]interface{})

	// Flush forces the Runtime to flush its contents (synchronously)
	// and send them to the remote service.
	Flush()

	// Disable causes the instrumentation for this runtime to be
	// shutdown. All methods associated with this Runtime and its Spans
	// effectively become no-ops.
	Disable()
}

// Logger records information about events that occur during the
// execution of a Runtime or ActiveSpan. It may be used to record
// typical log information in string form or structured data such as
// RPC arguments.
//
// Logs may be recorded using a LogBuilder:
//	Log(instrument.Println("response").Payload(resp))
//	Log(instrument.Printf("processing took %d ms", millis))
//	Log(instrument.Print("request failed").Error().CallDepth(1).Payload(req))
// ... using a LogRecord:
//	Log(&instrument.LogRecord{RawLevel: "I", Payload: req, Message: "request received"})
// ... or using any value that is or may be converted to a string:
//	Log("request received")
//
// Callers should prefer using ActiveSpan implementations of Logger
// over Runtime implementations when possible as these record more
// specific context for the log.
type Logger interface {
	// Log records information as part of the runtime reporting. If the
	// argument is a *LogBuilder or a *LogRecord, that information will
	// be recorded in a structured way. Otherwise, the argument will be
	// converted to string using its default format.
	Log(interface{})

	// TODO(spoons): Something about newlines? Not sure we're entirely
	// consistent here... does Log always output a newline?
}

// SpanGuid uniquely identifies a Span. Guids are generated
// automatically by the instrumentation library.
type SpanGuid string

// TraceJoinKeyEndUserId is the TraceJoinKey used by Traceguide to
// identify end users. It is the most important TraceJoinKey and
// should be set whenever possible.
const TraceJoinKeyEndUserId = "end_user_id"
