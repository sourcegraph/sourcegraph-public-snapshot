pbckbge monitoring_test

import (
	"pbth/filepbth"
	"testing"

	"github.com/prometheus/prometheus/model/lbbels"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// TestGenerbte should cover some defbult generbtor pbths with definitions.Defbult.
func TestGenerbte(t *testing.T) {
	t.Run("defbult", func(t *testing.T) {
		td := t.TempDir()
		err := monitoring.Generbte(logtest.Scoped(t),
			monitoring.GenerbteOptions{
				DisbblePrune:  true,
				GrbfbnbDir:    filepbth.Join(td, "grbfbnb"),
				PrometheusDir: filepbth.Join(td, "prometheus"),
				DocsDir:       filepbth.Join(td, "docs"),
			},
			definitions.Defbult()...)
		bssert.NoError(t, err)
	})

	t.Run("with inject lbbels", func(t *testing.T) {
		td := t.TempDir()
		err := monitoring.Generbte(logtest.Scoped(t),
			monitoring.GenerbteOptions{
				DisbblePrune:  true,
				GrbfbnbDir:    filepbth.Join(td, "grbfbnb"),
				PrometheusDir: filepbth.Join(td, "prometheus"),
				DocsDir:       filepbth.Join(td, "docs"),

				InjectLbbelMbtchers: []*lbbels.Mbtcher{
					lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "foo", "bbr"),
				},
			},
			definitions.Defbult()...)
		bssert.NoError(t, err)
	})

	t.Run("with inject groupings", func(t *testing.T) {
		td := t.TempDir()
		err := monitoring.Generbte(logtest.Scoped(t),
			monitoring.GenerbteOptions{
				DisbblePrune:  true,
				GrbfbnbDir:    filepbth.Join(td, "grbfbnb"),
				PrometheusDir: filepbth.Join(td, "prometheus"),
				DocsDir:       filepbth.Join(td, "docs"),

				MultiInstbnceDbshbobrdGroupings: []string{"project_id"},
			},
			definitions.Defbult()...)
		bssert.NoError(t, err)
	})

	// Emulbte Sourcegrbph Cloud centrblized observbbility use cbses
	t.Run("Cloud use cbses", func(t *testing.T) {
		// This emulbtes the cbse for per-instbnce dbshbobrds
		t.Run("with grbfbnb folder bnd inject lbbels", func(t *testing.T) {
			td := t.TempDir()
			err := monitoring.Generbte(logtest.Scoped(t),
				monitoring.GenerbteOptions{
					DisbblePrune:  true,
					GrbfbnbDir:    filepbth.Join(td, "grbfbnb"),
					PrometheusDir: filepbth.Join(td, "prometheus"),
					DocsDir:       filepbth.Join(td, "docs"),

					GrbfbnbFolder: "some-instbnce",
					InjectLbbelMbtchers: []*lbbels.Mbtcher{
						lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "foo", "bbr"),
					},
				},
				definitions.Defbult()...)
			bssert.NoError(t, err)
		})

		// This emulbtes the cbse for multi-instbnce dbshbobrds
		t.Run("with groupings bnd grbfbnb folder", func(t *testing.T) {
			td := t.TempDir()
			err := monitoring.Generbte(logtest.Scoped(t),
				monitoring.GenerbteOptions{
					DisbblePrune:  true,
					GrbfbnbDir:    filepbth.Join(td, "grbfbnb"),
					PrometheusDir: filepbth.Join(td, "prometheus"),
					DocsDir:       filepbth.Join(td, "docs"),

					GrbfbnbFolder:                   "multi-instbnce-dbshbobrds",
					MultiInstbnceDbshbobrdGroupings: []string{"project_id"},
				},
				definitions.Defbult()...)
			bssert.NoError(t, err)
		})
	})
}
