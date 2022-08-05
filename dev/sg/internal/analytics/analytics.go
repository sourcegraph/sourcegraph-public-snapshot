package analytics

import (
	"bufio"
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	sgAnalyticsVersionResourceKey = "sg.analytics_version"
	// Increment to make breaking changes to spans and discard old spans
	sgAnalyticsVersion = "v1.1"
)

const (
	honeycombEndpoint  = "grpc://api.honeycomb.io:443"
	otlpEndpointEnvKey = "OTEL_EXPORTER_OTLP_ENDPOINT"
)

// Submit pushes all persisted events to Honeycomb if OTEL_EXPORTER_OTLP_ENDPOINT is not
// set.
func Submit(ctx context.Context, honeycombToken string) error {
	spans, err := Load()
	if err != nil {
		return err
	}
	if len(spans) == 0 {
		return errors.New("no spans to submit")
	}

	// if endpoint is not set, point to Honeycomb
	var otlpOptions []otlptracegrpc.Option
	if _, exists := os.LookupEnv(otlpEndpointEnvKey); !exists {
		os.Setenv(otlpEndpointEnvKey, honeycombEndpoint)
		otlpOptions = append(otlpOptions, otlptracegrpc.WithHeaders(map[string]string{
			"x-honeycomb-team": honeycombToken,
		}))
	}

	// Set up a trace exporter
	client := otlptracegrpc.NewClient(otlpOptions...)
	if err := client.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to initialize export client")
	}

	// send spans and shut down
	if err := client.UploadTraces(ctx, spans); err != nil {
		return errors.Wrap(err, "failed to export spans")
	}
	if err := client.Stop(ctx); err != nil {
		return errors.Wrap(err, "failed to flush span exporter")
	}

	return nil
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
	p, err := spansPath()
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
func Load() (spans []*tracepb.ResourceSpans, errs error) {
	p, err := spansPath()
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
		var req coltracepb.ExportTraceServiceRequest
		if err := protojson.Unmarshal(scanner.Bytes(), &req); err != nil {
			errs = errors.Append(errs, err)
			continue // drop malformed data
		}

		for _, s := range req.GetResourceSpans() {
			if !isValidVersion(s) {
				continue
			}
			spans = append(spans, s)
		}
	}
	return
}
