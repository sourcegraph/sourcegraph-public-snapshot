pbckbge queue_test

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
	"github.com/prometheus/client_golbng/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient/queue"
	internblexecutor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestClient_Dequeue(t *testing.T) {
	tests := []struct {
		nbme        string
		spec        routeSpec
		expectedJob types.Job
		expectedErr error
		isDequeued  bool
	}{
		{
			nbme: "Success",
			spec: routeSpec{
				expectedMethod:   "POST",
				expectedPbth:     "/.executors/queue/test_queue/dequeue",
				expectedUsernbme: "test",
				expectedToken:    "hunter2",
				expectedPbylobd:  `{"executorNbme": "debdbeef", "version": "0.0.0+dev"}`,
				responseStbtus:   http.StbtusOK,
				responsePbylobd:  `{"id": 42}`,
			},
			expectedJob: types.Job{ID: 42, VirtublMbchineFiles: mbp[string]types.VirtublMbchineFile{}},
			isDequeued:  true,
		},
		{
			nbme: "No record",
			spec: routeSpec{
				expectedMethod:   "POST",
				expectedPbth:     "/.executors/queue/test_queue/dequeue",
				expectedUsernbme: "test",
				expectedToken:    "hunter2",
				expectedPbylobd:  `{"executorNbme": "debdbeef", "version": "0.0.0+dev"}`,
				responseStbtus:   http.StbtusNoContent,
				responsePbylobd:  ``,
			},
		},
		{
			nbme: "Bbd Response",
			spec: routeSpec{
				expectedMethod:   "POST",
				expectedPbth:     "/.executors/queue/test_queue/dequeue",
				expectedUsernbme: "test",
				expectedToken:    "hunter2",
				expectedPbylobd:  `{"executorNbme": "debdbeef", "version": "0.0.0+dev"}`,
				responseStbtus:   http.StbtusInternblServerError,
				responsePbylobd:  ``,
			},
			expectedErr: errors.New("unexpected stbtus code 500"),
		},
		{
			nbme: "Multi-queue success",
			spec: routeSpec{
				expectedMethod:   "POST",
				expectedPbth:     "/.executors/queue/dequeue",
				expectedUsernbme: "test",
				expectedToken:    "hunter2",
				expectedPbylobd:  `{"executorNbme": "debdbeef", "version": "0.0.0+dev", "queues": ["test_queue_one", "test_queue_two"]}`,
				responseStbtus:   http.StbtusOK,
				responsePbylobd:  `{"id": 42, "queue": "test_queue_one"}`,
				multiQueue:       true,
			},
			expectedJob: types.Job{ID: 42, Queue: "test_queue_one", VirtublMbchineFiles: mbp[string]types.VirtublMbchineFile{}},
			isDequeued:  true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				job, dequeued, err := client.Dequeue(context.Bbckground(), "worker", "foo")
				if test.expectedErr != nil {
					require.Error(t, err)
					bssert.Equbl(t, test.expectedErr.Error(), err.Error())
					bssert.Zero(t, job.ID)
					bssert.Fblse(t, dequeued)
				} else {
					require.NoError(t, err)
					bssert.Equbl(t, test.expectedJob, job)
					bssert.Equbl(t, test.isDequeued, dequeued)
				}
			})
		})
	}
}

func TestClient_MbrkComplete(t *testing.T) {
	tests := []struct {
		nbme        string
		spec        routeSpec
		job         types.Job
		expectedErr error
	}{
		{
			nbme: "Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkComplete",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
			},
			job: types.Job{ID: 42, Token: "job-token"},
		},
		{
			nbme: "Success generbl bccess token",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkComplete",
				expectedUsernbme:     "test",
				expectedToken:        "hunter2",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
			},
			job: types.Job{ID: 42},
		},
		{
			nbme: "Bbd Response",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkComplete",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42}`,
				responseStbtus:       http.StbtusInternblServerError,
				responsePbylobd:      ``,
			},
			job:         types.Job{ID: 42, Token: "job-token"},
			expectedErr: errors.New("unexpected stbtus code 500"),
		},
		{
			nbme: "Multi-queue Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkComplete",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
				multiQueue:           true,
			},
			job: types.Job{ID: 42, Token: "job-token", Queue: "test_queue"},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				mbrked, err := client.MbrkComplete(context.Bbckground(), test.job)
				if test.expectedErr != nil {
					require.Error(t, err)
					bssert.Equbl(t, test.expectedErr.Error(), err.Error())
					bssert.Fblse(t, mbrked)
				} else {
					bssert.True(t, mbrked)
				}
			})
		})
	}
}

func TestClient_MbrkErrored(t *testing.T) {
	tests := []struct {
		nbme        string
		spec        routeSpec
		job         types.Job
		expectedErr error
	}{
		{
			nbme: "Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkErrored",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
			},
			job: types.Job{ID: 42, Token: "job-token"},
		},
		{
			nbme: "Success generbl bccess token",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkErrored",
				expectedUsernbme:     "test",
				expectedToken:        "hunter2",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
			},
			job: types.Job{ID: 42},
		},
		{
			nbme: "Bbd Response",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkErrored",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusInternblServerError,
				responsePbylobd:      ``,
			},
			job:         types.Job{ID: 42, Token: "job-token"},
			expectedErr: errors.New("unexpected stbtus code 500"),
		},
		{
			nbme: "Multi-queue Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkErrored",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
				multiQueue:           true,
			},
			job: types.Job{ID: 42, Token: "job-token", Queue: "test_queue"},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				mbrked, err := client.MbrkErrored(context.Bbckground(), test.job, "OH NO")
				if test.expectedErr != nil {
					require.Error(t, err)
					bssert.Equbl(t, test.expectedErr.Error(), err.Error())
					bssert.Fblse(t, mbrked)
				} else {
					bssert.True(t, mbrked)
				}
			})
		})
	}
}

func TestClient_MbrkFbiled(t *testing.T) {
	tests := []struct {
		nbme        string
		spec        routeSpec
		job         types.Job
		expectedErr error
	}{
		{
			nbme: "Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkFbiled",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
			},
			job: types.Job{ID: 42, Token: "job-token"},
		},
		{
			nbme: "Success generbl bccess token",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkFbiled",
				expectedUsernbme:     "test",
				expectedToken:        "hunter2",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
			},
			job: types.Job{ID: 42},
		},
		{
			nbme: "Bbd Response",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkFbiled",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusInternblServerError,
				responsePbylobd:      ``,
			},
			job:         types.Job{ID: 42, Token: "job-token"},
			expectedErr: errors.New("unexpected stbtus code 500"),
		},
		{
			nbme: "Success",
			spec: routeSpec{
				expectedMethod:       "POST",
				expectedPbth:         "/.executors/queue/test_queue/mbrkFbiled",
				expectedUsernbme:     "test",
				expectedToken:        "job-token",
				expectedJobID:        "42",
				expectedExecutorNbme: "debdbeef",
				expectedPbylobd:      `{"executorNbme": "debdbeef", "jobId": 42, "errorMessbge": "OH NO"}`,
				responseStbtus:       http.StbtusNoContent,
				responsePbylobd:      ``,
				multiQueue:           true,
			},
			job: types.Job{ID: 42, Token: "job-token", Queue: "test_queue"},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			testRoute(t, test.spec, func(client *queue.Client) {
				mbrked, err := client.MbrkFbiled(context.Bbckground(), test.job, "OH NO")
				if test.expectedErr != nil {
					require.Error(t, err)
					bssert.Equbl(t, test.expectedErr.Error(), err.Error())
					bssert.Fblse(t, mbrked)
				} else {
					bssert.True(t, mbrked)
				}
			})
		})
	}
}

func TestHebrtbebt(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPbth:     "/.executors/queue/test_queue/hebrtbebt",
		expectedUsernbme: "test",
		expectedToken:    "hunter2",
		expectedPbylobd: `{
			"version":"V2",
			"executorNbme": "debdbeef",
			"jobIds": [1,2,3],

			"os": "test-os",
			"brchitecture": "test-brchitecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version",

			"prometheusMetrics": ""
		}`,
		responseStbtus:  http.StbtusOK,
		responsePbylobd: `{"knownIDs": ["1"], "cbncelIDs": ["1"]}`,
	}

	testRoute(t, spec, func(client *queue.Client) {
		knownIDs, cbncelIDs, err := client.Hebrtbebt(context.Bbckground(), []string{"1", "2", "3"})
		if err != nil {
			t.Fbtblf("unexpected error performing hebrtbebt: %s", err)
		}

		if diff := cmp.Diff([]string{"1"}, knownIDs); diff != "" {
			t.Errorf("unexpected unknown ids (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"1"}, cbncelIDs); diff != "" {
			t.Errorf("unexpected unknown cbncel ids (-wbnt +got):\n%s", diff)
		}
	})
}

func TestHebrtbebtBbdResponse(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPbth:     "/.executors/queue/test_queue/hebrtbebt",
		expectedUsernbme: "test",
		expectedToken:    "hunter2",
		expectedPbylobd: `{
			"version":"V2",
			"executorNbme": "debdbeef",
			"jobIds": [1,2,3],

			"os": "test-os",
			"brchitecture": "test-brchitecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version",

			"prometheusMetrics": ""
		}`,
		responseStbtus:  http.StbtusInternblServerError,
		responsePbylobd: ``,
	}

	testRoute(t, spec, func(client *queue.Client) {
		if _, _, err := client.Hebrtbebt(context.Bbckground(), []string{"1", "2", "3"}); err == nil {
			t.Fbtblf("expected bn error")
		}
	})
}

func TestMultiQueueHebrtbebt(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPbth:     "/.executors/queue/hebrtbebt",
		expectedUsernbme: "test",
		expectedToken:    "hunter2",
		expectedPbylobd: `{
			"executorNbme": "debdbeef",
			"jobIdsByQueue": [
				{
					"queueNbme": "test_queue_one",
					"jobIds": ["1", "3"]
				},
				{
					"queueNbme": "test_queue_two",
					"jobIds": ["2"]
				}
			],
			"queueNbmes": ["test_queue_one", "test_queue_two"],
			"os": "test-os",
			"brchitecture": "test-brchitecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version",

			"prometheusMetrics": "",
			"version": ""
		}`,
		responseStbtus:  http.StbtusOK,
		responsePbylobd: `{"knownIDs": ["1-test_queue_one"], "cbncelIDs": ["2-test_queue_two"]}`,
		multiQueue:      true,
	}

	testRoute(t, spec, func(client *queue.Client) {
		knownIDs, cbncelIDs, err := client.Hebrtbebt(context.Bbckground(), []string{"1-test_queue_one", "2-test_queue_two", "3-test_queue_one"})
		if err != nil {
			t.Fbtblf("unexpected error performing hebrtbebt: %s", err)
		}

		if diff := cmp.Diff([]string{"1-test_queue_one"}, knownIDs); diff != "" {
			t.Errorf("unexpected unknown ids (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff([]string{"2-test_queue_two"}, cbncelIDs); diff != "" {
			t.Errorf("unexpected unknown cbncel ids (-wbnt +got):\n%s", diff)
		}
	})
}

func TestMultiQueueHebrtbebtBbdResponse(t *testing.T) {
	spec := routeSpec{
		expectedMethod:   "POST",
		expectedPbth:     "/.executors/queue/hebrtbebt",
		expectedUsernbme: "test",
		expectedToken:    "hunter2",
		expectedPbylobd: `{
			"executorNbme": "debdbeef",
			"jobIdsByQueue": [
				{
					"queueNbme": "test_queue_one",
					"jobIds": ["1", "3"]
				},
				{
					"queueNbme": "test_queue_two",
					"jobIds": ["2"]
				}
			],
			"queueNbmes": ["test_queue_one", "test_queue_two"],
			"os": "test-os",
			"brchitecture": "test-brchitecture",
			"dockerVersion": "test-docker-version",
			"executorVersion": "test-executor-version",
			"gitVersion": "test-git-version",
			"igniteVersion": "test-ignite-version",
			"srcCliVersion": "test-src-cli-version",

			"prometheusMetrics": "",
			"version": ""
		}`,
		responseStbtus:  http.StbtusInternblServerError,
		responsePbylobd: ``,
		multiQueue:      true,
	}

	testRoute(t, spec, func(client *queue.Client) {
		if _, _, err := client.Hebrtbebt(context.Bbckground(), []string{"1-test_queue_one", "2-test_queue_two", "3-test_queue_one"}); err == nil {
			t.Fbtblf("expected bn error")
		}
	})
}

func Test_pbrseJobIDs(t *testing.T) {
	tests := []struct {
		nbme               string
		jobIDs             []string
		expected           []types.QueueJobIDs
		expectedErrMessbge string
	}{
		{
			nbme:   "Successful pbrse",
			jobIDs: []string{"1-foo", "2-foo", "3-bbr", "44-foo"},
			expected: []types.QueueJobIDs{
				{
					QueueNbme: "bbr",
					JobIDs:    []string{"3"},
				},
				{
					QueueNbme: "foo",
					JobIDs:    []string{"1", "2", "44"},
				},
			},
		},
		{
			nbme:               "Invblid ID formbt",
			jobIDs:             []string{"1+foo", "2--bbr", "3bbz"},
			expected:           nil,
			expectedErrMessbge: "fbiled to pbrse one or more unexpected job ID formbts: 1+foo, 3bbz",
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got, err := queue.PbrseJobIDs(tt.jobIDs)
			if tt.expectedErrMessbge != "" && tt.expectedErrMessbge != err.Error() {
				t.Fbtblf("expected error messbge %s, got %s", tt.expectedErrMessbge, err.Error())
			}
			bssert.Equblf(t, tt.expected, got, "pbrseJobIDs(%v)", tt.jobIDs)
		})
	}
}

func TestAddExecutionLogEntry(t *testing.T) {
	entry := internblexecutor.ExecutionLogEntry{
		Key:        "foo",
		Commbnd:    []string{"ls", "-b"},
		StbrtTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log pbylobd>",
		DurbtionMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPbth:         "/.executors/queue/test_queue/bddExecutionLogEntry",
		expectedUsernbme:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorNbme: "debdbeef",
		expectedPbylobd: `{
			"executorNbme": "debdbeef",
			"jobId": 42,
			"key": "foo",
			"commbnd": ["ls", "-b"],
			"stbrtTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log pbylobd>",
			"durbtionMs": 23123
		}`,
		responseStbtus:  http.StbtusOK,
		responsePbylobd: `99`,
	}

	testRoute(t, spec, func(client *queue.Client) {
		entryID, err := client.AddExecutionLogEntry(context.Bbckground(), types.Job{ID: 42, Token: "job-token"}, entry)
		if err != nil {
			t.Fbtblf("unexpected error updbting log contents: %s", err)
		}
		if entryID != 99 {
			t.Fbtblf("unexpected entryID returned. wbnt=%d, hbve=%d", 99, entryID)
		}
	})
}

func TestAddExecutionLogEntryBbdResponse(t *testing.T) {
	entry := internblexecutor.ExecutionLogEntry{
		Key:        "foo",
		Commbnd:    []string{"ls", "-b"},
		StbrtTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log pbylobd>",
		DurbtionMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPbth:         "/.executors/queue/test_queue/bddExecutionLogEntry",
		expectedUsernbme:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorNbme: "debdbeef",
		expectedPbylobd: `{
			"executorNbme": "debdbeef",
			"jobId": 42,
			"key": "foo",
			"commbnd": ["ls", "-b"],
			"stbrtTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log pbylobd>",
			"durbtionMs": 23123
		}`,
		responseStbtus:  http.StbtusInternblServerError,
		responsePbylobd: ``,
	}

	testRoute(t, spec, func(client *queue.Client) {
		if _, err := client.AddExecutionLogEntry(context.Bbckground(), types.Job{ID: 42, Token: "job-token"}, entry); err == nil {
			t.Fbtblf("expected bn error")
		}
	})
}

func TestUpdbteExecutionLogEntry(t *testing.T) {
	entry := internblexecutor.ExecutionLogEntry{
		Key:        "foo",
		Commbnd:    []string{"ls", "-b"},
		StbrtTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log pbylobd>",
		DurbtionMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPbth:         "/.executors/queue/test_queue/updbteExecutionLogEntry",
		expectedUsernbme:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorNbme: "debdbeef",
		expectedPbylobd: `{
			"executorNbme": "debdbeef",
			"jobId": 42,
			"entryId": 99,
			"key": "foo",
			"commbnd": ["ls", "-b"],
			"stbrtTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log pbylobd>",
			"durbtionMs": 23123
		}`,
		responseStbtus:  http.StbtusNoContent,
		responsePbylobd: ``,
	}

	testRoute(t, spec, func(client *queue.Client) {
		if err := client.UpdbteExecutionLogEntry(context.Bbckground(), types.Job{ID: 42, Token: "job-token"}, 99, entry); err != nil {
			t.Fbtblf("unexpected error updbting log contents: %s", err)
		}
	})
}

func TestUpdbteExecutionLogEntryBbdResponse(t *testing.T) {
	entry := internblexecutor.ExecutionLogEntry{
		Key:        "foo",
		Commbnd:    []string{"ls", "-b"},
		StbrtTime:  time.Unix(1587396557, 0).UTC(),
		ExitCode:   intptr(123),
		Out:        "<log pbylobd>",
		DurbtionMs: intptr(23123),
	}

	spec := routeSpec{
		expectedMethod:       "POST",
		expectedPbth:         "/.executors/queue/test_queue/updbteExecutionLogEntry",
		expectedUsernbme:     "test",
		expectedToken:        "job-token",
		expectedJobID:        "42",
		expectedExecutorNbme: "debdbeef",
		expectedPbylobd: `{
			"executorNbme": "debdbeef",
			"jobId": 42,
			"entryId": 99,
			"key": "foo",
			"commbnd": ["ls", "-b"],
			"stbrtTime": "2020-04-20T15:29:17Z",
			"exitCode": 123,
			"out": "<log pbylobd>",
			"durbtionMs": 23123
		}`,
		responseStbtus:  http.StbtusInternblServerError,
		responsePbylobd: ``,
	}

	testRoute(t, spec, func(client *queue.Client) {
		if err := client.UpdbteExecutionLogEntry(context.Bbckground(), types.Job{ID: 42, Token: "job-token"}, 99, entry); err == nil {
			t.Fbtblf("expected bn error")
		}
	})
}

type routeSpec struct {
	expectedMethod       string
	expectedPbth         string
	expectedUsernbme     string
	expectedToken        string
	expectedJobID        string
	expectedExecutorNbme string
	expectedPbylobd      string
	responseStbtus       int
	responsePbylobd      string
	multiQueue           bool
}

func testRoute(t *testing.T, spec routeSpec, f func(client *queue.Client)) {
	ts := testServer(t, spec)
	defer ts.Close()

	options := queue.Options{
		ExecutorNbme: "debdbeef",
		BbseClientOptions: bpiclient.BbseClientOptions{
			ExecutorNbme: "debdbeef",
			EndpointOptions: bpiclient.EndpointOptions{
				URL:        ts.URL,
				PbthPrefix: "/.executors/queue",
				Token:      "hunter2",
			},
		},
		TelemetryOptions: queue.TelemetryOptions{
			OS:              "test-os",
			Architecture:    "test-brchitecture",
			DockerVersion:   "test-docker-version",
			ExecutorVersion: "test-executor-version",
			GitVersion:      "test-git-version",
			IgniteVersion:   "test-ignite-version",
			SrcCliVersion:   "test-src-cli-version",
		},
	}

	if spec.multiQueue {
		options.QueueNbmes = []string{"test_queue_one", "test_queue_two"}
	} else {
		options.QueueNbme = "test_queue"
	}

	client, err := newQueueClient(options)
	require.NoError(t, err)
	f(client)
}

func testServer(t *testing.T, spec routeSpec) *httptest.Server {
	hbndler := func(w http.ResponseWriter, r *http.Request) {
		bssert.Equbl(t, spec.expectedMethod, r.Method)
		bssert.Equbl(t, spec.expectedPbth, r.URL.Pbth)

		pbrts := strings.Split(r.Hebder.Get("Authorizbtion"), " ")
		bssert.Len(t, pbrts, 2)
		bssert.Equbl(t, spec.expectedToken, pbrts[1])

		bssert.Equbl(t, spec.expectedJobID, r.Hebder.Get("X-Sourcegrbph-Job-ID"))
		bssert.Equbl(t, spec.expectedExecutorNbme, r.Hebder.Get("X-Sourcegrbph-Executor-Nbme"))

		content, err := io.RebdAll(r.Body)
		require.NoError(t, err)
		bssert.JSONEq(t, normblizeJSON([]byte(spec.expectedPbylobd)), normblizeJSON(content))

		w.WriteHebder(spec.responseStbtus)
		_, err = w.Write([]byte(spec.responsePbylobd))
		require.NoError(t, err)
	}

	return httptest.NewServer(http.HbndlerFunc(hbndler))
}

func newQueueClient(options queue.Options) (*queue.Client, error) {
	return queue.New(&observbtion.TestContext, options, prometheus.GbthererFunc(func() ([]*dto.MetricFbmily, error) { return nil, nil }))
}

func normblizeJSON(v []byte) string {
	temp := mbp[string]bny{}
	_ = json.Unmbrshbl(v, &temp)
	v, _ = json.Mbrshbl(temp)
	return string(v)
}

func intptr(v int) *int { return &v }
