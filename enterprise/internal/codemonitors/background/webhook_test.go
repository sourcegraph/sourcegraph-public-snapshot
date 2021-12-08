package background

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestWebhook(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		action := actionArgs{
			MonitorDescription: "My test monitor", MonitorURL: "https://google.com",
			Query:      "repo:camdentest -file:id_rsa.pub BEGIN",
			QueryURL:   "https://youtube.com",
			NumResults: 31313,
		}

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", false, b)
			w.WriteHeader(200)
		}))
		defer s.Close()

		client := s.Client()
		err := postWebhook(context.Background(), client, s.URL, action)
		require.NoError(t, err)
	})

	t.Run("error is returned", func(t *testing.T) {
		action := actionArgs{
			MonitorDescription: "My test monitor", MonitorURL: "https://google.com",
			Query:      "repo:camdentest -file:id_rsa.pub BEGIN",
			QueryURL:   "https://youtube.com",
			NumResults: 31313,
		}

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			testutil.AssertGolden(t, "testdata/"+t.Name()+".json", false, b)
			w.WriteHeader(500)
		}))
		defer s.Close()

		client := s.Client()
		err := postWebhook(context.Background(), client, s.URL, action)
		require.Error(t, err)
	})
}
