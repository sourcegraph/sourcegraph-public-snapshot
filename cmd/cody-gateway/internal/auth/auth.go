package auth

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayevents"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Authenticator struct {
	Logger      log.Logger
	EventLogger events.Logger
	Sources     *actor.Sources
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := trace.Logger(r.Context(), a.Logger)
		token, err := authbearer.ExtractBearer(r.Header)
		if err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, err)
			return
		}

		act, err := a.Sources.Get(r.Context(), token)
		if err != nil {
			// Didn't even match to a source at all
			if actor.IsErrNotFromSource(err) {
				logger.Debug("received token with unknown source",
					log.String("token", token)) // unknown token, log for debug purposes
				response.JSONError(logger, w, http.StatusUnauthorized, err)
				return
			}

			// Matched to a source, but was denied
			var e actor.ErrAccessTokenDenied
			if errors.As(err, &e) {
				response.JSONError(logger, w, http.StatusUnauthorized, err)

				if err := a.EventLogger.LogEvent(
					r.Context(),
					events.Event{
						Name:       codygatewayevents.EventNameUnauthorized,
						Source:     e.Source,
						Identifier: "unknown",
						Metadata: map[string]any{
							"reason": e.Reason,
						},
					},
				); err != nil {
					logger.Error("failed to log event", log.Error(err))
				}
				return
			}

			// Fallback case: some mysterious error happened, likely upstream
			// service unavailability
			response.JSONError(logger, w, http.StatusServiceUnavailable, err)
			return
		}

		if !isAccessEnabled(act, r.URL.Path) {
			response.JSONError(
				logger,
				w,
				http.StatusForbidden,
				errors.New("Cody Gateway access not enabled"),
			)

			err := a.EventLogger.LogEvent(
				r.Context(),
				events.Event{
					Name:       codygatewayevents.EventNameAccessDenied,
					Source:     act.Source.Name(),
					Identifier: act.ID,
				},
			)
			if err != nil {
				logger.Error("failed to log event", log.Error(err))
			}
			return
		}

		r = r.WithContext(actor.WithActor(r.Context(), act))
		// Continue with the chain.
		next.ServeHTTP(w, r)
	})
}

func isAccessEnabled(act *actor.Actor, path string) bool {
	if act.AccessEnabled {
		return true
	}
	if act.EndpointAccess == nil {
		return false
	}
	path = strings.TrimPrefix(path, "/")
	for prefix, enabled := range act.EndpointAccess {
		prefix = strings.TrimPrefix(prefix, "/")
		if strings.HasPrefix(path, prefix) {
			return enabled
		}
	}
	return false
}
