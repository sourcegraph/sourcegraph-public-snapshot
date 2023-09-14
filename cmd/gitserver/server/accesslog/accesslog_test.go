package accesslog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRecord(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		ctx = withContext(ctx, &paramsContext{})

		meta := []log.Field{log.String("cmd", "git"), log.String("args", "grep foo")}

		Record(ctx, "github.com/foo/bar", meta...)

		pc := fromContext(ctx)
		require.NotNil(t, pc)
		assert.Equal(t, "github.com/foo/bar", pc.repo)
		assert.Equal(t, meta, pc.metadata)
	})

	t.Run("OK not initialized context", func(t *testing.T) {
		ctx := context.Background()

		meta := []log.Field{log.String("cmd", "git"), log.String("args", "grep foo")}

		Record(ctx, "github.com/foo/bar", meta...)
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
			Record(r.Context(), "github.com/foo/bar", log.String("cmd", "git"), log.String("args", "grep foo"))
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
			Record(r.Context(), "github.com/foo/bar", log.String("cmd", "git"), log.String("args", "grep foo"))
			handled = true
		})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req = req.WithContext(actor.WithActor(context.Background(), actor.FromUser(32)))

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

func TestAccessLogGRPC(t *testing.T) {
	var (
		fakeIP             = "192.168.1.1"
		fakeRepositoryName = "github.com/foo/bar"
	)

	t.Run("basic recording and audit fields", func(t *testing.T) {
		t.Run("unary", func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)

			configuration := &accessLogConf{}
			client := &requestclient.Client{IP: fakeIP}

			interceptor := chainUnaryInterceptors(
				mockClientUnaryInterceptor(client),
				UnaryServerInterceptor(logger, configuration),
			)

			handlerCalled := false
			handler := func(ctx context.Context, req any) (any, error) {
				Record(ctx, fakeRepositoryName, log.String("foo", "bar"))
				handlerCalled = true

				return req, nil
			}

			req := struct{}{}
			info := &grpc.UnaryServerInfo{}
			_, err := interceptor(context.Background(), req, info, handler)
			if err != nil {
				t.Fatalf("failed to call interceptor: %v", err)
			}

			if !handlerCalled {
				t.Fatal("handler not called")
			}

			logs := exportLogs()

			require.Len(t, logs, 2)
			assert.Equal(t, accessLoggingEnabledMessage, logs[0].Message)
			assert.Contains(t, logs[1].Message, accessEventMessage)
			assert.Equal(t, fakeRepositoryName, logs[1].Fields["params"].(map[string]any)["repo"])

			auditFields := logs[1].Fields["audit"].(map[string]interface{})
			assert.Equal(t, "gitserver", auditFields["entity"])
			assert.NotEmpty(t, auditFields["auditId"])

			actorFields := auditFields["actor"].(map[string]interface{})
			assert.Equal(t, "unknown", actorFields["actorUID"])
			assert.Equal(t, fakeIP, actorFields["ip"])
			assert.Equal(t, "", actorFields["X-Forwarded-For"])
		})

		t.Run("stream", func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)

			configuration := &accessLogConf{}
			client := &requestclient.Client{IP: fakeIP}

			streamInterceptor := chainStreamInterceptors(
				mockClientStreamInterceptor(client),
				StreamServerInterceptor(logger, configuration),
			)

			handlerCalled := false
			handler := func(srv interface{}, stream grpc.ServerStream) error {
				ctx := stream.Context()

				Record(ctx, fakeRepositoryName, log.String("foo", "bar"))
				handlerCalled = true
				return nil
			}

			srv := struct{}{}
			ss := &testServerStream{ctx: context.Background()}
			info := &grpc.StreamServerInfo{}

			err := streamInterceptor(srv, ss, info, handler)
			if err != nil {
				t.Fatal(err)
			}

			if !handlerCalled {
				t.Fatal("handler not called")
			}

			logs := exportLogs()

			require.Len(t, logs, 2)
			assert.Equal(t, accessLoggingEnabledMessage, logs[0].Message)
			assert.Contains(t, logs[1].Message, accessEventMessage)
			assert.Equal(t, fakeRepositoryName, logs[1].Fields["params"].(map[string]any)["repo"])

			auditFields := logs[1].Fields["audit"].(map[string]interface{})
			assert.Equal(t, "gitserver", auditFields["entity"])
			assert.NotEmpty(t, auditFields["auditId"])

			actorFields := auditFields["actor"].(map[string]interface{})
			assert.Equal(t, "unknown", actorFields["actorUID"])
			assert.Equal(t, fakeIP, actorFields["ip"])
			assert.Equal(t, "", actorFields["X-Forwarded-For"])
		})
	})

	t.Run("handler, no recording", func(t *testing.T) {
		t.Run("unary", func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)

			configuration := &accessLogConf{}
			client := &requestclient.Client{IP: fakeIP}

			interceptor := chainUnaryInterceptors(
				mockClientUnaryInterceptor(client),
				UnaryServerInterceptor(logger, configuration),
			)

			handlerCalled := false
			handler := func(ctx context.Context, req any) (any, error) {
				handlerCalled = true
				return req, nil
			}

			req := struct{}{}
			info := &grpc.UnaryServerInfo{}
			_, err := interceptor(context.Background(), req, info, handler)
			if err != nil {
				t.Fatalf("failed to call interceptor: %v", err)
			}

			if !handlerCalled {
				t.Fatal("handler not called")
			}

			logs := exportLogs()

			// Should have handled but not logged
			require.Len(t, logs, 1)
			assert.NotEqual(t, accessEventMessage, logs[0].Message)
		})
	})

	t.Run("stream", func(t *testing.T) {
		logger, exportLogs := logtest.Captured(t)

		configuration := &accessLogConf{}
		client := &requestclient.Client{IP: fakeIP}

		streamInterceptor := chainStreamInterceptors(
			mockClientStreamInterceptor(client),
			StreamServerInterceptor(logger, configuration),
		)

		handlerCalled := false
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			handlerCalled = true
			return nil
		}

		srv := struct{}{}
		ss := &testServerStream{ctx: context.Background()}
		info := &grpc.StreamServerInfo{}

		err := streamInterceptor(srv, ss, info, handler)
		if err != nil {
			t.Fatalf("failed to call interceptor: %v", err)
		}

		if !handlerCalled {
			t.Fatal("handler not called")
		}

		logs := exportLogs()

		// Should have handled but not logged
		require.Len(t, logs, 1)
		assert.NotEqual(t, accessEventMessage, logs[0].Message)
	})

	t.Run("disabled, then enabled", func(t *testing.T) {
		t.Run("unary", func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)

			configuration := &accessLogConf{disabled: true}
			client := &requestclient.Client{IP: fakeIP}

			interceptor := chainUnaryInterceptors(
				mockClientUnaryInterceptor(client),
				UnaryServerInterceptor(logger, configuration),
			)

			handlerCalled := false
			handler := func(ctx context.Context, req any) (any, error) {
				Record(ctx, fakeRepositoryName, log.String("foo", "bar"))
				handlerCalled = true

				return req, nil
			}

			req := struct{}{}
			info := &grpc.UnaryServerInfo{}
			_, err := interceptor(context.Background(), req, info, handler)
			if err != nil {
				t.Fatalf("failed to call interceptor: %v", err)
			}

			if !handlerCalled {
				t.Fatal("handler not called")
			}

			// Disabled, should have been handled but without a log message
			logs := exportLogs()
			require.Len(t, logs, 0)

			// Now we re-enable
			handlerCalled = false
			configuration.disabled = false
			configuration.callback()
			_, err = interceptor(context.Background(), req, info, handler)
			if err != nil {
				t.Fatalf("failed to call interceptor: %v", err)
			}

			if !handlerCalled {
				t.Fatal("handler not called")
			}

			// Enabled, should have handled AND generated a log message
			logs = exportLogs()
			require.Len(t, logs, 2)
			assert.Equal(t, accessLoggingEnabledMessage, logs[0].Message)
			assert.Contains(t, logs[1].Message, accessEventMessage)
		})

		t.Run("stream", func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)

			configuration := &accessLogConf{disabled: true}
			client := &requestclient.Client{IP: fakeIP}

			interceptor := chainStreamInterceptors(
				mockClientStreamInterceptor(client),
				StreamServerInterceptor(logger, configuration),
			)

			handlerCalled := false
			handler := func(srv interface{}, stream grpc.ServerStream) error {
				ctx := stream.Context()

				Record(ctx, fakeRepositoryName, log.String("foo", "bar"))
				handlerCalled = true
				return nil
			}

			srv := struct{}{}
			ss := &testServerStream{ctx: context.Background()}
			info := &grpc.StreamServerInfo{}

			err := interceptor(srv, ss, info, handler)
			if err != nil {
				t.Fatal(err)
			}

			if !handlerCalled {
				t.Fatal("handler not called")
			}

			// Disabled, should have been handled but without a log message
			logs := exportLogs()
			require.Len(t, logs, 0)

			// Now we re-enable
			handlerCalled = false
			configuration.disabled = false
			configuration.callback()
			err = interceptor(srv, ss, info, handler)
			if err != nil {
				t.Fatalf("failed to call interceptor: %v", err)
			}

			if !handlerCalled {
				t.Fatal("handler not called")
			}

			// Enabled, should have handled AND generated a log message
			logs = exportLogs()
			require.Len(t, logs, 2)
			assert.Equal(t, accessLoggingEnabledMessage, logs[0].Message)
			assert.Contains(t, logs[1].Message, accessEventMessage)
		})
	})
}

func mockClientUnaryInterceptor(client *requestclient.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = requestclient.WithClient(ctx, client)
		return handler(ctx, req)
	}
}

func mockClientStreamInterceptor(client *requestclient.Client) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := requestclient.WithClient(ss.Context(), client)
		return handler(srv, &wrappedServerStream{ss, ctx})
	}
}

func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if len(interceptors) == 0 {
			return handler(ctx, req)
		}

		return interceptors[0](ctx, req, info, func(ctx context.Context, req any) (any, error) {
			return chainUnaryInterceptors(interceptors[1:]...)(ctx, req, info, handler)
		})
	}
}

func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if len(interceptors) == 0 {
			return handler(srv, ss)
		}

		return interceptors[0](srv, ss, info, func(srv any, ss grpc.ServerStream) error {
			return chainStreamInterceptors(interceptors[1:]...)(srv, ss, info, handler)
		})
	}
}

// testServerStream is a mock implementation of grpc.ServerStream for testing.
type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *testServerStream) Context() context.Context {
	return m.ctx
}

var _ grpc.ServerStream = &testServerStream{}
