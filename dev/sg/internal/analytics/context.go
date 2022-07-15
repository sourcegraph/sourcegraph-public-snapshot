package analytics

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/otfields"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

type eventStoreKey struct{}

// WithContext enables analytics in this context.
func WithContext(ctx context.Context, sgVersion string) (context.Context, error) {
	processor, err := newSpanToDiskProcessor()
	if err != nil {
		return ctx, errors.Wrap(err, "newSpanToDiskProcessor")
	}

	provider := oteltracesdk.NewTracerProvider(
		oteltracesdk.WithResource(newResource(otfields.Resource{
			Name:       "sg",
			Namespace:  "dev",
			Version:    sgVersion,
			InstanceID: "", // TODO 'git config user.name' or 'user.email' maybe?
		})),
		oteltracesdk.WithSampler(oteltracesdk.AlwaysSample()),
		oteltracesdk.WithSpanProcessor(processor),
	)

	otel.SetTracerProvider(provider)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		std.Out.WriteWarningf("opentelemetry: %s", err.Error())
	}))

	return context.WithValue(ctx, eventStoreKey{}, &eventStore{
		processor: processor,
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
		semconv.ServiceVersionKey.String(r.Version))
}

// getStore retrieves the events store from context if it exists. Callers should check
// that the store is non-nil before attempting to use it.
func getStore(ctx context.Context) *eventStore {
	store, ok := ctx.Value(eventStoreKey{}).(*eventStore)
	if !ok {
		return nil
	}
	return store
}

func tracer() trace.Tracer {
	return otel.Tracer("sg")
}
