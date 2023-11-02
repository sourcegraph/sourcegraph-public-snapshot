package tracer

import (
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

var (
	// Use upstream samplers to ensure we return the right thing in our
	// custom Sampler implementation.
	alwaysSampleSampler = oteltracesdk.AlwaysSample()
	neverSampleSampler  = oteltracesdk.NeverSample()
)

// tracePolicySampler implements the oteltrace.Sampler interface and indicates
// whether a trace should be sampled or not based on the global trace policy
// and comparing it against the policy indicated in the parent context where
// relevant.
type tracePolicySampler struct{}

var _ oteltracesdk.Sampler = tracePolicySampler{}

func (tracePolicySampler) ShouldSample(p oteltracesdk.SamplingParameters) oteltracesdk.SamplingResult {
	switch policy.GetTracePolicy() {
	case policy.TraceAll:
		// Retain and export all events.
		return alwaysSampleSampler.ShouldSample(p)

	case policy.TraceNone:
		// Drop all events.
		return neverSampleSampler.ShouldSample(p)

	default:
		// By default, enforce policy.TraceSelective, which means that we only
		// sample if the parent context is marked for tracing.
		if policy.ShouldTrace(p.ParentContext) {
			return alwaysSampleSampler.ShouldSample(p)
		}
	}

	// Otherwise, indicate this span should be dropped and not exported.
	return neverSampleSampler.ShouldSample(p)
}

func (tracePolicySampler) Description() string { return "internal/tracer.tracePolicySampler" }
