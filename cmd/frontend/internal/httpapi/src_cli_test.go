package httpapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/derision-test/glock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	srccli "github.com/sourcegraph/sourcegraph/internal/src-cli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSrcCliVersionHandler_ServeHTTP(t *testing.T) {
	minimumBranch := minimumVersionBranch(t)

	clock := glock.NewMockClock()
	logger, _ := logtest.Captured(t)

	doer := NewMockDoer()
	doer.DoFunc.SetDefaultHook(func(r *http.Request) (*http.Response, error) {
		assert.Contains(t, r.URL.Path, minimumBranch)
		return &http.Response{
			Body:       io.NopCloser(bytes.NewBufferString(`"3.42.1"`)),
			StatusCode: http.StatusOK,
		}, nil
	})

	handler := &srcCliVersionHandler{
		clock:    clock,
		doer:     doer,
		logger:   logger,
		maxStale: srcCliCacheLifetime,
	}

	t.Run("no mux vars", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/no-vars", nil)
		require.NoError(t, err)

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"rest": "unknown"})

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("version", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/version", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"rest": "version"})

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"version":"3.42.1"}`+"\n", rec.Body.String())
	})

	t.Run("download", func(t *testing.T) {
		for _, filename := range allowedFilenames {
			t.Run(filename, func(t *testing.T) {
				rec := httptest.NewRecorder()
				req, err := http.NewRequest(http.MethodGet, "/"+filename, nil)
				require.NoError(t, err)

				req = mux.SetURLVars(req, map[string]string{"rest": filename})

				handler.ServeHTTP(rec, req)
				assert.Equal(t, http.StatusFound, rec.Code)
				assert.Equal(
					t,
					srcCliDownloadsURL+"/3.42.1/"+filename,
					rec.Header().Get("Location"),
				)
			})
		}
	})
}

func TestSrcCliVersionHandler_Version(t *testing.T) {
	minimumBranch := minimumVersionBranch(t)

	t.Run("error response", func(t *testing.T) {
		// Basically, we're going to ensure that a failure in an upstream HTTP
		// request still results in srccli.MinimumVersion being returned.
		clock := glock.NewMockClock()
		logger, _ := logtest.Captured(t)

		doer := NewMockDoer()
		doer.DoFunc.SetDefaultHook(func(r *http.Request) (*http.Response, error) {
			assert.Contains(t, r.URL.Path, minimumBranch)
			return &http.Response{
				Body:       io.NopCloser(bytes.NewBufferString(`"3.42.1"`)),
				StatusCode: http.StatusInternalServerError,
			}, nil
		})

		handler := &srcCliVersionHandler{
			clock:    clock,
			doer:     doer,
			logger:   logger,
			maxStale: srcCliCacheLifetime,
		}

		version := handler.Version()
		assert.Equal(t, srccli.MinimumVersion, version)
	})

	t.Run("transport error", func(t *testing.T) {
		clock := glock.NewMockClock()
		logger, _ := logtest.Captured(t)

		doer := NewMockDoer()
		doer.DoFunc.SetDefaultHook(func(r *http.Request) (*http.Response, error) {
			assert.Contains(t, r.URL.Path, minimumBranch)
			return nil, errors.New("transport error")
		})

		handler := &srcCliVersionHandler{
			clock:    clock,
			doer:     doer,
			logger:   logger,
			maxStale: srcCliCacheLifetime,
		}

		version := handler.Version()
		assert.Equal(t, srccli.MinimumVersion, version)
	})

	t.Run("success", func(t *testing.T) {
		clock := glock.NewMockClock()
		logger, exportLogs := logtest.Captured(t)

		doFuncHookSuccess := func(r *http.Request) (*http.Response, error) {
			assert.Contains(t, r.URL.Path, minimumBranch)
			return &http.Response{
				Body:       io.NopCloser(bytes.NewBufferString(`"3.42.1"`)),
				StatusCode: http.StatusOK,
			}, nil
		}

		doer := NewMockDoer()
		doer.DoFunc.SetDefaultHook(doFuncHookSuccess)

		handler := &srcCliVersionHandler{
			clock:    clock,
			doer:     doer,
			logger:   logger,
			maxStale: srcCliCacheLifetime,
		}

		version := handler.Version()
		assert.Equal(t, "3.42.1", version)
		assert.Len(t, doer.DoFunc.History(), 1)

		// Make another request with a poisoned Do hook to ensure no HTTP
		// request is made.
		doer.DoFunc.SetDefaultHook(func(r *http.Request) (*http.Response, error) {
			assert.Fail(t, "unexpected request to a warm cache")
			return nil, errors.New("unexpected request to a warm cache")
		})

		version = handler.Version()
		assert.Equal(t, "3.42.1", version)
		assert.Len(t, doer.DoFunc.History(), 1)
		assert.Empty(t, exportLogs())

		// Finally, advance the clock and ensure the Do hook is invoked again.
		clock.Advance(2 * srcCliCacheLifetime)
		doer.DoFunc.SetDefaultHook(doFuncHookSuccess)

		version = handler.Version()
		assert.Equal(t, "3.42.1", version)
		assert.Len(t, doer.DoFunc.History(), 2)
	})
}

func minimumVersionBranch(t *testing.T) string {
	t.Helper()

	minimumVersion, err := semver.NewVersion(srccli.MinimumVersion)
	require.NoError(t, err)
	return fmt.Sprintf("%d.%d", minimumVersion.Major(), minimumVersion.Minor())
}
