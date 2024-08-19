// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"errors"
)

var (
	// errNilConfig is returned when an empty name is given.
	errNilConfig = errors.New("nil config")
	// errNilLogger is returned when a logger is nil
	errNilLogger = errors.New("nil logger")
	// errNilPushTraceData is returned when a nil PushTraces is given.
	errNilPushTraceData = errors.New("nil PushTraces")
	// errNilPushMetricsData is returned when a nil PushMetrics is given.
	errNilPushMetricsData = errors.New("nil PushMetrics")
	// errNilPushLogsData is returned when a nil PushLogs is given.
	errNilPushLogsData = errors.New("nil PushLogs")
	// errNilTracesConverter is returned when a nil RequestFromTracesFunc is given.
	errNilTracesConverter = errors.New("nil RequestFromTracesFunc")
	// errNilMetricsConverter is returned when a nil RequestFromMetricsFunc is given.
	errNilMetricsConverter = errors.New("nil RequestFromMetricsFunc")
	// errNilLogsConverter is returned when a nil RequestFromLogsFunc is given.
	errNilLogsConverter = errors.New("nil RequestFromLogsFunc")
)
