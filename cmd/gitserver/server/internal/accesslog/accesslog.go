// accesslog provides instrumentation to record logs of access made by a given actor to a repo at
// the http handler level.
// access logs may optionally (as per site configuration) be included in the audit log.
package accesslog

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"go.uber.org/atomic"

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
	logger     log.Logger
	next       http.HandlerFunc
	logEnabled *atomic.Bool
}

var _ http.Handler = &accessLogger{}

// messages are defined here to make assertions in testing.
const (
	accessEventMessage          = "access"
	accessLoggingEnabledMessage = "access logging enabled"
)

func (a *accessLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Prepare the context to hold the params which the handler is going to set.
	ctx := r.Context()
	r = r.WithContext(withContext(ctx, &paramsContext{}))
	a.next(w, r)

	// If access logging is not enabled, we are done
	if !a.logEnabled.Load() {
		return
	}

	// Otherwise, log this access
	var fields []log.Field

	// Now we've gone through the handler, we can get the params that the handler
	// got from the request body.
	paramsCtx := fromContext(r.Context())
	if paramsCtx == nil {
		return
	}
	if paramsCtx.repo == "" {
		return
	}

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

// HTTPMiddleware will extract actor information and params collected by Record that has
// been stored in the context, in order to log a trace of the access.
func HTTPMiddleware(logger log.Logger, watcher conftypes.WatchableSiteConfig, next http.HandlerFunc) http.HandlerFunc {
	handler := &accessLogger{
		logger:     logger,
		next:       next,
		logEnabled: atomic.NewBool(audit.IsEnabled(watcher.SiteConfig(), audit.GitserverAccess)),
	}
	if handler.logEnabled.Load() {
		logger.Info(accessLoggingEnabledMessage)
	}

	// Allow live toggling of access logging
	watcher.Watch(func() {
		newShouldLog := audit.IsEnabled(watcher.SiteConfig(), audit.GitserverAccess)
		changed := handler.logEnabled.Swap(newShouldLog) != newShouldLog
		if changed {
			if newShouldLog {
				logger.Info(accessLoggingEnabledMessage)
			} else {
				logger.Info("access logging disabled")
			}
		}
	})

	return handler.ServeHTTP
}
