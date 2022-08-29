package tracer

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"go.opentelemetry.io/otel/sdk/trace"
)

func Test_shouldTracePolicySampler(t *testing.T) {
	t.Run("with shouldTrace", func(t *testing.T) {
		s := shouldTracePolicySampler{}
		ctx := policy.WithShouldTrace(context.Background(), true)

		res := s.ShouldSample(trace.SamplingParameters{ParentContext: ctx})
		want := trace.RecordAndSample
		if got := res.Decision; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("without shouldTrace", func(t *testing.T) {
		s := shouldTracePolicySampler{}
		res := s.ShouldSample(trace.SamplingParameters{ParentContext: context.Background()})
		want := trace.Drop
		if got := res.Decision; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
