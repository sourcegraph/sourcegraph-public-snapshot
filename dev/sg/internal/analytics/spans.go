package analytics

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newSpanToDiskProcessor creates an OpenTelemetry span processor that persists spans
// to disk in protojson format.
func newSpanToDiskProcessor(ctx context.Context) (tracesdk.SpanProcessor, error) {
	exporter, err := otlptrace.New(ctx, &otlpDiskClient{})
	if err != nil {
		return nil, errors.Wrap(err, "create exporter")
	}
	return tracesdk.NewBatchSpanProcessor(exporter), nil
}

type spansStoreKey struct{}

// spansStore manages the OpenTelemetry tracer provider that manages all events associated
// with a run of sg.
type spansStore struct {
	rootSpan    trace.Span
	provider    *oteltracesdk.TracerProvider
	persistOnce sync.Once
}

// getStore retrieves the events store from context if it exists. Callers should check
// that the store is non-nil before attempting to use it.
func getStore(ctx context.Context) *spansStore {
	store, ok := ctx.Value(spansStoreKey{}).(*spansStore)
	if !ok {
		return nil
	}
	return store
}

// Persist is called once per sg run, at the end, to save events
func (s *spansStore) Persist(ctx context.Context) error {
	var err error
	s.persistOnce.Do(func() {
		s.rootSpan.End()
		err = s.provider.Shutdown(ctx)
	})
	return err
}

func spansPath() (string, error) {
	home, err := root.GetSGHomePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "spans"), nil
}

// otlpDiskClient is an OpenTelemetry trace client that "sends" spans to disk, instead of
// to an external collector.
type otlpDiskClient struct {
	f         *os.File
	uploadMux sync.Mutex
}

var _ otlptrace.Client = &otlpDiskClient{}

// Start should establish connection(s) to endpoint(s). It is
// called just once by the exporter, so the implementation
// does not need to worry about idempotence and locking.
func (c *otlpDiskClient) Start(ctx context.Context) error {
	p, err := spansPath()
	if err != nil {
		return err
	}
	c.f, err = os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	return err
}

// Stop should close the connections. The function is called
// only once by the exporter, so the implementation does not
// need to worry about idempotence, but it may be called
// concurrently with UploadTraces, so proper
// locking is required. The function serves as a
// synchronization point - after the function returns, the
// process of closing connections is assumed to be finished.
func (c *otlpDiskClient) Stop(ctx context.Context) error {
	c.uploadMux.Lock()
	defer c.uploadMux.Unlock()

	if err := c.f.Sync(); err != nil {
		return errors.Wrap(err, "file.Sync")
	}
	return c.f.Close()
}

// UploadTraces should transform the passed traces to the wire
// format and send it to the collector. May be called
// concurrently.
func (c *otlpDiskClient) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	c.uploadMux.Lock()
	defer c.uploadMux.Unlock()

	// Create a request we can marshal
	req := coltracepb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	}
	b, err := protojson.Marshal(&req)
	if err != nil {
		return errors.Wrap(err, "protojson.Marshal")
	}
	if _, err := c.f.Write(append(b, '\n')); err != nil {
		return errors.Wrap(err, "Write")
	}
	return c.f.Sync()
}
