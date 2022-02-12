package run

import (
	"strings"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
)

func prettyPrint(job Job) string {
	switch j := job.(type) {
	case
		*RepoSearch,
		*textsearch.RepoSubsetTextSearch,
		*textsearch.RepoUniverseTextSearch,
		*structural.StructuralSearch,
		*commit.CommitSearch,
		*symbol.RepoSubsetSymbolSearch,
		*symbol.RepoUniverseSymbolSearch,
		*repos.ComputeExcludedRepos:
		return j.Name()
	case *AndJob:
		var jobs []string
		for _, child := range j.children {
			jobs = append(jobs, prettyPrint(child))
		}
		return "and(" + strings.Join(jobs, ",") + ")"
	case *OrJob:
		var jobs []string
		for _, child := range j.children {
			jobs = append(jobs, prettyPrint(child))
		}
		return "or(" + strings.Join(jobs, ",") + ")"
	case
		*PriorityJob,
		*ParallelJob,
		*TimeoutJob,
		*LimitJob,
		*subRepoPermsFilterJob:
		return j.Name()
	case *noopJob:
		return "NoOp"
	}
	// Unreachable.
	return ""
}

func TestMap(t *testing.T) {
	test := func(job Job, mapper Mapper) string {
		return prettyPrint(mapper.Map(job))
	}

	andMapper := Mapper{
		MapAndJob: func(children []Job) []Job {
			return append(children, NewOrJob(NewNoopJob(), NewNoopJob()))
		},
	}
	autogold.Want("basic and-job mapper", "and(NoOp,NoOp,or(NoOp,NoOp))").Equal(t, test(NewAndJob(NewNoopJob(), NewNoopJob()), andMapper))
}
