package attribution

import (
	"encoding/json"
	"net/http"

	"github.com/Khan/genqlient/graphql"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// LimitUpperBound is the maximum (and default) value allowed for Limit request parameter.
// If a higher value is given, then this default is set.
const LimitUpperBound = 4

// NewHandler creates a REST handler for attribution search.
// graphql.Client can be nil which disables the search.
func NewHandler(client graphql.Client, baseLogger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		a := actor.FromContext(ctx)
		logger := a.Logger(trace.Logger(ctx, baseLogger))
		if got, want := a.GetSource(), codygatewayactor.ActorSourceEnterpriseSubscription; got != want {
			response.JSONError(logger, w, http.StatusUnauthorized, errors.New("only available for enterprise product subscriptions"))
			return
		}
		if client == nil {
			response.JSONError(logger, w, http.StatusServiceUnavailable, errors.New("attribution search not enabled"))
			return
		}
		var request codygateway.AttributionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, err)
			return
		}
		limit := request.Limit
		if limit == 0 || limit > LimitUpperBound {
			limit = LimitUpperBound
		}
		searchResponse, err := dotcom.SnippetAttribution(ctx, client, request.Snippet, limit)
		if err != nil {
			response.JSONError(logger, w, http.StatusBadGateway, errors.Wrap(err, "fetching SnippetAttribution from sourcegraph.com"))
			return
		}
		var rs []codygateway.AttributionRepository
		for _, n := range searchResponse.SnippetAttribution.Nodes {
			rs = append(rs, codygateway.AttributionRepository{Name: n.RepositoryName})
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(&codygateway.AttributionResponse{
			Repositories: rs,
			TotalCount:   searchResponse.SnippetAttribution.TotalCount,
			LimitHit:     searchResponse.SnippetAttribution.LimitHit,
		}); err != nil {
			baseLogger.Debug("failed to marshal json response", log.Error(err))
		}
	})
}
