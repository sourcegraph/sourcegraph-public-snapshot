// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package sharedcomponent exposes functionality for components
// to register against a shared key, such as a configuration object, in order to be reused across signal types.
// This is particularly useful when the component relies on a shared resource such as os.File or http.Server.
package sharedcomponent // import "go.opentelemetry.io/collector/internal/sharedcomponent"

import (
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
)

func NewMap[K comparable, V component.Component]() *Map[K, V] {
	return &Map[K, V]{
		components: map[K]*Component[V]{},
	}
}

// Map keeps reference of all created instances for a given shared key such as a component configuration.
type Map[K comparable, V component.Component] struct {
	lock       sync.Mutex
	components map[K]*Component[V]
}

// LoadOrStore returns the already created instance if exists, otherwise creates a new instance
// and adds it to the map of references.
func (m *Map[K, V]) LoadOrStore(key K, create func() (V, error), telemetrySettings *component.TelemetrySettings) (*Component[V], error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if c, ok := m.components[key]; ok {
		// If we haven't already seen this telemetry settings, this shared component represents
		// another instance. Wrap ReportStatus to report for all instances this shared
		// component represents.
		if _, ok := c.seenSettings[telemetrySettings]; !ok {
			c.seenSettings[telemetrySettings] = struct{}{}
			prev := c.telemetry.ReportStatus
			c.telemetry.ReportStatus = func(ev *component.StatusEvent) {
				telemetrySettings.ReportStatus(ev)
				prev(ev)
			}
		}
		return c, nil
	}
	comp, err := create()
	if err != nil {
		return nil, err
	}

	newComp := &Component[V]{
		component: comp,
		removeFunc: func() {
			m.lock.Lock()
			defer m.lock.Unlock()
			delete(m.components, key)
		},
		telemetry: telemetrySettings,
		seenSettings: map[*component.TelemetrySettings]struct{}{
			telemetrySettings: {},
		},
	}
	m.components[key] = newComp
	return newComp, nil
}

// Component ensures that the wrapped component is started and stopped only once.
// When stopped it is removed from the Map.
type Component[V component.Component] struct {
	component V

	startOnce  sync.Once
	stopOnce   sync.Once
	removeFunc func()

	telemetry    *component.TelemetrySettings
	seenSettings map[*component.TelemetrySettings]struct{}
}

// Unwrap returns the original component.
func (c *Component[V]) Unwrap() V {
	return c.component
}

// Start starts the underlying component if it never started before.
func (c *Component[V]) Start(ctx context.Context, host component.Host) error {
	var err error
	c.startOnce.Do(func() {
		// It's important that status for a shared component is reported through its
		// telemetry settings to keep status in sync and avoid race conditions. This logic duplicates
		// and takes priority over the automated status reporting that happens in graph, making the
		// status reporting in graph a no-op.
		c.telemetry.ReportStatus(component.NewStatusEvent(component.StatusStarting))
		if err = c.component.Start(ctx, host); err != nil {
			c.telemetry.ReportStatus(component.NewPermanentErrorEvent(err))
		}
	})
	return err
}

// Shutdown shuts down the underlying component.
func (c *Component[V]) Shutdown(ctx context.Context) error {
	var err error
	c.stopOnce.Do(func() {
		// It's important that status for a shared component is reported through its
		// telemetry settings to keep status in sync and avoid race conditions. This logic duplicates
		// and takes priority over the automated status reporting that happens in graph, making the
		// status reporting in graph a no-op.
		c.telemetry.ReportStatus(component.NewStatusEvent(component.StatusStopping))
		err = c.component.Shutdown(ctx)
		if err != nil {
			c.telemetry.ReportStatus(component.NewPermanentErrorEvent(err))
		} else {
			c.telemetry.ReportStatus(component.NewStatusEvent(component.StatusStopped))
		}
		c.removeFunc()
	})
	return err
}
