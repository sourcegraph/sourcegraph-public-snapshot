package job

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

// Job is an interface shared by all individual search operations in the
// backend (e.g., text vs commit vs symbol search are represented as different
// jobs) as well as combinations over those searches (run a set in parallel,
// timeout). Calling Run on a job object runs a search.
//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/search/job -i Job -o job_mock_test.go
type Job interface {
	Run(context.Context, database.DB, streaming.Sender) (*search.Alert, error)
	Name() string
}

var allJobs = []Job{
	&zoekt.ZoektRepoSubsetSearch{},
	&searcher.Searcher{},
	&run.RepoSearch{},
	&textsearch.RepoUniverseTextSearch{},
	&structural.StructuralSearch{},
	&commit.CommitSearch{},
	&symbol.RepoSubsetSymbolSearch{},
	&symbol.RepoUniverseSymbolSearch{},
	&repos.ComputeExcludedRepos{},
	&noopJob{},

	&repoPagerJob{},

	&AndJob{},
	&OrJob{},
	&PriorityJob{},
	&ParallelJob{},

	&TimeoutJob{},
	&LimitJob{},
	&subRepoPermsFilterJob{},
	&selectJob{},
	&alertJob{},
}
