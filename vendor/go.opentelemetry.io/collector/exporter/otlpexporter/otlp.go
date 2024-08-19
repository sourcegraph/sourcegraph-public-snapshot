// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otlpexporter // import "go.opentelemetry.io/collector/exporter/otlpexporter"

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

type baseExporter struct {
	// Input configuration.
	config *Config

	// gRPC clients and connection.
	traceExporter  ptraceotlp.GRPCClient
	metricExporter pmetricotlp.GRPCClient
	logExporter    plogotlp.GRPCClient
	clientConn     *grpc.ClientConn
	metadata       metadata.MD
	callOptions    []grpc.CallOption

	settings component.TelemetrySettings

	// Default user-agent header.
	userAgent string
}

func newExporter(cfg component.Config, set exporter.Settings) *baseExporter {
	oCfg := cfg.(*Config)

	userAgent := fmt.Sprintf("%s/%s (%s/%s)",
		set.BuildInfo.Description, set.BuildInfo.Version, runtime.GOOS, runtime.GOARCH)

	return &baseExporter{config: oCfg, settings: set.TelemetrySettings, userAgent: userAgent}
}

// start actually creates the gRPC connection. The client construction is deferred till this point as this
// is the only place we get hold of Extensions which are required to construct auth round tripper.
func (e *baseExporter) start(ctx context.Context, host component.Host) (err error) {
	if e.clientConn, err = e.config.ClientConfig.ToClientConn(ctx, host, e.settings, grpc.WithUserAgent(e.userAgent)); err != nil {
		return err
	}
	e.traceExporter = ptraceotlp.NewGRPCClient(e.clientConn)
	e.metricExporter = pmetricotlp.NewGRPCClient(e.clientConn)
	e.logExporter = plogotlp.NewGRPCClient(e.clientConn)
	headers := map[string]string{}
	for k, v := range e.config.ClientConfig.Headers {
		headers[k] = string(v)
	}
	e.metadata = metadata.New(headers)
	e.callOptions = []grpc.CallOption{
		grpc.WaitForReady(e.config.ClientConfig.WaitForReady),
	}

	return
}

func (e *baseExporter) shutdown(context.Context) error {
	if e.clientConn != nil {
		return e.clientConn.Close()
	}
	return nil
}

func (e *baseExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	req := ptraceotlp.NewExportRequestFromTraces(td)
	resp, respErr := e.traceExporter.Export(e.enhanceContext(ctx), req, e.callOptions...)
	if err := processError(respErr); err != nil {
		return err
	}
	partialSuccess := resp.PartialSuccess()
	if !(partialSuccess.ErrorMessage() == "" && partialSuccess.RejectedSpans() == 0) {
		e.settings.Logger.Warn("Partial success response",
			zap.String("message", resp.PartialSuccess().ErrorMessage()),
			zap.Int64("dropped_spans", resp.PartialSuccess().RejectedSpans()),
		)
	}
	return nil
}

func (e *baseExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	req := pmetricotlp.NewExportRequestFromMetrics(md)
	resp, respErr := e.metricExporter.Export(e.enhanceContext(ctx), req, e.callOptions...)
	if err := processError(respErr); err != nil {
		return err
	}
	partialSuccess := resp.PartialSuccess()
	if !(partialSuccess.ErrorMessage() == "" && partialSuccess.RejectedDataPoints() == 0) {
		e.settings.Logger.Warn("Partial success response",
			zap.String("message", resp.PartialSuccess().ErrorMessage()),
			zap.Int64("dropped_data_points", resp.PartialSuccess().RejectedDataPoints()),
		)
	}
	return nil
}

func (e *baseExporter) pushLogs(ctx context.Context, ld plog.Logs) error {
	req := plogotlp.NewExportRequestFromLogs(ld)
	resp, respErr := e.logExporter.Export(e.enhanceContext(ctx), req, e.callOptions...)
	if err := processError(respErr); err != nil {
		return err
	}
	partialSuccess := resp.PartialSuccess()
	if !(partialSuccess.ErrorMessage() == "" && partialSuccess.RejectedLogRecords() == 0) {
		e.settings.Logger.Warn("Partial success response",
			zap.String("message", resp.PartialSuccess().ErrorMessage()),
			zap.Int64("dropped_log_records", resp.PartialSuccess().RejectedLogRecords()),
		)
	}
	return nil
}

func (e *baseExporter) enhanceContext(ctx context.Context) context.Context {
	if e.metadata.Len() > 0 {
		return metadata.NewOutgoingContext(ctx, e.metadata)
	}
	return ctx
}

func processError(err error) error {
	if err == nil {
		// Request is successful, we are done.
		return nil
	}

	// We have an error, check gRPC status code.
	st := status.Convert(err)
	if st.Code() == codes.OK {
		// Not really an error, still success.
		return nil
	}

	// Now, this is a real error.
	retryInfo := getRetryInfo(st)

	if !shouldRetry(st.Code(), retryInfo) {
		// It is not a retryable error, we should not retry.
		return consumererror.NewPermanent(err)
	}

	// Check if server returned throttling information.
	throttleDuration := getThrottleDuration(retryInfo)
	if throttleDuration != 0 {
		// We are throttled. Wait before retrying as requested by the server.
		return exporterhelper.NewThrottleRetry(err, throttleDuration)
	}

	// Need to retry.
	return err
}

func shouldRetry(code codes.Code, retryInfo *errdetails.RetryInfo) bool {
	switch code {
	case codes.Canceled,
		codes.DeadlineExceeded,
		codes.Aborted,
		codes.OutOfRange,
		codes.Unavailable,
		codes.DataLoss:
		// These are retryable errors.
		return true
	case codes.ResourceExhausted:
		// Retry only if RetryInfo was supplied by the server.
		// This indicates that the server can still recover from resource exhaustion.
		return retryInfo != nil
	}
	// Don't retry on any other code.
	return false
}

func getRetryInfo(status *status.Status) *errdetails.RetryInfo {
	for _, detail := range status.Details() {
		if t, ok := detail.(*errdetails.RetryInfo); ok {
			return t
		}
	}
	return nil
}

func getThrottleDuration(t *errdetails.RetryInfo) time.Duration {
	if t == nil || t.RetryDelay == nil {
		return 0
	}
	if t.RetryDelay.Seconds > 0 || t.RetryDelay.Nanos > 0 {
		return time.Duration(t.RetryDelay.Seconds)*time.Second + time.Duration(t.RetryDelay.Nanos)*time.Nanosecond
	}
	return 0
}
