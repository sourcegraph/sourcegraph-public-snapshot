package run

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoSearchJob struct {
	RepoOpts                     search.RepoOptions
	FilePatternsReposMustInclude []string
	FilePatternsReposMustExclude []string
	Features                     search.Features

	Mode search.GlobalSearchMode
}

func (s *RepoSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	repos := searchrepos.NewResolver(clients.DB)
	err = repos.Paginate(ctx, s.RepoOpts, func(page *searchrepos.Resolved) error {
		tr.LogFields(log.Int("resolved.len", len(page.RepoRevs)))

		// Filter the repos if there is a repohasfile: or -repohasfile field.
		if len(s.FilePatternsReposMustExclude) > 0 || len(s.FilePatternsReposMustInclude) > 0 {
			// Fallback to batch for reposToAdd
			page.RepoRevs, err = s.reposToAdd(ctx, clients, page.RepoRevs)
			if err != nil {
				return err
			}
		}

		stream.Send(streaming.SearchEvent{
			Results: repoRevsToRepoMatches(ctx, clients.DB, page.RepoRevs),
		})

		return nil
	})

	// Do not error with no results for repo search. For text search, this is an
	// actionable error, but for repo search, it is not.
	err = errors.Ignore(err, errors.IsPred(searchrepos.ErrNoResolvedRepos))
	return nil, err
}

func (*RepoSearchJob) Name() string {
	return "RepoSearchJob"
}

func (s *RepoSearchJob) Fields(v job.Verbosity) (res []log.Field) {
	switch v {
	case job.VerbosityMax:
		res = append(res,
			log.Bool("contentBasedLangFilters", s.Features.ContentBasedLangFilters),
			trace.Strings("filePatternsReposMustInclude", s.FilePatternsReposMustInclude),
			trace.Strings("filePatternsReposMustExclude", s.FilePatternsReposMustExclude),
		)
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Scoped("repoOpts", s.RepoOpts.Tags()...),
			trace.Stringer("mode", s.Mode),
		)
	}
	return res
}

func (s *RepoSearchJob) Children() []job.Describer       { return nil }
func (s *RepoSearchJob) MapChildren(job.MapFunc) job.Job { return s }

func repoRevsToRepoMatches(ctx context.Context, db database.DB, repos []*search.RepositoryRevisions) []result.Match {
	matches := make([]result.Match, 0, len(repos))
	for _, r := range repos {
		revs, err := r.ExpandedRevSpecs(ctx, db)
		if err != nil { // fallback to just return revspecs
			revs = r.RevSpecs()
		}
		for _, rev := range revs {
			matches = append(matches, &result.RepoMatch{
				Name: r.Repo.Name,
				ID:   r.Repo.ID,
				Rev:  rev,
			})
		}
	}
	return matches
}

func matchesToFileMatches(matches []result.Match) ([]*result.FileMatch, error) {
	fms := make([]*result.FileMatch, 0, len(matches))
	for _, match := range matches {
		fm, ok := match.(*result.FileMatch)
		if !ok {
			return nil, errors.Errorf("expected only file match results")
		}
		fms = append(fms, fm)
	}
	return fms, nil
}
