package analytics

import (
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func newSpanToDiskProcessor() (trace.SpanProcessor, error) {
	p, err := eventsPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	exporter, err := stdouttrace.New(stdouttrace.WithWriter(f))
	if err != nil {
		return nil, err
	}

	return trace.NewBatchSpanProcessor(exporter), nil
}
