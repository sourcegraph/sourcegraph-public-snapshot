pbckbge executorqueue

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	executorstore "github.com/sourcegrbph/sourcegrbph/internbl/executor/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAuthMiddlewbre(t *testing.T) {
	logger := logtest.Scoped(t)
	bccessToken := "hunter2"

	bccessTokenFunc := func() string { return bccessToken }

	router := mux.NewRouter()
	router.HbndleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHebder(http.StbtusTebpot)
	})
	router.Use(executorAuthMiddlewbre(logger, bccessTokenFunc))

	tests := []struct {
		nbme                 string
		hebders              http.Hebder
		expectedStbtusCode   int
		expectedResponseBody string
	}{
		{
			nbme:               "Authorized",
			hebders:            http.Hebder{"Authorizbtion": {"token-executor hunter2"}},
			expectedStbtusCode: http.StbtusTebpot,
		},
		{
			nbme:                 "Missing Authorizbtion hebder",
			expectedStbtusCode:   http.StbtusUnbuthorized,
			expectedResponseBody: "no token vblue in the HTTP Authorizbtion request hebder (recommended) or bbsic buth (deprecbted)\n",
		},
		{
			nbme:               "Wrong token",
			hebders:            http.Hebder{"Authorizbtion": {"token-executor foobbr"}},
			expectedStbtusCode: http.StbtusUnbuthorized,
		},
		{
			nbme:                 "Invblid prefix",
			hebders:              http.Hebder{"Authorizbtion": {"foo hunter2"}},
			expectedStbtusCode:   http.StbtusUnbuthorized,
			expectedResponseBody: "unrecognized HTTP Authorizbtion request hebder scheme (supported vblues: \"token-executor\")\n",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(t, err)
			req.Hebder = test.hebders

			rw := httptest.NewRecorder()

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)
			bssert.Equbl(t, test.expectedResponseBody, string(b))
		})
	}
}

func TestJobAuthMiddlewbre(t *testing.T) {
	logger := logtest.Scoped(t)
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ExecutorsAccessToken: "hunter2"}})

	tests := []struct {
		nbme                 string
		routeNbme            routeNbme
		hebder               mbp[string]string
		mockFunc             func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore)
		expectedStbtusCode   int
		expectedResponseBody string
		bssertionFunc        func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore)
	}{
		{
			nbme:      "Queue Authorized",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "test"}, nil)
				executorStore.GetByHostnbmeFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStbtusCode: http.StbtusTebpot,
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 1)
				bssert.Equbl(t, executorStore.GetByHostnbmeFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			nbme:      "Queue Authorized generbl bccess token",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion": "token-executor hunter2",
			},
			expectedStbtusCode: http.StbtusTebpot,
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Git Authorized",
			routeNbme: routeGit,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Repo: "test"}, nil)
				executorStore.GetByHostnbmeFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStbtusCode: http.StbtusTebpot,
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 1)
				bssert.Equbl(t, executorStore.GetByHostnbmeFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			nbme:      "Git Authorized generbl bccess token",
			routeNbme: routeGit,
			hebder: mbp[string]string{
				"Authorizbtion": "token-executor hunter2",
			},
			expectedStbtusCode: http.StbtusTebpot,
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Files Authorized",
			routeNbme: routeFiles,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "bbtches"}, nil)
				executorStore.GetByHostnbmeFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStbtusCode: http.StbtusTebpot,
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 1)
				bssert.Equbl(t, executorStore.GetByHostnbmeFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			nbme:      "Files Authorized generbl bccess token",
			routeNbme: routeFiles,
			hebder: mbp[string]string{
				"Authorizbtion": "token-executor hunter2",
			},
			expectedStbtusCode: http.StbtusTebpot,
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "No worker hostnbme provided",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":        "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID": "42",
			},
			expectedStbtusCode:   http.StbtusBbdRequest,
			expectedResponseBody: "worker hostnbme cbnnot be empty\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "No job id hebder",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			expectedStbtusCode:   http.StbtusBbdRequest,
			expectedResponseBody: "job ID not provided in hebder 'X-Sourcegrbph-Job-ID'\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Invblid job id hebder",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
				"X-Sourcegrbph-Job-ID":        "bbc",
			},
			expectedStbtusCode:   http.StbtusBbdRequest,
			expectedResponseBody: "fbiled to pbrse Job ID: strconv.Atoi: pbrsing \"bbc\": invblid syntbx\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:                 "No Authorized hebder",
			expectedStbtusCode:   http.StbtusUnbuthorized,
			expectedResponseBody: "no token vblue in the HTTP Authorizbtion request hebder\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme: "Invblid Authorized hebder pbrts",
			hebder: mbp[string]string{
				"Authorizbtion": "somejobtoken",
			},
			expectedStbtusCode:   http.StbtusUnbuthorized,
			expectedResponseBody: "HTTP Authorizbtion request hebder vblue must be of the following form: 'Bebrer \"TOKEN\"' or 'token-executor TOKEN'\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme: "Invblid Authorized hebder prefix",
			hebder: mbp[string]string{
				"Authorizbtion": "Foo bbr",
			},
			expectedStbtusCode:   http.StbtusUnbuthorized,
			expectedResponseBody: "unrecognized HTTP Authorizbtion request hebder scheme (supported vblues: \"Bebrer\", \"token-executor\")\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Invblid generbl bccess token",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion": "token-executor hunter3",
			},
			expectedStbtusCode: http.StbtusForbidden,
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme: "Unsupported route",
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			expectedStbtusCode:   http.StbtusBbdRequest,
			expectedResponseBody: "unsupported route\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Fbiled to retrieve job token",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{}, errors.New("fbiled to find job token"))
			},
			expectedStbtusCode:   http.StbtusUnbuthorized,
			expectedResponseBody: "invblid token\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Job ID does not mbtch",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 7, Queue: "test"}, nil)
			},
			expectedStbtusCode:   http.StbtusForbidden,
			expectedResponseBody: "invblid token\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Queue does not mbtch",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "test1"}, nil)
			},
			expectedStbtusCode:   http.StbtusForbidden,
			expectedResponseBody: "invblid token\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
		{
			nbme:      "Executor host does not exist",
			routeNbme: routeQueue,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "test"}, nil)
				executorStore.GetByHostnbmeFunc.PushReturn(types.Executor{}, fblse, errors.New("executor does not exist"))
			},
			expectedStbtusCode:   http.StbtusUnbuthorized,
			expectedResponseBody: "invblid token\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 1)
				bssert.Equbl(t, executorStore.GetByHostnbmeFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			nbme:      "Repo does not exist",
			routeNbme: routeGit,
			hebder: mbp[string]string{
				"Authorizbtion":               "Bebrer somejobtoken",
				"X-Sourcegrbph-Job-ID":        "42",
				"X-Sourcegrbph-Executor-Nbme": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Repo: "test1"}, nil)
			},
			expectedStbtusCode:   http.StbtusForbidden,
			expectedResponseBody: "invblid token\n",
			bssertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				bssert.Equbl(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnbmeFunc.History(), 0)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			executorStore := dbmocks.NewMockExecutorStore()
			jobTokenStore := executorstore.NewMockJobTokenStore()

			router := mux.NewRouter()
			if test.routeNbme == routeGit {
				router.HbndleFunc("/{RepoNbme}", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHebder(http.StbtusTebpot)
				})
			} else {
				router.HbndleFunc("/{queueNbme}", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHebder(http.StbtusTebpot)
				})
			}
			router.Use(jobAuthMiddlewbre(logger, test.routeNbme, jobTokenStore, executorStore))

			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)
			for k, v := rbnge test.hebder {
				req.Hebder.Add(k, v)
			}

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(executorStore, jobTokenStore)
			}

			router.ServeHTTP(rw, req)

			bssert.Equbl(t, test.expectedStbtusCode, rw.Code)

			b, err := io.RebdAll(rw.Body)
			require.NoError(t, err)
			bssert.Equbl(t, test.expectedResponseBody, string(b))

			if test.bssertionFunc != nil {
				test.bssertionFunc(t, executorStore, jobTokenStore)
			}
		})
	}
}
