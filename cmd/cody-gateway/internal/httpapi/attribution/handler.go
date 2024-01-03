package attribution

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Request for attribution search. Expected in JSON form as the body of POST request.
type Request struct {
	// Snippet is the text to search attribution of.
	Snippet string
	// Limit is the upper bound of number of responses we want to get.
	Limit int
}

// Response of attribution search. Contains some repositories to which the snippet can be attributed to.
type Response struct {
	// Repositories which contain code matching search snippet.
	Repositories []Repository
	// TotalCount denotes how many total matches there were (including listed repositories).
	TotalCount int
	// LimitHit is true if the number of search hits goes beyond limit specified in request.
	LimitHit bool
}

// Repository represents matching of search content against a repository.
type Repository struct {
	// Name of the repo on dotcom. Like github.com/sourcegraph/sourcegraph.
	Name string
}

// NewHandler creates a REST handler for attribution search.
// graphql.Client can be nil which disables the search.
func NewHandler(client graphql.Client, baseLogger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := actor.FromContext(r.Context())
		if got, want := a.GetSource(), codygateway.ActorSourceProductSubscription; got != want {
            response.JSONError(logger, w, errors.New("only available for enterprise product subscriptions")
			return
		}
		if client == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "attribution search not enabled")
			return
		}
		logger := a.Logger(trace.Logger(r.Context(), baseLogger))
		var request Request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "cannot understand the request")
			logger.Debug("failed to unmarshal json request", log.Error(err))
			return
		}
		// TODO(#59244): Actually query dotcom for attribution.
		response := &Response{
			TotalCount: 0,
			LimitHit:   false,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			baseLogger.Debug("failed to marshal json response", log.Error(err))
		}
	})
}
