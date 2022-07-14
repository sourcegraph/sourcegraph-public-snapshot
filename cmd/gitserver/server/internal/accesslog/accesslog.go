// accesslog provides instrumentation to record logs of access made by a given actor to a repo at
// the http handler level.
package accesslog

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/userip"
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

// HTTPMiddleware will extract actor information and params collected by Record that has
// been stored in the context, in order to log a trace of the access.
func HTTPMiddleware(logger log.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger
		ctx := r.Context()
		userIP := userip.FromContext(ctx)
		act := actor.FromContext(ctx)

		if userIP != nil {
			l = l.With(log.Object(
				"actor",
				log.String("ip", userIP.IP),
				log.String("X-Forwarded-For", userIP.XForwardedFor),
				log.Int32("actor", act.UID),
			))
		}

		// Prepare the context to hold the params which the handler is going to set.
		r = r.WithContext(withContext(ctx, &paramsContext{}))
		next(w, r)

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
			l = l.With(log.Object(
				"params",
				log.String("repo", paramsCtx.repo),
				log.String("cmd", paramsCtx.cmd),
				log.Strings("args", paramsCtx.args),
			))
		} else {
			l = l.With(log.String("params", "nil"))
		}

		l.Info("acces request")
		return
	}
}
