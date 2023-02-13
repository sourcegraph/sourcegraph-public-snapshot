package instrumentation

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// HTTPMiddleware wraps the handler with the following:
//
//   - If the HTTP header, X-Sourcegraph-Should-Trace, is set to a truthy value, set the
//     shouldTraceKey context.Context value to true
//   - go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp, which applies the
//     desired instrumentation.
//
// The provided operation name is used to add details to spans.
func HTTPMiddleware(operation string, h http.Handler, opts ...otelhttp.Option) http.Handler {
	instrumentedHandler := otelhttp.NewHandler(h, operation,
		append(
			[]otelhttp.Option{
				otelhttp.WithTracerProvider(&samplingRetainTracerProvider{}),
				otelhttp.WithFilter(func(r *http.Request) bool {
					return policy.ShouldTrace(r.Context())
				}),
				otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
					if operation != "" {
						return fmt.Sprintf("%s.%s %s", operation, r.Method, r.URL.Path)
					}
					return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
				}),
			},
			opts...,
		)...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
type samplingRetainTracerProvider struct{}
type samplingRetainTracer struct {
	tracer trace.Tracer
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
