package background

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
)

func TestEmail(t *testing.T) {
	template := txemail.MustParseTemplate(newSearchResultsEmailTemplates)

	t.Run("test message", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:                  "New",
			CodeMonitorURL:            "https://sourcegraph.com/your/code/monitor",
			SearchURL:                 "https://sourcegraph.com/search",
			Description:               "My test monitor",
			NumberOfResultsWithDetail: "There was 1 new search result for your query",
			IsTest:                    true,
		}

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".html", *update, buf.String())
		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".txt", *update, buf.String())
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			require.Equal(t, "Test: [New event] My test monitor", buf.String())
		})
	})

	t.Run("one result", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:                  "New",
			CodeMonitorURL:            "https://sourcegraph.com/your/code/monitor",
			SearchURL:                 "https://sourcegraph.com/search",
			Description:               "My test monitor",
			NumberOfResultsWithDetail: "There was 1 new search result for your query",
		}

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".html", *update, buf.String())
		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".txt", *update, buf.String())
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			require.Equal(t, "[New event] My test monitor", buf.String())
		})
	})

	t.Run("multiple results", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:                  "New",
			CodeMonitorURL:            "https://sourcegraph.com/your/code/monitor",
			SearchURL:                 "https://sourcegraph.com/search",
			Description:               "My test monitor",
			NumberOfResultsWithDetail: "There was 1 new search result for your query",
		}

		t.Run("html", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Html.Execute(&buf, templateData)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".html", *update, buf.String())
		})

		t.Run("text", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Text.Execute(&buf, templateData)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".txt", *update, buf.String())
		})

		t.Run("subject", func(t *testing.T) {
			var buf bytes.Buffer
			err := template.Subj.Execute(&buf, templateData)
			require.NoError(t, err)
			require.Equal(t, "[New event] My test monitor", buf.String())
		})
	})
}
