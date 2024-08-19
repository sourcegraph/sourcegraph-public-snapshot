// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterbatcher // import "go.opentelemetry.io/collector/exporter/exporterbatcher"

import (
	"errors"
	"time"
)

// Config defines a configuration for batching requests based on a timeout and a minimum number of items.
// MaxSizeItems defines batch splitting functionality if it's more than zero.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Config struct {
	// Enabled indicates whether to not enqueue batches before sending to the consumerSender.
	Enabled bool `mapstructure:"enabled"`

	// FlushTimeout sets the time after which a batch will be sent regardless of its size.
	FlushTimeout time.Duration `mapstructure:"flush_timeout"`

	MinSizeConfig `mapstructure:",squash"`
	MaxSizeConfig `mapstructure:",squash"`
}

// MinSizeConfig defines the configuration for the minimum number of items in a batch.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type MinSizeConfig struct {
	// MinSizeItems is the number of items (spans, data points or log records for OTLP) at which the batch should be
	// sent regardless of the timeout. There is no guarantee that the batch size always greater than this value.
	// This option requires the Request to implement RequestItemsCounter interface. Otherwise, it will be ignored.
	MinSizeItems int `mapstructure:"min_size_items"`
}

// MaxSizeConfig defines the configuration for the maximum number of items in a batch.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type MaxSizeConfig struct {
	// MaxSizeItems is the maximum number of the batch items, i.e. spans, data points or log records for OTLP.
	// If the batch size exceeds this value, it will be broken up into smaller batches if possible.
	// Setting this value to zero disables the maximum size limit.
	MaxSizeItems int `mapstructure:"max_size_items"`
}

func (c Config) Validate() error {
	if c.MinSizeItems < 0 {
		return errors.New("min_size_items must be greater than or equal to zero")
	}
	if c.MaxSizeItems < 0 {
		return errors.New("max_size_items must be greater than or equal to zero")
	}
	if c.MaxSizeItems != 0 && c.MaxSizeItems < c.MinSizeItems {
		return errors.New("max_size_items must be greater than or equal to min_size_items")
	}
	if c.FlushTimeout <= 0 {
		return errors.New("timeout must be greater than zero")
	}
	return nil
}

func NewDefaultConfig() Config {
	return Config{
		Enabled:      true,
		FlushTimeout: 200 * time.Millisecond,
		MinSizeConfig: MinSizeConfig{
			MinSizeItems: 8192,
		},
	}
}
