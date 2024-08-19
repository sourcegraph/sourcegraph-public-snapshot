// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package extension // import "go.opentelemetry.io/collector/extension"

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// Extension is the interface for objects hosted by the OpenTelemetry Collector that
// don't participate directly on data pipelines but provide some functionality
// to the service, examples: health check endpoint, z-pages, etc.
type Extension = component.Component

// Dependent is an optional interface that can be implemented by extensions
// that depend on other extensions and must be started only after their dependencies.
// See https://github.com/open-telemetry/opentelemetry-collector/pull/8768 for examples.
type Dependent interface {
	Extension
	Dependencies() []component.ID
}

// PipelineWatcher is an extra interface for Extension hosted by the OpenTelemetry
// Collector that is to be implemented by extensions interested in changes to pipeline
// states. Typically this will be used by extensions that change their behavior if data is
// being ingested or not, e.g.: a k8s readiness probe.
type PipelineWatcher interface {
	// Ready notifies the Extension that all pipelines were built and the
	// receivers were started, i.e.: the service is ready to receive data
	// (note that it may already have received data when this method is called).
	Ready() error

	// NotReady notifies the Extension that all receivers are about to be stopped,
	// i.e.: pipeline receivers will not accept new data.
	// This is sent before receivers are stopped, so the Extension can take any
	// appropriate actions before that happens.
	NotReady() error
}

// ConfigWatcher is an interface that should be implemented by an extension that
// wishes to be notified of the Collector's effective configuration.
type ConfigWatcher interface {
	// NotifyConfig notifies the extension of the Collector's current effective configuration.
	// The extension owns the `confmap.Conf`. Callers must ensure that it's safe for
	// extensions to store the `conf` pointer and use it concurrently with any other
	// instances of `conf`.
	NotifyConfig(ctx context.Context, conf *confmap.Conf) error
}

// StatusWatcher is an extra interface for Extension hosted by the OpenTelemetry
// Collector that is to be implemented by extensions interested in changes to component
// status.
type StatusWatcher interface {
	// ComponentStatusChanged notifies about a change in the source component status.
	// Extensions that implement this interface must be ready that the ComponentStatusChanged
	// may be called before, after or concurrently with calls to Component.Start() and Component.Shutdown().
	// The function may be called concurrently with itself.
	ComponentStatusChanged(source *component.InstanceID, event *component.StatusEvent)
}

// CreateSettings is passed to Factory.Create(...) function.
//
// Deprecated: [v0.103.0] Use extension.Settings instead.
type CreateSettings = Settings

// Settings is passed to Factory.Create(...) function.
type Settings struct {
	// ID returns the ID of the component that will be created.
	ID component.ID

	component.TelemetrySettings

	// BuildInfo can be used by components for informational purposes
	BuildInfo component.BuildInfo
}

// CreateFunc is the equivalent of Factory.Create(...) function.
type CreateFunc func(context.Context, Settings, component.Config) (Extension, error)

// CreateExtension implements Factory.Create.
func (f CreateFunc) CreateExtension(ctx context.Context, set Settings, cfg component.Config) (Extension, error) {
	return f(ctx, set, cfg)
}

type Factory interface {
	component.Factory

	// CreateExtension creates an extension based on the given config.
	CreateExtension(ctx context.Context, set Settings, cfg component.Config) (Extension, error)

	// ExtensionStability gets the stability level of the Extension.
	ExtensionStability() component.StabilityLevel

	unexportedFactoryFunc()
}

type factory struct {
	cfgType component.Type
	component.CreateDefaultConfigFunc
	CreateFunc
	extensionStability component.StabilityLevel
}

func (f *factory) Type() component.Type {
	return f.cfgType
}

func (f *factory) unexportedFactoryFunc() {}

func (f *factory) ExtensionStability() component.StabilityLevel {
	return f.extensionStability
}

// NewFactory returns a new Factory  based on this configuration.
func NewFactory(
	cfgType component.Type,
	createDefaultConfig component.CreateDefaultConfigFunc,
	createServiceExtension CreateFunc,
	sl component.StabilityLevel) Factory {
	return &factory{
		cfgType:                 cfgType,
		CreateDefaultConfigFunc: createDefaultConfig,
		CreateFunc:              createServiceExtension,
		extensionStability:      sl,
	}
}

// MakeFactoryMap takes a list of factories and returns a map with Factory type as keys.
// It returns a non-nil error when there are factories with duplicate type.
func MakeFactoryMap(factories ...Factory) (map[component.Type]Factory, error) {
	fMap := map[component.Type]Factory{}
	for _, f := range factories {
		if _, ok := fMap[f.Type()]; ok {
			return fMap, fmt.Errorf("duplicate extension factory %q", f.Type())
		}
		fMap[f.Type()] = f
	}
	return fMap, nil
}

// Builder extension is a helper struct that given a set of Configs and Factories helps with creating extensions.
type Builder struct {
	cfgs      map[component.ID]component.Config
	factories map[component.Type]Factory
}

// NewBuilder creates a new extension.Builder to help with creating components form a set of configs and factories.
func NewBuilder(cfgs map[component.ID]component.Config, factories map[component.Type]Factory) *Builder {
	return &Builder{cfgs: cfgs, factories: factories}
}

// Create creates an extension based on the settings and configs available.
func (b *Builder) Create(ctx context.Context, set Settings) (Extension, error) {
	cfg, existsCfg := b.cfgs[set.ID]
	if !existsCfg {
		return nil, fmt.Errorf("extension %q is not configured", set.ID)
	}

	f, existsFactory := b.factories[set.ID.Type()]
	if !existsFactory {
		return nil, fmt.Errorf("extension factory not available for: %q", set.ID)
	}

	sl := f.ExtensionStability()
	if sl >= component.StabilityLevelAlpha {
		set.Logger.Debug(sl.LogMessage())
	} else {
		set.Logger.Info(sl.LogMessage())
	}
	return f.CreateExtension(ctx, set, cfg)
}

func (b *Builder) Factory(componentType component.Type) component.Factory {
	return b.factories[componentType]
}
