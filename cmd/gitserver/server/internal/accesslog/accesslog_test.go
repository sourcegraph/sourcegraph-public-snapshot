package accesslog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/userip"
	"github.com/stretchr/testify/assert"
)

func TestRecord(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		ctx = withContext(ctx, &paramsContext{})
		Record(ctx, "github.com/foo/bar", []string{"git", "grep", "foo"})
		pc := fromContext(ctx)
		assert.NotNil(t, pc)
		assert.Equal(t, "github.com/foo/bar", pc.repo)
		assert.Equal(t, "git", pc.cmd)
		assert.Equal(t, []string{"grep", "foo"}, pc.args)
	})
	t.Run("OK not initialized context", func(t *testing.T) {
		ctx := context.Background()
		Record(ctx, "github.com/foo/bar", []string{"git", "grep", "foo"})
		pc := fromContext(ctx)
		assert.Nil(t, pc)
	})
	t.Run("OK no args", func(t *testing.T) {
		ctx := context.Background()
		ctx = withContext(ctx, &paramsContext{})
		Record(ctx, "github.com/foo/bar", []string{"git"})
		pc := fromContext(ctx)
		assert.NotNil(t, pc)
		assert.Equal(t, "git", pc.cmd)
		assert.Nil(t, pc.args)
	})
}

func TestMiddleware(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		logger, exportLogs := logtest.Captured(t)
		h := Middleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Record(r.Context(), "github.com/foo/bar", []string{"git", "grep", "foo"})
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = userip.WithUserIP(ctx, &userip.UserIP{IP: "192.168.1.1"})
		req = req.WithContext(ctx)

		h.ServeHTTP(rec, req)
		logs := exportLogs()
		assert.Len(t, logs, 1)
		assert.Equal(t, "github.com/foo/bar", logs[0].Fields["params"].(map[string]any)["repo"])
		assert.Equal(t, "192.168.1.1", logs[0].Fields["actor"].(map[string]any)["ip"])
	})
	t.Run("OK, no recording", func(t *testing.T) {
		logger, exportLogs := logtest.Captured(t)
		h := Middleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("called")
		}))
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		h.ServeHTTP(rec, req)
		logs := exportLogs()
		assert.Len(t, logs, 1)
		assert.Equal(t, "called", logs[0].Message)
	})
}
