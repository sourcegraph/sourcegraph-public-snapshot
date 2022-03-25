package background

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
)

func TestEmail(t *testing.T) {
	template := txemail.MustParseTemplate(newSearchResultsEmailTemplates)

	t.Run("test message", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:         "",
			CodeMonitorURL:   "https://sourcegraph.com/your/code/monitor",
			SearchURL:        "https://sourcegraph.com/search",
			Description:      "My test monitor",
			TotalCount:       1,
			IsTest:           true,
			ResultPluralized: "result",
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
			require.Equal(t, "Sourcegraph code monitor My test monitor detected 2 new results", buf.String())
		})
	})

	t.Run("one result with results", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:         "",
			CodeMonitorURL:   "https://sourcegraph.com/your/code/monitor",
			SearchURL:        "https://sourcegraph.com/search",
			Description:      "My test monitor",
			TotalCount:       1,
			ResultPluralized: "result",
			IncludeResults: true,
			TruncatedCount: 0,
			TruncatedResults: []*result.CommitMatch{&commitResultMock},
			TruncatedResultPluralized: "result",
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
			require.Equal(t, "Sourcegraph code monitor My test monitor detected 1 new result", buf.String())
		})
	})

	t.Run("multiple results with results", func(t *testing.T) {
		templateData := &TemplateDataNewSearchResults{
			Priority:         "",
			CodeMonitorURL:   "https://sourcegraph.com/your/code/monitor",
			SearchURL:        "https://sourcegraph.com/search",
			Description:      "My test monitor",
			TotalCount:       6,
			TruncatedCount: 1,
			ResultPluralized: "results",
			IncludeResults: true,
			TruncatedResults: []*result.CommitMatch{&diffResultMock, &commitResultMock, &diffResultMock},
			TruncatedResultPluralized: "result",
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
			require.Equal(t, "Sourcegraph code monitor My test monitor detected 6 new results", buf.String())
		})
	})

}
