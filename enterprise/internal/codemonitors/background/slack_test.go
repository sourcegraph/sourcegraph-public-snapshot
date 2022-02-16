package background

import (
	"context"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	cmtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/types"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var update = flag.Bool("update", false, "Update goldenfiles of tests")

func TestSlackWebhook(t *testing.T) {
	t.Parallel()
	eu, err := url.Parse("https://sourcegraph.com")
	require.NoError(t, err)

	action := actionArgs{
		MonitorDescription: "My test monitor",
		MonitorOwnerName:   "Camden Cheek",
		ExternalURL:        eu,
		Query:              "repo:camdentest -file:id_rsa.pub BEGIN",
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
				Value: "summary line\n\nvery\nlong\nmessage\nbody\nwith\nmore\nthan\nten\nlines\nthat\nwill\nbe\ntruncated\n",
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
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", *update, b)
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
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", *update, b)
			w.WriteHeader(500)
		}))
		defer s.Close()

		client := s.Client()
		err := postSlackWebhook(context.Background(), client, s.URL, slackPayload(action))
		require.Error(t, err)
	})

	// If these tests fail, be sure to check that the changes are correct here:
	// https://app.slack.com/block-kit-builder/T02FSM7DL#%7B%22blocks%22:%5B%5D%7D
	t.Run("golden with results", func(t *testing.T) {
		actionCopy := action
		actionCopy.IncludeResults = true
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", *update, slackPayload(actionCopy))
	})

	t.Run("golden with truncated results", func(t *testing.T) {
		actionCopy := action
		actionCopy.IncludeResults = true
		// quadruple the number of results
		actionCopy.Results = append(actionCopy.Results, actionCopy.Results...)
		actionCopy.Results = append(actionCopy.Results, actionCopy.Results...)
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", *update, slackPayload(actionCopy))
	})

	t.Run("golden without results", func(t *testing.T) {
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", *update, slackPayload(action))
	})
}

func TestTriggerTestSlackWebhookAction(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", *update, b)
		w.WriteHeader(200)
	}))
	defer s.Close()

	client := s.Client()
	err := SendTestSlackWebhook(context.Background(), client, "My test monitor", s.URL)
	require.NoError(t, err)
}
