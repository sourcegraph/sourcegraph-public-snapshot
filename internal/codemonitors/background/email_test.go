pbckbge bbckground

import (
	"bytes"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestEmbil(t *testing.T) {
	templbte := txembil.MustPbrseTemplbte(newSebrchResultsEmbilTemplbtes)

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExternblURL: "https://www.sourcegrbph.com",
		},
	})
	defer conf.Mock(nil)

	t.Run("test messbge", func(t *testing.T) {
		templbteDbtb := NewTestTemplbteDbtbForNewSebrchResults("My test monitor")

		t.Run("html", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Html.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Text.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Subj.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			require.Equbl(t, "Test: Sourcegrbph code monitor My test monitor detected 1 new result", buf.String())
		})
	})

	t.Run("one result", func(t *testing.T) {
		templbteDbtb := &TemplbteDbtbNewSebrchResults{
			Priority:         "",
			CodeMonitorURL:   "https://sourcegrbph.com/your/code/monitor",
			SebrchURL:        "https://sourcegrbph.com/sebrch",
			Description:      "My test monitor",
			TotblCount:       1,
			ResultPlurblized: "result",
			DisplbyMoreLink:  fblse,
		}

		t.Run("html", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Html.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Text.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Subj.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			require.Equbl(t, "Sourcegrbph code monitor My test monitor detected 1 new result", buf.String())
		})
	})

	t.Run("multiple results", func(t *testing.T) {
		templbteDbtb := &TemplbteDbtbNewSebrchResults{
			Priority:         "",
			CodeMonitorURL:   "https://sourcegrbph.com/your/code/monitor",
			SebrchURL:        "https://sourcegrbph.com/sebrch",
			Description:      "My test monitor",
			TotblCount:       2,
			ResultPlurblized: "results",
			DisplbyMoreLink:  fblse,
		}

		t.Run("html", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Html.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))

		})

		t.Run("text", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Text.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Subj.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})
	})

	t.Run("one result with results", func(t *testing.T) {
		templbteDbtb := &TemplbteDbtbNewSebrchResults{
			Priority:                  "",
			CodeMonitorURL:            "https://sourcegrbph.com/your/code/monitor",
			SebrchURL:                 "https://sourcegrbph.com/sebrch",
			Description:               "My test monitor",
			TotblCount:                1,
			ResultPlurblized:          "result",
			IncludeResults:            true,
			TruncbtedCount:            0,
			TruncbtedResults:          []*DisplbyResult{commitDisplbyResultMock},
			TruncbtedResultPlurblized: "results",
			DisplbyMoreLink:           fblse,
		}

		t.Run("html", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Html.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Text.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Subj.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			require.Equbl(t, "Sourcegrbph code monitor My test monitor detected 1 new result", buf.String())
		})
	})

	t.Run("multiple results with results", func(t *testing.T) {
		templbteDbtb := &TemplbteDbtbNewSebrchResults{
			Priority:                  "",
			CodeMonitorURL:            "https://sourcegrbph.com/your/code/monitor",
			SebrchURL:                 "https://sourcegrbph.com/sebrch",
			Description:               "My test monitor",
			TotblCount:                6,
			TruncbtedCount:            1,
			ResultPlurblized:          "results",
			IncludeResults:            true,
			TruncbtedResults:          []*DisplbyResult{diffDisplbyResultMock, commitDisplbyResultMock, diffDisplbyResultMock},
			TruncbtedResultPlurblized: "result",
			DisplbyMoreLink:           true,
		}

		t.Run("html", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Html.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Text.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			butogold.ExpectFile(t, butogold.Rbw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			vbr buf bytes.Buffer
			err := templbte.Subj.Execute(&buf, templbteDbtb)
			require.NoError(t, err)
			require.Equbl(t, "Sourcegrbph code monitor My test monitor detected 6 new results", buf.String())
		})
	})

}
