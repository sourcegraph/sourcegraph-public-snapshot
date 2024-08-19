// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package component // import "go.opentelemetry.io/collector/component"

// Host represents the entity that is hosting a Component. It is used to allow communication
// between the Component and its host (normally the service.Collector is the host).
type Host interface {
	// GetFactory of the specified kind. Returns the factory for a component type.
	// This allows components to create other components. For example:
	//   func (r MyReceiver) Start(host component.Host) error {
	//     apacheFactory := host.GetFactory(KindReceiver,"apache").(receiver.Factory)
	//     receiver, err := apacheFactory.CreateMetrics(...)
	//     ...
	//   }
	//
	// GetFactory can be called by the component anytime after Component.Start() begins and
	// until Component.Shutdown() ends. Note that the component is responsible for destroying
	// other components that it creates.
	GetFactory(kind Kind, componentType Type) Factory

	// GetExtensions returns the map of extensions. Only enabled and created extensions will be returned.
	// Typically, it is used to find an extension by type or by full config name. Both cases
	// can be done by iterating the returned map. There are typically very few extensions,
	// so there are no performance implications due to iteration.
	//
	// GetExtensions can be called by the component anytime after Component.Start() begins and
	// until Component.Shutdown() ends.
	GetExtensions() map[ID]Component
}
