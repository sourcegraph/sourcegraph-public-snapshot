package apiclient

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func TestDequeue(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/dequeue",
		expectedUsername: "test",
		expectedPassword: "hunter2",
		expectedPayload:  `{"executorName": "deadbeef"}`,
		responseStatus:   http.StatusOK,
		responsePayload:  `{"id": 42}`,
	}

	testRoute(t, spec, func(client *Client) {
		var job Job
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
		expectedPassword: "hunter2",
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
		expectedPassword: "hunter2",
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
		ExitCode:   123,
		Out:        "<log payload>",
		DurationMs: 23123,
	}

	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/addExecutionLogEntry",
		expectedUsername: "test",
		expectedPassword: "hunter2",
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
		responseStatus:  http.StatusNoContent,
		responsePayload: ``,
	}

	testRoute(t, spec, func(client *Client) {
		if err := client.AddExecutionLogEntry(context.Background(), "test_queue", 42, entry); err != nil {
			t.Fatalf("unexpected error updating log contents: %s", err)
		}
	})
}

func TestAddExecutionLogEntryBadResponse(t *testing.T) {
	entry := workerutil.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   123,
		Out:        "<log payload>",
		DurationMs: 23123,
	}

	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/addExecutionLogEntry",
		expectedUsername: "test",
		expectedPassword: "hunter2",
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
		if err := client.AddExecutionLogEntry(context.Background(), "test_queue", 42, entry); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestMarkComplete(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/test_queue/markComplete",
		expectedUsername: "test",
		expectedPassword: "hunter2",
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
		expectedPassword: "hunter2",
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
		expectedPassword: "hunter2",
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
		expectedPassword: "hunter2",
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
		expectedPassword: "hunter2",
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

func TestHeartbeat(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/heartbeat",
		expectedUsername: "test",
		expectedPassword: "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "jobIds": [1, 2, 3]}`,
		responseStatus:   http.StatusNoContent,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *Client) {
		if err := client.Heartbeat(context.Background(), []int{1, 2, 3}); err != nil {
			t.Fatalf("unexpected error performing heartbeat: %s", err)
		}
	})
}

func TestHeartbeatBadResponse(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPath:     "/.executors/queue/heartbeat",
		expectedUsername: "test",
		expectedPassword: "hunter2",
		expectedPayload:  `{"executorName": "deadbeef", "jobIds": [1, 2, 3]}`,
		responseStatus:   http.StatusInternalServerError,
		responsePayload:  ``,
	}

	testRoute(t, spec, func(client *Client) {
		if err := client.Heartbeat(context.Background(), []int{1, 2, 3}); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

type routeSpec struct {
	expectedMethod   string
	expectedPath     string
	expectedUsername string
	expectedPassword string
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
			URL:      ts.URL,
			Username: "test",
			Password: "hunter2",
		},
	}
	f(New(options))
}

func testServer(t *testing.T, spec routeSpec) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != spec.expectedMethod {
			t.Errorf("unexpected method. want=%s have=%s", spec.expectedMethod, r.Method)
		}
		if r.URL.Path != spec.expectedPath {
			t.Errorf("unexpected method. want=%s have=%s", spec.expectedPath, r.URL.Path)
		}

		username, password, _ := r.BasicAuth()
		if username != spec.expectedUsername {
			t.Errorf("unexpected username. want=%s have=%s", spec.expectedUsername, username)
		}
		if password != spec.expectedPassword {
			t.Errorf("unexpected password. want=%s have=%s", spec.expectedPassword, password)
		}

		content, err := ioutil.ReadAll(r.Body)
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
	temp := map[string]interface{}{}
	_ = json.Unmarshal(v, &temp)
	v, _ = json.Marshal(temp)
	return string(v)
}
