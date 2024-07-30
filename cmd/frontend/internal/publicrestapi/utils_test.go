package publicrestapi

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/golly"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
)

func init() {
	txemail.DisableSilently()
}

type publicrestTest struct {
	t           *testing.T
	Handler     http.Handler
	Golly       *golly.Golly
	Credentials *golly.TestingCredentials
}

func newTest(t *testing.T, name string) *publicrestTest {
	MockUUID = "mocked-publicrestapi-uuid"
	gollyDoer := golly.NewGollyDoer(t, name, httpcli.TestExternalClient)
	recordReplayHandler := newRecordReplayHandler(gollyDoer, gollyDoer.DotcomCredentials())
	publicrestHandler := NewHandler(recordReplayHandler)
	return &publicrestTest{
		t:           t,
		Handler:     publicrestHandler,
		Golly:       gollyDoer,
		Credentials: gollyDoer.DotcomCredentials(),
	}

}

func newRecordReplayHandler(gollyDoer *golly.Golly, credentials *golly.TestingCredentials) http.Handler {
	m := mux.NewRouter()
	m.PathPrefix("/").Handler(dotcomProxyHandlerFunc(gollyDoer, credentials))
	return m
}

func dotcomProxyHandlerFunc(doer httpcli.Doer, credentials *golly.TestingCredentials) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := credentials.Endpoint + r.URL.RequestURI()
		req, err := http.NewRequest(r.Method, url, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for name, values := range r.Header {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}

		resp, err := doer.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resp == nil {
			http.Error(w, "nil response", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(resp.StatusCode)

		if _, err := io.Copy(w, resp.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (c *publicrestTest) chatCompletions(t *testing.T, body string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST",
		"/api/v1/chat/completions",
		strings.NewReader(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Credentials.AccessToken()))

	rr := httptest.NewRecorder()
	c.Handler.ServeHTTP(rr, req)
	return rr
}
