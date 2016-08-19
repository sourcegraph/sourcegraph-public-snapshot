package traceutil

import (
	"context"
	"fmt"
	"os"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc/metadata"

	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

// SpanURL returns the URL to the tracing UI for the given span. The span must be non-nil.
func SpanURL(span opentracing.Span) string {
	if spanCtx, ok := span.Context().(basictracer.SpanContext); ok {
		if project := os.Getenv("LIGHTSTEP_PROJECT"); project != "" {
			t := span.(basictracer.Span).Start().UnixNano() / 1000
			return fmt.Sprintf("https://app.lightstep.com/%s/trace?span_guid=%x&at_micros=%d#span-%x", project, spanCtx.SpanID, t, spanCtx.SpanID)
		}
		log15.Warn("LIGHTSTEP_PROJECT is not set")
	}
	return "#lightstep-not-enabled"
}

// InjectGRPCMetadata injects the span context into the GRPC metadata, which in turn is stored in the context.Context.
func InjectGRPCMetadata(ctx context.Context, spanCtx opentracing.SpanContext) context.Context {
	m, ok := metadata.FromContext(ctx)
	if !ok {
		m = make(metadata.MD)
	}
	if err := opentracing.GlobalTracer().Inject(spanCtx, opentracing.TextMap, metadataRW(m)); err != nil {
		log15.Error("injecting span context failed", "error", err)
		return ctx
	}
	return metadata.NewContext(ctx, m)
}

// ExtractGRPCMetadata extracts the span context from the GRPC metadata.
func ExtractGRPCMetadata(ctx context.Context) opentracing.SpanContext {
	m, _ := metadata.FromContext(ctx)
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, metadataRW(m))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		log15.Error("extracting span context failed", "error", err)
	}
	return spanCtx
}

type metadataRW metadata.MD

func (m metadataRW) Set(key, val string) {
	m[key] = []string{val}
}

func (m metadataRW) ForeachKey(handler func(key, val string) error) error {
	for k, v := range m {
		if err := handler(k, v[0]); err != nil {
			return err
		}
	}
	return nil
}
