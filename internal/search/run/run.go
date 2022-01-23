package run

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
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

	// DefaultLimit is the default limit to use if not specified in query.
	DefaultLimit int
}

// Job is an interface shared by all individual search operations in the backend
// (e.g., text vs commit vs symbol search are represented as different jobs).
// Calling Run on a job object runs a search. The third argument accepts resolved repositories (which may or may
// not be required, depending on the job. E.g., a global search job does not
// require upfront repository resolution).
type Job interface {
	Run(context.Context, streaming.Sender, searchrepos.Pager) error
	Name() string

	// Required sets whether the results of this job are required. If true,
	// we must wait for its routines to complete. If false, the job is
	// optional, expressing that we may cancel the job. We typically run a
	// set of required and optional jobs concurrently, and cancel optional
	// jobs once we've guaranteed some required results, or after a timeout.
	Required() bool
}

// Routine represents all inputs to run multiple search operations (i.e.,
// multiple Jobs) in a single search routine. In other words, it executes all
// jobs that may be implemented by different search engines (Zoekt vs Searcher)
// or return different result types (text vs. symbols). The relation with
// SearchInputs and Routine is that SearchInputs are static values, parsed and
// validated, to produce one or more Routines. Routines express the complete
// information to execute the runtime semantics for particular search
// operations.
type Routine struct {
	Jobs        []Job
	RepoOptions search.RepoOptions
	Timeout     time.Duration
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
