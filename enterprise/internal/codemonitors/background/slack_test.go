package background

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	cmtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/types"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestSlackWebhook(t *testing.T) {
	t.Parallel()
	action := actionArgs{
		MonitorDescription: "My test monitor",
		MonitorURL:         "https://google.com",
		Query:              "repo:camdentest -file:id_rsa.pub BEGIN",
		QueryURL:           "https://youtube.com",
		Results: cmtypes.CommitSearchResults{{
			Commit: cmtypes.Commit{
				Repository: cmtypes.Repository{Name: "github.com/test/test"},
				Oid:        "7815187511872asbasdfgasd",
			},
			DiffPreview: &cmtypes.HighlightedString{
				Value: "file1.go file2.go\n@ -97,5 +97,5 @ func Test() {\n leading context\n+matched added\n-matched removed\n trailing context\n",
				Highlights: []cmtypes.Highlight{{
					Line:      3,
					Character: 1,
					Length:    7,
				}, {
					Line:      4,
					Character: 1,
					Length:    7,
				}},
			},
		}, {
			Commit: cmtypes.Commit{
				Repository: cmtypes.Repository{Name: "github.com/test/test"},
				Oid:        "7815187511872asbasdfgasd",
			},
			MessagePreview: &cmtypes.HighlightedString{
				Value: "summary line\n\nmessage body\n",
				Highlights: []cmtypes.Highlight{{
					Line:      1,
					Character: 0,
					Length:    7,
				}},
			},
		}},
		IncludeResults: false,
	}

	t.Run("no error", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", true, b)
			w.WriteHeader(200)
		}))
		defer s.Close()

		client := s.Client()
		err := postSlackWebhook(context.Background(), client, s.URL, slackPayload(action))
		require.NoError(t, err)
	})

	t.Run("error is returned", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", true, b)
			w.WriteHeader(500)
		}))
		defer s.Close()

		client := s.Client()
		err := postSlackWebhook(context.Background(), client, s.URL, slackPayload(action))
		require.Error(t, err)
	})

	t.Run("golden with results", func(t *testing.T) {
		actionCopy := action
		actionCopy.IncludeResults = true
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", true, slackPayload(actionCopy))
	})

	t.Run("golden with truncated results", func(t *testing.T) {
		actionCopy := action
		actionCopy.IncludeResults = true
		// quadruple the number of results
		actionCopy.Results = append(actionCopy.Results, actionCopy.Results...)
		actionCopy.Results = append(actionCopy.Results, actionCopy.Results...)
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", true, slackPayload(actionCopy))
	})

	t.Run("golden without results", func(t *testing.T) {
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", true, slackPayload(action))
	})
}
