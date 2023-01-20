package worker_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue/worker"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDequeue(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/dequeue",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "version": "0.0.0+dev"}`,
		responseStatus:   http.StatusOK,
		responsePayload:  `{"id": 42}`,
	}

	testRoute(t, spec, func(client *worker.Client) {
		var job executor.Job
		dequeued, err := client.Dequeue(context.Background(), "test_queue", &job)
		if err != nil {
			t.Fatalf("unexpected error dequeueing record: %s", err)
		}
		if !dequeued {
			t.Fatalf("expected record to be dequeued")
		}
		if job.ID != 42 {
			t.Errorf("unexpected id. want=%d have=%d", 42, job.ID)
		}
	})
}

func TestDequeueNoRecord(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/dequeue",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "version": "0.0.0+dev"}`,
		responseStatus:   http.StatusNoContent,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		dequeued, err := client.Dequeue(context.Background(), "test_queue", nil)
		if err != nil {
			t.Fatalf("unexpected error dequeueing record: %s", err)
		}
		if dequeued {
			t.Fatalf("did not expect a record to be dequeued")
		}
	})
}

func TestDequeueBadResponse(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/dequeue",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "version": "0.0.0+dev"}`,
		responseStatus:   http.StatusInternalServerError,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if _, err := client.Dequeue(context.Background(), "test_queue", nil); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestMarkComplete(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/markComplete",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "jobId": 42}`,
		responseStatus:   http.StatusNoContent,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if err := client.MarkComplete(context.Background(), "test_queue", 42); err != nil {
			t.Fatalf("unexpected error completing job: %s", err)
		}
	})
}

func TestMarkCompleteBadResponse(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/markComplete",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "jobId": 42}`,
		responseStatus:   http.StatusInternalServerError,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if err := client.MarkComplete(context.Background(), "test_queue", 42); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestMarkErrored(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/markErrored",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
		responseStatus:   http.StatusNoContent,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if err := client.MarkErrored(context.Background(), "test_queue", 42, "OH NO"); err != nil {
			t.Fatalf("unexpected error completing job: %s", err)
		}
	})
}

func TestMarkErroredBadResponse(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/markErrored",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
		responseStatus:   http.StatusInternalServerError,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if err := client.MarkErrored(context.Background(), "test_queue", 42, "OH NO"); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestMarkFailed(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/markFailed",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
		responseStatus:   http.StatusNoContent,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if err := client.MarkFailed(context.Background(), "test_queue", 42, "OH NO"); err != nil {
			t.Fatalf("unexpected error completing job: %s", err)
		}
	})
}

func TestCanceledJobs(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/canceledJobs",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef","knownJobIds":[1]}`,
		responseStatus:   http.StatusOK,
		responsePayload:  `[1]`,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if ids, err := client.CanceledJobs(context.Background(), "test_queue", []int{1}); err != nil {
			t.Fatalf("unexpected error completing job: %s", err)
		} else if diff := cmp.Diff(ids, []int{1}); diff != "" {
			t.Fatalf("unexpected set of IDs returned: %s", diff)
		}
	})
}

func TestHeartbeat(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/heartbeat",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload: `{
			"executorName": "deadbeef",
			"jobIds": [1,2,3],
			"version": "V2",

			"os": "test-os",
			"architecture": "test-architecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version",

			"prometheusMetrics": ""
		}`,
		responseStatus:  http.StatusOK,
		responsePayload: `{"knownIDs": [1], "cancelIDs": [1]}`,
	}

	testRoute(t, spec, func(client *worker.Client) {
		unknownIDs, cancelIDs, err := client.Heartbeat(context.Background(), "test_queue", []int{1, 2, 3})
		if err != nil {
			t.Fatalf("unexpected error performing heartbeat: %s", err)
		}

		if diff := cmp.Diff([]int{1}, unknownIDs); diff != "" {
			t.Errorf("unexpected unknown ids (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{1}, cancelIDs); diff != "" {
			t.Errorf("unexpected unknown cancel ids (-want +got):\n%s", diff)
		}
	})
}

func TestHeartbeatBadResponse(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/heartbeat",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload: `{
			"executorName": "deadbeef",
			"jobIds": [1,2,3],
			"version": "V2",

			"os": "test-os",
			"architecture": "test-architecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version",

			"prometheusMetrics": ""
		}`,
		responseStatus:  http.StatusInternalServerError,
		responsePayload: ``,
	}

	testRoute(t, spec, func(client *worker.Client) {
		if _, _, err := client.Heartbeat(context.Background(), "test_queue", []int{1, 2, 3}); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

type routeSpec struct {
	expectedMethod   string
	expectedPath     string
	expectedUsername string
	expectedToken    string
	expectedPayload  string
	responseStatus   int
	responsePayload  string
}

func testRoute(t *testing.T, spec routeSpec, f func(client *worker.Client)) {
	ts := testServer(t, spec)
	defer ts.Close()

	options := queue.Options{
		ExecutorName: "deadbeef",
		BaseClientOptions: apiclient.BaseClientOptions{
			EndpointOptions: apiclient.EndpointOptions{
				URL:        ts.URL,
				PathPrefix: "/.executors/queue",
				Token:      "hunter2",
			},
		},
		TelemetryOptions: queue.TelemetryOptions{
			OS:              "test-os",
			Architecture:    "test-architecture",
			DockerVersion:   "test-docker-version",
			ExecutorVersion: "test-executor-version",
			GitVersion:      "test-git-version",
			IgniteVersion:   "test-ignite-version",
			SrcCliVersion:   "test-src-cli-version",
		},
	}

	client, err := worker.New(&observation.TestContext, options, prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) { return nil, nil }))
	require.NoError(t, err)
	f(client)
}

func testServer(t *testing.T, spec routeSpec) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != spec.expectedMethod {
			t.Errorf("unexpected method. want=%s have=%s", spec.expectedMethod, r.Method)
		}
		if r.URL.Path != spec.expectedPath {
			t.Errorf("unexpected method. want=%s have=%s", spec.expectedPath, r.URL.Path)
		}

		parts := strings.Split(r.Header.Get("Authorization"), " ")
		if len(parts) != 2 || parts[0] != "token-executor" {
			if parts[1] != spec.expectedToken {
				t.Errorf("unexpected token`. want=%s have=%s", spec.expectedToken, parts[1])
			}
		}

		content, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error reading payload: %s", err)
		}
		if diff := cmp.Diff(normalizeJSON([]byte(spec.expectedPayload)), normalizeJSON(content)); diff != "" {
			t.Errorf("unexpected request payload (-want +got):\n%s", diff)
		}

		w.WriteHeader(spec.responseStatus)
		w.Write([]byte(spec.responsePayload))
	}

	return httptest.NewServer(http.HandlerFunc(handler))
}

func normalizeJSON(v []byte) string {
	temp := map[string]any{}
	_ = json.Unmarshal(v, &temp)
	v, _ = json.Marshal(temp)
	return string(v)
}

func intptr(v int) *int { return &v }
