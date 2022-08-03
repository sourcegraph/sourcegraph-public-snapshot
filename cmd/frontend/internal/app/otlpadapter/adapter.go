package otlpadapter

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"go.opentelemetry.io/collector/component"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/std"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type signalAdapter struct {
	// Exporter should send signals using the configured protocol to the configured
	// backend.
	component.Exporter
	// Receiver should receive http/json signals and pass it to the Exporter
	component.Receiver
}

// Start initializes the exporter and receiver of this adapter.
func (a *signalAdapter) Start(ctx context.Context, host component.Host) error {
	if err := a.Exporter.Start(ctx, host); err != nil {
		return errors.Wrap(err, "Exporter.Start")
	}
	if err := a.Receiver.Start(ctx, host); err != nil {
		return errors.Wrap(err, "Receiver.Start")
	}
	return nil
}

type adaptedSignal struct {
	// PathPrefix is the path for this signal (e.g. '/v1/traces')
	//
	// Specification: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#endpoint-urls-for-otlphttp
	PathPrefix string
	// CreateAdapter creates the receiver for this signal that redirects to the
	// appropriate exporter.
	CreateAdapter func() (*signalAdapter, error)
}

// Register attaches a route to the router that adapts requests on the `/otlp` path.
func (sig *adaptedSignal) Register(ctx context.Context, logger log.Logger, r *mux.Router, receiverURL *url.URL) {
	adapterLogger := logger.Scoped(path.Base(sig.PathPrefix), "OpenTelemetry signal-specific tunnel")

	// Set up an http/json -> ${configured_protocol} adapter
	adapter, err := sig.CreateAdapter()
	if err != nil {
		adapterLogger.Fatal("CreateAdapter", log.Error(err))
	}
	if err := adapter.Start(ctx, &otelHost{logger: logger}); err != nil {
		adapterLogger.Fatal("adapter.Start", log.Error(err))
	}

	// The redirector starts up a receiver service running at receiverEndpoint,
	// so now we have to reverse-proxy incoming requests to it so that things get
	// exported correctly.
	r.PathPrefix("/otlp" + sig.PathPrefix).Handler(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = receiverURL.Scheme
			req.URL.Host = receiverURL.Host
			req.URL.Path = sig.PathPrefix
		},
		ErrorLog: std.NewLogger(adapterLogger, log.LevelWarn),
	})

	adapterLogger.Info("signal adapter registered")
}
