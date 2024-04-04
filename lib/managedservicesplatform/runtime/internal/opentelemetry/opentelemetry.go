package opentelemetry

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/std"
	gcpdetector "go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"

	// Use semconv version matching the one used by go.opentelemetry.io/otel/sdk/resource
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	GCPProjectID string
	// OtelSDKDisabled disables the OpenTelemetry SDK integration. We manually
	// implement this spec as the Go SDK does not:
	//
	// - https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#general-sdk-configuration
	// - https://github.com/open-telemetry/opentelemetry-go/issues/3559
	OtelSDKDisabled bool
}

// Init initializes OpenTelemetry integrations. If config.GCPProjectID is set,
// all OpenTelemetry integrations will point to a GCP exporter - otherwise, a
// local dev default is chosen:
//
//   - traces: OTLP exporter
//   - metrics: Prometheus exporter
func Init(ctx context.Context, logger log.Logger, config Config, r log.Resource) (func(), error) {
	if config.OtelSDKDisabled {
		logger.Warn("OpenTelemetry SDK integration disabled by configuration")
		return func() {}, nil
	}

	// Set globals
	otel.SetTextMapPropagator(defaultPropagator())

	// Set logging hooks.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		if config.GCPProjectID != "" {
			logger.Error("OpenTelemetry error", log.Error(err))
		} else {
			logger.Warn("OpenTelemetry error", log.Error(err))
		}
	}))
	if config.GCPProjectID != "" {
		otel.SetLogger(
			// logr library levels are annoying to deal with, so we just use
			// a single level (info), as it's all diagnostics output to us anyway.
			logr.New(stdr.New(std.NewLogger(logger, log.LevelInfo)).GetSink()),
		)
	}

	res, err := getOpenTelemetryResource(ctx, r)
	if err != nil {
		return nil, errors.Wrap(err, "init resource")
	}

	shutdownTracing, err := maybeEnableTracing(ctx, logger.Scoped("tracing"), config, res)
	if err != nil {
		return nil, errors.Wrap(err, "enable tracing")
	}

	shutdownMetrics, err := maybeEnableMetrics(ctx, logger.Scoped("metrics"), config, res)
	if err != nil {
		return nil, errors.Wrap(err, "enable metrics")
	}

	return func() {
		logger.Debug("shutting down OpenTelemetry")
		var wg conc.WaitGroup
		wg.Go(shutdownTracing)
		wg.Go(shutdownMetrics)
		wg.Wait()
	}, nil
}

func getOpenTelemetryResource(ctx context.Context, r log.Resource) (*resource.Resource, error) {
	// Identify your application using resource detection
	res, err := resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcpdetector.NewDetector()),
		// Use the default detectors
		resource.WithTelemetrySDK(),
		// Add our own attributes
		resource.WithAttributes(
			semconv.ServiceNameKey.String(r.Name),
			semconv.ServiceVersionKey.String(r.Version),
			semconv.ServiceInstanceIDKey.String(r.InstanceID),
			semconv.ServiceNamespaceKey.String(r.Namespace),
		),
	)

	if errors.Is(err, resource.ErrSchemaURLConflict) {
		// Ignore the conflict error, the resource is still safe to use
		// https://github.com/open-telemetry/opentelemetry-go/pull/4876
		return res, nil
	}
	return res, err
}
