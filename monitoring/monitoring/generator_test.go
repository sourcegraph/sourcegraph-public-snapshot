package monitoring_test

import (
	"path/filepath"
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// TestGenerate should cover some default generator paths with definitions.Default.
func TestGenerate(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		td := t.TempDir()
		err := monitoring.Generate(logtest.Scoped(t),
			monitoring.GenerateOptions{
				DisablePrune:  true,
				GrafanaDir:    filepath.Join(td, "grafana"),
				PrometheusDir: filepath.Join(td, "prometheus"),
				DocsDir:       filepath.Join(td, "docs"),
			},
			definitions.Default()...)
		assert.NoError(t, err)
	})

	t.Run("with inject labels", func(t *testing.T) {
		td := t.TempDir()
		err := monitoring.Generate(logtest.Scoped(t),
			monitoring.GenerateOptions{
				DisablePrune:  true,
				GrafanaDir:    filepath.Join(td, "grafana"),
				PrometheusDir: filepath.Join(td, "prometheus"),
				DocsDir:       filepath.Join(td, "docs"),

				InjectLabelMatchers: []*labels.Matcher{
					labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
				},
			},
			definitions.Default()...)
		assert.NoError(t, err)
	})

	t.Run("with inject groupings", func(t *testing.T) {
		td := t.TempDir()
		err := monitoring.Generate(logtest.Scoped(t),
			monitoring.GenerateOptions{
				DisablePrune:  true,
				GrafanaDir:    filepath.Join(td, "grafana"),
				PrometheusDir: filepath.Join(td, "prometheus"),
				DocsDir:       filepath.Join(td, "docs"),

				MultiInstanceDashboardGroupings: []string{"project_id"},
			},
			definitions.Default()...)
		assert.NoError(t, err)
	})

	// Emulate Sourcegraph Cloud centralized observability use cases
	t.Run("Cloud use cases", func(t *testing.T) {
		// This emulates the case for per-instance dashboards
		t.Run("with grafana folder and inject labels", func(t *testing.T) {
			td := t.TempDir()
			err := monitoring.Generate(logtest.Scoped(t),
				monitoring.GenerateOptions{
					DisablePrune:  true,
					GrafanaDir:    filepath.Join(td, "grafana"),
					PrometheusDir: filepath.Join(td, "prometheus"),
					DocsDir:       filepath.Join(td, "docs"),

					GrafanaFolder: "some-instance",
					InjectLabelMatchers: []*labels.Matcher{
						labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
					},
				},
				definitions.Default()...)
			assert.NoError(t, err)
		})

		// This emulates the case for multi-instance dashboards
		t.Run("with groupings and grafana folder", func(t *testing.T) {
			td := t.TempDir()
			err := monitoring.Generate(logtest.Scoped(t),
				monitoring.GenerateOptions{
					DisablePrune:  true,
					GrafanaDir:    filepath.Join(td, "grafana"),
					PrometheusDir: filepath.Join(td, "prometheus"),
					DocsDir:       filepath.Join(td, "docs"),

					GrafanaFolder:                   "multi-instance-dashboards",
					MultiInstanceDashboardGroupings: []string{"project_id"},
				},
				definitions.Default()...)
			assert.NoError(t, err)
		})
	})
}
