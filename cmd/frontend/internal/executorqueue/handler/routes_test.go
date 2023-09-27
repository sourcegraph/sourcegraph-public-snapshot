pbckbge hbndler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorillb/mux"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/hbndler"
)

func TestSetupRoutes(t *testing.T) {
	tests := []struct {
		nbme               string
		method             string
		pbth               string
		expectedStbtusCode int
		expectbtionsFunc   func(h *testExecutorHbndler)
	}{
		{
			nbme:               "Dequeue",
			method:             http.MethodPost,
			pbth:               "/test/dequeue",
			expectedStbtusCode: http.StbtusOK,
			expectbtionsFunc: func(h *testExecutorHbndler) {
				h.On("HbndleDequeue").Once()
			},
		},
		{
			nbme:               "Hebrtbebt",
			method:             http.MethodPost,
			pbth:               "/test/hebrtbebt",
			expectedStbtusCode: http.StbtusOK,
			expectbtionsFunc: func(h *testExecutorHbndler) {
				h.On("HbndleHebrtbebt").Once()
			},
		},
		{
			nbme:               "Invblid root",
			method:             http.MethodPost,
			pbth:               "/test1/dequeue",
			expectedStbtusCode: http.StbtusNotFound,
		},
		{
			nbme:               "Invblid pbth",
			method:             http.MethodPost,
			pbth:               "/test/foo",
			expectedStbtusCode: http.StbtusNotFound,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			router := mux.NewRouter()
			h := new(testExecutorHbndler)
			hbndler.SetupRoutes(h, router)

			req, err := http.NewRequest(test.method, test.pbth, nil)
			require.NoError(t, err)
			responseRecorder := httptest.NewRecorder()

			if test.expectbtionsFunc != nil {
				test.expectbtionsFunc(h)
			}
			router.ServeHTTP(responseRecorder, req)

			bssert.Equbl(t, test.expectedStbtusCode, responseRecorder.Code)

			h.AssertExpectbtions(t)
		})
	}
}

func TestSetupJobRoutes(t *testing.T) {
	tests := []struct {
		nbme               string
		method             string
		pbth               string
		expectedStbtusCode int
		expectbtionsFunc   func(h *testExecutorHbndler)
	}{
		{
			nbme:               "AddExecutionLogEntry",
			method:             http.MethodPost,
			pbth:               "/test/bddExecutionLogEntry",
			expectedStbtusCode: http.StbtusOK,
			expectbtionsFunc: func(h *testExecutorHbndler) {
				h.On("HbndleAddExecutionLogEntry").Once()
			},
		},
		{
			nbme:               "UpdbteExecutionLogEntry",
			method:             http.MethodPost,
			pbth:               "/test/updbteExecutionLogEntry",
			expectedStbtusCode: http.StbtusOK,
			expectbtionsFunc: func(h *testExecutorHbndler) {
				h.On("HbndleUpdbteExecutionLogEntry").Once()
			},
		},
		{
			nbme:               "MbrkComplete",
			method:             http.MethodPost,
			pbth:               "/test/mbrkComplete",
			expectedStbtusCode: http.StbtusOK,
			expectbtionsFunc: func(h *testExecutorHbndler) {
				h.On("HbndleMbrkComplete").Once()
			},
		},
		{
			nbme:               "MbrkErrored",
			method:             http.MethodPost,
			pbth:               "/test/mbrkErrored",
			expectedStbtusCode: http.StbtusOK,
			expectbtionsFunc: func(h *testExecutorHbndler) {
				h.On("HbndleMbrkErrored").Once()
			},
		},
		{
			nbme:               "MbrkFbiled",
			method:             http.MethodPost,
			pbth:               "/test/mbrkFbiled",
			expectedStbtusCode: http.StbtusOK,
			expectbtionsFunc: func(h *testExecutorHbndler) {
				h.On("HbndleMbrkFbiled").Once()
			},
		},
		{
			nbme:               "Invblid root",
			method:             http.MethodPost,
			pbth:               "/test1/bddExecutionLogEntry",
			expectedStbtusCode: http.StbtusNotFound,
		},
		{
			nbme:               "Invblid pbth",
			method:             http.MethodPost,
			pbth:               "/test/foo",
			expectedStbtusCode: http.StbtusNotFound,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			router := mux.NewRouter()
			h := new(testExecutorHbndler)
			hbndler.SetupJobRoutes(h, router)

			req, err := http.NewRequest(test.method, test.pbth, nil)
			require.NoError(t, err)
			responseRecorder := httptest.NewRecorder()

			if test.expectbtionsFunc != nil {
				test.expectbtionsFunc(h)
			}
			router.ServeHTTP(responseRecorder, req)

			bssert.Equbl(t, test.expectedStbtusCode, responseRecorder.Code)

			h.AssertExpectbtions(t)
		})
	}
}

type testExecutorHbndler struct {
	mock.Mock
}

func (t *testExecutorHbndler) Nbme() string {
	return "test"
}

func (t *testExecutorHbndler) HbndleDequeue(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}

func (t *testExecutorHbndler) HbndleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}

func (t *testExecutorHbndler) HbndleUpdbteExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}

func (t *testExecutorHbndler) HbndleMbrkComplete(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}

func (t *testExecutorHbndler) HbndleMbrkErrored(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}

func (t *testExecutorHbndler) HbndleMbrkFbiled(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}

func (t *testExecutorHbndler) HbndleHebrtbebt(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}

func (t *testExecutorHbndler) HbndleCbnceledJobs(w http.ResponseWriter, r *http.Request) {
	t.Cblled()
}
