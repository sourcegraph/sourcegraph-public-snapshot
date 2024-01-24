package requestlogger

import (
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Middleware logs all requests. Should be placed underneath all instrumentation
// and/or actor extraction.
func Middleware(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		response := response.NewStatusHeaderRecorder(w, logger)
		next.ServeHTTP(response, r)

		ctx := r.Context()
		rc := requestclient.FromContext(ctx)
		logFields := append(rc.LogFields(),
			log.String("method", r.Method),
			log.String("path", r.URL.Path),
			log.Int("response.statusCode", response.StatusCode),
			log.Duration("duration", time.Since(start)))

		actor.FromContext(ctx).
			Logger(trace.Logger(ctx, logger)).
			Debug("Request", logFields...)
	})
}
