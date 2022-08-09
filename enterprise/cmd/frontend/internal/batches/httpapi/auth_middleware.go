package httpapi

import (
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func authMiddleware(next http.Handler, store *store.Store, operation *observation.Operation) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, err := func() (_ int, err error) {
			ctx, trace, endObservation := operation.With(r.Context(), &err, observation.Args{})
			defer endObservation(1, observation.Args{})
			_ = ctx

			trace.Log(log.Event("bypassing code host auth check"))
			return 0, nil
		}()
		if err != nil {
			if statusCode >= 500 {
				operation.Logger.Error("batches.httpapi: failed to authorize request", sglog.Error(err))
			}

			http.Error(w, fmt.Sprintf("failed to authorize request: %s", err.Error()), statusCode)
			return
		}

		next.ServeHTTP(w, r)
	})
}
