package server

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/userip"
	"go.uber.org/zap/zapcore"
)

type repoContextKey struct{}

type RepoContext struct {
	Repo string
	Cmd  string
	Args []string
}

func WithRepoContext(ctx context.Context, rc *RepoContext) context.Context {
	return context.WithValue(ctx, repoContextKey{}, rc)
}

func FromRepoContext(ctx context.Context) *RepoContext {
	rc, ok := ctx.Value(repoContextKey{}).(*RepoContext)
	if !ok || rc == nil {
		return nil
	}
	return rc
}

func LogRequest(logger log.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		println("before")
		ctx := r.Context()
		userIP := userip.FromContext(ctx)
		act := actor.FromContext(ctx)

		// TODO DEVX We have a bug, we should able to use With, but it's mutating the parent logger which is very wrong
		var fields []zapcore.Field
		fields = append(fields, log.Object(
			"actor",
			log.String("ip", userIP.IP),
			log.String("X-Forwarded-For", userIP.XForwardedFor),
			log.Int32("actor", act.UID),
		))

		// Prepare the context to hold our values
		r = r.WithContext(WithRepoContext(ctx, &RepoContext{}))
		next(w, r)
		println("after")

		// Now we've gone through the handler, we can get our stuff
		repoCtx := FromRepoContext(r.Context())

		if repoCtx != nil {
			fields = append(fields, log.Object(
				"repo",
				log.String("repo", repoCtx.Repo),
				log.String("cmd", repoCtx.Cmd),
				log.Strings("args", repoCtx.Args),
			))
		} else {
			fields = append(fields, log.String("repo", "nil"))
		}

		logger.Warn("log request", fields...)
		return
	}
}
