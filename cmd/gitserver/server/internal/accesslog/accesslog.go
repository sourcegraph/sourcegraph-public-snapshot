// accesslog provides instrumentation to record logs of access made by a given actor to a repo at
// the http handler level.
package accesslog

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/schema"
)

type contextKey struct{}

type paramsContext struct {
	repo string
	cmd  string
	args []string
}

// Record updates a mutable unexported field stored in the context,
// making it available for Middleware to log at the end of the middleware
// chain.
func Record(ctx context.Context, repo string, args []string) {
	pc := fromContext(ctx)
	if pc == nil {
		return
	}

	pc.repo = repo
	if len(args) > 0 {
		pc.cmd = args[0]
	}
	if len(args) > 1 {
		pc.args = args[1:]
	}
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

const accessEventMessage = "access"

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
	var (
		cli    = requestclient.FromContext(ctx)
		act    = actor.FromContext(ctx)
		fields []log.Field
	)

	// Log the actor and client
	if cli != nil {
		fields = append(fields, log.Object(
			"actor",
			log.String("ip", cli.IP),
			log.String("X-Forwarded-For", cli.ForwardedFor),
			log.Int32("actor", act.UID),
		))
	}

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
		fields = append(fields, log.Object(
			"params",
			log.String("repo", paramsCtx.repo),
			log.String("cmd", paramsCtx.cmd),
			log.Strings("args", paramsCtx.args),
		))
	} else {
		fields = append(fields, log.String("params", "nil"))
	}

	a.logger.Info(accessEventMessage, fields...)
	return
}

// HTTPMiddleware will extract actor information and params collected by Record that has
// been stored in the context, in order to log a trace of the access.
func HTTPMiddleware(logger log.Logger, watcher conftypes.WatchableSiteConfig, next http.HandlerFunc) http.HandlerFunc {
	handler := &accessLogger{
		logger:     logger,
		next:       next,
		logEnabled: atomic.NewBool(shouldLog(watcher.SiteConfig())),
	}
	watcher.Watch(func() {
		handler.logEnabled.Store(shouldLog(watcher.SiteConfig()))
	})
	return http.HandlerFunc(handler.ServeHTTP)
}

func shouldLog(c schema.SiteConfiguration) bool {
	if c.Log == nil {
		return false
	}
	return c.Log.GitserverAccessLog
}
