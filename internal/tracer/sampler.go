package tracer

import (
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// shouldTracePolicySampler checks policy.GetTracePolicy to determine if it should drop or
// keep all traces, or check the context associated with a trace with policy.ShouldTrace
// to determine whether or not to drop the trace or keep it.
type shouldTracePolicySampler struct{}

var _ oteltracesdk.Sampler = &shouldTracePolicySampler{}

var (
	alwaysSample = oteltracesdk.AlwaysSample()
	neverSample  = oteltracesdk.NeverSample()
)

func (s *shouldTracePolicySampler) ShouldSample(parameters oteltracesdk.SamplingParameters) oteltracesdk.SamplingResult {
	switch policy.GetTracePolicy() {
	// Keep nothing
	case policy.TraceNone:
		return neverSample.ShouldSample(parameters)

	// Keep everything
	case policy.TraceAll:
		return alwaysSample.ShouldSample(parameters)

	// Keep only traces that are specifically requested
	default:
		if policy.ShouldTrace(parameters.ParentContext) {
			return alwaysSample.ShouldSample(parameters)
		}
		return neverSample.ShouldSample(parameters)
	}
}

func (s *shouldTracePolicySampler) Description() string { return "shouldTracePolicySampler" }
