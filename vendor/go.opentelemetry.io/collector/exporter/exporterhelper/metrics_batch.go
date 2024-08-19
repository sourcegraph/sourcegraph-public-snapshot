// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/exporter/exporterbatcher"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// mergeMetrics merges two metrics requests into one.
func mergeMetrics(_ context.Context, r1 Request, r2 Request) (Request, error) {
	mr1, ok1 := r1.(*metricsRequest)
	mr2, ok2 := r2.(*metricsRequest)
	if !ok1 || !ok2 {
		return nil, errors.New("invalid input type")
	}
	mr2.md.ResourceMetrics().MoveAndAppendTo(mr1.md.ResourceMetrics())
	return mr1, nil
}

// mergeSplitMetrics splits and/or merges the metrics into multiple requests based on the MaxSizeConfig.
func mergeSplitMetrics(_ context.Context, cfg exporterbatcher.MaxSizeConfig, r1 Request, r2 Request) ([]Request, error) {
	var (
		res          []Request
		destReq      *metricsRequest
		capacityLeft = cfg.MaxSizeItems
	)
	for _, req := range []Request{r1, r2} {
		if req == nil {
			continue
		}
		srcReq, ok := req.(*metricsRequest)
		if !ok {
			return nil, errors.New("invalid input type")
		}
		if srcReq.md.DataPointCount() <= capacityLeft {
			if destReq == nil {
				destReq = srcReq
			} else {
				srcReq.md.ResourceMetrics().MoveAndAppendTo(destReq.md.ResourceMetrics())
			}
			capacityLeft -= destReq.md.DataPointCount()
			continue
		}

		for {
			extractedMetrics := extractMetrics(srcReq.md, capacityLeft)
			if extractedMetrics.DataPointCount() == 0 {
				break
			}
			capacityLeft -= extractedMetrics.DataPointCount()
			if destReq == nil {
				destReq = &metricsRequest{md: extractedMetrics, pusher: srcReq.pusher}
			} else {
				extractedMetrics.ResourceMetrics().MoveAndAppendTo(destReq.md.ResourceMetrics())
			}
			// Create new batch once capacity is reached.
			if capacityLeft == 0 {
				res = append(res, destReq)
				destReq = nil
				capacityLeft = cfg.MaxSizeItems
			}
		}
	}

	if destReq != nil {
		res = append(res, destReq)
	}

	return res, nil
}

// extractMetrics extracts metrics from srcMetrics until count of data points is reached.
func extractMetrics(srcMetrics pmetric.Metrics, count int) pmetric.Metrics {
	destMetrics := pmetric.NewMetrics()
	srcMetrics.ResourceMetrics().RemoveIf(func(srcRM pmetric.ResourceMetrics) bool {
		if count == 0 {
			return false
		}
		needToExtract := resourceDataPointsCount(srcRM) > count
		if needToExtract {
			srcRM = extractResourceMetrics(srcRM, count)
		}
		count -= resourceDataPointsCount(srcRM)
		srcRM.MoveTo(destMetrics.ResourceMetrics().AppendEmpty())
		return !needToExtract
	})
	return destMetrics
}

// extractResourceMetrics extracts resource metrics and returns a new resource metrics with the specified number of data points.
func extractResourceMetrics(srcRM pmetric.ResourceMetrics, count int) pmetric.ResourceMetrics {
	destRM := pmetric.NewResourceMetrics()
	destRM.SetSchemaUrl(srcRM.SchemaUrl())
	srcRM.Resource().CopyTo(destRM.Resource())
	srcRM.ScopeMetrics().RemoveIf(func(srcSM pmetric.ScopeMetrics) bool {
		if count == 0 {
			return false
		}
		needToExtract := scopeDataPointsCount(srcSM) > count
		if needToExtract {
			srcSM = extractScopeMetrics(srcSM, count)
		}
		count -= scopeDataPointsCount(srcSM)
		srcSM.MoveTo(destRM.ScopeMetrics().AppendEmpty())
		return !needToExtract
	})
	return destRM
}

// extractScopeMetrics extracts scope metrics and returns a new scope metrics with the specified number of data points.
func extractScopeMetrics(srcSM pmetric.ScopeMetrics, count int) pmetric.ScopeMetrics {
	destSM := pmetric.NewScopeMetrics()
	destSM.SetSchemaUrl(srcSM.SchemaUrl())
	srcSM.Scope().CopyTo(destSM.Scope())
	srcSM.Metrics().RemoveIf(func(srcMetric pmetric.Metric) bool {
		if count == 0 {
			return false
		}
		needToExtract := metricDataPointCount(srcMetric) > count
		if needToExtract {
			srcMetric = extractMetricDataPoints(srcMetric, count)
		}
		count -= metricDataPointCount(srcMetric)
		srcMetric.MoveTo(destSM.Metrics().AppendEmpty())
		return !needToExtract
	})
	return destSM
}

func extractMetricDataPoints(srcMetric pmetric.Metric, count int) pmetric.Metric {
	destMetric := pmetric.NewMetric()
	switch srcMetric.Type() {
	case pmetric.MetricTypeGauge:
		extractGaugeDataPoints(srcMetric.Gauge(), count, destMetric.SetEmptyGauge())
	case pmetric.MetricTypeSum:
		extractSumDataPoints(srcMetric.Sum(), count, destMetric.SetEmptySum())
	case pmetric.MetricTypeHistogram:
		extractHistogramDataPoints(srcMetric.Histogram(), count, destMetric.SetEmptyHistogram())
	case pmetric.MetricTypeExponentialHistogram:
		extractExponentialHistogramDataPoints(srcMetric.ExponentialHistogram(), count,
			destMetric.SetEmptyExponentialHistogram())
	case pmetric.MetricTypeSummary:
		extractSummaryDataPoints(srcMetric.Summary(), count, destMetric.SetEmptySummary())
	}
	return destMetric
}

func extractGaugeDataPoints(srcGauge pmetric.Gauge, count int, destGauge pmetric.Gauge) {
	srcGauge.DataPoints().RemoveIf(func(srcDP pmetric.NumberDataPoint) bool {
		if count == 0 {
			return false
		}
		srcDP.MoveTo(destGauge.DataPoints().AppendEmpty())
		count--
		return true
	})
}

func extractSumDataPoints(srcSum pmetric.Sum, count int, destSum pmetric.Sum) {
	srcSum.DataPoints().RemoveIf(func(srcDP pmetric.NumberDataPoint) bool {
		if count == 0 {
			return false
		}
		srcDP.MoveTo(destSum.DataPoints().AppendEmpty())
		count--
		return true
	})
}

func extractHistogramDataPoints(srcHistogram pmetric.Histogram, count int, destHistogram pmetric.Histogram) {
	srcHistogram.DataPoints().RemoveIf(func(srcDP pmetric.HistogramDataPoint) bool {
		if count == 0 {
			return false
		}
		srcDP.MoveTo(destHistogram.DataPoints().AppendEmpty())
		count--
		return true
	})
}

func extractExponentialHistogramDataPoints(srcExponentialHistogram pmetric.ExponentialHistogram, count int, destExponentialHistogram pmetric.ExponentialHistogram) {
	srcExponentialHistogram.DataPoints().RemoveIf(func(srcDP pmetric.ExponentialHistogramDataPoint) bool {
		if count == 0 {
			return false
		}
		srcDP.MoveTo(destExponentialHistogram.DataPoints().AppendEmpty())
		count--
		return true
	})
}

func extractSummaryDataPoints(srcSummary pmetric.Summary, count int, destSummary pmetric.Summary) {
	srcSummary.DataPoints().RemoveIf(func(srcDP pmetric.SummaryDataPoint) bool {
		if count == 0 {
			return false
		}
		srcDP.MoveTo(destSummary.DataPoints().AppendEmpty())
		count--
		return true
	})
}

func resourceDataPointsCount(rm pmetric.ResourceMetrics) (count int) {
	for i := 0; i < rm.ScopeMetrics().Len(); i++ {
		count += scopeDataPointsCount(rm.ScopeMetrics().At(i))
	}
	return count
}

func scopeDataPointsCount(sm pmetric.ScopeMetrics) (count int) {
	for i := 0; i < sm.Metrics().Len(); i++ {
		count += metricDataPointCount(sm.Metrics().At(i))
	}
	return count
}

func metricDataPointCount(m pmetric.Metric) int {
	switch m.Type() {
	case pmetric.MetricTypeGauge:
		return m.Gauge().DataPoints().Len()
	case pmetric.MetricTypeSum:
		return m.Sum().DataPoints().Len()
	case pmetric.MetricTypeHistogram:
		return m.Histogram().DataPoints().Len()
	case pmetric.MetricTypeExponentialHistogram:
		return m.ExponentialHistogram().DataPoints().Len()
	case pmetric.MetricTypeSummary:
		return m.Summary().DataPoints().Len()
	}
	return 0
}
