package analytics

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const defaultOtlpEndpoint = "api.honeycomb.io:443"
const otlpEndpointKey = "OTEL_EXPORTER_OTLP_ENDPOINT"

// Submit pushes all persisted events to OkayHQ.
func Submit(ctx context.Context, honeycombToken string, gitHubLogin string) error {
	spans, err := Load()
	if err != nil {
		return err
	}

	if _, exists := os.LookupEnv(otlpEndpointKey); !exists {
		os.Setenv(otlpEndpointKey, defaultOtlpEndpoint)
		// TODO auth
	}

	// Set up a trace exporter
	client := otlptracegrpc.NewClient()
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return errors.Wrap(err, "failed to create trace exporter")
	}

	// send spans and shut down
	if err := exporter.ExportSpans(ctx, spans); err != nil {
		return err
	}

	return exporter.Shutdown(ctx)
}

// Persist stores all events in context to disk.
func Persist(ctx context.Context) error {
	store := getStore(ctx)
	if store == nil {
		return nil
	}
	return store.Persist(ctx)
}

// Reset deletes all persisted events.
func Reset() error {
	p, err := eventsPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		// don't have to remove something that doesn't exist
		return nil
	}
	return os.Remove(p)
}

// Load retrieves all persisted events.
func Load() (spans []trace.ReadOnlySpan, err error) {
	p, err := eventsPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Don't worry too much about malformed events, analytics are relatively optional
		// so just grab what we can.
		var span tracetest.SpanStub
		if err := json.Unmarshal(scanner.Bytes(), &span); err != nil {
			continue // drop malformed data
		}

		// Now we ensure that this event is something we want to keep
		if !span.SpanContext.IsValid() {
			continue
		}

		spans = append(spans, span.Snapshot())
	}
	return
}
