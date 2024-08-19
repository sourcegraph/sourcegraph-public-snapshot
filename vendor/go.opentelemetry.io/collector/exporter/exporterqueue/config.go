// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterqueue // import "go.opentelemetry.io/collector/exporter/exporterqueue"

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

// Config defines configuration for queueing requests before exporting.
// It's supposed to be used with the new exporter helpers New[Traces|Metrics|Logs]RequestExporter.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Config struct {
	// Enabled indicates whether to not enqueue batches before exporting.
	Enabled bool `mapstructure:"enabled"`
	// NumConsumers is the number of consumers from the queue.
	NumConsumers int `mapstructure:"num_consumers"`
	// QueueSize is the maximum number of requests allowed in queue at any given time.
	QueueSize int `mapstructure:"queue_size"`
}

// NewDefaultConfig returns the default Config.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
func NewDefaultConfig() Config {
	return Config{
		Enabled:      true,
		NumConsumers: 10,
		QueueSize:    1_000,
	}
}

// Validate checks if the QueueSettings configuration is valid
func (qCfg *Config) Validate() error {
	if !qCfg.Enabled {
		return nil
	}
	if qCfg.NumConsumers <= 0 {
		return errors.New("number of consumers must be positive")
	}
	if qCfg.QueueSize <= 0 {
		return errors.New("queue size must be positive")
	}
	return nil
}

// PersistentQueueConfig defines configuration for queueing requests in a persistent storage.
// The struct is provided to be added in the exporter configuration as one struct under the "sending_queue" key.
// The exporter helper Go interface requires the fields to be provided separately to WithRequestQueue and
// NewPersistentQueueFactory.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type PersistentQueueConfig struct {
	Config `mapstructure:",squash"`
	// StorageID if not empty, enables the persistent storage and uses the component specified
	// as a storage extension for the persistent queue
	StorageID *component.ID `mapstructure:"storage"`
}
