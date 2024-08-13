package llmapi

import (
	"encoding/json"
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
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

type publicrestTest struct {
	t           *testing.T
	Handler     http.Handler
	Golly       *golly.Golly
	Credentials *golly.TestingCredentials
	HttpClient  http.Handler
}

func newTest(t *testing.T, getModelConfigFunc GetModelConfigurationFunc) *publicrestTest {
	MockUUID = "mocked-llmapi-uuid"
	gollyDoer := golly.NewGollyDoer(t, httpcli.TestExternalClient)
	recordReplayHandler := newRecordReplayHandler(gollyDoer, gollyDoer.DotcomCredentials())
	apiHandler := mux.NewRouter().PathPrefix("/.api/llm/").Subrouter()

	RegisterHandlers(apiHandler, recordReplayHandler, getModelConfigFunc)

	return &publicrestTest{
		t:           t,
		Handler:     apiHandler,
		Golly:       gollyDoer,
		Credentials: gollyDoer.DotcomCredentials(),
		HttpClient:  recordReplayHandler,
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
		"/.api/llm/chat/completions",
		strings.NewReader(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Credentials.AccessToken))

	rr := httptest.NewRecorder()
	c.Handler.ServeHTTP(rr, req)
	return rr
}

func (c *publicrestTest) getChatModels() []types.Model {
	modelConfig := c.getModelConfig()
	chatModels := []types.Model{}
	for _, model := range modelConfig.Models {
		for _, capability := range model.Capabilities {
			if capability == "chat" {
				chatModels = append(chatModels, model)
			}
		}
	}
	return chatModels
}

func (c *publicrestTest) getModelConfig() *types.ModelConfiguration {
	req, err := http.NewRequest("GET", c.Credentials.Endpoint+"/.api/modelconfig/supported-models.json", nil)
	req.Header.Set("Authorization", "token "+c.Credentials.AccessToken)
	assert.NoError(c.t, err)
	res, err := c.Golly.Do(req)
	assert.NoError(c.t, err)
	var modelConfig types.ModelConfiguration
	assert.Equal(c.t, http.StatusOK, res.StatusCode, "Failed to get model config %s", res.Body)
	assert.NoError(c.t, json.NewDecoder(res.Body).Decode(&modelConfig))
	return &modelConfig
}
