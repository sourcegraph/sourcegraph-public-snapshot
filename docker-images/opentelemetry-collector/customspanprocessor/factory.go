package spanattrprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	typeStr = "spanattrcount"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithTraces(
			createTracer,
			component.StabilityLevelAlpha,
		),
	)
}
func createTracer(ctx context.Context, settings processor.CreateSettings, cfg component.Config, next consumer.Traces) (processor.Traces, error) {
	return &TraceProcessor{
		logger: settings.Logger,
		next:   next,
	}, nil
}

func createDefaultConfig() component.Config {
	return &Config{
		processor.CreateSettings{
			ID:                component.NewIDWithName(component.Type("processor"), "spanattrcount"),
			TelemetrySettings: component.TelemetrySettings{},
			BuildInfo:         component.NewDefaultBuildInfo(),
		},
	}
}
