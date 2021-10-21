package run

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SearchInputs contains fields we set before kicking off search.
type SearchInputs struct {
	Plan          query.Plan // the comprehensive query plan
	Query         query.Q    // the current basic query being evaluated, one part of query.Plan
	OriginalQuery string     // the raw string of the original search query
	PatternType   query.SearchType
	UserSettings  *schema.Settings

	// DefaultLimit is the default limit to use if not specified in query.
	DefaultLimit int
}

// Job is an interface shared by all search backends. Calling Run on a job
// object runs a search. The relation with SearchInputs and Jobs is that
// SearchInputs are static values, parsed and validated, to produce Jobs. Jobs
// express semantic behavior at runtime across different backends and system
// architecture.
type Job interface {
	Run(context.Context, streaming.Sender) error
	Name() string
}

// MaxResults computes the limit for the query.
func (inputs SearchInputs) MaxResults() int {
	if inputs.Query == nil {
		return 0
	}

	if count := inputs.Query.Count(); count != nil {
		return *count
	}

	if inputs.DefaultLimit != 0 {
		return inputs.DefaultLimit
	}

	return search.DefaultMaxSearchResults
}
