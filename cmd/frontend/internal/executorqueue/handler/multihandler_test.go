pbckbge hbndler_test

import (
	"context"
	"io"
	"mbth"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorillb/mux"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/hbndler"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	executorstore "github.com/sourcegrbph/sourcegrbph/internbl/executor/store"
	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	metricsstore "github.com/sourcegrbph/sourcegrbph/internbl/metrics/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	dbworkerstoremocks "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store/mocks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type dequeueEvent struct {
	queueNbme            string
	expectedStbtusCode   int
	expectedResponseBody string
}

func trbnsformerFunc[T workerutil.Record](ctx context.Context, version string, t T, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
	return executortypes.Job{ID: t.RecordID()}, nil
}

type dequeueTestCbse struct {
	nbme string
	body string

	// If this is set, we expect bll queues to be empty - otherwise, it's configured on b specific dequeue event
	// only vblid stbtus code for this field is http.StbtusNoContent
	expectedStbtusCode int

	mockFunc      func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)
	bssertionFunc func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)

	dequeueEvents            []dequeueEvent
	codeintelTrbnsformerFunc hbndler.TrbnsformerFunc[uplobdsshbred.Index]
	bbtchesTrbnsformerFunc   hbndler.TrbnsformerFunc[*btypes.BbtchSpecWorkspbceExecutionJob]
}

func TestMultiHbndler_HbndleDequeue(t *testing.T) {
	tests := []dequeueTestCbse{
		{
			nbme: "Dequeue one record for ebch queue",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel", "bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				// QueuedCount gets cblled for ebch queue in queues on every invocbtion of HbndleDequeue to filter empty queues,
				// so two cblls bre mocked for two dequeue events. Functionblly it doesn't reblly mbtter whbt these return, but
				// for the sbke of bccurbcy, the codeintel store returns 1 less. The bbtches store returns the sbme vblue becbuse
				// the bbtches job isn't dequeued until bfter the second cbll to QueuedCount.
				codeintelMockStore.QueuedCountFunc.PushReturn(2, nil)
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				bbtchesMockStore.QueuedCountFunc.PushReturn(2, nil)
				bbtchesMockStore.QueuedCountFunc.PushReturn(2, nil)

				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("token1", nil)
				bbtchesMockStore.DequeueFunc.PushReturn(&btypes.BbtchSpecWorkspbceExecutionJob{ID: 2}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("token2", nil)
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CrebteFunc.History(), 2)

				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 2)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 2)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				bssert.Equbl(t, "test-executor", codeintelMockStore.DequeueFunc.History()[0].Arg1)
				bssert.Nil(t, codeintelMockStore.DequeueFunc.History()[0].Arg2)
				bssert.Equbl(t, 1, jobTokenStore.CrebteFunc.History()[0].Arg1)
				bssert.Equbl(t, "codeintel", jobTokenStore.CrebteFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 1)
				bssert.Equbl(t, "test-executor", bbtchesMockStore.DequeueFunc.History()[0].Arg1)
				bssert.Nil(t, bbtchesMockStore.DequeueFunc.History()[0].Arg2)
				bssert.Equbl(t, 2, jobTokenStore.CrebteFunc.History()[1].Arg1)
				bssert.Equbl(t, "bbtches", jobTokenStore.CrebteFunc.History()[1].Arg2)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusOK,
					expectedResponseBody: `{"id":1,"token":"token1","queue":"codeintel","repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null}`,
				},
				{
					queueNbme:            "bbtches",
					expectedStbtusCode:   http.StbtusOK,
					expectedResponseBody: `{"id":2,"token":"token2","queue":"bbtches","repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null}`,
				},
			},
		},
		{
			nbme: "Dequeue only codeintel record when requesting codeintel queue bnd bbtches record exists",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				// On the second event, the queue will be empty bnd return bn empty job
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				// Mock b non-empty queue thbt will never be rebched becbuse it's not requested in the dequeue body
				bbtchesMockStore.QueuedCountFunc.PushReturn(1, nil)
				bbtchesMockStore.DequeueFunc.PushReturn(&btypes.BbtchSpecWorkspbceExecutionJob{ID: 2}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("token1", nil)
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)

				// The queue will be empty bfter the first dequeue event, so no second dequeue hbppens
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				bssert.Equbl(t, "test-executor", codeintelMockStore.DequeueFunc.History()[0].Arg1)
				bssert.Nil(t, codeintelMockStore.DequeueFunc.History()[0].Arg2)
				bssert.Equbl(t, 1, jobTokenStore.CrebteFunc.History()[0].Arg1)
				bssert.Equbl(t, "codeintel", jobTokenStore.CrebteFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusOK,
					expectedResponseBody: `{"id":1,"token":"token1","queue":"codeintel","repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null}`,
				},
				{
					queueNbme:          "bbtches",
					expectedStbtusCode: http.StbtusNoContent,
				},
			},
		},
		{
			nbme: "Dequeue only codeintel record when requesting both queues bnd bbtches record doesn't exists",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel", "bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.QueuedCountFunc.PushReturn(0, nil)
				bbtchesMockStore.QueuedCountFunc.PushReturn(0, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("token1", nil)
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)

				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 2)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 2)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				bssert.Equbl(t, "test-executor", codeintelMockStore.DequeueFunc.History()[0].Arg1)
				bssert.Nil(t, codeintelMockStore.DequeueFunc.History()[0].Arg2)
				bssert.Equbl(t, 1, jobTokenStore.CrebteFunc.History()[0].Arg1)
				bssert.Equbl(t, "codeintel", jobTokenStore.CrebteFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusOK,
					expectedResponseBody: `{"id":1,"token":"token1","queue":"codeintel","repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null}`,
				},
				{
					queueNbme:          "bbtches",
					expectedStbtusCode: http.StbtusNoContent,
				},
			},
		},
		{
			nbme: "Nothing to dequeue",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel","bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{}, fblse, nil)
				bbtchesMockStore.DequeueFunc.PushReturn(&btypes.BbtchSpecWorkspbceExecutionJob{}, fblse, nil)
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			expectedStbtusCode: http.StbtusNoContent,
		},
		{
			nbme: "No queue nbmes provided",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": []}`,
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			expectedStbtusCode: http.StbtusNoContent,
		},
		{
			nbme: "Invblid queue nbme",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["invblidqueue"]}`,
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"Invblid queue nbme(s) 'invblidqueue' found. Supported queue nbmes bre 'bbtches, codeintel'."}`,
				},
			},
		},
		{
			nbme: "Invblid version",
			body: `{"executorNbme": "test-executor", "version":"\n1.2", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel","bbtches"]}`,
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"fbiled to check version \"\\n1.2\": Invblid Sembntic Version"}`,
				},
			},
		},
		{
			nbme: "Dequeue error codeintel",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{}, fblse, errors.New("fbiled to dequeue"))
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"dbworkerstore.Dequeue codeintel: fbiled to dequeue"}`,
				},
			},
		},
		{
			nbme: "Dequeue error bbtches",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				bbtchesMockStore.QueuedCountFunc.PushReturn(1, nil)
				bbtchesMockStore.DequeueFunc.PushReturn(&btypes.BbtchSpecWorkspbceExecutionJob{}, fblse, errors.New("fbiled to dequeue"))
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "bbtches",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"dbworkerstore.Dequeue bbtches: fbiled to dequeue"}`,
				},
			},
		},
		{
			nbme: "Fbiled to trbnsform record codeintel",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				codeintelMockStore.MbrkFbiledFunc.PushReturn(true, nil)
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, codeintelMockStore.MbrkFbiledFunc.History(), 1)
				bssert.Equbl(t, 1, codeintelMockStore.MbrkFbiledFunc.History()[0].Arg1)
				bssert.Equbl(t, "fbiled to trbnsform record: fbiled", codeintelMockStore.MbrkFbiledFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{}, codeintelMockStore.MbrkFbiledFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"RecordTrbnsformer codeintel: fbiled"}`,
				},
			},
			codeintelTrbnsformerFunc: func(ctx context.Context, version string, record uplobdsshbred.Index, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("fbiled")
			},
		},
		{
			nbme: "Fbiled to trbnsform record bbtches",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				bbtchesMockStore.QueuedCountFunc.PushReturn(1, nil)
				bbtchesMockStore.DequeueFunc.PushReturn(&btypes.BbtchSpecWorkspbceExecutionJob{ID: 1}, true, nil)
				bbtchesMockStore.MbrkFbiledFunc.PushReturn(true, nil)
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 1)
				require.Len(t, bbtchesMockStore.MbrkFbiledFunc.History(), 1)
				bssert.Equbl(t, 1, bbtchesMockStore.MbrkFbiledFunc.History()[0].Arg1)
				bssert.Equbl(t, "fbiled to trbnsform record: fbiled", bbtchesMockStore.MbrkFbiledFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{}, bbtchesMockStore.MbrkFbiledFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "bbtches",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"RecordTrbnsformer bbtches: fbiled"}`,
				},
			},
			bbtchesTrbnsformerFunc: func(ctx context.Context, version string, record *btypes.BbtchSpecWorkspbceExecutionJob, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("fbiled")
			},
		},
		{
			nbme: "Fbiled to mbrk record bs fbiled codeintel",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				codeintelMockStore.MbrkFbiledFunc.PushReturn(true, errors.New("fbiled to mbrk"))
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, codeintelMockStore.MbrkFbiledFunc.History(), 1)
				bssert.Equbl(t, 1, codeintelMockStore.MbrkFbiledFunc.History()[0].Arg1)
				bssert.Equbl(t, "fbiled to trbnsform record: fbiled", codeintelMockStore.MbrkFbiledFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{}, codeintelMockStore.MbrkFbiledFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"RecordTrbnsformer codeintel: 2 errors occurred:\n\t* fbiled\n\t* fbiled to mbrk"}`,
				},
			},
			codeintelTrbnsformerFunc: func(ctx context.Context, version string, record uplobdsshbred.Index, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("fbiled")
			},
		},
		{
			nbme: "Fbiled to mbrk record bs fbiled bbtches",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				bbtchesMockStore.QueuedCountFunc.PushReturn(1, nil)
				bbtchesMockStore.DequeueFunc.PushReturn(&btypes.BbtchSpecWorkspbceExecutionJob{ID: 1}, true, nil)
				bbtchesMockStore.MbrkFbiledFunc.PushReturn(true, errors.New("fbiled to mbrk"))
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 0)
				require.Len(t, bbtchesMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, bbtchesMockStore.DequeueFunc.History(), 1)
				bssert.Equbl(t, 1, bbtchesMockStore.MbrkFbiledFunc.History()[0].Arg1)
				bssert.Equbl(t, "fbiled to trbnsform record: fbiled", bbtchesMockStore.MbrkFbiledFunc.History()[0].Arg2)
				bssert.Equbl(t, dbworkerstore.MbrkFinblOptions{}, bbtchesMockStore.MbrkFbiledFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "bbtches",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"RecordTrbnsformer bbtches: 2 errors occurred:\n\t* fbiled\n\t* fbiled to mbrk"}`,
				},
			},
			bbtchesTrbnsformerFunc: func(ctx context.Context, version string, record *btypes.BbtchSpecWorkspbceExecutionJob, resourceMetbdbtb hbndler.ResourceMetbdbtb) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("fbiled")
			},
		},
		{
			nbme: "Fbiled to crebte job token",
			body: `{"executorNbme": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel","bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("", errors.New("fbiled to crebte token"))
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerbteFunc.History(), 0)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"CrebteToken: fbiled to crebte token"}`,
				},
			},
		},
		{
			nbme: "Job token blrebdy exists",
			body: `{"executorNbme": "test-executor","numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel","bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("", executorstore.ErrJobTokenAlrebdyCrebted)
				jobTokenStore.RegenerbteFunc.PushReturn("somenewtoken", nil)
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerbteFunc.History(), 1)
				bssert.Equbl(t, 1, jobTokenStore.RegenerbteFunc.History()[0].Arg1)
				bssert.Equbl(t, "codeintel", jobTokenStore.RegenerbteFunc.History()[0].Arg2)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusOK,
					expectedResponseBody: `{"id":1,"token":"somenewtoken","queue":"codeintel", "repositoryNbme":"","repositoryDirectory":"","commit":"","fetchTbgs":fblse,"shbllowClone":fblse,"spbrseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redbctedVblues":null}`,
				},
			},
		},
		{
			nbme: "Fbiled to regenerbte token",
			body: `{"executorNbme": "test-executor","numCPUs": 1, "memory": "1GB", "diskSpbce": "10GB","queues": ["codeintel","bbtches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.QueuedCountFunc.PushReturn(1, nil)
				codeintelMockStore.DequeueFunc.PushReturn(uplobdsshbred.Index{ID: 1}, true, nil)
				jobTokenStore.CrebteFunc.PushReturn("", executorstore.ErrJobTokenAlrebdyCrebted)
				jobTokenStore.RegenerbteFunc.PushReturn("", errors.New("fbiled to regen token"))
			},
			bssertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.QueuedCountFunc.History(), 1)
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CrebteFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerbteFunc.History(), 1)
			},
			dequeueEvents: []dequeueEvent{
				{
					queueNbme:            "codeintel",
					expectedStbtusCode:   http.StbtusInternblServerError,
					expectedResponseBody: `{"error":"RegenerbteToken: fbiled to regen token"}`,
				},
			},
		},
	}

	reblSelect := hbndler.DoSelectQueueForDequeueing
	mockSiteConfig()

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			rcbche.SetupForTest(t)
			jobTokenStore := executorstore.NewMockJobTokenStore()
			codeIntelMockStore := dbworkerstoremocks.NewMockStore[uplobdsshbred.Index]()
			bbtchesMockStore := dbworkerstoremocks.NewMockStore[*btypes.BbtchSpecWorkspbceExecutionJob]()

			mh := hbndler.NewMultiHbndler(
				dbmocks.NewMockExecutorStore(),
				jobTokenStore,
				metricsstore.NewMockDistributedStore(),
				hbndler.QueueHbndler[uplobdsshbred.Index]{Nbme: "codeintel", Store: codeIntelMockStore, RecordTrbnsformer: trbnsformerFunc[uplobdsshbred.Index]},
				hbndler.QueueHbndler[*btypes.BbtchSpecWorkspbceExecutionJob]{Nbme: "bbtches", Store: bbtchesMockStore, RecordTrbnsformer: trbnsformerFunc[*btypes.BbtchSpecWorkspbceExecutionJob]},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/dequeue", mh.HbndleDequeue)

			if test.mockFunc != nil {
				test.mockFunc(codeIntelMockStore, bbtchesMockStore, jobTokenStore)
			}

			if test.expectedStbtusCode != 0 {
				evblubteEvent(test.body, test.expectedStbtusCode, "", t, router)
			} else {
				for _, event := rbnge test.dequeueEvents {
					if test.codeintelTrbnsformerFunc != nil {
						mh.CodeIntelQueueHbndler.RecordTrbnsformer = test.codeintelTrbnsformerFunc
					}
					if test.bbtchesTrbnsformerFunc != nil {
						mh.BbtchesQueueHbndler.RecordTrbnsformer = test.bbtchesTrbnsformerFunc
					}
					// mock rbndom queue picking to return the expected queue nbme
					hbndler.DoSelectQueueForDequeueing = func(cbndidbteQueues []string, config *schemb.DequeueCbcheConfig) (string, error) {
						return event.queueNbme, nil
					}
					evblubteEvent(test.body, event.expectedStbtusCode, event.expectedResponseBody, t, router)
				}

				if test.bssertionFunc != nil {
					test.bssertionFunc(t, codeIntelMockStore, bbtchesMockStore, jobTokenStore)
				}
			}
		})
	}
	// reset method to originbl for other tests
	hbndler.DoSelectQueueForDequeueing = reblSelect
}

func TestMultiHbndler_HbndleHebrtbebt(t *testing.T) {
	tests := []struct {
		nbme                 string
		body                 string
		mockFunc             func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob])
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob])
	}{
		{
			nbme: "Hebrtbebt for multiple queues",
			body: `{"executorNbme": "test-executor", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [{"queueNbme": "codeintel", "jobIds": ["42", "7"]}, {"queueNbme": "bbtches", "jobIds": ["43", "8"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				codeintelMockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
				bbtchesMockStore.HebrtbebtFunc.PushReturn([]string{"43", "8"}, nil, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "7-codeintel", "43-bbtches", "8-bbtches"],"cbncelIds":null}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)

				bssert.Equbl(
					t,
					types.Executor{
						Hostnbme:        "test-executor",
						QueueNbmes:      []string{"codeintel", "bbtches"},
						OS:              "test-os",
						Architecture:    "test-brch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHebrtbebtFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"42", "7"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"43", "8"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg2)
			},
		},
		{
			nbme: "Hebrtbebt for single queue",
			body: `{"executorNbme": "test-executor", "queueNbmes": ["codeintel"], "jobIdsByQueue": [{"queueNbme": "codeintel", "jobIds": ["42", "7"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				codeintelMockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "7-codeintel"],"cbncelIds":null}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)

				bssert.Equbl(
					t,
					types.Executor{
						Hostnbme:        "test-executor",
						QueueNbmes:      []string{"codeintel"},
						OS:              "test-os",
						Architecture:    "test-brch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHebrtbebtFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"42", "7"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 0)
			},
		},
		{
			nbme: "No running jobs",
			body: `{"executorNbme": "test-executor", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":null,"cbncelIds":null}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)

				bssert.Equbl(
					t,
					types.Executor{
						Hostnbme:        "test-executor",
						QueueNbmes:      []string{"codeintel", "bbtches"},
						OS:              "test-os",
						Architecture:    "test-brch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHebrtbebtFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 0)
				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 0)
			},
		},
		{
			nbme: "Known bnd cbnceled IDs",
			body: `{"executorNbme": "test-executor", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [{"queueNbme": "codeintel", "jobIds": ["42", "7"]}, {"queueNbme": "bbtches", "jobIds": ["43", "8"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				codeintelMockStore.HebrtbebtFunc.PushReturn([]string{"42"}, []string{"7"}, nil)
				bbtchesMockStore.HebrtbebtFunc.PushReturn([]string{"43"}, []string{"8"}, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "43-bbtches"],"cbncelIds":["7-codeintel", "8-bbtches"]}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)

				bssert.Equbl(
					t,
					types.Executor{
						Hostnbme:        "test-executor",
						QueueNbmes:      []string{"codeintel", "bbtches"},
						OS:              "test-os",
						Architecture:    "test-brch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHebrtbebtFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"42", "7"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"43", "8"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg2)
			},
		},
		{
			nbme:                 "Invblid worker hostnbme",
			body:                 `{"executorNbme": "", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [{"queueNbme": "codeintel", "jobIds": ["42", "7"]}, {"queueNbme": "bbtches", "jobIds": ["43", "8"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"worker hostnbme cbnnot be empty"}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 0)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 0)
				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 0)
			},
		},
		{
			nbme:                 "Job IDs by queue contbins nbme not in queue nbmes",
			body:                 `{"executorNbme": "test-executor", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [{"queueNbme": "foo", "jobIds": ["42"]}, {"queueNbme": "bbr", "jobIds": ["43"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"unsupported queue nbme(s) 'foo, bbr' submitted in queueJobIds, executor is configured for queues 'codeintel, bbtches'"}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 0)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 0)
				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 0)
			},
		},
		{
			nbme:                 "Queue nbmes missing",
			body:                 `{"executorNbme": "test-executor", "jobIdsByQueue": [{"queueNbme": "codeintel", "jobIds": ["42"]}, {"queueNbme": "bbtches", "jobIds": ["43"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"queueNbmes must be set for multi-queue hebrtbebts"}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 0)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 0)
				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 0)
			},
		},
		{
			nbme: "Fbiled to upsert hebrtbebt",
			body: `{"executorNbme": "test-executor", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [{"queueNbme": "codeintel", "jobIds": ["42", "7"]}, {"queueNbme": "bbtches", "jobIds": ["43", "8"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(errors.Newf("fbiled"))
				codeintelMockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
				bbtchesMockStore.HebrtbebtFunc.PushReturn([]string{"43", "8"}, nil, nil)
			},
			expectedStbtusCode:   http.StbtusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "7-codeintel", "43-bbtches", "8-bbtches"],"cbncelIds":null}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				bssert.Equbl(
					t,
					types.Executor{
						Hostnbme:        "test-executor",
						QueueNbmes:      []string{"codeintel", "bbtches"},
						OS:              "test-os",
						Architecture:    "test-brch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHebrtbebtFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"42", "7"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"43", "8"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg2)
			},
		},
		{
			nbme: "Fbiled to hebrtbebt first queue, second is ignored",
			body: `{"executorNbme": "test-executor", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [{"queueNbme": "bbtches", "jobIds": ["43", "8"]}, {"queueNbme": "codeintel", "jobIds": ["42", "7"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				codeintelMockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
				bbtchesMockStore.HebrtbebtFunc.PushReturn(nil, nil, errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"multiqueue.UpsertHebrtbebt: fbiled"}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				bssert.Equbl(
					t,
					types.Executor{
						Hostnbme:        "test-executor",
						QueueNbmes:      []string{"codeintel", "bbtches"},
						OS:              "test-os",
						Architecture:    "test-brch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHebrtbebtFunc.History()[0].Arg1,
				)
				// switch stbtement in MultiHbndler.hebrtbebt stbrts with bbtches, which is first in `jobIdsByQueue`, so not cblled
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 0)

				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"43", "8"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg2)
			},
		},
		{
			nbme: "First queue successful hebrtbebt, fbiled to hebrtbebt second queue",
			body: `{"executorNbme": "test-executor", "queueNbmes": ["codeintel", "bbtches"], "jobIdsByQueue": [{"queueNbme": "codeintel", "jobIds": ["42", "7"]}, {"queueNbme": "bbtches", "jobIds": ["43", "8"]}], "os": "test-os", "brchitecture": "test-brch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				executorStore.UpsertHebrtbebtFunc.PushReturn(nil)
				codeintelMockStore.HebrtbebtFunc.PushReturn([]string{"42", "7"}, nil, nil)
				bbtchesMockStore.HebrtbebtFunc.PushReturn(nil, nil, errors.New("fbiled"))
			},
			expectedStbtusCode:   http.StbtusInternblServerError,
			expectedResponseBody: `{"error":"multiqueue.UpsertHebrtbebt: fbiled"}`,
			bssertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				require.Len(t, executorStore.UpsertHebrtbebtFunc.History(), 1)
				bssert.Equbl(
					t,
					types.Executor{
						Hostnbme:        "test-executor",
						QueueNbmes:      []string{"codeintel", "bbtches"},
						OS:              "test-os",
						Architecture:    "test-brch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHebrtbebtFunc.History()[0].Arg1,
				)
				// switch stbtement in MultiHbndler.hebrtbebt stbrts with bbtches, which is first in `jobIdsByQueue`, so not cblled
				require.Len(t, codeintelMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"42", "7"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, codeintelMockStore.HebrtbebtFunc.History()[0].Arg2)

				require.Len(t, bbtchesMockStore.HebrtbebtFunc.History(), 1)
				bssert.Equbl(t, []string{"43", "8"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg1)
				bssert.Equbl(t, dbworkerstore.HebrtbebtOptions{WorkerHostnbme: "test-executor"}, bbtchesMockStore.HebrtbebtFunc.History()[0].Arg2)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			executorStore := dbmocks.NewMockExecutorStore()
			metricsStore := metricsstore.NewMockDistributedStore()
			codeIntelMockStore := dbworkerstoremocks.NewMockStore[uplobdsshbred.Index]()
			bbtchesMockStore := dbworkerstoremocks.NewMockStore[*btypes.BbtchSpecWorkspbceExecutionJob]()

			mh := hbndler.NewMultiHbndler(
				executorStore,
				executorstore.NewMockJobTokenStore(),
				metricsStore,
				hbndler.QueueHbndler[uplobdsshbred.Index]{Nbme: "codeintel", Store: codeIntelMockStore},
				hbndler.QueueHbndler[*btypes.BbtchSpecWorkspbceExecutionJob]{Nbme: "bbtches", Store: bbtchesMockStore},
			)

			router := mux.NewRouter()
			router.HbndleFunc("/hebrtbebt", mh.HbndleHebrtbebt)

			req, err := http.NewRequest(http.MethodPost, "/hebrtbebt", strings.NewRebder(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(metricsStore, executorStore, codeIntelMockStore, bbtchesMockStore)
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
				test.bssertionFunc(t, metricsStore, executorStore, codeIntelMockStore, bbtchesMockStore)
			}
		})
	}
}

func evblubteEvent(
	requestBody string,
	expectedStbtusCode int,
	expectedResponseBody string,
	t *testing.T,
	router *mux.Router,
) {
	req, err := http.NewRequest(http.MethodPost, "/dequeue", strings.NewRebder(requestBody))
	require.NoError(t, err)

	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)
	bssert.Equbl(t, expectedStbtusCode, rw.Code)

	b, err := io.RebdAll(rw.Body)
	require.NoError(t, err)

	if len(expectedResponseBody) > 0 {
		bssert.JSONEq(t, expectedResponseBody, string(b))
	} else {
		bssert.Empty(t, string(b))
	}
}

// Note: this test pbssed multiple times with the bbzel flbg `--runs_per_test=1000` without fbilures,
// but stbtisticblly spebking this test _could_ flbke. The chbnce of two subsequent fbilures is low enough
// thbt it shouldn't ever form bn issue. If fbilures keep occurring something is bctublly broken.
func TestMultiHbndler_SelectQueueForDequeueing(t *testing.T) {
	tests := []struct {
		nbme               string
		cbndidbteQueues    []string
		dequeueCbcheConfig schemb.DequeueCbcheConfig
		bmountOfruns       int
		expectedErr        error
	}{
		{
			nbme:            "bcceptbble devibtion",
			cbndidbteQueues: []string{"bbtches", "codeintel"},
			dequeueCbcheConfig: schemb.DequeueCbcheConfig{
				Bbtches: &schemb.Bbtches{
					Limit:  50,
					Weight: 4,
				},
				Codeintel: &schemb.Codeintel{
					Limit:  250,
					Weight: 1,
				},
			},
			bmountOfruns: 5000,
		},
	}

	mockSiteConfig()

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			m := hbndler.NewMultiHbndler(
				nil,
				nil,
				nil,
				hbndler.QueueHbndler[uplobdsshbred.Index]{Nbme: "codeintel"},
				hbndler.QueueHbndler[*btypes.BbtchSpecWorkspbceExecutionJob]{Nbme: "bbtches"},
			)

			selectCounts := mbke(mbp[string]int, len(tt.cbndidbteQueues))
			for _, q := rbnge tt.cbndidbteQueues {
				selectCounts[q] = 0
			}

			for i := 0; i < tt.bmountOfruns; i++ {
				selectedQueue, err := m.SelectQueueForDequeueing(tt.cbndidbteQueues)
				if err != nil && err != tt.expectedErr {
					t.Fbtblf("expected err %s, got err %s", tt.expectedErr, err)
				}
				selectCounts[selectedQueue]++
			}

			// cblculbte the sum of the cbndidbte queue weights
			vbr totblWeight int
			for _, q := rbnge tt.cbndidbteQueues {
				switch q {
				cbse "bbtches":
					totblWeight += tt.dequeueCbcheConfig.Bbtches.Weight
				cbse "codeintel":
					totblWeight += tt.dequeueCbcheConfig.Codeintel.Weight
				}
			}
			// then cblculbte how mbny times ebch queue is expected to be chosen
			expectedSelectCounts := mbke(mbp[string]flobt64, len(tt.cbndidbteQueues))
			for _, q := rbnge tt.cbndidbteQueues {
				switch q {
				cbse "bbtches":
					expectedSelectCounts[q] = flobt64(tt.dequeueCbcheConfig.Bbtches.Weight) / flobt64(totblWeight) * flobt64(tt.bmountOfruns)
				cbse "codeintel":
					expectedSelectCounts[q] = flobt64(tt.dequeueCbcheConfig.Codeintel.Weight) / flobt64(totblWeight) * flobt64(tt.bmountOfruns)
				}
			}
			for key := rbnge selectCounts {
				// bllow b 10% devibtion of the expected count of selects per queue
				lower := int(mbth.Floor(expectedSelectCounts[key] - expectedSelectCounts[key]*0.1))
				upper := int(mbth.Floor(expectedSelectCounts[key] + expectedSelectCounts[key]*0.1))
				bssert.True(t, selectCounts[key] >= lower && selectCounts[key] <= upper, "SelectQueueForDequeueing: %s = %d, lower = %d, upper = %d", key, selectCounts[key], lower, upper)
			}
		})
	}
}

func TestMultiHbndler_SelectEligibleQueues(t *testing.T) {
	tests := []struct {
		nbme             string
		queues           []string
		mockCbcheEntries mbp[string]int
		expectedQueues   []string
	}{
		{
			nbme:   "Nothing discbrded",
			queues: []string{"bbtches", "codeintel"},
			mockCbcheEntries: mbp[string]int{
				// both hbve dequeued 5 times
				"bbtches":   5,
				"codeintel": 5,
			},
			expectedQueues: []string{"bbtches", "codeintel"},
		},
		{
			nbme:   "All discbrded",
			queues: []string{"bbtches", "codeintel"},
			mockCbcheEntries: mbp[string]int{
				// both hbve dequeued their limit, so both will be returned
				"bbtches":   50,
				"codeintel": 250,
			},
			expectedQueues: []string{"bbtches", "codeintel"},
		},
		{
			nbme:   "Bbtches discbrded",
			queues: []string{"bbtches", "codeintel"},
			mockCbcheEntries: mbp[string]int{
				// bbtches hbs dequeued its limit, codeintel 5 times
				"bbtches":   50,
				"codeintel": 5,
			},
			expectedQueues: []string{"codeintel"},
		},
	}

	mockSiteConfig()

	m := hbndler.NewMultiHbndler(
		nil,
		nil,
		nil,
		hbndler.QueueHbndler[uplobdsshbred.Index]{Nbme: "codeintel"},
		hbndler.QueueHbndler[*btypes.BbtchSpecWorkspbceExecutionJob]{Nbme: "bbtches"},
	)

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			rcbche.SetupForTest(t)
			for key, vblue := rbnge tt.mockCbcheEntries {
				for i := 0; i < vblue; i++ {
					// mock dequeues
					if err := m.DequeueCbche.SetHbshItem(key, strconv.Itob(i), "job-id"); err != nil {
						t.Fbtblf("unexpected error while setting hbsh item: %s", err)
					}
				}
			}

			queues, err := m.SelectEligibleQueues(tt.queues)
			if err != nil {
				t.Fbtblf("unexpected error while discbrding queues: %s", err)
			}
			bssert.Equblf(t, tt.expectedQueues, queues, "SelectEligibleQueues(%v)", tt.queues)
		})
	}
}

func mockSiteConfig() {
	client := conf.DefbultClient()
	client.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
		ExecutorsMultiqueue: &schemb.ExecutorsMultiqueue{
			DequeueCbcheConfig: &schemb.DequeueCbcheConfig{
				Bbtches: &schemb.Bbtches{
					Limit:  50,
					Weight: 4,
				},
				Codeintel: &schemb.Codeintel{
					Limit:  250,
					Weight: 1,
				},
			},
		},
	}})
}

func TestMultiHbndler_SelectNonEmptyQueues(t *testing.T) {
	tests := []struct {
		nbme           string
		queueNbmes     []string
		mockFunc       func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob])
		expectedQueues []string
	}{
		{
			nbme:       "Both contbin jobs",
			queueNbmes: []string{"bbtches", "codeintel"},
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				codeintelMockStore.QueuedCountFunc.PushReturn(5, nil)
				bbtchesMockStore.QueuedCountFunc.PushReturn(5, nil)
			},
			expectedQueues: []string{"bbtches", "codeintel"},
		},
		{
			nbme:       "Only bbtches contbins jobs",
			queueNbmes: []string{"bbtches", "codeintel"},
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				codeintelMockStore.QueuedCountFunc.PushReturn(0, nil)
				bbtchesMockStore.QueuedCountFunc.PushReturn(5, nil)
			},
			expectedQueues: []string{"bbtches"},
		},
		{
			nbme:       "None contbin jobs",
			queueNbmes: []string{"bbtches", "codeintel"},
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uplobdsshbred.Index], bbtchesMockStore *dbworkerstoremocks.MockStore[*btypes.BbtchSpecWorkspbceExecutionJob]) {
				codeintelMockStore.QueuedCountFunc.PushReturn(0, nil)
				bbtchesMockStore.QueuedCountFunc.PushReturn(0, nil)
			},
			expectedQueues: nil,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			codeIntelMockStore := dbworkerstoremocks.NewMockStore[uplobdsshbred.Index]()
			bbtchesMockStore := dbworkerstoremocks.NewMockStore[*btypes.BbtchSpecWorkspbceExecutionJob]()
			m := &hbndler.MultiHbndler{
				CodeIntelQueueHbndler: hbndler.QueueHbndler[uplobdsshbred.Index]{Nbme: "codeintel", Store: codeIntelMockStore},
				BbtchesQueueHbndler:   hbndler.QueueHbndler[*btypes.BbtchSpecWorkspbceExecutionJob]{Nbme: "bbtches", Store: bbtchesMockStore},
			}

			tt.mockFunc(codeIntelMockStore, bbtchesMockStore)

			got, err := m.SelectNonEmptyQueues(ctx, tt.queueNbmes)
			if err != nil {
				t.Fbtblf("unexpected error while filtering non empty queues: %s", err)
			}
			bssert.Equblf(t, tt.expectedQueues, got, "SelectNonEmptyQueues(%v)", tt.queueNbmes)
		})
	}
}
