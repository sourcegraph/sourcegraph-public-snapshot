package run

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
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

// Job is an interface shared by all individual search operations in the
// backend (e.g., text vs commit vs symbol search are represented as different
// jobs) as well as combinations over those searches (run a set in parallel,
// timeout). Calling Run on a job object runs a search.
//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/search/run -i Job -o job_mock_test.go
type Job interface {
	Run(context.Context, database.DB, streaming.Sender) (*search.Alert, error)
	Name() string
}

type Mapper struct {
	MapJob func(job Job) Job

	// Search Jobs (leaf nodes)
	MapRepoSearchJob               func(*RepoSearch) *RepoSearch
	MapRepoSubsetTextSearchJob     func(*textsearch.RepoSubsetTextSearch) *textsearch.RepoSubsetTextSearch
	MapRepoUniverseTextSearchJob   func(*textsearch.RepoUniverseTextSearch) *textsearch.RepoUniverseTextSearch
	MapStructuralSearchJob         func(*structural.StructuralSearch) *structural.StructuralSearch
	MapCommitSearchJob             func(*commit.CommitSearch) *commit.CommitSearch
	MapRepoSubsetSymbolSearchJob   func(*symbol.RepoSubsetSymbolSearch) *symbol.RepoSubsetSymbolSearch
	MapRepoUniverseSymbolSearchJob func(*symbol.RepoUniverseSymbolSearch) *symbol.RepoUniverseSymbolSearch
	MapComputeExcludedReposJob     func(*repos.ComputeExcludedRepos) *repos.ComputeExcludedRepos

	// Expression Jobs
	MapAndJob func(children []Job) []Job
	MapOrJob  func(children []Job) []Job

	// Combinator Jobs
	MapParallelJob func(children []Job) []Job
	MapPriorityJob func(required, optional Job) (Job, Job)
	MapTimeoutJob  func(timeout time.Duration, child Job) (time.Duration, Job)
	MapLimitJob    func(limit int, child Job) (int, Job)

	// Filter Jobs
	MapSubRepoPermsFilterJob func(child Job) Job
}

func (m *Mapper) Map(job Job) Job {
	if m.MapJob != nil {
		job = m.MapJob(job)
	}

	switch j := job.(type) {
	case *RepoSearch:
		if m.MapRepoSearchJob != nil {
			j = m.MapRepoSearchJob(j)
		}
		return j

	case *textsearch.RepoSubsetTextSearch:
		if m.MapRepoSubsetTextSearchJob != nil {
			j = m.MapRepoSubsetTextSearchJob(j)
		}
		return j

	case *textsearch.RepoUniverseTextSearch:
		if m.MapRepoUniverseTextSearchJob != nil {
			j = m.MapRepoUniverseTextSearchJob(j)
		}
		return j

	case *structural.StructuralSearch:
		if m.MapStructuralSearchJob != nil {
			j = m.MapStructuralSearchJob(j)
		}
		return j

	case *commit.CommitSearch:
		if m.MapCommitSearchJob != nil {
			j = m.MapCommitSearchJob(j)
		}
		return j

	case *symbol.RepoSubsetSymbolSearch:
		if m.MapRepoSubsetSymbolSearchJob != nil {
			j = m.MapRepoSubsetSymbolSearchJob(j)
		}
		return j

	case *symbol.RepoUniverseSymbolSearch:
		if m.MapRepoUniverseSymbolSearchJob != nil {
			j = m.MapRepoUniverseSymbolSearchJob(j)
		}
		return j

	case *repos.ComputeExcludedRepos:
		if m.MapComputeExcludedReposJob != nil {
			j = m.MapComputeExcludedReposJob(j)
		}
		return j

	case *AndJob:
		children := make([]Job, 0, len(j.children))
		for _, child := range j.children {
			children = append(children, m.Map(child))
		}
		if m.MapAndJob != nil {
			children = m.MapAndJob(children)
		}
		return NewAndJob(children...)

	case *OrJob:
		children := make([]Job, 0, len(j.children))
		for _, child := range j.children {
			children = append(children, m.Map(child))
		}
		if m.MapOrJob != nil {
			children = m.MapOrJob(children)
		}
		return NewOrJob(children...)

	case *ParallelJob:
		children := make([]Job, 0, len(*j))
		for _, child := range *j {
			children = append(children, m.Map(child))
		}
		if m.MapParallelJob != nil {
			children = m.MapParallelJob(children)
		}
		return NewParallelJob(children...)

	case *PriorityJob:
		required := m.Map(j.required)
		optional := m.Map(j.optional)
		if m.MapPriorityJob != nil {
			required, optional = m.MapPriorityJob(required, optional)
		}
		return NewPriorityJob(required, optional)

	case *TimeoutJob:
		child := m.Map(j.child)
		timeout := j.timeout
		if m.MapTimeoutJob != nil {
			timeout, child = m.MapTimeoutJob(timeout, child)
		}
		return NewTimeoutJob(timeout, child)

	case *LimitJob:
		child := m.Map(j.child)
		limit := j.limit
		if m.MapLimitJob != nil {
			limit, child = m.MapLimitJob(limit, child)
		}
		return NewLimitJob(limit, child)

	case *subRepoPermsFilterJob:
		child := m.Map(j.child)
		if m.MapSubRepoPermsFilterJob != nil {
			child = m.MapSubRepoPermsFilterJob(child)
		}
		return NewFilterJob(child)

	case *noopJob:
		return j

	}
	// Unreachable
	return job
}
