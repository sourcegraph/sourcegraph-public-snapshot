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
}
