package zoekt

import (
	"context"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type ZoektSymbolSearch struct {
	Repos          *IndexedRepoRevs // the set of indexed repository revisions to search.
	Query          zoektquery.Q
	FileMatchLimit int32
	Select         filter.SelectPath
	Zoekt          zoekt.Streamer
	Since          func(time.Time) time.Duration `json:"-"` // since if non-nil will be used instead of time.Since. For tests
}

// Run calls the zoekt backend to search symbols
func (z *ZoektSymbolSearch) Run(ctx context.Context, _ database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx := jobutil.StartSpan(ctx, z)
	defer func() { jobutil.FinishSpan(tr, alert, err) }()

	if z.Repos == nil {
		return nil, nil
	}
	if len(z.Repos.RepoRevs) == 0 {
		return nil, nil
	}

	since := time.Since
	if z.Since != nil {
		since = z.Since
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err = zoektSearch(ctx, z.Repos, z.Query, search.SymbolRequest, z.Zoekt, z.FileMatchLimit, z.Select, since, stream)
	if err != nil {
		tr.LogFields(otlog.Error(err))
		// Only record error if we haven't timed out.
		if ctx.Err() == nil {
			cancel()
			return nil, err
		}
	}
	return nil, nil
}

func (z *ZoektSymbolSearch) Name() string {
	return "ZoektSymbolSearch"
}
