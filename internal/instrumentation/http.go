package instrumentation

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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
		var trace bool
		switch policy.GetTracePolicy() {
		case policy.TraceSelective:
			trace = policy.RequestWantsTracing(r)
		case policy.TraceAll:
			trace = true
		default:
			trace = false
		}
		// Pass through to instrumented handler with trace policy in context
		instrumentedHandler.ServeHTTP(w, r.WithContext(policy.WithShouldTrace(r.Context(), trace)))
	})
}
