package jobutil

import (
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/structural"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

var allJobs = []job.Job{
	&zoekt.RepoSubsetTextSearchJob{},
	&zoekt.SymbolSearchJob{},
	&searcher.TextSearchJob{},
	&searcher.SymbolSearchJob{},
	&run.RepoSearchJob{},
	&zoekt.GlobalTextSearchJob{},
	&structural.SearchJob{},
	&commit.SearchJob{},
	&zoekt.GlobalSymbolSearchJob{},
	&repos.ComputeExcludedJob{},
	&NoopJob{},

	&repoPagerJob{},
	&generatedSearchJob{},
	&FeelingLuckySearchJob{},

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
