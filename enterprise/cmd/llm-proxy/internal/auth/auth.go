package auth

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/response"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Authenticator struct {
	Log     log.Logger
	Sources actor.Sources
	Next    http.Handler
}

var _ http.Handler = &Authenticator{}

func (a *Authenticator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	act, err := a.Sources.Get(r.Context(), token)
	if err != nil {
		response.JSONError(a.Log, w, http.StatusUnauthorized, err)
		return
	}

	if !act.AccessEnabled {
		response.JSONError(
			a.Log,
			w,
			http.StatusForbidden,
			errors.New("LLM proxy access not enabled"),
		)
		return
	}

	r = r.WithContext(actor.WithActor(r.Context(), act))
	a.Next.ServeHTTP(w, r)
}
