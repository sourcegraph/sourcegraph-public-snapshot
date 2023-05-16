package auth

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/response"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Authenticator struct {
	Logger      log.Logger
	EventLogger events.Logger
	Sources     actor.Sources
	Next        http.Handler
}

var _ http.Handler = &Authenticator{}

func (a *Authenticator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := ExtractBearer(r.Header)
	if err != nil {
		response.JSONError(a.Logger, w, http.StatusBadRequest, err)
		return
	}

	act, err := a.Sources.Get(r.Context(), token)
	if err != nil {
		response.JSONError(a.Logger, w, http.StatusUnauthorized, err)

		err := a.EventLogger.LogEvent(
			events.Event{
				Name:       llmproxy.EventNameUnauthorized,
				Source:     "anonymous",
				Identifier: "anonymous",
			},
		)
		if err != nil {
			a.Logger.Error("failed to log event", log.Error(err))
		}
		return
	}

	if !act.AccessEnabled {
		response.JSONError(
			a.Logger,
			w,
			http.StatusForbidden,
			errors.New("LLM proxy access not enabled"),
		)

		err := a.EventLogger.LogEvent(
			events.Event{
				Name:       llmproxy.EventNameAccessDenied,
				Source:     act.Source.Name(),
				Identifier: act.ID,
			},
		)
		if err != nil {
			a.Logger.Error("failed to log event", log.Error(err))
		}
		return
	}

	r = r.WithContext(actor.WithActor(r.Context(), act))
	a.Next.ServeHTTP(w, r)
}
