package apiclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func TestDequeue(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/dequeue",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef"}`,
		responseStatus:   http.StatusOK,
		responsePayload:  `{"id": 42}`,
	}

	testRoute(t, spec, func(client *Client) {
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
		expectedPayload:  `{"executorName": "deadbeef"}`,
		responseStatus:   http.StatusNoContent,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *Client) {
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
		expectedPayload:  `{"executorName": "deadbeef"}`,
		responseStatus:   http.StatusInternalServerError,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *Client) {
		if _, err := client.Dequeue(context.Background(), "test_queue", nil); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestAddExecutionLogEntry(t *testing.T) {
	entry := workerutil.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/addExecutionLogEntry",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload: `{
			"executorName": "deadbeef",
			"jobId": 42,
			"key": "foo",
			"command": ["ls", "-a"],
			"startTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log payload>",
			"durationMs": 23123
		}`,
		responseStatus:  http.StatusOK,
		responsePayload: `99`,
	}

	testRoute(t, spec, func(client *Client) {
		entryID, err := client.AddExecutionLogEntry(context.Background(), "test_queue", 42, entry)
		if err != nil {
			t.Fatalf("unexpected error updating log contents: %s", err)
		}
		if entryID != 99 {
			t.Fatalf("unexpected entryID returned. want=%d, have=%d", 99, entryID)
		}
	})
}

func TestAddExecutionLogEntryBadResponse(t *testing.T) {
	entry := workerutil.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/addExecutionLogEntry",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload: `{
			"executorName": "deadbeef",
			"jobId": 42,
			"key": "foo",
			"command": ["ls", "-a"],
			"startTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log payload>",
			"durationMs": 23123
		}`,
		responseStatus:  http.StatusInternalServerError,
		responsePayload: ``,
	}

	testRoute(t, spec, func(client *Client) {
		if _, err := client.AddExecutionLogEntry(context.Background(), "test_queue", 42, entry); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestUpdateExecutionLogEntry(t *testing.T) {
	entry := workerutil.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/updateExecutionLogEntry",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload: `{
			"executorName": "deadbeef",
			"jobId": 42,
			"entryId": 99,
			"key": "foo",
			"command": ["ls", "-a"],
			"startTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log payload>",
			"durationMs": 23123
		}`,
		responseStatus:  http.StatusNoContent,
		responsePayload: ``,
	}

	testRoute(t, spec, func(client *Client) {
		if err := client.UpdateExecutionLogEntry(context.Background(), "test_queue", 42, 99, entry); err != nil {
			t.Fatalf("unexpected error updating log contents: %s", err)
		}
	})
}

func TestUpdateExecutionLogEntryBadResponse(t *testing.T) {
	entry := workerutil.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/updateExecutionLogEntry",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload: `{
			"executorName": "deadbeef",
			"jobId": 42,
			"entryId": 99,
			"key": "foo",
			"command": ["ls", "-a"],
			"startTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log payload>",
			"durationMs": 23123
		}`,
		responseStatus:  http.StatusInternalServerError,
		responsePayload: ``,
	}

	testRoute(t, spec, func(client *Client) {
		if err := client.UpdateExecutionLogEntry(context.Background(), "test_queue", 42, 99, entry); err == nil {
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

	testRoute(t, spec, func(client *Client) {
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

	testRoute(t, spec, func(client *Client) {
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

	testRoute(t, spec, func(client *Client) {
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

	testRoute(t, spec, func(client *Client) {
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

	testRoute(t, spec, func(client *Client) {
		if err := client.MarkFailed(context.Background(), "test_queue", 42, "OH NO"); err != nil {
			t.Fatalf("unexpected error completing job: %s", err)
		}
	})
}

func TestCanceled(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/canceled",
		expectedUsername: "test",
		expectedToken:    "hunter2",
		expectedPayload:  `{"executorName": "deadbeef"}`,
		responseStatus:   http.StatusOK,
		responsePayload:  `[1]`,
	}

	testRoute(t, spec, func(client *Client) {
		if ids, err := client.Canceled(context.Background(), "test_queue"); err != nil {
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

			"os": "test-os",
			"architecture": "test-architecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version"
		}`,
		responseStatus:  http.StatusOK,
		responsePayload: `[1]`,
	}

	testRoute(t, spec, func(client *Client) {
		unknownIDs, err := client.Heartbeat(context.Background(), "test_queue", []int{1, 2, 3})
		if err != nil {
			t.Fatalf("unexpected error performing heartbeat: %s", err)
		}

		if diff := cmp.Diff([]int{1}, unknownIDs); diff != "" {
			t.Errorf("unexpected unknown ids (-want +got):\n%s", diff)
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

			"os": "test-os",
			"architecture": "test-architecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version"
		}`,
		responseStatus:  http.StatusInternalServerError,
		responsePayload: ``,
	}

	testRoute(t, spec, func(client *Client) {
		if _, err := client.Heartbeat(context.Background(), "test_queue", []int{1, 2, 3}); err == nil {
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

func testRoute(t *testing.T, spec routeSpec, f func(client *Client)) {
	ts := testServer(t, spec)
	defer ts.Close()

	options := Options{
		ExecutorName: "deadbeef",
		PathPrefix:   "/.executors/queue",
		EndpointOptions: EndpointOptions{
			URL:   ts.URL,
			Token: "hunter2",
		},
		TelemetryOptions: TelemetryOptions{
			OS:              "test-os",
			Architecture:    "test-architecture",
			DockerVersion:   "test-docker-version",
			ExecutorVersion: "test-executor-version",
			GitVersion:      "test-git-version",
			IgniteVersion:   "test-ignite-version",
			SrcCliVersion:   "test-src-cli-version",
		},
	}

	f(New(options, &observation.TestContext))
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
