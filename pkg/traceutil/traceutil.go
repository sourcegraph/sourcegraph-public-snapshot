package traceutil

import (
	"fmt"
	"os"
	"strconv"

	log15 "gopkg.in/inconshreveable/log15.v2"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	"github.com/neelance/graphql-go"
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
)

var graphqlFieldHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "graphql",
	Name:      "field_seconds",
	Help:      "GraphQL field resolver latencies in seconds.",
	Buckets:   []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
}, []string{"type", "field", "error"})

func init() {
	prometheus.MustRegister(graphqlFieldHistogram)
}

type graphqlFieldRecorder struct {
}

func (r *graphqlFieldRecorder) RecordSpan(span basictracer.RawSpan) {
	if field, _ := span.Tags[graphql.OpenTracingTagField].(string); field != "" {
		typ, _ := span.Tags[graphql.OpenTracingTagType].(string)
		err, _ := span.Tags[graphql.OpenTracingTagError].(string)
		graphqlFieldHistogram.WithLabelValues(typ, field, strconv.FormatBool(err != "")).Observe(span.Duration.Seconds())
	}
}

type trivialFieldsFilter struct {
	rec basictracer.SpanRecorder
}

func (f *trivialFieldsFilter) RecordSpan(span basictracer.RawSpan) {
	if b, ok := span.Tags[graphql.OpenTracingTagTrivial].(bool); ok && b {
		return
	}
	f.rec.RecordSpan(span)
}

type multiRecorder []basictracer.SpanRecorder

func (mr multiRecorder) RecordSpan(span basictracer.RawSpan) {
	for _, r := range mr {
		r.RecordSpan(span)
	}
}

// MultiRecorder creates a recorder that duplicates its writes to all the provided recorders.
func MultiRecorder(recorders ...basictracer.SpanRecorder) basictracer.SpanRecorder {
	mr := make(multiRecorder, len(recorders))
	copy(mr, recorders)
	return mr
}

func InitTracer() {
	if t := os.Getenv("LIGHTSTEP_ACCESS_TOKEN"); t != "" {
		options := basictracer.DefaultOptions()
		options.ShouldSample = func(_ uint64) bool { return true }
		options.Recorder = MultiRecorder(
			&trivialFieldsFilter{lightstep.NewRecorder(lightstep.Options{
				AccessToken: t,
			})},
			&graphqlFieldRecorder{},
		)
		opentracing.InitGlobalTracer(basictracer.NewWithOptions(options))
	}
}

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
