package tracer

import (
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

type mockConfig struct {
	get func() schema.SiteConfiguration
}

var _ conftypes.SiteConfigQuerier = &mockConfig{}

func (m *mockConfig) SiteConfig() schema.SiteConfiguration { return m.get() }

func TestConfigWatcher(t *testing.T) {
	var (
		logger         = logtest.Scoped(t)
		switchableOTel = newSwitchableOtelTracerProvider(logger.Scoped("otel", ""))
		switchableOT   = newSwitchableOTTracer(logger.Scoped("ot", ""))
		confQuerier    = &mockConfig{}
	)

	update := newConfWatcher(
		logger,
		confQuerier,
		switchableOTel,
		switchableOT,
		options{},
	)

	t.Run("tracing disabled", func(t *testing.T) {
		confQuerier.get = func() schema.SiteConfiguration {
			return schema.SiteConfiguration{}
		}

		update()

		// should all be no-op
		assert.Equal(t, oteltrace.NewNoopTracerProvider().Tracer(""), switchableOTel.Tracer(""))
		assert.Equal(t, opentracing.NoopTracer{}, switchableOT.tracer)
	})

	t.Run("enable tracing with 'observability.tracing: {}'", func(t *testing.T) {
		confQuerier.get = func() schema.SiteConfiguration {
			return schema.SiteConfiguration{
				ObservabilityTracing: &schema.ObservabilityTracing{},
			}
		}

		// fetch updated conf
		update()

		// should not be no-op
		assert.NotEqual(t, oteltrace.NewNoopTracerProvider().Tracer(""), switchableOTel.Tracer(""))
		assert.NotEqual(t, opentracing.NoopTracer{}, switchableOT.tracer)

		// should have debug set to false
		assert.False(t, switchableOTel.current.Load().(*otelTracerProviderCarrier).debug)
		assert.False(t, switchableOT.debug)
	})

	t.Run("update tracing with debug", func(t *testing.T) {
		confQuerier.get = func() schema.SiteConfiguration {
			return schema.SiteConfiguration{
				ObservabilityTracing: &schema.ObservabilityTracing{
					Debug: true,
				},
			}
		}

		// fetch updated conf
		update()

		// should not be no-op
		assert.NotEqual(t, oteltrace.NewNoopTracerProvider().Tracer(""), switchableOTel.Tracer(""))
		assert.NotEqual(t, opentracing.NoopTracer{}, switchableOT.tracer)

		// should have debug set
		assert.True(t, switchableOTel.current.Load().(*otelTracerProviderCarrier).debug)
		assert.True(t, switchableOT.debug)
	})
}
