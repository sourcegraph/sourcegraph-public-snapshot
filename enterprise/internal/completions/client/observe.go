package client

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"go.opentelemetry.io/otel/attribute"
)

func newObservedClient(inner types.CompletionsClient) *observedClient {
	observationCtx := observation.NewContext(log.Scoped("completions", "completions client"))
	ops := newOperations(observationCtx)
	return &observedClient{
		inner: inner,
		ops:   ops,
	}
}

type observedClient struct {
	inner types.CompletionsClient
	ops   *operations
}

var _ types.CompletionsClient = (*observedClient)(nil)

func (o *observedClient) Stream(ctx context.Context, feature types.CompletionsFeature, params types.CompletionRequestParameters, send types.SendCompletionEvent) (err error) {
	ctx, tr, endObservation := o.ops.stream.With(ctx, &err, observation.Args{
		Attrs:             append(params.Attrs(), attribute.String("feature", string(feature))),
		MetricLabelValues: []string{params.Model},
	})
	defer endObservation(1, observation.Args{})

	tracedSend := func(event types.CompletionResponse) error {
		if event.StopReason != "" {
			tr.AddEvent("stopped", attribute.String("reason", event.StopReason))
		} else {
			tr.AddEvent("completion", attribute.Int("len", len(event.Completion)))
		}
		return send(event)
	}

	return o.inner.Stream(ctx, feature, params, tracedSend)
}

func (o *observedClient) Complete(ctx context.Context, feature types.CompletionsFeature, params types.CompletionRequestParameters) (resp *types.CompletionResponse, err error) {
	ctx, _, endObservation := o.ops.stream.With(ctx, &err, observation.Args{
		Attrs:             append(params.Attrs(), attribute.String("feature", string(feature))),
		MetricLabelValues: []string{params.Model},
	})
	defer endObservation(1, observation.Args{})

	return o.inner.Complete(ctx, feature, params)
}

type operations struct {
	stream   *observation.Operation
	complete *observation.Operation
}

var (
	streamMetrics = metrics.NewREDMetrics(
		prometheus.DefaultRegisterer,
		"completions_stream",
		metrics.WithLabels("model"),
	)
	completeMetrics = metrics.NewREDMetrics(
		prometheus.DefaultRegisterer,
		"completions_complete",
		metrics.WithLabels("model"),
	)
)

func newOperations(observationCtx *observation.Context) *operations {
	streamOp := observationCtx.Operation(observation.Op{
		Metrics: streamMetrics,
		Name:    "completions.stream",
	})
	completeOp := observationCtx.Operation(observation.Op{
		Metrics: streamMetrics,
		Name:    "completions.complete",
	})
	return &operations{
		stream:   streamOp,
		complete: completeOp,
	}
}
