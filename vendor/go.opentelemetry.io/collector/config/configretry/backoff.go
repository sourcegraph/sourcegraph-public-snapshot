// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configretry // import "go.opentelemetry.io/collector/config/configretry"

import (
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// NewDefaultBackOffConfig returns the default settings for RetryConfig.
func NewDefaultBackOffConfig() BackOffConfig {
	return BackOffConfig{
		Enabled:             true,
		InitialInterval:     5 * time.Second,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         30 * time.Second,
		MaxElapsedTime:      5 * time.Minute,
	}
}

// BackOffConfig defines configuration for retrying batches in case of export failure.
// The current supported strategy is exponential backoff.
type BackOffConfig struct {
	// Enabled indicates whether to not retry sending batches in case of export failure.
	Enabled bool `mapstructure:"enabled"`
	// InitialInterval the time to wait after the first failure before retrying.
	InitialInterval time.Duration `mapstructure:"initial_interval"`
	// RandomizationFactor is a random factor used to calculate next backoffs
	// Randomized interval = RetryInterval * (1 ± RandomizationFactor)
	RandomizationFactor float64 `mapstructure:"randomization_factor"`
	// Multiplier is the value multiplied by the backoff interval bounds
	Multiplier float64 `mapstructure:"multiplier"`
	// MaxInterval is the upper bound on backoff interval. Once this value is reached the delay between
	// consecutive retries will always be `MaxInterval`.
	MaxInterval time.Duration `mapstructure:"max_interval"`
	// MaxElapsedTime is the maximum amount of time (including retries) spent trying to send a request/batch.
	// Once this value is reached, the data is discarded. If set to 0, the retries are never stopped.
	MaxElapsedTime time.Duration `mapstructure:"max_elapsed_time"`
}

func (bs *BackOffConfig) Validate() error {
	if !bs.Enabled {
		return nil
	}
	if bs.InitialInterval < 0 {
		return errors.New("'initial_interval' must be non-negative")
	}
	if bs.RandomizationFactor < 0 || bs.RandomizationFactor > 1 {
		return errors.New("'randomization_factor' must be within [0, 1]")
	}
	if bs.Multiplier < 0 {
		return errors.New("'multiplier' must be non-negative")
	}
	if bs.MaxInterval < 0 {
		return errors.New("'max_interval' must be non-negative")
	}
	if bs.MaxElapsedTime < 0 {
		return errors.New("'max_elapsed_time' must be non-negative")
	}
	if bs.MaxElapsedTime > 0 {
		if bs.MaxElapsedTime < bs.InitialInterval {
			return errors.New("'max_elapsed_time' must not be less than 'initial_interval'")
		}
		if bs.MaxElapsedTime < bs.MaxInterval {
			return errors.New("'max_elapsed_time' must not be less than 'max_interval'")
		}

	}
	return nil
}
