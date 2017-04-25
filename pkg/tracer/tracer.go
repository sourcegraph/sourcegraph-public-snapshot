package tracer

import (
	"fmt"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

type tagsAndLogsFilter struct {
	rec basictracer.SpanRecorder
}

func (f *tagsAndLogsFilter) RecordSpan(span basictracer.RawSpan) {
	span.Tags = nil
	span.Logs = nil
	f.rec.RecordSpan(span)
}

var lightstepAccessToken = env.Get("LIGHTSTEP_ACCESS_TOKEN", "", "access token for sending traces to LightStep")
var lightstepProject = env.Get("LIGHTSTEP_PROJECT", "", "the project id on LightStep, only used for creating links to traces")
var lightstepIncludeSensitive, _ = strconv.ParseBool(env.Get("LIGHTSTEP_INCLUDE_SENSITIVE", "", "send span tags and logs to LightStep"))

func Init() {
	if lightstepAccessToken != "" {
		var rec basictracer.SpanRecorder = lightstep.NewRecorder(lightstep.Options{
			AccessToken: lightstepAccessToken,
			UseGRPC:     true,
		})
		if !lightstepIncludeSensitive {
			rec = &tagsAndLogsFilter{rec}
		}

		options := basictracer.DefaultOptions()
		options.ShouldSample = func(_ uint64) bool { return true }
		options.Recorder = rec
		opentracing.InitGlobalTracer(basictracer.NewWithOptions(options))

		traceutil.SpanURL = spanURL
	}
}

func spanURL(span opentracing.Span) string {
	spanCtx := span.Context().(basictracer.SpanContext)
	t := span.(basictracer.Span).Start().UnixNano() / 1000
	return fmt.Sprintf("https://app.lightstep.com/%s/trace?span_guid=%x&at_micros=%d#span-%x", lightstepProject, spanCtx.SpanID, t, spanCtx.SpanID)
}
