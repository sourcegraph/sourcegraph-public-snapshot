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

// NewHandler creates a REST handler for attribution search.
// graphql.Client can be nil which disables the search.
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

// SnippetAttributions is holds the collection of attributions for a snippet.
type SnippetAttributions struct {
	// RepositoryNames is the list of repository names. We intend on mixing
	// names from both the local instance as well as from sourcegraph.com. So
	// we intentionally use a string since the name may not represent a
	// repository available on this instance.
	//
	// Note: for now this is a simple slice, we likely will expand what is
	// represented here and it will change into a struct capturing more
	// information.
	RepositoryNames []string

	// TotalCount is the total number of repository attributions we found
	// before stopping the search.
	//
	// Note: if we didn't finish searching the full corpus then LimitHit will
	// be true. For filtering use case this means if LimitHit is true you need
	// to be conservative with TotalCount and assume it could be higher.
	TotalCount int

	// LimitHit is true if we stopped searching before looking into the full
	// corpus. If LimitHit is true then it is possible there are more than
	// TotalCount attributions.
	LimitHit bool
}

func snippetAttributionDotCom(ctx context.Context, snippet string, limit int) (result *SnippetAttributions, err error) {
	ctx, traceLogger, endObservation := c.operations.snippetAttributionDotCom.With(ctx, &err, observation.Args{})
	defer endObservationWithResult(traceLogger, endObservation, &result)()

	resp, err := dotcom.SnippetAttribution(ctx, c.SourcegraphDotComClient, snippet, limit)
	if err != nil {
		return nil, err
	}

	var repoNames []string
	for _, node := range resp.SnippetAttribution.Nodes {
		repoNames = append(repoNames, node.RepositoryName)
	}

	return &SnippetAttributions{
		RepositoryNames: repoNames,
		TotalCount:      resp.SnippetAttribution.TotalCount,
		LimitHit:        resp.SnippetAttribution.LimitHit,
	}, nil
}
