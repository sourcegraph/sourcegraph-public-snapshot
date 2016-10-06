package traceutil

import (
	"fmt"
	"os"

	log15 "gopkg.in/inconshreveable/log15.v2"

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
