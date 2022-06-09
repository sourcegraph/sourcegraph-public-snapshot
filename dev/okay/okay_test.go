package okay_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/okay"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestMain(m *testing.M) {
	logtest.Init(m)
	m.Run()
	os.Exit(0)
}

func TestPush(t *testing.T) {
	t.Run("serialization", func(t *testing.T) {
		wantJSON := `{
  "event": "custom",
  "metrics": {
    "elapsed": {
      "type": "durationMs",
      "value": 112728000
    }
  },
  "identity": {
    "type": "sourceControlLogin",
    "user": "bobheadxi"
  },
  "timestamp": "2022-04-21T12:37:00Z",
  "uniqueKey": [
    "unique_key"
  ],
  "properties": {
    "unique_key":"preprod,34159,frontend,github-proxy,precise-code-intel,searcher,symbols,syntect-server,worker",
    "okay.url": "https://github.com/sourcegraph/sourcegraph/pull/34159",
    "environment": "preprod",
    "pull_request.url": "https://github.com/sourcegraph/sourcegraph/pull/34159",
    "pull_request.title": "httptestutil: delete all VCR headers that look risky",
    "pull_request.number": "34159",
    "pull_request.revision": "abef4220236c7114d84933574fadba43beaf54f5"
  },
  "customEventName": "qa.deployment",
  "labels": [
      "frontend",
      "github-proxy",
      "precise-code-intel",
      "searcher",
      "symbols",
      "syntect-server",
      "worker"
   ]
}`

		svr, cli := newTestServer(t, eventHandler(func(body []byte) int {
			assert.JSONEq(t, wantJSON, string(body))
			return 200
		}))
		defer svr.Close()

		err := cli.Push(&okay.Event{
			Name:      "qa.deployment",
			Timestamp: time.Date(2022, 04, 21, 12, 37, 0, 0, time.UTC),
			Metrics: map[string]okay.Metric{
				"elapsed": {
					Type:  "durationMs",
					Value: 112728000,
				},
			},
			GitHubLogin: "bobheadxi",
			OkayURL:     "https://github.com/sourcegraph/sourcegraph/pull/34159",
			UniqueKey:   []string{"unique_key"},
			Properties: map[string]string{
				"unique_key":            "preprod,34159,frontend,github-proxy,precise-code-intel,searcher,symbols,syntect-server,worker",
				"environment":           "preprod",
				"pull_request.url":      "https://github.com/sourcegraph/sourcegraph/pull/34159",
				"pull_request.title":    "httptestutil: delete all VCR headers that look risky",
				"pull_request.number":   "34159",
				"pull_request.revision": "abef4220236c7114d84933574fadba43beaf54f5",
			},
			Labels: []string{
				"frontend",
				"github-proxy",
				"precise-code-intel",
				"searcher",
				"symbols",
				"syntect-server",
				"worker",
			},
		})
		assert.NoError(t, err)
		assert.NoError(t, cli.Flush())
	})

	t.Run("validation", func(t *testing.T) {
		svr, cli := newTestServer(t)
		defer svr.Close()

		t.Run("NOK blank name", func(t *testing.T) {
			event := dummyEvent()
			event.Name = ""
			err := cli.Push(event)
			assert.Error(t, err)

			// No error because we never reached the server.
			assert.NoError(t, cli.Flush())
		})

		t.Run("NOK zero value timestamp", func(t *testing.T) {
			event := dummyEvent()
			event.Timestamp = time.Time{}
			err := cli.Push(event)
			assert.Error(t, err)

			// No error because we never reached the server.
			assert.NoError(t, cli.Flush())
		})

		t.Run("NOK no metrics", func(t *testing.T) {
			event := dummyEvent()
			event.Metrics = nil
			err := cli.Push(event)
			assert.Error(t, err)

			// No error because we never reached the server.
			assert.NoError(t, cli.Flush())
		})

		t.Run("NOK absent uniqueKey", func(t *testing.T) {
			event := dummyEvent()
			event.UniqueKey = []string{"not-existing"}
			err := cli.Push(event)
			assert.Error(t, err)

			// No error because we never reached the server.
			assert.NoError(t, cli.Flush())
		})

	})
}

func TestFlush(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		svr, cli := newTestServer(t,
			eventHandlerMap(func(rawEvent map[string]any) int {
				assert.Equal(t, "one", rawEvent["customEventName"])
				return 200
			}),
			eventHandlerMap(func(rawEvent map[string]any) int {
				assert.Equal(t, "two", rawEvent["customEventName"])
				return 200
			}),
			eventHandlerMap(func(rawEvent map[string]any) int {
				assert.Equal(t, "three", rawEvent["customEventName"])
				return 200
			}),
		)
		defer svr.Close()

		for _, name := range []string{"one", "two", "three"} {
			event := dummyEvent()
			event.Name = name
			err := cli.Push(event)
			assert.NoError(t, err)
		}
		assert.NoError(t, cli.Flush())
	})

	t.Run("NOK http errors", func(t *testing.T) {
		svr, cli := newTestServer(t,
			eventHandlerMap(func(rawEvent map[string]any) int {
				assert.Equal(t, "one", rawEvent["customEventName"])
				return 200
			}),
			eventHandlerMap(func(rawEvent map[string]any) int {
				assert.Equal(t, "two", rawEvent["customEventName"])
				return 500
			}),
			eventHandlerMap(func(rawEvent map[string]any) int {
				assert.Equal(t, "three", rawEvent["customEventName"])
				return 500
			}),
			eventHandlerMap(func(rawEvent map[string]any) int {
				assert.Equal(t, "four", rawEvent["customEventName"])
				return 200
			}),
		)
		defer svr.Close()

		for _, name := range []string{"one", "two", "three", "four"} {
			event := dummyEvent()
			event.Name = name
			err := cli.Push(event)
			assert.NoError(t, err)
		}
		err := cli.Flush()
		assert.NotNil(t, err)

		assert.Contains(t, err.Error(), "three")
		assert.Contains(t, err.Error(), "two")
	})
}

func newTestServer(t *testing.T, handlers ...eventHandler) (*httptest.Server, *okay.Client) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if len(handlers) == 0 {
			t.Fatalf("unexpected request")
		}
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var h eventHandler
		h, handlers = handlers[0], handlers[1:]

		status := h(b)
		w.WriteHeader(status)
	}))

	cli := okay.NewClient(svr.Client(), "foobar")
	cli.SetEndpoint(svr.URL)

	return svr, cli
}

func timestamp() time.Time {
	t, _ := time.Parse(time.RFC822, time.RFC822)
	return t
}

type eventHandler func([]byte) int

func eventHandlerMap(h func(map[string]any) int) eventHandler {
	return func(body []byte) int {
		var rawEvent map[string]any
		_ = json.Unmarshal(body, &rawEvent)
		return h(rawEvent)
	}
}

func dummyEvent() *okay.Event {
	return &okay.Event{
		Name:      "dummy",
		Timestamp: timestamp(),
		Metrics: map[string]okay.Metric{
			"fooMetric": {
				Type:  "count",
				Value: 1,
			},
		},
		UniqueKey: []string{"foo"},
		Properties: map[string]string{
			"foo": "bar",
			"bar": "baz",
		},
	}
}
