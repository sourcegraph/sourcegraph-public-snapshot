package instrument

import "sync"

var (
	defaultRuntime     Runtime
	defaultRuntimeLock sync.RWMutex
)

// DefaultRuntime returns the global Runtime for this process. Most
// users will not invoke it directly; it is used internally by many of
// the functions in this package. If SetDefaultRuntime has not been
// called, the default Runtime will be initialized to a "no-op"
// implementation.
func DefaultRuntime() Runtime {
	// Try to keep the common case fast.
	defaultRuntimeLock.RLock()
	if defaultRuntime != nil {
		defer defaultRuntimeLock.RUnlock()
		return defaultRuntime
	}
	// Exchange reader lock for writer lock.
	defaultRuntimeLock.RUnlock()
	defaultRuntimeLock.Lock()
	defer defaultRuntimeLock.Unlock()
	// Double check that it's still nil.
	if defaultRuntime == nil {
		defaultRuntime = NewNopRuntime()
		defaultRuntime.Log(FileLine(1).Info().
			Print("Traceguide default Runtime set to NOP Runtime"))
	}
	return defaultRuntime
}

// SetDefaultRuntime sets the global Runtime used by many of the
// functions in this package.
func SetDefaultRuntime(runtime Runtime) {
	// TODO: bhs: Maybe clearer as Initialize(runtime Runtime) since
	// that's how people will think of this, I expect.

	defaultRuntimeLock.Lock()
	defer defaultRuntimeLock.Unlock()
	defaultRuntime = runtime
}

// Also include global implementation of Runtime method that dispatch
// to DefaultRuntime().

// --- Runtime-proper methods ---

// RunInSpan runs f in a new Span using the default Runtime.
func RunInSpan(f func(s ActiveSpan) error, options ...SpanOption) error {
	return DefaultRuntime().RunInSpan(f, options...)
}

// StartSpan creates a new Span in the default Runtime.
func StartSpan() ActiveSpan {
	return DefaultRuntime().StartSpan()
}

// AddTraceJoinIdToSpansInStack adds TraceJoinIds using the default
// Runtime.
func AddTraceJoinIdToSpansInStack(key string, value interface{}) error {
	return DefaultRuntime().AddTraceJoinIdToSpansInStack(key, value)
}

// RecordTraceJoin joins 2 or more TraceJoinIds using the default
// Runtime.
func RecordTraceJoin(joinIds ...interface{}) {
	DefaultRuntime().RecordTraceJoin(joinIds...)
}

// MergeAttributes updates attributes of the default Runtime.
func MergeAttributes(attrs map[string]interface{}) {
	DefaultRuntime().MergeAttributes(attrs)
}

// Flush invokes Flush on the default Runtime.
func Flush() { DefaultRuntime().Flush() }

// Disable invokes Disable on the default Runtime.
func Disable() { DefaultRuntime().Disable() }

// --- Logger and LogBuilder methods ---

// Log records log information in the default Runtime.
func Log(arg interface{}) {
	DefaultRuntime().Log(arg)
}

// See LogBuilder.Print.
func Print(args ...interface{}) *LogBuilder {
	return (&LogBuilder{}).Print(args...)
}

// See LogBuilder.Println.
func Println(args ...interface{}) *LogBuilder {
	return (&LogBuilder{}).Println(args...)
}

// See LogBuilder.Printf.
func Printf(format string, args ...interface{}) *LogBuilder {
	return (&LogBuilder{}).Printf(format, args...)
}

// See LogBuilder.Payload.
func Payload(payload interface{}) *LogBuilder {
	return (&LogBuilder{}).Payload(payload)
}

// See LogBuilder.EventName.
func EventName(name string) *LogBuilder {
	return (&LogBuilder{}).EventName(name)
}

// See LogBuilder.FileLine.
func FileLine(depth int) *LogBuilder {
	return (&LogBuilder{}).FileLine(depth)
}
