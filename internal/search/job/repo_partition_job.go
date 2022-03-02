package job

import (
	"context"
	"fmt"

	zoektstreamer "github.com/google/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type PartitionRepos struct {
	RepoOptions      search.RepoOptions
	UseIndex         query.YesNoOnly // whether to include indexed repos
	ContainsRefGlobs bool            // wether to include repositories with refs
	Jobs             []Job           // child jobs to receive partitioned repos.

	Zoekt zoektstreamer.Streamer
}

func (p *PartitionRepos) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "PartitionRepos", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repoResolver := &repos.Resolver{DB: db, Opts: p.RepoOptions}
	pager := func(page *repos.Resolved) error {
		indexed, unindexed, err := zoekt.PartitionRepos(
			ctx,
			page.RepoRevs,
			p.Zoekt,
			search.TextRequest,
			p.UseIndex,
			p.ContainsRefGlobs,
			zoekt.MissingRepoRevStatus(stream),
		)
		if err != nil {
			return err
		}

		for _, job := range p.Jobs {
			switch j := job.(type) {
			case *zoekt.ZoektRepoSubsetSearch:
				j.Repos = indexed

			case *searcher.Searcher:
				j.Repos = unindexed

			default:
				panic(fmt.Sprintf("job %T not supported for repo pagination", job))
			}

		}
		return nil
	}

	return nil, repoResolver.Paginate(ctx, nil, pager)
}
