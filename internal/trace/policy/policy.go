// Package policy exports functionality related to whether or not to trace.
package policy

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/atomic"
)

type TracePolicy string

const (
	// TraceNone turns off tracing.
	TraceNone TracePolicy = "none"

	// TraceSelective turns on tracing only for requests with the X-Sourcegraph-Should-Trace header
	// set to a truthy value.
	TraceSelective TracePolicy = "selective"

	// TraceAll turns on tracing for all requests.
	TraceAll TracePolicy = "all"
)

var trPolicy = atomic.NewString(string(TraceNone))

func SetTracePolicy(newTracePolicy TracePolicy) {
	trPolicy.Store(string(newTracePolicy))
}

func GetTracePolicy() TracePolicy {
	return TracePolicy(trPolicy.Load())
}

type key int

const shouldTraceKey key = iota

// ShouldTrace returns true if the shouldTraceKey context value is true. If the
// value is not set at all, we check if we should the global policy is set to
// TraceAll instead. It is generally used when propagating configured trace
// policies.
//
// It should generally NOT be used to determine if a span should be started at
// all, as you should always start a span if you might want one - internal
// samplers will drop traces that shouldn't be sampled on export for you based on
// trace policy.
//
// It should NOT be used to guarantee if a span is present in context. The OpenTelemetry
// library may provide a no-op span with trace.SpanFromContext, but the
// opentracing.SpanFromContext function in particular can provide a nil span if no span
// is provided.
func ShouldTrace(ctx context.Context) bool {
	v, ok := ctx.Value(shouldTraceKey).(bool)
	if !ok {
		// If ShouldTrace is not set, we respect TraceAll instead.
		return GetTracePolicy() == TraceAll
	}
	return v
}

// WithShouldTrace sets the shouldTraceKey context value.
func WithShouldTrace(ctx context.Context, shouldTrace bool) context.Context {
	return context.WithValue(ctx, shouldTraceKey, shouldTrace)
}

const (
	traceHeader = "X-Sourcegraph-Should-Trace"
	traceQuery  = "trace"
)

// Transport wraps an underlying HTTP RoundTripper, injecting the X-Sourcegraph-Should-Trace header
// into outgoing requests whenever the shouldTraceKey context value is true.
type Transport struct {
	RoundTripper http.RoundTripper
}

func (r *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(traceHeader, strconv.FormatBool(ShouldTrace(req.Context())))
	return r.RoundTripper.RoundTrip(req)
}

// requestWantsTrace returns true if a request is opting into tracing either
// via our HTTP Header or our URL Query.
func RequestWantsTracing(r *http.Request) bool {
	// Prefer header over query param.
	if v := r.Header.Get(traceHeader); v != "" {
		b, _ := strconv.ParseBool(v)
		return b
	}
	// PERF: Avoid parsing RawQuery if "trace=" is not present
	if strings.Contains(r.URL.RawQuery, "trace=") {
		v := r.URL.Query().Get(traceQuery)
		b, _ := strconv.ParseBool(v)
		return b
	}
	return false
}
