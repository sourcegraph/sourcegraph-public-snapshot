package zoekt

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type SymbolSearchJob struct {
	Repos          *IndexedRepoRevs // the set of indexed repository revisions to search.
	Query          zoektquery.Q
	FileMatchLimit int32
	Select         filter.SelectPath
	Since          func(time.Time) time.Duration `json:"-"` // since if non-nil will be used instead of time.Since. For tests
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

	err = zoektSearch(ctx, z.Repos, z.Query, nil, search.SymbolRequest, clients.Zoekt, z.FileMatchLimit, z.Select, since, stream)
	if err != nil {
		tr.LogFields(log.Error(err))
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

func (z *SymbolSearchJob) Fields(v job.Verbosity) (res []log.Field) {
	switch v {
	case job.VerbosityMax:
		res = append(res,
			log.Int32("fileMatchLimit", z.FileMatchLimit),
			trace.Stringer("select", z.Select),
		)
		// z.Repos is nil for un-indexed search
		if z.Repos != nil {
			res = append(res,
				log.Int("numRepoRevs", len(z.Repos.RepoRevs)),
				log.Int("numBranchRepos", len(z.Repos.branchRepos)),
			)
		}
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Stringer("query", z.Query),
		)
	}
	return res
}

func (z *SymbolSearchJob) Children() []job.Describer       { return nil }
func (z *SymbolSearchJob) MapChildren(job.MapFunc) job.Job { return z }

type GlobalSymbolSearchJob struct {
	GlobalZoektQuery *GlobalZoektQuery
	ZoektArgs        *search.ZoektParameters
	RepoOpts         search.RepoOptions
}

func (s *GlobalSymbolSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	userPrivateRepos := privateReposForActor(ctx, clients.Logger, clients.DB, s.RepoOpts)
	s.GlobalZoektQuery.ApplyPrivateFilter(userPrivateRepos)
	s.ZoektArgs.Query = s.GlobalZoektQuery.Generate()

	// always search for symbols in indexed repositories when searching the repo universe.
	err = DoZoektSearchGlobal(ctx, clients.Zoekt, s.ZoektArgs, nil, stream)
	if err != nil {
		tr.LogFields(log.Error(err))
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

func (s *GlobalSymbolSearchJob) Fields(v job.Verbosity) (res []log.Field) {
	switch v {
	case job.VerbosityMax:
		res = append(res,
			trace.Printf("repoScope", "%q", s.GlobalZoektQuery.RepoScope),
			log.Bool("includePrivate", s.GlobalZoektQuery.IncludePrivate),
			log.Int32("fileMatchLimit", s.ZoektArgs.FileMatchLimit),
			trace.Stringer("select", s.ZoektArgs.Select),
		)
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Stringer("query", s.GlobalZoektQuery.Query),
			log.String("type", string(s.ZoektArgs.Typ)),
			trace.Scoped("repoOpts", s.RepoOpts.Tags()...),
		)
	}
	return res
}

func (s *GlobalSymbolSearchJob) Children() []job.Describer       { return nil }
func (s *GlobalSymbolSearchJob) MapChildren(job.MapFunc) job.Job { return s }
