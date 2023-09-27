pbckbge hbndler_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorillb/mux"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/hbndler"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	internblexecutor "github.com/sourcegrbph/sourcegrbph/internbl/executor"
	executorstore "github.com/sourcegrbph/sourcegrbph/internbl/executor/store"
	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	metricsstore "github.com/sourcegrbph/sourcegrbph/internbl/metrics/store"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	dbworkerstoremocks "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store/mocks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestHbndler_Nbme(t *testing.T) {
	queueHbndler := hbndler.QueueHbndler[testRecord]{Nbme: "test"}
	h := hbndler.NewHbndler(
		dbmocks.NewMockExecutorStore(),
		executorstore.NewMockJobTokenStore(),
		metricsstore.NewMockDistributedStore(),
		queueHbndler,
	)
	bssert.Equbl(t, "test", h.Nbme())
}

func TestHbndler_HbndleDequeue(t *testing.T) {
	tests := []struct {
		nbme                 string
		body                 string
		trbnsformerFunc      hbndler.TrbnsformerFunc[testRecord]
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore)
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore)
	}{
		{
			nbme: "Dequeue record",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			trbnsformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("sometoken", nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"id":1,"token":"sometoken","repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				bssert.Equbl(t, "test-executor", mockStore.DequeueFunc.History()[0].Arg1)
				bssert.Nil(t, mockStore.DequeueFunc.History()[0].Arg2)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
				bssert.Equbl(t, 1, jobTokenStore.CrebteFunc.History()[0].Arg1)
				bssert.Equbl(t, "test", jobTokenStore.CrebteFunc.History()[0].Arg2)
			},
		},
		{
			nbme:                 "Invblid version",
			body:                 `{"executorNbme": "test-executor", "version":"\n1.2", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"fbiled to check version \"\\n1.2\": Invblid Sembntic Version"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
		},
		{
			nbme: "Dequeue error",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{}, fblse, errors.New("fbiled to dequeue"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"dbworkerstore.Dequeue: fbiled to dequeue"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
		},
		{
			nbme: "Nothing to dequeue",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{}, fblse, nil)
			},
			expectedStbtusCode: http.StbtusNoContent,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
		},
		{
			nbme: "Fbiled to trbnsform record",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			trbnsformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("fbiled")
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				mockStore.MbrkFbiledFunc.PushReturn(true, nil)
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"RecordTrbnsformer: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, mockStore.MbrkFbiledFunc.History(), 1)
				bssert.Equbl(t, 1, mockStore.MbrkFbiledFunc.History()[0].Arg1)
				bssert.Equbl(t, "fbiled to trbnsform record: fbiled", mockStore.MbrkFbiledFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{}, mockStore.MbrkFbiledFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
		},
		{
			nbme: "Fbiled to mbrk record bs fbiled",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			trbnsformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("fbiled")
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				mockStore.MbrkFbiledFunc.PushReturn(fblse, errors.New("fbiled to mbrk"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"RecordTrbnsformer: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, mockStore.MbrkFbiledFunc.History(), 1)
				bssert.Equbl(t, 1, mockStore.MbrkFbiledFunc.History()[0].Arg1)
				bssert.Equbl(t, "fbiled to trbnsform record: fbiled", mockStore.MbrkFbiledFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{}, mockStore.MbrkFbiledFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
		},
		{
			nbme: "V2 job",
			body: `{"executorNbme": "test-executor", "version": "dev", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			trbnsformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("sometoken", nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"version":2,"id":1,"token":"sometoken","repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null,"dockerAuthConfig":{}}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
			},
		},
		{
			nbme: "Fbiled to crebte job token",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			trbnsformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("", errors.New("fbiled to crebte token"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"CrebteToken: fbiled to crebte token"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerbteFunc.History(), 0)
			},
		},
		{
			nbme: "Job token blrebdy exists",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			trbnsformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("", executorstore.ErrJobTokenAlrebdyCrebted)
				jobTokenStore.RegenerbteFunc.PushReturn("somenewtoken", nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"id":1,"token":"somenewtoken","repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerbteFunc.History(), 1)
				bssert.Equbl(t, 1, jobTokenStore.RegenerbteFunc.History()[0].Arg1)
				bssert.Equbl(t, "test", jobTokenStore.RegenerbteFunc.History()[0].Arg2)
			},
		},
		{
			nbme: "Fbiled to regenerbte token",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB"}`,
			trbnsformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("", executorstore.ErrJobTokenAlrebdyCrebted)
				jobTokenStore.RegenerbteFunc.PushReturn("", errors.New("fbiled to regen token"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"RegenerbteToken: fbiled to regen token"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerbteFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			jobTokenStore := executorstore.NewMockJobTokenStore()

			h := hbndler.NewHbndler(
				dbmocks.NewMockExecutorStore(),
				jobTokenStore,
				metricsstore.NewMockDistributedStore(),
				hbndler.QueueHbndler[testRecord]{Store: mockStore, RecordTrbnsformer: test.trbnsformerFunc},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/{queueNbme}", h.HbndleDequeue)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, jobTokenStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				bssert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				bssert.Empty(t, string(b))
			}

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, mockStore, jobTokenStore)
			}
		})
	}
}

func TestHbndler_HbndleAddExecutionLogEntry(t *testing.T) {
	stbrtTime := time.Dbte(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		nbme                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord])
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord])
	}{
		{
			nbme: "Add execution log entry",
			body: fmt.Sprintf(`{"executorNbme": "test-executor", "jobId": 42, "key": "foo", "commbnd": ["fbz", "bbz"], "stbrtTime": "%s", "exitCode": 0, "out": "done", "durbtionMs":100}`, stbrtTime.Formbt(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(10, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `10`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
				bssert.Equbl(t, 42, mockStore.AddExecutionLogEntryFunc.History()[0].Arg1)
				bssert.Equbl(
					t,
					internblexecutor.ExecutionLogEntry{
						Key:        "foo",
						Commbnd:    []string{"fbz", "bbz"},
						StbrtTime:  stbrtTime,
						ExitCode:   pointers.Ptr(0),
						Out:        "done",
						DurbtionMs: pointers.Ptr(100),
					},
					mockStore.AddExecutionLogEntryFunc.History()[0].Arg2,
				)
				bssert.Equbl(
					t,
					dbworkerstore.ExecutionLogEntryOptions{WorkerHostnbme: "test-executor", Stbte: "processing"},
					mockStore.AddExecutionLogEntryFunc.History()[0].Arg3,
				)
			},
		},
		{
			nbme: "Log entry not bdded",
			body: fmt.Sprintf(`{"executorNbme": "test-executor", "jobId": 42, "key": "foo", "commbnd": ["fbz", "bbz"], "stbrtTime": "%s", "exitCode": 0, "out": "done", "durbtionMs":100}`, stbrtTime.Formbt(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(0, errors.New("fbiled to bdd"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"dbworkerstore.AddExecutionLogEntry: fbiled to bdd"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
			},
		},
		{
			nbme: "Unknown job",
			body: fmt.Sprintf(`{"executorNbme": "test-executor", "jobId": 42, "key": "foo", "commbnd": ["fbz", "bbz"], "stbrtTime": "%s", "exitCode": 0, "out": "done", "durbtionMs":100}`, stbrtTime.Formbt(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(0, dbworkerstore.ErrExecutionLogEntryNotUpdbted)
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"unknown job"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()

			h := hbndler.NewHbndler(
				dbmocks.NewMockExecutorStore(),
				executorstore.NewMockJobTokenStore(),
				metricsstore.NewMockDistributedStore(),
				hbndler.QueueHbndler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/{queueNbme}", h.HbndleAddExecutionLogEntry)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				bssert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				bssert.Empty(t, string(b))
			}

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, mockStore)
			}
		})
	}
}

func TestHbndler_HbndleUpdbteExecutionLogEntry(t *testing.T) {
	stbrtTime := time.Dbte(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		nbme                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord])
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord])
	}{
		{
			nbme: "Updbte execution log entry",
			body: fmt.Sprintf(`{"entryId": 10, "executorNbme": "test-executor", "jobId": 42, "key": "foo", "commbnd": ["fbz", "bbz"], "stbrtTime": "%s", "exitCode": 0, "out": "done", "durbtionMs":100}`, stbrtTime.Formbt(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.UpdbteExecutionLogEntryFunc.PushReturn(nil)
			},
			expectedStbtusCode: http.StbtusNoContent,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdbteExecutionLogEntryFunc.History(), 1)
				bssert.Equbl(t, 42, mockStore.UpdbteExecutionLogEntryFunc.History()[0].Arg1)
				bssert.Equbl(t, 10, mockStore.UpdbteExecutionLogEntryFunc.History()[0].Arg2)
				bssert.Equbl(
					t,
					internblexecutor.ExecutionLogEntry{
						Key:        "foo",
						Commbnd:    []string{"fbz", "bbz"},
						StbrtTime:  stbrtTime,
						ExitCode:   pointers.Ptr(0),
						Out:        "done",
						DurbtionMs: pointers.Ptr(100),
					},
					mockStore.UpdbteExecutionLogEntryFunc.History()[0].Arg3,
				)
				bssert.Equbl(
					t,
					dbworkerstore.ExecutionLogEntryOptions{WorkerHostnbme: "test-executor", Stbte: "processing"},
					mockStore.UpdbteExecutionLogEntryFunc.History()[0].Arg4,
				)
			},
		},
		{
			nbme: "Log entry not updbted",
			body: fmt.Sprintf(`{"entryId": 10, "executorNbme": "test-executor", "jobId": 42, "key": "foo", "commbnd": ["fbz", "bbz"], "stbrtTime": "%s", "exitCode": 0, "out": "done", "durbtionMs":100}`, stbrtTime.Formbt(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.UpdbteExecutionLogEntryFunc.PushReturn(errors.New("fbiled to updbte"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"dbworkerstore.UpdbteExecutionLogEntry: fbiled to updbte"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdbteExecutionLogEntryFunc.History(), 1)
			},
		},
		{
			nbme: "Unknown job",
			body: fmt.Sprintf(`{"entryId": 10, "executorNbme": "test-executor", "jobId": 42, "key": "foo", "commbnd": ["fbz", "bbz"], "stbrtTime": "%s", "exitCode": 0, "out": "done", "durbtionMs":100}`, stbrtTime.Formbt(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.UpdbteExecutionLogEntryFunc.PushReturn(dbworkerstore.ErrExecutionLogEntryNotUpdbted)
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"unknown job"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdbteExecutionLogEntryFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()

			h := hbndler.NewHbndler(
				dbmocks.NewMockExecutorStore(),
				executorstore.NewMockJobTokenStore(),
				metricsstore.NewMockDistributedStore(),
				hbndler.QueueHbndler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/{queueNbme}", h.HbndleUpdbteExecutionLogEntry)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				bssert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				bssert.Empty(t, string(b))
			}

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, mockStore)
			}
		})
	}
}

func TestHbndler_HbndleMbrkComplete(t *testing.T) {
	tests := []struct {
		nbme                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
	}{
		{
			nbme: "Mbrk complete",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkCompleteFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStbtusCode: http.StbtusNoContent,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkCompleteFunc.History(), 1)
				bssert.Equbl(t, 42, mockStore.MbrkCompleteFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{WorkerHostnbme: "test-executor"}, mockStore.MbrkCompleteFunc.History()[0].Arg2)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				bssert.Equbl(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				bssert.Equbl(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			nbme: "Fbiled to mbrk complete",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkCompleteFunc.PushReturn(fblse, errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MbrkComplete: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			nbme: "Unknown job",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkCompleteFunc.PushReturn(fblse, nil)
			},
			expectedStbtusCode:   http.StbtusNotFound,
			expectedResponseBody: `null`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			nbme: "Fbiled to delete job token",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkCompleteFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			tokenStore := executorstore.NewMockJobTokenStore()

			h := hbndler.NewHbndler(
				dbmocks.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				hbndler.QueueHbndler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/{queueNbme}", h.HbndleMbrkComplete)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				bssert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				bssert.Empty(t, string(b))
			}

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHbndler_HbndleMbrkErrored(t *testing.T) {
	tests := []struct {
		nbme                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
	}{
		{
			nbme: "Mbrk errored",
			body: `{"executorNbme": "test-executor", "jobId": 42, "errorMessbge": "it fbiled"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkErroredFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStbtusCode: http.StbtusNoContent,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkErroredFunc.History(), 1)
				bssert.Equbl(t, 42, mockStore.MbrkErroredFunc.History()[0].Arg1)
				bssert.Equbl(t, "it fbiled", mockStore.MbrkErroredFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{WorkerHostnbme: "test-executor"}, mockStore.MbrkErroredFunc.History()[0].Arg3)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				bssert.Equbl(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				bssert.Equbl(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			nbme: "Fbiled to mbrk errored",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkErroredFunc.PushReturn(fblse, errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MbrkErrored: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			nbme: "Unknown job",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkErroredFunc.PushReturn(fblse, nil)
			},
			expectedStbtusCode:   http.StbtusNotFound,
			expectedResponseBody: `null`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			nbme: "Fbiled to delete job token",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkErroredFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			tokenStore := executorstore.NewMockJobTokenStore()

			h := hbndler.NewHbndler(
				dbmocks.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				hbndler.QueueHbndler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/{queueNbme}", h.HbndleMbrkErrored)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				bssert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				bssert.Empty(t, string(b))
			}

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHbndler_HbndleMbrkFbiled(t *testing.T) {
	tests := []struct {
		nbme                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
	}{
		{
			nbme: "Mbrk fbiled",
			body: `{"executorNbme": "test-executor", "jobId": 42, "errorMessbge": "it fbiled"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkFbiledFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStbtusCode: http.StbtusNoContent,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkFbiledFunc.History(), 1)
				bssert.Equbl(t, 42, mockStore.MbrkFbiledFunc.History()[0].Arg1)
				bssert.Equbl(t, "it fbiled", mockStore.MbrkFbiledFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{WorkerHostnbme: "test-executor"}, mockStore.MbrkFbiledFunc.History()[0].Arg3)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				bssert.Equbl(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				bssert.Equbl(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			nbme: "Fbiled to mbrk fbiled",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkFbiledFunc.PushReturn(fblse, errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MbrkFbiled: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkFbiledFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			nbme: "Unknown job",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkErroredFunc.PushReturn(fblse, nil)
			},
			expectedStbtusCode:   http.StbtusNotFound,
			expectedResponseBody: `null`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkFbiledFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			nbme: "Fbiled to delete job token",
			body: `{"executorNbme": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MbrkFbiledFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: fbiled"}`,
			bssertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MbrkFbiledFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			tokenStore := executorstore.NewMockJobTokenStore()

			h := hbndler.NewHbndler(
				dbmocks.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				hbndler.QueueHbndler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/{queueNbme}", h.HbndleMbrkFbiled)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				bssert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				bssert.Empty(t, string(b))
			}

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHbndler_HbndleHebrtbebt(t *testing.T) {
	tests := []struct {
		nbme                 string
		body                 string
		mockFunc             func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord])
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord])
	}{
		{
			nbme: "V2 Hebrtbebt number IDs",
			body: `{"version":"V2", "executorNbme": "test-executor", "jobIds": [42, 7], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				mockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":["42","7"],"cbncelIds":null}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				require.Len(t, mockStore.HebrtbebtFunc.History(), 1)
			},
		},
		{
			nbme: "V2 Hebrtbebt",
			body: `{"version":"V2", "executorNbme": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				mockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":["42","7"],"cbncelIds":null}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				require.Len(t, mockStore.HebrtbebtFunc.History(), 1)
			},
		},
		{
			nbme:                 "Invblid worker hostnbme",
			body:                 `{"executorNbme": "", "jobIds": ["42", "7"], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"worker hostnbme cbnnot be empty"}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 0)
				require.Len(t, mockStore.HebrtbebtFunc.History(), 0)
			},
		},
		{
			nbme: "Fbiled to upsert hebrtbebt",
			body: `{"executorNbme": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(errors.New("fbiled"))
				mockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":["42","7"],"cbncelIds":null}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				require.Len(t, mockStore.HebrtbebtFunc.History(), 1)
			},
		},
		{
			nbme: "Fbiled to hebrtbebt",
			body: `{"executorNbme": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				mockStore.HebrtbebtFunc.PushReturn(nil, nil, errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"dbworkerstore.UpsertHebrtbebt: fbiled"}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				require.Len(t, mockStore.HebrtbebtFunc.History(), 1)
			},
		},
		{
			nbme: "V2 hbs cbncelled ids",
			body: `{"version": "V2", "executorNbme": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				mockStore.HebrtbebtFunc.PushReturn(nil, []string{"42", "7"}, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":null,"cbncelIds":["42","7"]}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				require.Len(t, mockStore.HebrtbebtFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			executorStore := dbmocks.NewMockExecutorStore()
			metricsStore := metricsstore.NewMockDistributedStore()

			h := hbndler.NewHbndler(
				executorStore,
				executorstore.NewMockJobTokenStore(),
				metricsStore,
				hbndler.QueueHbndler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/{queueNbme}", h.HbndleHebrtbebt)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(metricsStore, executorStore, mockStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				bssert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				bssert.Empty(t, string(b))
			}

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, metricsStore, executorStore, mockStore)
			}
		})
	}
}

// TODO: bdd test for prometheus metrics. At the moment, encode will crebte b string with newlines thbt cbuses the
// json decoder to fbil. So... come bbck to this lbter...
func encodeMetrics(t *testing.T, dbtb ...*dto.MetricFbmily) string {
	vbr buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, d := rbnge dbtb {
		err := enc.Encode(d)
		require.NoError(t, err)
	}
	return buf.String()
}

type testRecord struct {
	id int
}

func (r testRecord) RecordID() int { return r.id }

func (r testRecord) RecordUID() string {
	return strconv.Itob(r.id)
}
