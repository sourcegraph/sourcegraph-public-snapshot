package tracer

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/schema"
)

type mockConfig struct {
	get func() Configuration
}

var _ ConfigurationSource = &mockConfig{}

func (m *mockConfig) Config() Configuration { return m.get() }

func TestConfigWatcher(t *testing.T) {
	var (
		ctx           = context.Background()
		logger        = logtest.Scoped(t)
		provider      = oteltracesdk.NewTracerProvider()
		debugMode     = &atomic.Bool{}
		noopProcessor = oteltracesdk.NewBatchSpanProcessor(tracetest.NewNoopExporter())
	)

	otelTracerProvider := newLoggedOtelTracerProvider(logger, provider, debugMode)
	// otelTracer represents a tracer a caller might hold. All tracers should be updated
	// by updating the underlying provider.
	otelTracer := otelTracerProvider.Tracer(t.Name())

	t.Run("tracing disabled", func(t *testing.T) {
		var updated bool
		doUpdate := newConfWatcher(
			logger,
			&mockConfig{
				get: func() Configuration {
					return Configuration{
						ObservabilityTracing: nil,
					}
				},
			},
			provider,
			func(logger log.Logger, opts options, debug bool) (oteltracesdk.SpanProcessor, error) {
				updated = true
				assert.Equal(t, opts.TracerType, None)
				assert.False(t, debug)
				return noopProcessor, nil
			},
			debugMode,
		)

		doUpdate()
		assert.True(t, updated)
		// should set global policy
		assert.Equal(t, policy.TraceNone, policy.GetTracePolicy())
	})

	t.Run("enable tracing with 'observability.tracing: {}'", func(t *testing.T) {
		mockConfig := &mockConfig{
			get: func() Configuration {
				return Configuration{
					ObservabilityTracing: &schema.ObservabilityTracing{},
				}
			},
		}

		var updated bool
		expectTracerType := DefaultTracerType
		spansRecorder := tracetest.NewSpanRecorder()
		doUpdate := newConfWatcher(
			logger,
			mockConfig,
			provider,
			func(logger log.Logger, opts options, debug bool) (oteltracesdk.SpanProcessor, error) {
				// must be set to default
				updated = assert.Equal(t, opts.TracerType, expectTracerType)
				assert.False(t, debug)
				if opts.TracerType == "none" {
					return noopProcessor, nil
				}
				return spansRecorder, nil
			},
			debugMode,
		)

		// fetch updated conf
		doUpdate()
		assert.True(t, updated)

		// should update global policy
		assert.Equal(t, policy.TraceSelective, policy.GetTracePolicy())

		// span recorder must be registered, and spans from both tracers must go to it
		var spanCount int
		t.Run("otel tracer spans go to new processor", func(t *testing.T) {
			_, span := otelTracer.Start(policy.WithShouldTrace(ctx, true), "foo")
			span.End()
			spanCount++
			assert.Len(t, spansRecorder.Ended(), spanCount)
		})
		t.Run("otel tracerprovider new tracers go to new processor", func(t *testing.T) {
			_, span := otelTracerProvider.Tracer(t.Name()).
				Start(policy.WithShouldTrace(ctx, true), "bar")
			span.End()
			spanCount++
			assert.Len(t, spansRecorder.Ended(), spanCount)
		})

		t.Run("disable tracing after enabling it", func(t *testing.T) {
			mockConfig.get = func() Configuration {
				return Configuration{
					ObservabilityTracing: &schema.ObservabilityTracing{Sampling: "none"},
				}
			}
			expectTracerType = "none"

			// fetch updated conf
			doUpdate()

			// no new spans should register
			t.Run("otel tracer spans not go to processor", func(t *testing.T) {
				_, span := otelTracer.Start(policy.WithShouldTrace(ctx, true), "foo")
				span.End()
				assert.Len(t, spansRecorder.Ended(), spanCount)
			})
			t.Run("otel tracerprovider not go to processor", func(t *testing.T) {
				_, span := otelTracerProvider.Tracer(t.Name()).
					Start(policy.WithShouldTrace(ctx, true), "bar")
				span.End()
				assert.Len(t, spansRecorder.Ended(), spanCount)
			})
		})
	})

	t.Run("update tracing with debug and sampling all", func(t *testing.T) {
		mockConf := &mockConfig{
			get: func() Configuration {
				return Configuration{
					ObservabilityTracing: &schema.ObservabilityTracing{
						Debug:    true,
						Sampling: "all",
					},
				}
			},
		}
		spansRecorder1 := tracetest.NewSpanRecorder()
		updatedSpanProcessor := spansRecorder1
		doUpdate := newConfWatcher(
			logger,
			mockConf,
			provider,
			func(logger log.Logger, opts options, debug bool) (oteltracesdk.SpanProcessor, error) {
				return updatedSpanProcessor, nil
			},
			debugMode,
		)

		// fetch updated conf
		doUpdate()

		// span recorder must be registered, and spans from both tracers must go to it
		var spanCount1 int
		{
			_, span := otelTracer.Start(ctx, "foo") // does not need ShouldTrace due to policy
			span.End()
			spanCount1++
			assert.Len(t, spansRecorder1.Ended(), spanCount1)
		}

		// should have debug set
		assert.True(t, otelTracerProvider.debug.Load())

		// should set global policy
		assert.Equal(t, policy.TraceAll, policy.GetTracePolicy())

		t.Run("sanity check - swap existing processor with another", func(t *testing.T) {
			spansRecorder2 := tracetest.NewSpanRecorder()
			updatedSpanProcessor = spansRecorder2
			mockConf.get = func() Configuration {
				return Configuration{
					ObservabilityTracing: &schema.ObservabilityTracing{
						Debug:    true,
						Sampling: "all",
					},
				}
			}

			// fetch updated conf
			doUpdate()

			// span recorder must be registered, and spans from both tracers must go to it
			var spanCount2 int
			{
				_, span := otelTracer.Start(ctx, "foo")
				span.End()
				spanCount2++
				assert.Len(t, spansRecorder2.Ended(), spanCount2)
			}

			// old span recorder gets no more spans, because it should be removed
			assert.Len(t, spansRecorder1.Ended(), spanCount1)
		})
	})
}
