package run

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
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
	Features      featureflag.FlagSet
	CodeMonitorID *int64

	// DefaultLimit is the default limit to use if not specified in query.
	DefaultLimit int
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

// Job is an interface shared by all individual search operations in the
// backend (e.g., text vs commit vs symbol search are represented as different
// jobs) as well as combinations over those searches (run a set in parallel,
// timeout). Calling Run on a job object runs a search.
//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/search/run -i Job -o job_mock_test.go
type Job interface {
	Run(context.Context, database.DB, streaming.Sender) (*search.Alert, error)
	Name() string
}
