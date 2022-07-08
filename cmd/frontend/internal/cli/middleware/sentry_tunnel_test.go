package middleware_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli/middleware"
)

var sentryPayload = []byte(`{"event_id":"6af2790372f046689a858b1d914fe0d5","sent_at":"2022-07-07T17:38:47.215Z","sdk":{"name":"sentry.javascript.browser","version":"6.19.7"},"dsn":"https://randomkey@o19358.ingest.sentry.io/1391511"}
{"type":"event","sample_rates":[{}]}
{"message":"foopff","level":"info","event_id":"6af2790372f046689a858b1d914fe0d5","platform":"javascript","timestamp":1657215527.214,"environment":"production","sdk":{"integrations":["InboundFilters","FunctionToString","TryCatch","Breadcrumbs","GlobalHandlers","LinkedErrors","Dedupe","UserAgent"],"name":"sentry.javascript.browser","version":"6.19.7","packages":[{"name":"npm:@sentry/browser","version":"6.19.7"}]},"request":{"url":"https://sourcegraph.test:3443/search","headers":{"Referer":"https://sourcegraph.test:3443/search","User-Agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.53 Safari/537.36"}},"tags":{},"extra":{}}`)

func TestSentryTunnel(t *testing.T) {
	createTestServer := func() (*httptest.Server, chan struct{}) {
		ch := make(chan struct{})
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				// If we're pretending to be the sentry ingest server
				select {
				case ch <- struct{}{}:
				case <-time.After(time.Second):
				}
				return
			}
			w.WriteHeader(http.StatusTeapot)
		})
		server := httptest.NewServer(middleware.SentryTunnel(h))
		envvar.SentryTunnelEndpoint = server.URL
		return server, ch
	}

	t.Run("POST /_sentry_tunnel", func(t *testing.T) {
		t.Run("With a valid event", func(t *testing.T) {
			server, ch := createTestServer()
			resp, err := http.Post(fmt.Sprintf("%s/_sentry_tunnel", server.URL), "text/plain;charset=UTF-8", bytes.NewReader(sentryPayload))
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			select {
			case <-ch:
			case <-time.After(time.Second):
				t.Fatalf("expected senty ingester to be called")
			}
		})
		t.Run("With an invalid event", func(t *testing.T) {
			server, _ := createTestServer()
			resp, err := http.Post(fmt.Sprintf("%s/_sentry_tunnel", server.URL), "text/plain;charset=UTF-8", bytes.NewReader([]byte("foobar")))
			assert.NoError(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		})
		t.Run("With an invalid project id", func(t *testing.T) {
			server, _ := createTestServer()
			invalidProjectIDpayload := bytes.Replace(sentryPayload, []byte("1391511"), []byte("10000"), 1)
			resp, err := http.Post(fmt.Sprintf("%s/_sentry_tunnel", server.URL), "text/plain;charset=UTF-8", bytes.NewReader(invalidProjectIDpayload))
			assert.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})
	t.Run("GET /_sentry_tunnel", func(t *testing.T) {
		server, _ := createTestServer()
		resp, err := http.Get(fmt.Sprintf("%s/_sentry_tunnel", server.URL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
	t.Run("GET /foobar", func(t *testing.T) {
		server, _ := createTestServer()
		resp, err := http.Get(fmt.Sprintf("%s/foobar", server.URL))
		assert.NoError(t, err)
		// Teapot, because we're hitting the handler at the end of our chain.
		assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	})
}
