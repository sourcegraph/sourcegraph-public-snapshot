package resolvers

import (
	"github.com/sourcegraph/log"
)

type relatedInsightsResolver struct {
	baseInsightResolver

	logger log.Logger
}

// in resolver:
// fetch all relevant insights from db (using `GetAll` and the Repo arg)
// for each insight query, append repo and revision, call query.TabulationDecoder (we can make our own RepoTabulationDecoder too)
// group series to line matches
// return information in useful format
func (r *relatedInsightsResolver) 
