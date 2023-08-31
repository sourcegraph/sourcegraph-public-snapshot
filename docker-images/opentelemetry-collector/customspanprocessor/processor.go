package spanattrprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

var _ processor.Traces = (*TraceProcessor)(nil)
var _ consumer.Traces = (*TraceProcessor)(nil)

type TraceProcessor struct {
	next   consumer.Traces
	logger *zap.Logger
}

// Capabilities implements processor.Traces.
func (tp *TraceProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

// ConsumeTraces implements processor.Traces.
func (tp *TraceProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	tp.logger.Log(zap.InfoLevel, "CONSUME TRACES")
	resSpans := td.ResourceSpans()
	tp.logger.Log(zap.InfoLevel, "ResourceSpan Count", zap.Int("Count", resSpans.Len()))
	tp.logger.Log(zap.InfoLevel, "Span Count", zap.Int("Span Count", td.SpanCount()))
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		span := resSpans.At(i)
		tp.logger.Log(zap.InfoLevel, "span scope atrtibutes", zap.Int("count", span.ScopeSpans().Len()))
		tp.logger.Log(zap.InfoLevel, "span resource scope atrtibutes", zap.Int("count", span.Resource().Attributes().Len()))
	}
	return nil
}

// Shutdown implements processor.Traces.
func (tp *TraceProcessor) Shutdown(ctx context.Context) error {
	tp.logger.Log(zap.InfoLevel, "SHUTTING DOWN CUSTOM ATTRIBUTE PROCESSOR")
	return nil
}

// Start implements processor.Traces.
func (tp *TraceProcessor) Start(ctx context.Context, host component.Host) error {
	tp.logger.Log(zap.InfoLevel, "STARTING CUSTOM ATTRIBUTE PROCESSOR")
	return nil
}
