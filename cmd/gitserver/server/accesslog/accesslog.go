// accesslog provides instrumentation to record logs of access made by a given actor to a repo at
// the http handler level.
// access logs may optionally (as per site configuration) be included in the audit log.
package accesslog

import (
	"context"
	"net/http"
	"sync"

	"github.com/sourcegraph/log"
	"go.uber.org/atomic"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/audit"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

type contextKey struct{}

type paramsContext struct {
	repo     string
	metadata []log.Field
}

// Record updates a mutable unexported field stored in the context,
// making it available for Middleware to log at the end of the middleware
// chain.
func Record(ctx context.Context, repo string, meta ...log.Field) {
	pc := fromContext(ctx)
	if pc == nil {
		return
	}

	pc.repo = repo
	pc.metadata = meta
}

func withContext(ctx context.Context, pc *paramsContext) context.Context {
	return context.WithValue(ctx, contextKey{}, pc)
}

func fromContext(ctx context.Context) *paramsContext {
	pc, ok := ctx.Value(contextKey{}).(*paramsContext)
	if !ok || pc == nil {
		return nil
	}
	return pc
}

// accessLogger handles HTTP requests and, if logEnabled, logs accesses.
type accessLogger struct {
	logger log.Logger

	logEnabled       *atomic.Bool
	watcher          conftypes.WatchableSiteConfig
	watchEnabledOnce sync.Once
}

func newAccessLogger(logger log.Logger, watcher conftypes.WatchableSiteConfig) *accessLogger {
	return &accessLogger{
		logger: logger,

		logEnabled: atomic.NewBool(false),
		watcher:    watcher,
	}
}

// messages are defined here to make assertions in testing.
const (
	accessEventMessage          = "access"
	accessLoggingEnabledMessage = "access logging enabled"
)

func (a *accessLogger) maybeLog(ctx context.Context) {
	// If access logging is not enabled, we are done
	if !a.isEnabled() {
		return
	}

	// Otherwise, log this access

	// Now we've gone through the handler, we can get the params that the handler
	// got from the request body.
	paramsCtx := fromContext(ctx)
	if paramsCtx == nil {
		return
	}
	if paramsCtx.repo == "" {
		return
	}

	var fields []log.Field

	if paramsCtx != nil {
		params := append([]log.Field{log.String("repo", paramsCtx.repo)}, paramsCtx.metadata...)
		fields = append(fields, log.Object("params", params...))
	} else {
		fields = append(fields, log.String("params", "nil"))
	}

	audit.Log(ctx, a.logger, audit.Record{
		Entity: "gitserver",
		Action: "access",
		Fields: fields,
	})
}

func (a *accessLogger) isEnabled() bool {
	a.watchEnabledOnce.Do(func() {
		// Initialize the logEnabled field with the current value
		logEnabled := audit.IsEnabled(a.watcher.SiteConfig(), audit.GitserverAccess)
		if logEnabled {
			a.logger.Info(accessLoggingEnabledMessage)
		}

		a.logEnabled.Store(logEnabled)

		// Watch for changes to the site config
		a.watcher.Watch(func() {
			newShouldLog := audit.IsEnabled(a.watcher.SiteConfig(), audit.GitserverAccess)
			changed := a.logEnabled.Swap(newShouldLog) != newShouldLog
			if changed {
				if newShouldLog {
					a.logger.Info(accessLoggingEnabledMessage)
				} else {
					a.logger.Info("access logging disabled")
				}
			}
		})
	})

	return a.logEnabled.Load()
}

// HTTPMiddleware will extract actor information and params collected by Record that has
// been stored in the context, in order to log a trace of the access.
func HTTPMiddleware(logger log.Logger, watcher conftypes.WatchableSiteConfig, next http.HandlerFunc) http.HandlerFunc {
	a := newAccessLogger(logger, watcher)

	return func(w http.ResponseWriter, r *http.Request) {
		// Prepare the context to hold the params which the handler is going to set.
		ctx := withContext(r.Context(), &paramsContext{})
		r = r.WithContext(ctx)

		// Call the next handler in the chain.
		next(w, r)

		// Log the access
		a.maybeLog(ctx)
	}
}

// GRPCMethodFilter is a function that returns true if the given method should be logged.
type GRPCMethodFilter func(method string) bool

// AllowAllGRPCMethodsFilter is a GRPCMethodFilter that returns true for all methods.
func AllowAllGRPCMethodsFilter(_ string) bool { return true }

// AllowListGRPCMethodsFilter is a GRPCMethodFilter that returns true for only methods in the allowList.
func AllowListGRPCMethodsFilter(allowList []string) GRPCMethodFilter {
	allMethods := make(map[string]struct{}, len(allowList))
	for _, m := range allowList {
		allMethods[m] = struct{}{}
	}

	return func(method string) bool {
		_, ok := allMethods[method]
		return ok
	}
}

func UnaryServerInterceptor(logger log.Logger, watcher conftypes.WatchableSiteConfig) grpc.UnaryServerInterceptor {
	a := newAccessLogger(logger, watcher)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = withContext(ctx, &paramsContext{})
		resp, err = handler(ctx, req)

		a.maybeLog(ctx)
		return resp, err
	}
}

func StreamServerInterceptor(logger log.Logger, watcher conftypes.WatchableSiteConfig) grpc.StreamServerInterceptor {
	a := newAccessLogger(logger, watcher)

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := withContext(ss.Context(), &paramsContext{})

		ss = &wrappedServerStream{ServerStream: ss, ctx: ctx}
		err := handler(srv, ss)

		a.maybeLog(ctx)
		return err
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
