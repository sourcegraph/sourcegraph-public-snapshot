package shared

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
)

func TestAuthenticate(t *testing.T) {
	logger := logtest.Scoped(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	t.Run("unauthenticated and allow anonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		authenticate(logger, nil, nil, next, authenticateOptions{AllowAnonymous: true}).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthenticated but disallow anonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		authenticate(logger, nil, nil, next, authenticateOptions{AllowAnonymous: false}).ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("authenticated without cache hit", func(t *testing.T) {
		cache := NewMockCache()
		doer := NewMockDoer()
		doer.DoFunc.SetDefaultReturn(
			&http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`
{
  "data": {
    "dotcom": {
      "productSubscriptionByAccessToken": {
        "id": "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
        "uuid": "6452a8fc-e650-45a7-a0a2-357f776b3b46",
        "isArchived": false,
        "llmProxyAccess": {
          "enabled": true,
          "rateLimit": {
            "limit": 0,
            "intervalSeconds": 0
          }
        }
      }
    }
  }
}`)),
			},
			nil,
		)
		client := graphql.NewClient("https://sourcegraph.com/.api/graphql", doer)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, actor.FromContext(r.Context()).Subscription)
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer abc123")
		authenticate(logger, cache, client, next, authenticateOptions{AllowAnonymous: false}).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		mockrequire.Called(t, doer.DoFunc)
	})

	t.Run("authenticated with cache hit", func(t *testing.T) {
		cache := NewMockCache()
		cache.GetFunc.SetDefaultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","accessEnabled":true,"rateLimit":null}`),
			true,
		)
		doer := NewMockDoer()
		doer.DoFunc.SetDefaultReturn(
			&http.Response{
				StatusCode: http.StatusServiceUnavailable,
			},
			nil,
		)
		client := graphql.NewClient("https://sourcegraph.com/.api/graphql", doer)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, actor.FromContext(r.Context()).Subscription)
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer abc123")
		authenticate(logger, cache, client, next, authenticateOptions{AllowAnonymous: false}).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		mockrequire.NotCalled(t, doer.DoFunc)
	})

	t.Run("authenticated but not enabled", func(t *testing.T) {
		cache := NewMockCache()
		cache.GetFunc.SetDefaultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","accessEnabled":false,"rateLimit":null}`),
			true,
		)
		doer := NewMockDoer()
		client := graphql.NewClient("https://sourcegraph.com/.api/graphql", doer)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer abc123")
		authenticate(logger, cache, client, next, authenticateOptions{AllowAnonymous: false}).ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
