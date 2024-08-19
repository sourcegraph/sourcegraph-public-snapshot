// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package obsmetrics defines the obsreport metrics for each components
// all the metrics is in OpenCensus format which will be replaced with OTEL Metrics
// in the future
package obsmetrics // import "go.opentelemetry.io/collector/internal/obsreportconfig/obsmetrics"

const (
	SpanNameSep   = "/"
	MetricNameSep = "_"
	Scope         = "go.opentelemetry.io/collector/obsreport"
)
