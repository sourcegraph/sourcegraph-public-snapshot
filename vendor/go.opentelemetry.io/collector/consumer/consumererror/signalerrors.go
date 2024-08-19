// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package consumererror // import "go.opentelemetry.io/collector/consumer/consumererror"

import (
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type retryable[V ptrace.Traces | pmetric.Metrics | plog.Logs] struct {
	error
	data V
}

// Unwrap returns the wrapped error for functions Is and As in standard package errors.
func (err retryable[V]) Unwrap() error {
	return err.error
}

// Data returns the telemetry data that failed to be processed or sent.
func (err retryable[V]) Data() V {
	return err.data
}

// Traces is an error that may carry associated Trace data for a subset of received data
// that failed to be processed or sent.
type Traces struct {
	retryable[ptrace.Traces]
}

// NewTraces creates a Traces that can encapsulate received data that failed to be processed or sent.
func NewTraces(err error, data ptrace.Traces) error {
	return Traces{
		retryable: retryable[ptrace.Traces]{
			error: err,
			data:  data,
		},
	}
}

// Logs is an error that may carry associated Log data for a subset of received data
// that failed to be processed or sent.
type Logs struct {
	retryable[plog.Logs]
}

// NewLogs creates a Logs that can encapsulate received data that failed to be processed or sent.
func NewLogs(err error, data plog.Logs) error {
	return Logs{
		retryable: retryable[plog.Logs]{
			error: err,
			data:  data,
		},
	}
}

// Metrics is an error that may carry associated Metrics data for a subset of received data
// that failed to be processed or sent.
type Metrics struct {
	retryable[pmetric.Metrics]
}

// NewMetrics creates a Metrics that can encapsulate received data that failed to be processed or sent.
func NewMetrics(err error, data pmetric.Metrics) error {
	return Metrics{
		retryable: retryable[pmetric.Metrics]{
			error: err,
			data:  data,
		},
	}
}
