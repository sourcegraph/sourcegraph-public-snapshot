// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper/internal/metadata"
	"go.opentelemetry.io/collector/internal/obsreportconfig/obsmetrics"
)

// ObsReport is a helper to add observability to an exporter.
type ObsReport struct {
	level          configtelemetry.Level
	spanNamePrefix string
	tracer         trace.Tracer

	otelAttrs        []attribute.KeyValue
	telemetryBuilder *metadata.TelemetryBuilder
}

// ObsReportSettings are settings for creating an ObsReport.
type ObsReportSettings struct {
	ExporterID             component.ID
	ExporterCreateSettings exporter.Settings
}

// NewObsReport creates a new Exporter.
func NewObsReport(cfg ObsReportSettings) (*ObsReport, error) {
	return newExporter(cfg)
}

func newExporter(cfg ObsReportSettings) (*ObsReport, error) {
	telemetryBuilder, err := metadata.NewTelemetryBuilder(cfg.ExporterCreateSettings.TelemetrySettings,
		metadata.WithAttributeSet(attribute.NewSet(attribute.String(obsmetrics.ExporterKey, cfg.ExporterID.String()))),
	)
	if err != nil {
		return nil, err
	}

	return &ObsReport{
		level:          cfg.ExporterCreateSettings.TelemetrySettings.MetricsLevel,
		spanNamePrefix: obsmetrics.ExporterPrefix + cfg.ExporterID.String(),
		tracer:         cfg.ExporterCreateSettings.TracerProvider.Tracer(cfg.ExporterID.String()),

		otelAttrs: []attribute.KeyValue{
			attribute.String(obsmetrics.ExporterKey, cfg.ExporterID.String()),
		},
		telemetryBuilder: telemetryBuilder,
	}, nil
}

// StartTracesOp is called at the start of an Export operation.
// The returned context should be used in other calls to the Exporter functions
// dealing with the same export operation.
func (or *ObsReport) StartTracesOp(ctx context.Context) context.Context {
	return or.startOp(ctx, obsmetrics.ExportTraceDataOperationSuffix)
}

// EndTracesOp completes the export operation that was started with StartTracesOp.
func (or *ObsReport) EndTracesOp(ctx context.Context, numSpans int, err error) {
	numSent, numFailedToSend := toNumItems(numSpans, err)
	or.recordMetrics(context.WithoutCancel(ctx), component.DataTypeTraces, numSent, numFailedToSend)
	endSpan(ctx, err, numSent, numFailedToSend, obsmetrics.SentSpansKey, obsmetrics.FailedToSendSpansKey)
}

// StartMetricsOp is called at the start of an Export operation.
// The returned context should be used in other calls to the Exporter functions
// dealing with the same export operation.
func (or *ObsReport) StartMetricsOp(ctx context.Context) context.Context {
	return or.startOp(ctx, obsmetrics.ExportMetricsOperationSuffix)
}

// EndMetricsOp completes the export operation that was started with
// StartMetricsOp.
func (or *ObsReport) EndMetricsOp(ctx context.Context, numMetricPoints int, err error) {
	numSent, numFailedToSend := toNumItems(numMetricPoints, err)
	or.recordMetrics(context.WithoutCancel(ctx), component.DataTypeMetrics, numSent, numFailedToSend)
	endSpan(ctx, err, numSent, numFailedToSend, obsmetrics.SentMetricPointsKey, obsmetrics.FailedToSendMetricPointsKey)
}

// StartLogsOp is called at the start of an Export operation.
// The returned context should be used in other calls to the Exporter functions
// dealing with the same export operation.
func (or *ObsReport) StartLogsOp(ctx context.Context) context.Context {
	return or.startOp(ctx, obsmetrics.ExportLogsOperationSuffix)
}

// EndLogsOp completes the export operation that was started with StartLogsOp.
func (or *ObsReport) EndLogsOp(ctx context.Context, numLogRecords int, err error) {
	numSent, numFailedToSend := toNumItems(numLogRecords, err)
	or.recordMetrics(context.WithoutCancel(ctx), component.DataTypeLogs, numSent, numFailedToSend)
	endSpan(ctx, err, numSent, numFailedToSend, obsmetrics.SentLogRecordsKey, obsmetrics.FailedToSendLogRecordsKey)
}

// startOp creates the span used to trace the operation. Returning
// the updated context and the created span.
func (or *ObsReport) startOp(ctx context.Context, operationSuffix string) context.Context {
	spanName := or.spanNamePrefix + operationSuffix
	ctx, _ = or.tracer.Start(ctx, spanName)
	return ctx
}

func (or *ObsReport) recordMetrics(ctx context.Context, dataType component.DataType, sent, failed int64) {
	if or.level == configtelemetry.LevelNone {
		return
	}
	var sentMeasure, failedMeasure metric.Int64Counter
	switch dataType {
	case component.DataTypeTraces:
		sentMeasure = or.telemetryBuilder.ExporterSentSpans
		failedMeasure = or.telemetryBuilder.ExporterSendFailedSpans
	case component.DataTypeMetrics:
		sentMeasure = or.telemetryBuilder.ExporterSentMetricPoints
		failedMeasure = or.telemetryBuilder.ExporterSendFailedMetricPoints
	case component.DataTypeLogs:
		sentMeasure = or.telemetryBuilder.ExporterSentLogRecords
		failedMeasure = or.telemetryBuilder.ExporterSendFailedLogRecords
	}

	sentMeasure.Add(ctx, sent, metric.WithAttributes(or.otelAttrs...))
	failedMeasure.Add(ctx, failed, metric.WithAttributes(or.otelAttrs...))
}

func endSpan(ctx context.Context, err error, numSent, numFailedToSend int64, sentItemsKey, failedToSendItemsKey string) {
	span := trace.SpanFromContext(ctx)
	// End the span according to errors.
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int64(sentItemsKey, numSent),
			attribute.Int64(failedToSendItemsKey, numFailedToSend),
		)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
	}
	span.End()
}

func toNumItems(numExportedItems int, err error) (int64, int64) {
	if err != nil {
		return 0, int64(numExportedItems)
	}
	return int64(numExportedItems), 0
}

func (or *ObsReport) recordEnqueueFailure(ctx context.Context, dataType component.DataType, failed int64) {
	var enqueueFailedMeasure metric.Int64Counter
	switch dataType {
	case component.DataTypeTraces:
		enqueueFailedMeasure = or.telemetryBuilder.ExporterEnqueueFailedSpans
	case component.DataTypeMetrics:
		enqueueFailedMeasure = or.telemetryBuilder.ExporterEnqueueFailedMetricPoints
	case component.DataTypeLogs:
		enqueueFailedMeasure = or.telemetryBuilder.ExporterEnqueueFailedLogRecords
	}

	enqueueFailedMeasure.Add(ctx, failed, metric.WithAttributes(or.otelAttrs...))
}
