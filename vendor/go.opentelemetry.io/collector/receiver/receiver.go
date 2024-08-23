// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package receiver // import "go.opentelemetry.io/collector/receiver"

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
)

var (
	errNilNextConsumer = errors.New("nil next Consumer")
)

// Traces receiver receives traces.
// Its purpose is to translate data from any format to the collector's internal trace format.
// TracesReceiver feeds a consumer.Traces with data.
//
// For example, it could be Zipkin data source which translates Zipkin spans into ptrace.Traces.
type Traces interface {
	component.Component
}

// Metrics receiver receives metrics.
// Its purpose is to translate data from any format to the collector's internal metrics format.
// MetricsReceiver feeds a consumer.Metrics with data.
//
// For example, it could be Prometheus data source which translates Prometheus metrics into pmetric.Metrics.
type Metrics interface {
	component.Component
}

// Logs receiver receives logs.
// Its purpose is to translate data from any format to the collector's internal logs data format.
// LogsReceiver feeds a consumer.Logs with data.
//
// For example, it could be a receiver that reads syslogs and convert them into plog.Logs.
type Logs interface {
	component.Component
}

// CreateSettings configures Receiver creators.
//
// Deprecated: [v0.103.0] Use receiver.Settings instead.
type CreateSettings = Settings

// Settings configures Receiver creators.
type Settings struct {
	// ID returns the ID of the component that will be created.
	ID component.ID

	component.TelemetrySettings

	// BuildInfo can be used by components for informational purposes.
	BuildInfo component.BuildInfo
}

// Factory is factory interface for receivers.
//
// This interface cannot be directly implemented. Implementations must
// use the NewReceiverFactory to implement it.
type Factory interface {
	component.Factory

	// CreateTracesReceiver creates a TracesReceiver based on this config.
	// If the receiver type does not support tracing or if the config is not valid
	// an error will be returned instead. `nextConsumer` is never nil.
	CreateTracesReceiver(ctx context.Context, set Settings, cfg component.Config, nextConsumer consumer.Traces) (Traces, error)

	// TracesReceiverStability gets the stability level of the TracesReceiver.
	TracesReceiverStability() component.StabilityLevel

	// CreateMetricsReceiver creates a MetricsReceiver based on this config.
	// If the receiver type does not support metrics or if the config is not valid
	// an error will be returned instead. `nextConsumer` is never nil.
	CreateMetricsReceiver(ctx context.Context, set Settings, cfg component.Config, nextConsumer consumer.Metrics) (Metrics, error)

	// MetricsReceiverStability gets the stability level of the MetricsReceiver.
	MetricsReceiverStability() component.StabilityLevel

	// CreateLogsReceiver creates a LogsReceiver based on this config.
	// If the receiver type does not support the data type or if the config is not valid
	// an error will be returned instead. `nextConsumer` is never nil.
	CreateLogsReceiver(ctx context.Context, set Settings, cfg component.Config, nextConsumer consumer.Logs) (Logs, error)

	// LogsReceiverStability gets the stability level of the LogsReceiver.
	LogsReceiverStability() component.StabilityLevel

	unexportedFactoryFunc()
}

// FactoryOption apply changes to ReceiverOptions.
type FactoryOption interface {
	// applyOption applies the option.
	applyOption(o *factory)
}

// factoryOptionFunc is an ReceiverFactoryOption created through a function.
type factoryOptionFunc func(*factory)

func (f factoryOptionFunc) applyOption(o *factory) {
	f(o)
}

// CreateTracesFunc is the equivalent of Factory.CreateTraces.
type CreateTracesFunc func(context.Context, Settings, component.Config, consumer.Traces) (Traces, error)

// CreateTracesReceiver implements Factory.CreateTracesReceiver().
func (f CreateTracesFunc) CreateTracesReceiver(
	ctx context.Context,
	set Settings,
	cfg component.Config,
	nextConsumer consumer.Traces) (Traces, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

// CreateMetricsFunc is the equivalent of Factory.CreateMetrics.
type CreateMetricsFunc func(context.Context, Settings, component.Config, consumer.Metrics) (Metrics, error)

// CreateMetricsReceiver implements Factory.CreateMetricsReceiver().
func (f CreateMetricsFunc) CreateMetricsReceiver(
	ctx context.Context,
	set Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (Metrics, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

// CreateLogsFunc is the equivalent of ReceiverFactory.CreateLogsReceiver().
type CreateLogsFunc func(context.Context, Settings, component.Config, consumer.Logs) (Logs, error)

// CreateLogsReceiver implements Factory.CreateLogsReceiver().
func (f CreateLogsFunc) CreateLogsReceiver(
	ctx context.Context,
	set Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (Logs, error) {
	if f == nil {
		return nil, component.ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
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

func (f *factory) TracesReceiverStability() component.StabilityLevel {
	return f.tracesStabilityLevel
}

func (f *factory) MetricsReceiverStability() component.StabilityLevel {
	return f.metricsStabilityLevel
}

func (f *factory) LogsReceiverStability() component.StabilityLevel {
	return f.logsStabilityLevel
}

// WithTraces overrides the default "error not supported" implementation for CreateTracesReceiver and the default "undefined" stability level.
func WithTraces(createTracesReceiver CreateTracesFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.tracesStabilityLevel = sl
		o.CreateTracesFunc = createTracesReceiver
	})
}

// WithMetrics overrides the default "error not supported" implementation for CreateMetricsReceiver and the default "undefined" stability level.
func WithMetrics(createMetricsReceiver CreateMetricsFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.metricsStabilityLevel = sl
		o.CreateMetricsFunc = createMetricsReceiver
	})
}

// WithLogs overrides the default "error not supported" implementation for CreateLogsReceiver and the default "undefined" stability level.
func WithLogs(createLogsReceiver CreateLogsFunc, sl component.StabilityLevel) FactoryOption {
	return factoryOptionFunc(func(o *factory) {
		o.logsStabilityLevel = sl
		o.CreateLogsFunc = createLogsReceiver
	})
}

// NewFactory returns a Factory.
func NewFactory(cfgType component.Type, createDefaultConfig component.CreateDefaultConfigFunc, options ...FactoryOption) Factory {
	f := &factory{
		cfgType:                 cfgType,
		CreateDefaultConfigFunc: createDefaultConfig,
	}
	for _, opt := range options {
		opt.applyOption(f)
	}
	return f
}

// MakeFactoryMap takes a list of receiver factories and returns a map with factory type as keys.
// It returns a non-nil error when there are factories with duplicate type.
func MakeFactoryMap(factories ...Factory) (map[component.Type]Factory, error) {
	fMap := map[component.Type]Factory{}
	for _, f := range factories {
		if _, ok := fMap[f.Type()]; ok {
			return fMap, fmt.Errorf("duplicate receiver factory %q", f.Type())
		}
		fMap[f.Type()] = f
	}
	return fMap, nil
}

// Builder receiver is a helper struct that given a set of Configs and Factories helps with creating receivers.
type Builder struct {
	cfgs      map[component.ID]component.Config
	factories map[component.Type]Factory
}

// NewBuilder creates a new receiver.Builder to help with creating components form a set of configs and factories.
func NewBuilder(cfgs map[component.ID]component.Config, factories map[component.Type]Factory) *Builder {
	return &Builder{cfgs: cfgs, factories: factories}
}

// CreateTraces creates a Traces receiver based on the settings and config.
func (b *Builder) CreateTraces(ctx context.Context, set Settings, next consumer.Traces) (Traces, error) {
	if next == nil {
		return nil, errNilNextConsumer
	}
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("receiver %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("receiver factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.TracesReceiverStability())
	return f.CreateTracesReceiver(ctx, set, cfg, next)
}

// CreateMetrics creates a Metrics receiver based on the settings and config.
func (b *Builder) CreateMetrics(ctx context.Context, set Settings, next consumer.Metrics) (Metrics, error) {
	if next == nil {
		return nil, errNilNextConsumer
	}
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("receiver %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("receiver factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.MetricsReceiverStability())
	return f.CreateMetricsReceiver(ctx, set, cfg, next)
}

// CreateLogs creates a Logs receiver based on the settings and config.
func (b *Builder) CreateLogs(ctx context.Context, set Settings, next consumer.Logs) (Logs, error) {
	if next == nil {
		return nil, errNilNextConsumer
	}
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("receiver %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("receiver factory not available for: %q", set.ID)
	}

	logStabilityLevel(set.Logger, f.LogsReceiverStability())
	return f.CreateLogsReceiver(ctx, set, cfg, next)
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
