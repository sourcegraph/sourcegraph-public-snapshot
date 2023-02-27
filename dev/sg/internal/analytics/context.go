package analytics

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/run"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

// WithContext enables analytics in this context.
func WithContext(ctx context.Context, sgVersion string) (context.Context, error) {
	processor, err := newSpanToDiskProcessor(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "disk exporter")
	}

	// Loose attempt at getting identity - if we fail, just discard
	identity, _ := run.Cmd(ctx, "git config user.email").StdOut().Run().String()

	// Create a provider with configuration and resource specification
	provider := oteltracesdk.NewTracerProvider(
		oteltracesdk.WithResource(newResource(log.Resource{
			Name:       "sg",
			Namespace:  sgVersion,
			Version:    sgVersion,
			InstanceID: identity,
		})),
		oteltracesdk.WithSampler(oteltracesdk.AlwaysSample()),
		oteltracesdk.WithSpanProcessor(processor),
	)

	// Configure OpenTelemetry defaults
	otel.SetTracerProvider(provider)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		std.Out.WriteWarningf("opentelemetry: %s", err.Error())
	}))

	// Create a root span for an execution of sg for all spans to be grouped under
	var rootSpan *Span
	ctx, rootSpan = StartSpan(ctx, "sg", "root")

	return context.WithValue(ctx, spansStoreKey{}, &spansStore{
		rootSpan: rootSpan.Span,
		provider: provider,
	}), nil
}

// newResource adapts sourcegraph/log.Resource into the OpenTelemetry package's Resource
// type.
func newResource(r log.Resource) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(r.Name),
		semconv.ServiceNamespaceKey.String(r.Namespace),
		semconv.ServiceInstanceIDKey.String(r.InstanceID),
		semconv.ServiceVersionKey.String(r.Version),
		attribute.String(sgAnalyticsVersionResourceKey, sgAnalyticsVersion))
}

func isValidVersion(spans *tracepb.ResourceSpans) bool {
	for _, attrib := range spans.GetResource().GetAttributes() {
		if attrib.GetKey() == sgAnalyticsVersionResourceKey {
			return attrib.Value.GetStringValue() == sgAnalyticsVersion
		}
	}
	return false
}
