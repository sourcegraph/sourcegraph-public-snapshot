package otlpadapter

import (
	"context"

	"go.opentelemetry.io/collector/component"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type adaptedSignal struct {
	// PathPrefix is the path for this signal (e.g. '/v1/traces')
	//
	// Specification: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#endpoint-urls-for-otlphttp
	PathPrefix string
	// CreateAdapter creates the receiver for this signal that redirects to the
	// appropriate exporter.
	CreateAdapter func() (*signalAdapter, error)
}

type signalAdapter struct {
	// Exporter should send signals using the configured protocol to the configured
	// backend.
	component.Exporter
	// Receiver should receive http/json signals and pass it to the Exporter
	component.Receiver
}

func (a *signalAdapter) Start(ctx context.Context, host component.Host) error {
	if err := a.Exporter.Start(ctx, host); err != nil {
		return errors.Wrap(err, "Exporter.Start")
	}
	if err := a.Receiver.Start(ctx, host); err != nil {
		return errors.Wrap(err, "Receiver.Start")
	}
	return nil
}
