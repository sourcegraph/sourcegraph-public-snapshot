package attribution

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

func NewHandler() http.Handler {
	// TODO: Logger scoped and wired with actor.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := actor.FromContext(r.Context())
		if got, want := a.GetSource(), codygateway.ActorSourceProductSubscription; got != want {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "only available for enterprise product subscriptions")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
}
