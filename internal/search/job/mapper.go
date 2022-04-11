package job

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

type Mapper struct {
	MapJob func(job Job) Job

	// Search Jobs (leaf nodes)
	MapZoektRepoSubsetSearchJob    func(*zoekt.ZoektRepoSubsetSearch) *zoekt.ZoektRepoSubsetSearch
	MapZoektSymbolSearchJob        func(*zoekt.ZoektSymbolSearch) *zoekt.ZoektSymbolSearch
	MapSearcherJob                 func(*searcher.Searcher) *searcher.Searcher
	MapSymbolSearcherJob           func(*searcher.SymbolSearcher) *searcher.SymbolSearcher
	MapRepoSearchJob               func(*run.RepoSearch) *run.RepoSearch
	MapRepoUniverseTextSearchJob   func(*zoekt.GlobalSearch) *zoekt.GlobalSearch
	MapStructuralSearchJob         func(*structural.StructuralSearch) *structural.StructuralSearch
	MapCommitSearchJob             func(*commit.CommitSearch) *commit.CommitSearch
	MapRepoUniverseSymbolSearchJob func(*symbol.RepoUniverseSymbolSearch) *symbol.RepoUniverseSymbolSearch
	MapComputeExcludedReposJob     func(*repos.ComputeExcludedRepos) *repos.ComputeExcludedRepos

	// Repo pager Job (pre-step for some Search Jobs)
	MapRepoPagerJob func(*repoPagerJob) *repoPagerJob

	// Expression Jobs
	MapAndJob func(children []Job) []Job
	MapOrJob  func(children []Job) []Job

	// Combinator Jobs
	MapParallelJob func(children []Job) []Job
	MapPriorityJob func(required, optional Job) (Job, Job)
	MapTimeoutJob  func(timeout time.Duration, child Job) (time.Duration, Job)
	MapLimitJob    func(limit int, child Job) (int, Job)
	MapSelectJob   func(path filter.SelectPath, child Job) (filter.SelectPath, Job)
	MapAlertJob    func(inputs *run.SearchInputs, child Job) (*run.SearchInputs, Job)

	// Filter Jobs
	MapSubRepoPermsFilterJob func(child Job) Job
}

func (m *Mapper) Map(job Job) Job {
	if job == nil {
		return nil
	}

	if m.MapJob != nil {
		job = m.MapJob(job)
	}

	switch j := job.(type) {
	case *zoekt.ZoektRepoSubsetSearch:
		if m.MapZoektRepoSubsetSearchJob != nil {
			j = m.MapZoektRepoSubsetSearchJob(j)
		}
		return j

	case *zoekt.ZoektSymbolSearch:
		if m.MapZoektSymbolSearchJob != nil {
			j = m.MapZoektSymbolSearchJob(j)
		}
		return j

	case *searcher.Searcher:
		if m.MapSearcherJob != nil {
			j = m.MapSearcherJob(j)
		}
		return j

	case *searcher.SymbolSearcher:
		if m.MapSymbolSearcherJob != nil {
			j = m.MapSymbolSearcherJob(j)
		}
		return j

	case *run.RepoSearch:
		if m.MapRepoSearchJob != nil {
			j = m.MapRepoSearchJob(j)
		}
		return j

	case *zoekt.GlobalSearch:
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

	case *repoPagerJob:
		child := m.Map(j.child)
		j.child = child
		if m.MapRepoPagerJob != nil {
			j = m.MapRepoPagerJob(j)
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
		children := make([]Job, 0, len(j.children))
		for _, child := range j.children {
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

	case *selectJob:
		child := m.Map(j.child)
		filter := j.path
		if m.MapLimitJob != nil {
			filter, child = m.MapSelectJob(filter, child)
		}
		return NewSelectJob(filter, child)

	case *alertJob:
		child := m.Map(j.child)
		inputs := j.inputs
		if m.MapLimitJob != nil {
			inputs, child = m.MapAlertJob(inputs, child)
		}
		return NewAlertJob(inputs, child)

	case *subRepoPermsFilterJob:
		child := m.Map(j.child)
		if m.MapSubRepoPermsFilterJob != nil {
			child = m.MapSubRepoPermsFilterJob(child)
		}
		return NewFilterJob(child)

	case *noopJob:
		return j

	default:
		panic(fmt.Sprintf("unsupported job %T for job.Mapper", job))
	}
}

func MapAtom(j Job, f func(Job) Job) Job {
	mapper := Mapper{
		MapJob: func(currentJob Job) Job {
			switch typedJob := currentJob.(type) {
			case
				*zoekt.ZoektRepoSubsetSearch,
				*zoekt.ZoektSymbolSearch,
				*searcher.Searcher,
				*searcher.SymbolSearcher,
				*run.RepoSearch,
				*structural.StructuralSearch,
				*commit.CommitSearch,
				*symbol.RepoUniverseSymbolSearch,
				*repos.ComputeExcludedRepos,
				*noopJob:
				return f(typedJob)
			default:
				return currentJob
			}
		},
	}
	return mapper.Map(j)
}
