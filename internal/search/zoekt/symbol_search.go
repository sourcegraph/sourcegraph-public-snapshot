package zoekt

import (
	"context"
	"time"

	zoektquery "github.com/sourcegraph/zoekt/query"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type SymbolSearchJob struct {
	Repos       *IndexedRepoRevs // the set of indexed repository revisions to search.
	Query       zoektquery.Q
	ZoektParams *search.ZoektParameters
	Since       func(time.Time) time.Duration `json:"-"` // since if non-nil will be used instead of time.Since. For tests
}

// Run calls the zoekt backend to search symbols
func (z *SymbolSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, z)
	defer func() { finish(alert, err) }()

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

	err = zoektSearch(ctx, z.Repos, z.Query, nil, search.SymbolRequest, clients.Zoekt, z.ZoektParams, since, stream)
	if err != nil {
		tr.SetAttributes(trace.Error(err))
		// Only record error if we haven't timed out.
		if ctx.Err() == nil {
			cancel()
			return nil, err
		}
	}
	return nil, nil
}

func (z *SymbolSearchJob) Name() string {
	return "ZoektSymbolSearchJob"
}

func (z *SymbolSearchJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		res = append(res,
			attribute.Int("fileMatchLimit", int(z.ZoektParams.FileMatchLimit)),
			attribute.Stringer("select", z.ZoektParams.Select),
		)
		// z.Repos is nil for un-indexed search
		if z.Repos != nil {
			res = append(res,
				attribute.Int("numRepoRevs", len(z.Repos.RepoRevs)),
				attribute.Int("numBranchRepos", len(z.Repos.branchRepos)),
			)
		}
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			attribute.Stringer("query", z.Query),
		)
	}
	return res
}

func (z *SymbolSearchJob) Children() []job.Describer       { return nil }
func (z *SymbolSearchJob) MapChildren(job.MapFunc) job.Job { return z }

type GlobalSymbolSearchJob struct {
	GlobalZoektQuery *GlobalZoektQuery
	ZoektParams      *search.ZoektParameters
	RepoOpts         search.RepoOptions
}

func (s *GlobalSymbolSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	userPrivateRepos := privateReposForActor(ctx, clients.Logger, clients.DB, s.RepoOpts)
	s.GlobalZoektQuery.ApplyPrivateFilter(userPrivateRepos)
	s.ZoektParams.Query = s.GlobalZoektQuery.Generate()

	// always search for symbols in indexed repositories when searching the repo universe.
	err = DoZoektSearchGlobal(ctx, clients.Zoekt, s.ZoektParams, nil, stream)
	if err != nil {
		tr.SetAttributes(trace.Error(err))
		// Only record error if we haven't timed out.
		if ctx.Err() == nil {
			return nil, err
		}
	}

	return nil, nil
}

func (*GlobalSymbolSearchJob) Name() string {
	return "ZoektGlobalSymbolSearchJob"
}

func (s *GlobalSymbolSearchJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		res = append(res,
			trace.Stringers("repoScope", s.GlobalZoektQuery.RepoScope),
			attribute.Bool("includePrivate", s.GlobalZoektQuery.IncludePrivate),
			attribute.Int("fileMatchLimit", int(s.ZoektParams.FileMatchLimit)),
			attribute.Stringer("select", s.ZoektParams.Select),
		)
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			attribute.Stringer("query", s.GlobalZoektQuery.Query),
			attribute.String("type", string(s.ZoektParams.Typ)),
		)
		res = append(res, trace.Scoped("repoOpts", s.RepoOpts.Attributes()...)...)
	}
	return res
}

func (s *GlobalSymbolSearchJob) Children() []job.Describer       { return nil }
func (s *GlobalSymbolSearchJob) MapChildren(job.MapFunc) job.Job { return s }
