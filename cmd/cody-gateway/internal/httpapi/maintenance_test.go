package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/sams"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Fake for the sams.Client interface.
type fakeSAMSClient struct {
	// Value to return when IntrospectToken is called.
	// Will be reset to nil after each call.
	retToken *sams.TokenIntrospection
	retError error
}

func (f *fakeSAMSClient) IntrospectToken(_ context.Context, _ string) (*sams.TokenIntrospection, error) {
	t, e := f.retToken, f.retError
	f.retToken = nil
	f.retError = nil
	return t, e
}

func TestNewMaintenanceHandler(t *testing.T) {
	logger := logtest.NoOp(t)

	// Confirm that the handler won't do anything unless SAMS is configured.
	t.Run("NoOpWithoutSAMSConfig", func(t *testing.T) {
		var timesNextCalled int
		next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			timesNextCalled++
		})
		mHandler := NewMaintenanceHandler(logger, next, &config.Config{}, nil)
		mHandler.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/maintenance/do-the-thing", nil))
		mHandler.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/non-maintenance/endpoint", nil))
		assert.Equal(t, 2, timesNextCalled)
	})

	t.Run("SAMSAuth", func(t *testing.T) {
		samsClient := &fakeSAMSClient{}

		// We don't supply next, config, or the redisKV because they won't
		// be needed for this test.
		noOpNext := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Test Passed", http.StatusOK)
		})
		mHandler := newMaintenanceHandler(logger, noOpNext, nil /* redis */, samsClient)

		// SAMS returns an error when checking inspecting the supplied access token.
		t.Run("Errors", func(t *testing.T) {
			// We don't fail with an auth error for /maintenance/404 because we don't
			// know what tokens are required, instead just serve a 404.
			t.Run("404", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/maintenance/endpoint", nil)
				mHandler.ServeHTTP(w, req)

				resp := w.Result()
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Equal(t, http.StatusNotFound, resp.StatusCode)
				assert.Equal(t, "404 page not found\n", string(respBody))
			})

			t.Run("AuthHeaderMissing", func(t *testing.T) {
				{
					w := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests", nil)
					mHandler.ServeHTTP(w, req)

					resp := w.Result()
					respBody, err := io.ReadAll(resp.Body)
					require.NoError(t, err)

					assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
					assert.Equal(t, "Unauthorized\n", string(respBody))
				}

				// Same as before, but now with the wrong auth scheme.
				{
					w := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests", nil)
					req.Header.Add("Authorization", "token xxxxxx")
					mHandler.ServeHTTP(w, req)

					resp := w.Result()
					respBody, err := io.ReadAll(resp.Body)
					require.NoError(t, err)

					assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
					assert.Equal(t, "Unauthorized\n", string(respBody))
				}
			})

			t.Run("SAMSError", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests", nil)
				req.Header.Add("Authorization", "Bearer xxxxxx")
				samsClient.retError = errors.New("sams error")
				mHandler.ServeHTTP(w, req)

				resp := w.Result()
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				assert.Equal(t, "Internal Server Error\n", string(respBody))
			})

			t.Run("MissingScope", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests", nil)
				req.Header.Add("Authorization", "Bearer xxxxxx")
				samsClient.retToken = &sams.TokenIntrospection{
					Active: true,
					Scope:  "svc::resource::action svc2::resource::action",
				}
				mHandler.ServeHTTP(w, req)

				resp := w.Result()
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				assert.Equal(t, http.StatusForbidden, resp.StatusCode)
				assert.Equal(t, "Forbidden: Missing required scope\n", string(respBody))
			})
		})

		t.Run("Works", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests/invalid-prompt-key", nil)
			req.Header.Add("Authorization", "Bearer xxxxxx")
			samsClient.retToken = &sams.TokenIntrospection{
				Active: true,
				Scope:  "svc::resource::action " + samsScopeFlaggedPromptRead + " svc2::resource::action",
			}
			mHandler.ServeHTTP(w, req)

			resp := w.Result()
			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// We didn't configure this test to actually interact with the Redis instance.
			// So we just confirm our maintenance handler is executed. (Confirming the auth
			// check looking for the specific scope was successful.)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			assert.Equal(t, "Invalid prompt key\n", string(respBody))
		})
	})

	// Returns a properly configured maintenance handler that will receive an
	// HTTP request with a pre-seeded Redis instance.
	newMaintenanceHandlerTest := func() http.Handler {
		// Set up mock Redis with some valid data.
		redisData := map[string]string{
			"non-promt-data":   "na",
			"non-promt-data:2": "na",
			"prompt:trace-id-1:chat_completion:user-id-1": "prompt message 1",
			"prompt:trace-id-2:chat_completion:user-id-2": "prompt message 2",
			"prompt:trace-id-3:code_completion:user-id-2": "prompt message 3",
		}
		// Actually supply fake implementations for some of the mock handlers used.
		mockRedis := redispool.NewMockKeyValue()
		mockRedis.KeysFunc.SetDefaultHook(func(filter string) ([]string, error) {
			filter = strings.TrimSuffix(filter, "*")
			var matches []string
			for key := range redisData {
				if strings.HasPrefix(key, filter) {
					matches = append(matches, key)
				}
			}
			return matches, nil
		})
		mockRedis.GetFunc.SetDefaultHook(func(key string) redispool.Value {
			value, ok := redisData[key]
			if !ok {
				return redispool.NewValue(nil, nil)
			}
			return redispool.NewValue(value, nil)
		})

		// Set up a Fake SAMS client that will authenticate one request.
		samsClient := &fakeSAMSClient{}
		samsClient.retToken = &sams.TokenIntrospection{
			Active: true,
			Scope:  "svc::resource::action " + samsScopeFlaggedPromptRead + " svc2::resource::action",
		}

		// We don't supply next, config, or the redisKV because they won't
		// be needed for this test.
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Didn't execute the maintenance endpoint?", http.StatusInternalServerError)
		})
		return newMaintenanceHandler(logger, next, mockRedis, samsClient)
	}

	t.Run("ListFlaggedPrompts", func(t *testing.T) {
		mHandler := newMaintenanceHandlerTest()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests", nil)
		req.Header.Add("Authorization", "Bearer xxxxxx")
		mHandler.ServeHTTP(w, req)

		resp := w.Result()
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var typedResponse listFlaggedPromptsResponse
		require.NoError(t, json.Unmarshal(respBody, &typedResponse))
		require.Equal(t, 3, len(typedResponse.Prompts))

		for _, prompt := range typedResponse.Prompts {
			switch prompt.Key {
			case "prompt:trace-id-1:chat_completion:user-id-1":
				assert.Equal(t, "trace-id-1", prompt.TraceID)
				assert.Equal(t, "chat_completion", prompt.Feature)
				assert.Equal(t, "user-id-1", prompt.UserID)
			case "prompt:trace-id-2:chat_completion:user-id-2":
				assert.Equal(t, "trace-id-2", prompt.TraceID)
				assert.Equal(t, "chat_completion", prompt.Feature)
				assert.Equal(t, "user-id-2", prompt.UserID)
			case "prompt:trace-id-3:code_completion:user-id-2":
				assert.Equal(t, "trace-id-3", prompt.TraceID)
				assert.Equal(t, "code_completion", prompt.Feature)
				assert.Equal(t, "user-id-2", prompt.UserID)
			default:
				t.Errorf("unexpected prompt key returned: %v", prompt.Key)
			}
		}
	})

	t.Run("GetFlaggedPrompt", func(t *testing.T) {
		t.Run("404", func(t *testing.T) {
			mHandler := newMaintenanceHandlerTest()

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests/prompt:404", nil)
			req.Header.Add("Authorization", "Bearer xxxxxx")
			mHandler.ServeHTTP(w, req)

			resp := w.Result()
			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			assert.Equal(t, "Prompt not found\n", string(respBody))
		})

		t.Run("OK", func(t *testing.T) {
			promptKey := "prompt:trace-id-3:code_completion:user-id-2"

			mHandler := newMaintenanceHandlerTest()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/maintenance/flagged-requests/"+url.PathEscape(promptKey), nil)
			req.Header.Add("Authorization", "Bearer xxxxxx")
			mHandler.ServeHTTP(w, req)

			resp := w.Result()
			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "prompt message 3", string(respBody))
		})
	})
}
