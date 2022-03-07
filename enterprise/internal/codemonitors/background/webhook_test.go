package background

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestWebhook(t *testing.T) {
	eu, err := url.Parse("https://sourcegraph.com")
	require.NoError(t, err)

	t.Run("no error", func(t *testing.T) {
		action := actionArgs{
			MonitorDescription: "My test monitor",
			ExternalURL:        eu,
			MonitorID:          42,
			Query:              "repo:camdentest -file:id_rsa.pub BEGIN",
			Results:            make([]*result.CommitMatch, 3),
			IncludeResults:     false,
		}

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", true, b)
			w.WriteHeader(200)
		}))
		defer s.Close()

		client := s.Client()
		err := postWebhook(context.Background(), client, s.URL, action)
		require.NoError(t, err)
	})

	t.Run("error is returned", func(t *testing.T) {
		action := actionArgs{
			MonitorDescription: "My test monitor",
			ExternalURL:        eu,
			Query:              "repo:camdentest -file:id_rsa.pub BEGIN",
			Results:            make([]*result.CommitMatch, 3),
			IncludeResults:     false,
		}

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", true, b)
			w.WriteHeader(500)
		}))
		defer s.Close()

		client := s.Client()
		err := postWebhook(context.Background(), client, s.URL, action)
		require.Error(t, err)
	})
}

func TestTriggerTestWebhookAction(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		testutil.AssertGolden(t, "testdata/"+t.Name()+".json", *update, b)
		w.WriteHeader(200)
	}))
	defer s.Close()

	client := s.Client()
	err := SendTestWebhook(context.Background(), client, "My test monitor", s.URL)
	require.NoError(t, err)
}
