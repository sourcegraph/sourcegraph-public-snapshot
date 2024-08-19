// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package obsmetrics // import "go.opentelemetry.io/collector/internal/obsreportconfig/obsmetrics"

const (
	// ScraperKey used to identify scrapers in metrics and traces.
	ScraperKey = "scraper"

	// ScrapedMetricPointsKey used to identify metric points scraped by the
	// Collector.
	ScrapedMetricPointsKey = "scraped_metric_points"
	// ErroredMetricPointsKey used to identify metric points errored (i.e.
	// unable to be scraped) by the Collector.
	ErroredMetricPointsKey = "errored_metric_points"
)

const (
	ScraperPrefix                 = ScraperKey + SpanNameSep
	ScraperMetricPrefix           = ScraperKey + MetricNameSep
	ScraperMetricsOperationSuffix = SpanNameSep + "MetricsScraped"
)
