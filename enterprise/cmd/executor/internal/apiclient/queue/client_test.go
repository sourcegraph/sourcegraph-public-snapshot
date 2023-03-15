package queue_test

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
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestClient_Dequeue(t *testing.T) {
	tests := []struct {
		name        string
		spec        routeSpec
		expectedJob types.Job
		expectedErr error
		isDequeued  bool
	}{
		{
			name: "Success",
			spec: routeSpec{
				expectedMethod:   "POST",
				expectedPath:     "/.executors/queue/test_queue/dequeue",
				expectedUsername: "test",
				expectedToken:    "hunter2",
				expectedPayload:  `{"executorName": "deadbeef", "version": "0.0.0+dev"}`,
				responseStatus:   http.StatusOK,
				responsePayload:  `{"id": 42}`,
			},
			expectedJob: types.Job{ID: 42, VirtualMachineFiles: map[string]types.VirtualMachineFile{}},
			isDequeued:  true,
		},
		{
			name: "No record",
			spec: routeSpec{
				expectedMethod:   "POST",
				expectedPath:     "/.executors/queue/test_queue/dequeue",
				expectedUsername: "test",
				expectedToken:    "hunter2",
				expectedPayload:  `{"executorName": "deadbeef", "version": "0.0.0+dev"}`,
				responseStatus:   http.StatusNoContent,
				responsePayload:  ``,
			},
		},
		{
			name: "Bad Response",
			spec: routeSpec{
				expectedMethod:   "POST",
				expectedPath:     "/.executors/queue/test_queue/dequeue",
				expectedUsername: "test",
				expectedToken:    "hunter2",
				expectedPayload:  `{"executorName": "deadbeef", "version": "0.0.0+dev"}`,
				responseStatus:   http.StatusInternalServerError,
				responsePayload:  ``,
			},
			expectedErr: errors.New("unexpected status code 500"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				job, dequeued, err := client.Dequeue(context.Background(), "worker", "foo")
				if test.expectedErr != nil {
					require.Error(t, err)
					assert.Equal(t, test.expectedErr.Error(), err.Error())
					assert.Zero(t, job.ID)
					assert.False(t, dequeued)
				} else {
					require.NoError(t, err)
					assert.Equal(t, test.expectedJob, job)
					assert.Equal(t, test.isDequeued, dequeued)
				}
			})
		})
	}
}

func TestClient_MarkComplete(t *testing.T) {
	tests := []struct {
		name        string
		spec        routeSpec
		job         types.Job
		expectedErr error
	}{
		{
			name: "Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markComplete",
				expectedUsername:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42}`,
				responseStatus:       http.StatusNoContent,
				responsePayload:      ``,
			},
			job: types.Job{ID: 42, Token: "job-token"},
		},
		{
			name: "Success general access token",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markComplete",
				expectedUsername:     "test",
				expectedToken:        "hunter2",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42}`,
				responseStatus:       http.StatusNoContent,
				responsePayload:      ``,
			},
			job: types.Job{ID: 42},
		},
		{
			name: "Bad Response",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markComplete",
				expectedUsername:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42}`,
				responseStatus:       http.StatusInternalServerError,
				responsePayload:      ``,
			},
			job:         types.Job{ID: 42, Token: "job-token"},
			expectedErr: errors.New("unexpected status code 500"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				marked, err := client.MarkComplete(context.Background(), test.job)
				if test.expectedErr != nil {
					require.Error(t, err)
					assert.Equal(t, test.expectedErr.Error(), err.Error())
					assert.False(t, marked)
				} else {
					assert.True(t, marked)
				}
			})
		})
	}
}

func TestClient_MarkErrored(t *testing.T) {
	tests := []struct {
		name        string
		spec        routeSpec
		job         types.Job
		expectedErr error
	}{
		{
			name: "Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markErrored",
				expectedUsername:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
				responseStatus:       http.StatusNoContent,
				responsePayload:      ``,
			},
			job: types.Job{ID: 42, Token: "job-token"},
		},
		{
			name: "Success general access token",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markErrored",
				expectedUsername:     "test",
				expectedToken:        "hunter2",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
				responseStatus:       http.StatusNoContent,
				responsePayload:      ``,
			},
			job: types.Job{ID: 42},
		},
		{
			name: "Bad Response",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markErrored",
				expectedUsername:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
				responseStatus:       http.StatusInternalServerError,
				responsePayload:      ``,
			},
			job:         types.Job{ID: 42, Token: "job-token"},
			expectedErr: errors.New("unexpected status code 500"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				marked, err := client.MarkErrored(context.Background(), test.job, "OH NO")
				if test.expectedErr != nil {
					require.Error(t, err)
					assert.Equal(t, test.expectedErr.Error(), err.Error())
					assert.False(t, marked)
				} else {
					assert.True(t, marked)
				}
			})
		})
	}
}

func TestClient_MarkFailed(t *testing.T) {
	tests := []struct {
		name        string
		spec        routeSpec
		job         types.Job
		expectedErr error
	}{
		{
			name: "Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markFailed",
				expectedUsername:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
				responseStatus:       http.StatusNoContent,
				responsePayload:      ``,
			},
			job: types.Job{ID: 42, Token: "job-token"},
		},
		{
			name: "Success general access token",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markFailed",
				expectedUsername:     "test",
				expectedToken:        "hunter2",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
				responseStatus:       http.StatusNoContent,
				responsePayload:      ``,
			},
			job: types.Job{ID: 42},
		},
		{
			name: "Bad Response",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPath:         "/.executors/queue/test_queue/markFailed",
				expectedUsername:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorName: "deadbeef",
				expectedPayload:      `{"executorName": "deadbeef", "jobId": 42, "errorMessage": "OH NO"}`,
				responseStatus:       http.StatusInternalServerError,
				responsePayload:      ``,
			},
			job:         types.Job{ID: 42, Token: "job-token"},
			expectedErr: errors.New("unexpected status code 500"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				marked, err := client.MarkFailed(context.Background(), test.job, "OH NO")
				if test.expectedErr != nil {
					require.Error(t, err)
					assert.Equal(t, test.expectedErr.Error(), err.Error())
					assert.False(t, marked)
				} else {
					assert.True(t, marked)
				}
			})
		})
	}
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

	testRoute(t, spec, func(client *queue.Client) {
		if ids, err := client.CanceledJobs(context.Background(), []int{1}); err != nil {
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

	testRoute(t, spec, func(client *queue.Client) {
		unknownIDs, cancelIDs, err := client.Heartbeat(context.Background(), []int{1, 2, 3})
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

	testRoute(t, spec, func(client *queue.Client) {
		if _, _, err := client.Heartbeat(context.Background(), []int{1, 2, 3}); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestAddExecutionLogEntry(t *testing.T) {
	entry := internalexecutor.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPath:         "/.executors/queue/test_queue/addExecutionLogEntry",
		expectedUsername:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorName: "deadbeef",
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

	testRoute(t, spec, func(client *queue.Client) {
		entryID, err := client.AddExecutionLogEntry(context.Background(), types.Job{ID: 42, Token: "job-token"}, entry)
		if err != nil {
			t.Fatalf("unexpected error updating log contents: %s", err)
		}
		if entryID != 99 {
			t.Fatalf("unexpected entryID returned. want=%d, have=%d", 99, entryID)
		}
	})
}

func TestAddExecutionLogEntryBadResponse(t *testing.T) {
	entry := internalexecutor.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPath:         "/.executors/queue/test_queue/addExecutionLogEntry",
		expectedUsername:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorName: "deadbeef",
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

	testRoute(t, spec, func(client *queue.Client) {
		if _, err := client.AddExecutionLogEntry(context.Background(), types.Job{ID: 42, Token: "job-token"}, entry); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

func TestUpdateExecutionLogEntry(t *testing.T) {
	entry := internalexecutor.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPath:         "/.executors/queue/test_queue/updateExecutionLogEntry",
		expectedUsername:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorName: "deadbeef",
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

	testRoute(t, spec, func(client *queue.Client) {
		if err := client.UpdateExecutionLogEntry(context.Background(), types.Job{ID: 42, Token: "job-token"}, 99, entry); err != nil {
			t.Fatalf("unexpected error updating log contents: %s", err)
		}
	})
}

func TestUpdateExecutionLogEntryBadResponse(t *testing.T) {
	entry := internalexecutor.ExecutionLogEntry{
		Key:        "foo",
		Command:    []string{"ls", "-a"},
		StartTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log payload>",
		DurationMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPath:         "/.executors/queue/test_queue/updateExecutionLogEntry",
		expectedUsername:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorName: "deadbeef",
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

	testRoute(t, spec, func(client *queue.Client) {
		if err := client.UpdateExecutionLogEntry(context.Background(), types.Job{ID: 42, Token: "job-token"}, 99, entry); err == nil {
			t.Fatalf("expected an error")
		}
	})
}

type routeSpec struct {
	expectedMethod       string
	expectedPath         string
	expectedUsername     string
	expectedToken        string
	expectedJobID        string
	expectedExecutorName string
	expectedPayload      string
	responseStatus       int
	responsePayload      string
}

func testRoute(t *testing.T, spec routeSpec, f func(client *queue.Client)) {
	ts := testServer(t, spec)
	defer ts.Close()

	options := queue.Options{
		ExecutorName: "deadbeef",
		QueueName:    "test_queue",
		BaseClientOptions: apiclient.BaseClientOptions{
			ExecutorName: "deadbeef",
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

	client, err := newQueueClient(options)
	require.NoError(t, err)
	f(client)
}

func testServer(t *testing.T, spec routeSpec) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, spec.expectedMethod, r.Method)
		assert.Equal(t, spec.expectedPath, r.URL.Path)

		parts := strings.Split(r.Header.Get("Authorization"), " ")
		assert.Len(t, parts, 2)
		assert.Equal(t, spec.expectedToken, parts[1])

		assert.Equal(t, spec.expectedJobID, r.Header.Get("X-Sourcegraph-Job-ID"))
		assert.Equal(t, spec.expectedExecutorName, r.Header.Get("X-Sourcegraph-Executor-Name"))

		content, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, normalizeJSON([]byte(spec.expectedPayload)), normalizeJSON(content))

		w.WriteHeader(spec.responseStatus)
		_, err = w.Write([]byte(spec.responsePayload))
		require.NoError(t, err)
	}

	return httptest.NewServer(http.HandlerFunc(handler))
}

func newQueueClient(options queue.Options) (*queue.Client, error) {
	return queue.New(&observation.TestContext, options, prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) { return nil, nil }))
}

func normalizeJSON(v []byte) string {
	temp := map[string]any{}
	_ = json.Unmarshal(v, &temp)
	v, _ = json.Marshal(temp)
	return string(v)
}

func intptr(v int) *int { return &v }
