package trace

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
)

// Middleware wraps an http.Handler to extract opentracing span information from the request headers.
// The opentracing.SpanContext is added to the request context, and can be retrieved by SpanContextFromContext.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err == nil {
			r = r.WithContext(ContextWithSpanContext(r.Context(), spanContext))
		}
		next.ServeHTTP(w, r)
	})
}

type spanContextKey struct{}

// SpanContextFromContext retrieves the opentracing.SpanContext set on the context by Middleware
func SpanContextFromContext(ctx context.Context) opentracing.SpanContext {
	if v := ctx.Value(spanContextKey{}); v != nil {
		return v.(opentracing.SpanContext)
	}
	return nil
}

// ContextWithSpanContext creates a new context with the opentracing.SpanContext set
func ContextWithSpanContext(ctx context.Context, sc opentracing.SpanContext) context.Context {
	return context.WithValue(ctx, spanContextKey{}, sc)
}
