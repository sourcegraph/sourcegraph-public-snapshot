package accesslog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRecord(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		ctx = withContext(ctx, &paramsContext{})

		meta := map[string]string{"cmd": "git", "args": "grep foo"}

		Record(ctx, "github.com/foo/bar", meta)

		pc := fromContext(ctx)
		require.NotNil(t, pc)
		assert.Equal(t, "github.com/foo/bar", pc.repo)
		assert.Equal(t, meta, pc.metadata)
	})

	t.Run("OK not initialized context", func(t *testing.T) {
		ctx := context.Background()

		meta := map[string]string{"cmd": "git", "args": "grep foo"}

		Record(ctx, "github.com/foo/bar", meta)
		pc := fromContext(ctx)
		assert.Nil(t, pc)
	})
}

type accessLogConf struct {
	disabled bool
	callback func()
}

var _ conftypes.WatchableSiteConfig = &accessLogConf{}

func (a *accessLogConf) Watch(cb func()) { a.callback = cb }
func (a *accessLogConf) SiteConfig() schema.SiteConfiguration {
	return schema.SiteConfiguration{
		Log: &schema.Log{
			AuditLog: &schema.AuditLog{
				GitserverAccess: !a.disabled,
				GraphQL:         false,
				InternalTraffic: false,
			},
		},
	}
}

func TestHTTPMiddleware(t *testing.T) {
	t.Run("OK for access log setting", func(t *testing.T) {
		logger, exportLogs := logtest.Captured(t)
		h := HTTPMiddleware(logger, &accessLogConf{}, func(w http.ResponseWriter, r *http.Request) {
			meta := map[string]string{"cmd": "git", "args": "grep foo"}
			Record(r.Context(), "github.com/foo/bar", meta)
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = requestclient.WithClient(ctx, &requestclient.Client{IP: "192.168.1.1"})
		req = req.WithContext(ctx)

		h.ServeHTTP(rec, req)
		logs := exportLogs()
		require.Len(t, logs, 2)
		assert.Equal(t, accessLoggingEnabledMessage, logs[0].Message)
		assert.Contains(t, logs[1].Message, accessEventMessage)
		assert.Equal(t, "github.com/foo/bar", logs[1].Fields["params"].(map[string]any)["repo"])

		auditFields := logs[1].Fields["audit"].(map[string]interface{})
		assert.Equal(t, "gitserver", auditFields["entity"])
		assert.NotEmpty(t, auditFields["auditId"])

		actorFields := auditFields["actor"].(map[string]interface{})
		assert.Equal(t, "unknown", actorFields["actorUID"])
		assert.Equal(t, "192.168.1.1", actorFields["ip"])
		assert.Equal(t, "", actorFields["X-Forwarded-For"])
	})

	t.Run("handle, no recording", func(t *testing.T) {
		logger, exportLogs := logtest.Captured(t)
		var handled bool
		h := HTTPMiddleware(logger, &accessLogConf{}, func(w http.ResponseWriter, r *http.Request) {
			handled = true
		})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		h.ServeHTTP(rec, req)

		// Should have handled but not logged
		assert.True(t, handled)
		logs := exportLogs()
		require.Len(t, logs, 1)
		assert.NotEqual(t, accessEventMessage, logs[0].Message)
	})

	t.Run("disabled, then enabled", func(t *testing.T) {
		logger, exportLogs := logtest.Captured(t)
		cfg := &accessLogConf{disabled: true}
		var handled bool
		h := HTTPMiddleware(logger, cfg, func(w http.ResponseWriter, r *http.Request) {
			meta := map[string]string{"cmd": "git", "args": "grep foo"}
			Record(r.Context(), "github.com/foo/bar", meta)
			handled = true
		})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		// Request with access logging disabled
		h.ServeHTTP(rec, req)

		// Disabled, should have been handled but without a log message
		assert.True(t, handled)
		logs := exportLogs()
		require.Len(t, logs, 0)

		// Now we re-enable
		handled = false
		cfg.disabled = false
		cfg.callback()
		h.ServeHTTP(rec, req)

		// Enabled, should have handled AND generated a log message
		assert.True(t, handled)
		logs = exportLogs()
		require.Len(t, logs, 2)
		assert.Equal(t, accessLoggingEnabledMessage, logs[0].Message)
		assert.Contains(t, logs[1].Message, accessEventMessage)
	})
}
