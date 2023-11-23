package instrumentation

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// defaultOTELHTTPOptions is a set of options shared between instrumetned HTTP middleware
// and HTTP clients for consistent Sourcegraph-preferred behaviour.
var defaultOTELHTTPOptions = []otelhttp.Option{
	// Supplemental trace policy management - our core trace policy management
	// is implemented in internal/tracer.
	otelhttp.WithTracerProvider(&samplingRetainTracerProvider{}),
	// Uniform span names
	otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
		// If incoming, just include the path since our own host is not
		// very interesting. If outgoing, include the host as well.
		target := r.URL.Path
		if r.RemoteAddr == "" { // no RemoteAddr indicates this is an outgoing request
			target = r.Host + target
		}
		if operation != "" {
			return fmt.Sprintf("%s.%s %s", operation, r.Method, target)
		}
		return fmt.Sprintf("%s %s", r.Method, target)
	}),
	// Disable OTEL metrics which can be quite high-cardinality by setting
	// a no-op MeterProvider.
	otelhttp.WithMeterProvider(noop.NewMeterProvider()),
	// Make sure we use the global propagator, which should be set up on
	// service initialization to support all our commonly used propagation
	// formats (OpenTelemetry, W3c, Jaeger, etc)
	otelhttp.WithPropagators(otel.GetTextMapPropagator()),
}

// HTTPMiddleware wraps the handler with the following:
//
//   - If the HTTP header, X-Sourcegraph-Should-Trace, is set to a truthy value, set the
//     shouldTraceKey context.Context value to true
//   - go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp, which applies the
//     desired instrumentation, including picking up traces propagated in the request headers
//     using the globally configured propagator.
//
// The provided operation name is used to add details to spans.
func HTTPMiddleware(operation string, next http.Handler, opts ...otelhttp.Option) http.Handler {
	afterInstrumentedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set X-Trace after otelhttp's handler which starts the trace. The
		// top-level trace should be an OTEL trace, so we use otel/trace to
		// extract it. Then, we add it to the header before next writes the
		// header back to client.
		span := trace.SpanContextFromContext(r.Context())
		if span.IsValid() {
			// We only set the trace ID here. The trace URL is set to
			// X-Trace-URL by httptrace.HTTPMiddleware that does some more
			// elaborate handling. In particular, we don't want to introduce
			// a conf.Get() dependency here to build the trace URL, since we
			// want this to be fairly bare-bones for use in standalone services
			// like Cody Gateway.
			w.Header().Set("X-Trace", span.TraceID().String())
			w.Header().Set("X-Trace-Span", span.SpanID().String())
		}

		next.ServeHTTP(w, r)
	})

	instrumentedHandler := otelhttp.NewHandler(afterInstrumentedHandler, operation,
		append(defaultOTELHTTPOptions, opts...)...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set up trace policy before instrumented handler
		var shouldTrace bool
		switch policy.GetTracePolicy() {
		case policy.TraceSelective:
			shouldTrace = policy.RequestWantsTracing(r)
		case policy.TraceAll:
			shouldTrace = true
		default:
			shouldTrace = false
		}
		// Pass through to instrumented handler with trace policy in context
		instrumentedHandler.ServeHTTP(w, r.WithContext(policy.WithShouldTrace(r.Context(), shouldTrace)))
	})
}

// Experimental: it order to mitigate the amount of traces sent by components which are not
// respecting the tracing policy, we can delegate the final decision to the collector,
// and merely indicate that when it's selective or all, we want requests to be retained.
//
// By setting "sampling.retain" attribute on the span, a sampling policy will match on the OTEL Collector
// and explicitly sample (i.e keep it) the present trace.
//
// To achieve that, it shims the default TracerProvider with samplingRetainTracerProvider to inject
// the attribute at the beginning of the span, which is mandatory to perform sampling.
type samplingRetainTracerProvider struct{ embedded.TracerProvider }
type samplingRetainTracer struct {
	tracer trace.Tracer
	embedded.Tracer
}

func (p *samplingRetainTracerProvider) Tracer(instrumentationName string, opts ...trace.TracerOption) trace.Tracer {
	return &samplingRetainTracer{tracer: otel.GetTracerProvider().Tracer(instrumentationName, opts...)}
}

// samplingRetainKey is the attribute key used to mark as span as to be retained.
var samplingRetainKey = "sampling.retain"

// Start will only inject the attribute if this trace has been explictly asked to be traced.
func (t *samplingRetainTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if policy.ShouldTrace(ctx) {
		attrOpts := []trace.SpanStartOption{
			trace.WithAttributes(attribute.String(samplingRetainKey, "true")),
		}
		return t.tracer.Start(ctx, spanName, append(attrOpts, opts...)...)
	}
	return t.tracer.Start(ctx, spanName, opts...)
}

// NewHTTPTransport creates an http.RoundTripper that instruments all requests using
// OpenTelemetry and a default set of OpenTelemetry options.
func NewHTTPTransport(base http.RoundTripper, opts ...otelhttp.Option) *otelhttp.Transport {
	return otelhttp.NewTransport(base, append(defaultOTELHTTPOptions, opts...)...)
}
