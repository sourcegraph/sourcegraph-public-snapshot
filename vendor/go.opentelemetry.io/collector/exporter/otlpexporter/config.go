// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otlpexporter // import "go.opentelemetry.io/collector/exporter/otlpexporter"

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for OTLP exporter.
type Config struct {
	exporterhelper.TimeoutSettings `mapstructure:",squash"`     // squash ensures fields are correctly decoded in embedded struct.
	QueueConfig                    exporterhelper.QueueSettings `mapstructure:"sending_queue"`
	RetryConfig                    configretry.BackOffConfig    `mapstructure:"retry_on_failure"`

	configgrpc.ClientConfig `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
}

func (c *Config) Validate() error {
	endpoint := c.sanitizedEndpoint()
	if endpoint == "" {
		return errors.New(`requires a non-empty "endpoint"`)
	}

	// Validate that the port is in the address
	_, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return err
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf(`invalid port "%s"`, port)
	}

	return nil
}

func (c *Config) sanitizedEndpoint() string {
	switch {
	case strings.HasPrefix(c.Endpoint, "http://"):
		return strings.TrimPrefix(c.Endpoint, "http://")
	case strings.HasPrefix(c.Endpoint, "https://"):
		return strings.TrimPrefix(c.Endpoint, "https://")
	case strings.HasPrefix(c.Endpoint, "dns:///"):
		return strings.TrimPrefix(c.Endpoint, "dns:///")
	default:
		return c.Endpoint
	}
}

var _ component.Config = (*Config)(nil)
