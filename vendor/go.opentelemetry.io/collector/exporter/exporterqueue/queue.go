// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterqueue // import "go.opentelemetry.io/collector/exporter/exporterqueue"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/internal/queue"
)

// ErrQueueIsFull is the error that Queue returns when full.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
var ErrQueueIsFull = queue.ErrQueueIsFull

// Queue defines a producer-consumer exchange which can be backed by e.g. the memory-based ring buffer queue
// (boundedMemoryQueue) or via a disk-based queue (persistentQueue)
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Queue[T any] queue.Queue[T]

// Settings defines settings for creating a queue.
type Settings struct {
	DataType         component.DataType
	ExporterSettings exporter.Settings
}

// Marshaler is a function that can marshal a request into bytes.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Marshaler[T any] func(T) ([]byte, error)

// Unmarshaler is a function that can unmarshal bytes into a request.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Unmarshaler[T any] func([]byte) (T, error)

// Factory is a function that creates a new queue.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Factory[T any] func(context.Context, Settings, Config) Queue[T]

// NewMemoryQueueFactory returns a factory to create a new memory queue.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
func NewMemoryQueueFactory[T itemsCounter]() Factory[T] {
	return func(_ context.Context, _ Settings, cfg Config) Queue[T] {
		return queue.NewBoundedMemoryQueue[T](queue.MemoryQueueSettings[T]{
			Sizer:    sizerFromConfig[T](cfg),
			Capacity: capacityFromConfig(cfg),
		})
	}
}

// PersistentQueueSettings defines developer settings for the persistent queue factory.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type PersistentQueueSettings[T any] struct {
	// Marshaler is used to serialize queue elements before storing them in the persistent storage.
	Marshaler Marshaler[T]
	// Unmarshaler is used to deserialize requests after reading them from the persistent storage.
	Unmarshaler Unmarshaler[T]
}

// NewPersistentQueueFactory returns a factory to create a new persistent queue.
// If cfg.StorageID is nil then it falls back to memory queue.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
func NewPersistentQueueFactory[T itemsCounter](storageID *component.ID, factorySettings PersistentQueueSettings[T]) Factory[T] {
	if storageID == nil {
		return NewMemoryQueueFactory[T]()
	}
	return func(_ context.Context, set Settings, cfg Config) Queue[T] {
		return queue.NewPersistentQueue[T](queue.PersistentQueueSettings[T]{
			Sizer:            sizerFromConfig[T](cfg),
			Capacity:         capacityFromConfig(cfg),
			DataType:         set.DataType,
			StorageID:        *storageID,
			Marshaler:        factorySettings.Marshaler,
			Unmarshaler:      factorySettings.Unmarshaler,
			ExporterSettings: set.ExporterSettings,
		})
	}
}

type itemsCounter interface {
	ItemsCount() int
}

func sizerFromConfig[T itemsCounter](Config) queue.Sizer[T] {
	// TODO: Handle other ways to measure the queue size once they are added.
	return &queue.RequestSizer[T]{}
}

func capacityFromConfig(cfg Config) int64 {
	// TODO: Handle other ways to measure the queue size once they are added.
	return int64(cfg.QueueSize)
}
