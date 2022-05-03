package jobutil

import (
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

var allJobs = []job.Job{
	&zoekt.ZoektRepoSubsetSearch{},
	&zoekt.ZoektSymbolSearch{},
	&searcher.Searcher{},
	&searcher.SymbolSearcher{},
	&run.RepoSearch{},
	&zoekt.GlobalSearch{},
	&structural.StructuralSearch{},
	&commit.CommitSearch{},
	&symbol.RepoUniverseSymbolSearch{},
	&repos.ComputeExcludedRepos{},
	&noopJob{},

	&repoPagerJob{},

	&AndJob{},
	&OrJob{},
	&ParallelJob{},
	&SequentialJob{},

	&TimeoutJob{},
	&LimitJob{},
	&subRepoPermsFilterJob{},
	&selectJob{},
	&alertJob{},
}
