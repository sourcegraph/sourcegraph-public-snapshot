package run

import (
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SearchInputs contains fields we set before kicking off search.
type SearchInputs struct {
	Plan          query.Plan // the comprehensive query plan
	Query         query.Q    // the current basic query being evaluated, one part of query.Plan
	OriginalQuery string     // the raw string of the original search query
	PatternType   query.SearchType
	UserSettings  *schema.Settings
	Features      featureflag.FlagSet
	CodeMonitorID *int64

	// DefaultLimit is the default limit to use if not specified in query.
	DefaultLimit int
}

// MaxResults computes the limit for the query.
func (inputs SearchInputs) MaxResults() int {
	return inputs.Query.MaxResults(inputs.DefaultLimit)
}
