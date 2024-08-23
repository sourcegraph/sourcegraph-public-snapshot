// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package obsmetrics // import "go.opentelemetry.io/collector/internal/obsreportconfig/obsmetrics"

const (
	// ExporterKey used to identify exporters in metrics and traces.
	ExporterKey = "exporter"

	// SentSpansKey used to track spans sent by exporters.
	SentSpansKey = "sent_spans"
	// FailedToSendSpansKey used to track spans that failed to be sent by exporters.
	FailedToSendSpansKey = "send_failed_spans"
	// FailedToEnqueueSpansKey used to track spans that failed to be enqueued by exporters.
	FailedToEnqueueSpansKey = "enqueue_failed_spans"

	// SentMetricPointsKey used to track metric points sent by exporters.
	SentMetricPointsKey = "sent_metric_points"
	// FailedToSendMetricPointsKey used to track metric points that failed to be sent by exporters.
	FailedToSendMetricPointsKey = "send_failed_metric_points"
	// FailedToEnqueueMetricPointsKey used to track metric points that failed to be enqueued by exporters.
	FailedToEnqueueMetricPointsKey = "enqueue_failed_metric_points"

	// SentLogRecordsKey used to track logs sent by exporters.
	SentLogRecordsKey = "sent_log_records"
	// FailedToSendLogRecordsKey used to track logs that failed to be sent by exporters.
	FailedToSendLogRecordsKey = "send_failed_log_records"
	// FailedToEnqueueLogRecordsKey used to track logs that failed to be enqueued by exporters.
	FailedToEnqueueLogRecordsKey = "enqueue_failed_log_records"
)

var (
	ExporterPrefix                 = ExporterKey + SpanNameSep
	ExporterMetricPrefix           = ExporterKey + MetricNameSep
	ExportTraceDataOperationSuffix = SpanNameSep + "traces"
	ExportMetricsOperationSuffix   = SpanNameSep + "metrics"
	ExportLogsOperationSuffix      = SpanNameSep + "logs"
)
