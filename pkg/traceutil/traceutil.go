// Package traceutil contains appdash-related utilities.
package traceutil

import (
	"log"
	"os"

	"context"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil/appdashctx"
)

var logger = log.New(os.Stderr, "appdash: ", log.LstdFlags)

func NewRecorder(span appdash.SpanID, c appdash.Collector) *appdash.Recorder {
	rec := appdash.NewRecorder(span, c)
	rec.Logger = logger
	return rec
}

// Recorder creates a new appdash Recorder for an existing span.
func Recorder(ctx context.Context) *appdash.Recorder {
	c := appdashctx.Collector(ctx)
	if c == nil {
		c = discardCollector{}
	}

	span := SpanIDFromContext(ctx)
	if span.Trace == 0 {
		// log.Println("no trace set in context")
	}
	return NewRecorder(span, c)
}

// DefaultCollector is the default Appdash collector to use. It is
// legacy and should not be used for additional things beyond the
// existing uses.
//
// TODO(sqs): remove this and make callers fetch the collector from
// the context instead of using a global here.
var DefaultCollector appdash.Collector

type discardCollector struct{}

func (discardCollector) Collect(appdash.SpanID, ...appdash.Annotation) error {
	return nil
}
