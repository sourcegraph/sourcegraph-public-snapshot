package job

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
)

type Mapper struct {
	MapJob func(job Job) Job

	// Search Jobs (leaf nodes)
	MapRepoSearchJob               func(*run.RepoSearch) *run.RepoSearch
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
	case *run.RepoSearch:
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
