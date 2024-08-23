// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/internal/experr"
	"go.opentelemetry.io/collector/internal/obsreportconfig/obsmetrics"
)

// TODO: Clean this by forcing all exporters to return an internal error type that always include the information about retries.
type throttleRetry struct {
	err   error
	delay time.Duration
}

func (t throttleRetry) Error() string {
	return "Throttle (" + t.delay.String() + "), error: " + t.err.Error()
}

func (t throttleRetry) Unwrap() error {
	return t.err
}

// NewThrottleRetry creates a new throttle retry error.
func NewThrottleRetry(err error, delay time.Duration) error {
	return throttleRetry{
		err:   err,
		delay: delay,
	}
}

type retrySender struct {
	baseRequestSender
	traceAttribute attribute.KeyValue
	cfg            configretry.BackOffConfig
	stopCh         chan struct{}
	logger         *zap.Logger
}

func newRetrySender(config configretry.BackOffConfig, set exporter.Settings) *retrySender {
	return &retrySender{
		traceAttribute: attribute.String(obsmetrics.ExporterKey, set.ID.String()),
		cfg:            config,
		stopCh:         make(chan struct{}),
		logger:         set.Logger,
	}
}

func (rs *retrySender) Shutdown(context.Context) error {
	close(rs.stopCh)
	return nil
}

// send implements the requestSender interface
func (rs *retrySender) send(ctx context.Context, req Request) error {
	// Do not use NewExponentialBackOff since it calls Reset and the code here must
	// call Reset after changing the InitialInterval (this saves an unnecessary call to Now).
	expBackoff := backoff.ExponentialBackOff{
		InitialInterval:     rs.cfg.InitialInterval,
		RandomizationFactor: rs.cfg.RandomizationFactor,
		Multiplier:          rs.cfg.Multiplier,
		MaxInterval:         rs.cfg.MaxInterval,
		MaxElapsedTime:      rs.cfg.MaxElapsedTime,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	expBackoff.Reset()
	span := trace.SpanFromContext(ctx)
	retryNum := int64(0)
	for {
		span.AddEvent(
			"Sending request.",
			trace.WithAttributes(rs.traceAttribute, attribute.Int64("retry_num", retryNum)))

		err := rs.nextSender.send(ctx, req)
		if err == nil {
			return nil
		}

		// Immediately drop data on permanent errors.
		if consumererror.IsPermanent(err) {
			return fmt.Errorf("not retryable error: %w", err)
		}

		req = extractPartialRequest(req, err)

		backoffDelay := expBackoff.NextBackOff()
		if backoffDelay == backoff.Stop {
			return fmt.Errorf("no more retries left: %w", err)
		}

		throttleErr := throttleRetry{}
		if errors.As(err, &throttleErr) {
			backoffDelay = max(backoffDelay, throttleErr.delay)
		}

		backoffDelayStr := backoffDelay.String()
		span.AddEvent(
			"Exporting failed. Will retry the request after interval.",
			trace.WithAttributes(
				rs.traceAttribute,
				attribute.String("interval", backoffDelayStr),
				attribute.String("error", err.Error())))
		rs.logger.Info(
			"Exporting failed. Will retry the request after interval.",
			zap.Error(err),
			zap.String("interval", backoffDelayStr),
		)
		retryNum++

		// back-off, but get interrupted when shutting down or request is cancelled or timed out.
		select {
		case <-ctx.Done():
			return fmt.Errorf("request is cancelled or timed out %w", err)
		case <-rs.stopCh:
			return experr.NewShutdownErr(err)
		case <-time.After(backoffDelay):
		}
	}
}

// max returns the larger of x or y.
func max(x, y time.Duration) time.Duration {
	if x < y {
		return y
	}
	return x
}
