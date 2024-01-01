package attribution

import (
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

type SearchResponse struct {
}

func NewHandler(client graphql.Client) http.Handler {
	// TODO: Logger scoped and wired with actor.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := actor.FromContext(r.Context())
		if got, want := a.GetSource(), codygateway.ActorSourceProductSubscription; got != want {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "only available for enterprise product subscriptions")
			return
		}
		if client == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "attribution search not enabled")
			return
		}
		// TODO: Actually query dotcom for attribution.
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
}
