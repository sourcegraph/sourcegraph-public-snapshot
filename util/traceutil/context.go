package traceutil

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
)

type contextKey int

const (
	spanIDKey contextKey = iota
)

func NewContext(parent context.Context, id appdash.SpanID) context.Context {
	return context.WithValue(parent, spanIDKey, id)
}

func SpanIDFromContext(ctx context.Context) appdash.SpanID {
	id, ok := ctx.Value(spanIDKey).(appdash.SpanID)
	if !ok {
		// log.Println("Warning: no span ID set in context")
	}
	return id
}

// span is an HTTP transport and gRPC credential provider that
// adds an Appdash span ID to outgoing requests.
type span struct {
	spanID appdash.SpanID
}

func (t *span) NewTransport(underlying http.RoundTripper) http.RoundTripper {
	if DefaultCollector == nil {
		return underlying
	}
	return &httptrace.Transport{
		Recorder:  appdash.NewRecorder(t.spanID, DefaultCollector),
		Transport: underlying,
	}
}

// GetRequestMetadata implements gRPC's credentials.Credentials
// interface.
func (t *span) GetRequestMetadata(ctx context.Context) (map[string]string, error) {
	return t.Metadata(), nil
}

// Metadata returns the gRPC metadata identifying this span to
// propagate it through a request tree.
func (t *span) Metadata() map[string]string {
	return map[string]string{parentSpanMDKey: t.spanID.String()}
}

func MiddlewareGRPC(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		// log.Println("Warning: no server context metadata")
		return ctx, nil
	}

	if s, ok := md[parentSpanMDKey]; ok && len(s) > 0 {
		parentSpan, err := appdash.ParseSpanID(s[0])
		if err != nil {
			return nil, err
		}
		ctx = NewContext(ctx, appdash.NewSpanID(*parentSpan))
	} else {
		// log.Println("Warning: no span ID set in server context")
	}

	return ctx, nil
}

// parentSpanMDKey is the gRPC metadata key for the appdash span.
const parentSpanMDKey = "x-appdash-parent-span-id"
