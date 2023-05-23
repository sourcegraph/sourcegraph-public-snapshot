package oteldefaults

import (
	jaegerpropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	otpropagator "go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel/propagation"
)

// Propagator returns a propagator that supports a bunch of common formats like
// W3C Trace Context, W3C Baggage, OpenTracing, and Jaeger (the latter two being
// the more commonly used legacy formats at Sourcegraph). This helps ensure
// propagation between services continues to work.
func Propagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		jaegerpropagator.Jaeger{},
		otpropagator.OT{},
		// W3C Trace Context format (https://www.w3.org/TR/trace-context/)
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
