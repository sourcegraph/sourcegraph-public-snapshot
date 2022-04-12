package background

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/txemail"
)

func TestEmail(t *testing.T) {
	template := txemail.MustParseTemplate(newSearchResultsEmailTemplates)

	MockExternalURL = func() *url.URL {
		externalURL, _ := url.Parse("https://www.sourcegraph.com")
		return externalURL
	}

	t.Run("test message", func(t *testing.T) {
		templateData := NewTestTemplateDataForNewSearchResults("My test monitor")

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			require.Equal(t, "Test: Sourcegraph code monitor My test monitor detected 1 new result", buf.String())
		})
	})

	t.Run("one result", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:         "",
			CodeMonitorURL:   "https://sourcegraph.com/your/code/monitor",
			SearchURL:        "https://sourcegraph.com/search",
			Description:      "My test monitor",
			TotalCount:       1,
			ResultPluralized: "result",
			DisplayMoreLink:  false,
		}

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			require.Equal(t, "Sourcegraph code monitor My test monitor detected 1 new result", buf.String())
		})
	})

	t.Run("multiple results", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:         "",
			CodeMonitorURL:   "https://sourcegraph.com/your/code/monitor",
			SearchURL:        "https://sourcegraph.com/search",
			Description:      "My test monitor",
			TotalCount:       2,
			ResultPluralized: "results",
			DisplayMoreLink:  false,
		}

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))

		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})
	})

	t.Run("one result with results", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:                  "",
			CodeMonitorURL:            "https://sourcegraph.com/your/code/monitor",
			SearchURL:                 "https://sourcegraph.com/search",
			Description:               "My test monitor",
			TotalCount:                1,
			ResultPluralized:          "result",
			IncludeResults:            true,
			TruncatedCount:            0,
			TruncatedResults:          []*DisplayResult{commitDisplayResultMock},
			TruncatedResultPluralized: "results",
			DisplayMoreLink:           false,
		}

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			require.Equal(t, "Sourcegraph code monitor My test monitor detected 1 new result", buf.String())
		})
	})

	t.Run("multiple results with results", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:                  "",
			CodeMonitorURL:            "https://sourcegraph.com/your/code/monitor",
			SearchURL:                 "https://sourcegraph.com/search",
			Description:               "My test monitor",
			TotalCount:                6,
			TruncatedCount:            1,
			ResultPluralized:          "results",
			IncludeResults:            true,
			TruncatedResults:          []*DisplayResult{diffDisplayResultMock, commitDisplayResultMock, diffDisplayResultMock},
			TruncatedResultPluralized: "result",
			DisplayMoreLink:           true,
		}

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			autogold.Equal(t, autogold.Raw(buf.String()))
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			require.Equal(t, "Sourcegraph code monitor My test monitor detected 6 new results", buf.String())
		})
	})

}
