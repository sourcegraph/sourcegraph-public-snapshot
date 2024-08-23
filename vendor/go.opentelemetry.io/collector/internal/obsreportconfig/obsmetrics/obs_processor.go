// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package obsmetrics // import "go.opentelemetry.io/collector/internal/obsreportconfig/obsmetrics"

const (
	// ProcessorKey is the key used to identify processors in metrics and traces.
	ProcessorKey = "processor"

	// DroppedSpansKey is the key used to identify spans dropped by the Collector.
	DroppedSpansKey = "dropped_spans"

	// DroppedMetricPointsKey is the key used to identify metric points dropped by the Collector.
	DroppedMetricPointsKey = "dropped_metric_points"

	// DroppedLogRecordsKey is the key used to identify log records dropped by the Collector.
	DroppedLogRecordsKey = "dropped_log_records"
)

var (
	ProcessorMetricPrefix = ProcessorKey + MetricNameSep
)
