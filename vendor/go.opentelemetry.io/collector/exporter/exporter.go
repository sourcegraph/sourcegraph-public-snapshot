// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporter // import "go.opentelemetry.io/collector/exporter"

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
)

// Traces is an exporter that can consume traces.
type Traces interface {
	component.Component
	consumer.Traces
}

// Metrics is an exporter that can consume metrics.
type Metrics interface {
	component.Component
	consumer.Metrics
}

// Logs is an exporter that can consume logs.
type Logs interface {
	component.Component
	consumer.Logs
}

// CreateSettings configures Exporter creators.
//
// Deprecated: [v0.103.0] Use exporter.Settings instead.
type CreateSettings = Settings

// Settings configures exporter creators.
type Settings struct {
	// ID returns the ID of the component that will be created.
	ID component.ID

	component.TelemetrySettings

	// BuildInfo can be used by components for informational purposes
	BuildInfo component.BuildInfo
}

// Factory is factory interface for exporters.
//
// This interface cannot be directly implemented. Implementations must
// use the NewFactory to implement it.
type Factory interface {
	component.Factory

	// CreateTracesExporter creates a TracesExporter based on this config.
	// If the exporter type does not support tracing or if the config is not valid,
	// an error will be returned instead.
	CreateTracesExporter(ctx context.Context, set Settings, cfg component.Config) (Traces, error)

	// TracesExporterStability gets the stability level of the TracesExporter.
	TracesExporterStability() component.StabilityLevel

	// CreateMetricsExporter creates a MetricsExporter based on this config.
	// If the exporter type does not support metrics or if the config is not valid,
	// an error will be returned instead.
	CreateMetricsExporter(ctx context.Context, set Settings, cfg component.Config) (Metrics, error)

	// MetricsExporterStability gets the stability level of the MetricsExporter.
	MetricsExporterStability() component.StabilityLevel

	// CreateLogsExporter creates a LogsExporter based on the config.
	// If the exporter type does not support logs or if the config is not valid,
	// an error will be returned instead.
	CreateLogsExporter(ctx context.Context, set Settings, cfg component.Config) (Logs, error)

	// LogsExporterStability gets the stability level of the LogsExporter.
	LogsExporterStability() component.StabilityLevel

	unexportedFactoryFunc()
}

// FactoryOption apply changes to Factory.
type FactoryOption interface {
	// applyExporterFactoryOption applies the option.
	applyExporterFactoryOption(o *factory)
}

var _ FactoryOption = (*factoryOptionFunc)(nil)

// factoryOptionFunc is an ExporterFactoryOption created through a function.
type factoryOptionFunc func(*factory)

func (f factoryOptionFunc) applyExporterFactoryOption(o *factory) {
	f(o)
}

// CreateTracesFunc is the equivalent of Factory.CreateTraces.
type CreateTracesFunc func(context.Context, Settings, component.Config) (Traces, error)

// CreateTracesExporter implements ExporterFactory.CreateTracesExporter().
func (f CreateTracesFunc) CreateTracesExporter(ctx context.Context, set Settings, cfg component.Config) (Traces, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg)
}

// CreateMetricsFunc is the equivalent of Factory.CreateMetrics.
type CreateMetricsFunc func(context.Context, Settings, component.Config) (Metrics, error)

// CreateMetricsExporter implements ExporterFactory.CreateMetricsExporter().
func (f CreateMetricsFunc) CreateMetricsExporter(ctx context.Context, set Settings, cfg component.Config) (Metrics, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg)
}

// CreateLogsFunc is the equivalent of Factory.CreateLogs.
type CreateLogsFunc func(context.Context, Settings, component.Config) (Logs, error)

// CreateLogsExporter implements Factory.CreateLogsExporter().
func (f CreateLogsFunc) CreateLogsExporter(ctx context.Context, set Settings, cfg component.Config) (Logs, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg)
}

type factory struct {
	cfgType component.Type
	component.CreateDefaultConfigFunc
	CreateTracesFunc
	tracesStabilityLevel component.StabilityLevel
	CreateMetricsFunc
	metricsStabilityLevel component.StabilityLevel
	CreateLogsFunc
	logsStabilityLevel component.StabilityLevel
}

func (f *factory) Type() component.Type {
	return f.cfgType
}

func (f *factory) unexportedFactoryFunc() {}

func (f *factory) TracesExporterStability() component.StabilityLevel {
	return f.tracesStabilityLevel
}

func (f *factory) MetricsExporterStability() component.StabilityLevel {
	return f.metricsStabilityLevel
}

func (f *factory) LogsExporterStability() component.StabilityLevel {
	return f.logsStabilityLevel
}

// WithTraces overrides the default "error not supported" implementation for CreateTracesExporter and the default "undefined" stability level.
func WithTraces(createTraces CreateTracesFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.tracesStabilityLevel = sl
		o.CreateTracesFunc = createTraces
	})
}

// WithMetrics overrides the default "error not supported" implementation for CreateMetricsExporter and the default "undefined" stability level.
func WithMetrics(createMetrics CreateMetricsFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.metricsStabilityLevel = sl
		o.CreateMetricsFunc = createMetrics
	})
}

// WithLogs overrides the default "error not supported" implementation for CreateLogsExporter and the default "undefined" stability level.
func WithLogs(createLogs CreateLogsFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.logsStabilityLevel = sl
		o.CreateLogsFunc = createLogs
	})
}

// NewFactory returns a Factory.
func NewFactory(cfgType component.Type, createDefaultConfig component.CreateDefaultConfigFunc, options ...FactoryOption) Factory {
	f := &factory{
		cfgType:                 cfgType,
		CreateDefaultConfigFunc: createDefaultConfig,
	}
	for _, opt := range options {
		opt.applyExporterFactoryOption(f)
	}
	return f
}

// MakeFactoryMap takes a list of factories and returns a map with Factory type as keys.
// It returns a non-nil error when there are factories with duplicate type.
func MakeFactoryMap(factories ...Factory) (map[component.Type]Factory, error) {
	fMap := map[component.Type]Factory{}
	for _, f := range factories {
		if _, ok := fMap[f.Type()]; ok {
			return fMap, fmt.Errorf("duplicate exporter factory %q", f.Type())
		}
		fMap[f.Type()] = f
	}
	return fMap, nil
}

// Builder exporter is a helper struct that given a set of Configs and Factories helps with creating exporters.
type Builder struct {
	cfgs      map[component.ID]component.Config
	factories map[component.Type]Factory
}

// NewBuilder creates a new exporter.Builder to help with creating components form a set of configs and factories.
func NewBuilder(cfgs map[component.ID]component.Config, factories map[component.Type]Factory) *Builder {
	return &Builder{cfgs: cfgs, factories: factories}
}

// CreateTraces creates a Traces exporter based on the settings and config.
func (b *Builder) CreateTraces(ctx context.Context, set Settings) (Traces, error) {
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("exporter %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("exporter factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.TracesExporterStability())
	return f.CreateTracesExporter(ctx, set, cfg)
}

// CreateMetrics creates a Metrics exporter based on the settings and config.
func (b *Builder) CreateMetrics(ctx context.Context, set Settings) (Metrics, error) {
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("exporter %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("exporter factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.MetricsExporterStability())
	return f.CreateMetricsExporter(ctx, set, cfg)
}

// CreateLogs creates a Logs exporter based on the settings and config.
func (b *Builder) CreateLogs(ctx context.Context, set Settings) (Logs, error) {
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("exporter %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("exporter factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.LogsExporterStability())
	return f.CreateLogsExporter(ctx, set, cfg)
}

func (b *Builder) Factory(componentType component.Type) component.Factory {
	return b.factories[componentType]
}

// logStabilityLevel logs the stability level of a component. The log level is set to info for
// undefined, unmaintained, deprecated and development. The log level is set to debug
// for alpha, beta and stable.
func logStabilityLevel(logger *zap.Logger, sl component.StabilityLevel) {
	if sl >= component.StabilityLevelAlpha {
		logger.Debug(sl.LogMessage())
	} else {
		logger.Info(sl.LogMessage())
	}
}
