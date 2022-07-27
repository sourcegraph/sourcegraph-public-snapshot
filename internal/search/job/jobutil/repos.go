package jobutil

import (
	"context"
	"github.com/grafana/regexp"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
	RepoOpts search.RepoOptions
}

func (s *RepoSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	repos := searchrepos.NewResolver(clients.Logger, clients.DB, clients.SearcherURLs, clients.Zoekt)
	err = repos.Paginate(ctx, s.RepoOpts, func(page *searchrepos.Resolved) (err error) {
		tr.LogFields(log.Int("resolved.len", len(page.RepoRevs)))

		descriptionMatches := map[api.RepoID][]result.Range{}
		if len(s.RepoOpts.DescriptionPatterns) > 0 {
			repoDescriptions, err := s.repoDescriptions(ctx, clients.DB, page.RepoRevs)
			if err != nil {
				return err
			}
			descriptionMatches = s.descriptionMatchRanges(repoDescriptions, s.RepoOpts.DescriptionPatterns)
		}

		stream.Send(streaming.SearchEvent{
			Results: repoRevsToRepoMatches(page.RepoRevs, descriptionMatches),
		})

		return nil
	})

	// Do not error with no results for repo search. For text search, this is an
	// actionable error, but for repo search, it is not.
	err = errors.Ignore(err, errors.IsPred(searchrepos.ErrNoResolvedRepos))
	return nil, err
}

func (s *RepoSearchJob) repoDescriptions(ctx context.Context, db database.DB, repoRevs []*search.RepositoryRevisions) (map[api.RepoID]string, error) {
	repoIDs := make([]api.RepoID, 0, len(repoRevs))
	for _, repoRev := range repoRevs {
		repoIDs = append(repoIDs, repoRev.Repo.ID)
	}

	repoDescriptions, err := db.Repos().GetRepoDescriptionsByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	return repoDescriptions, nil
}

func (s *RepoSearchJob) descriptionMatchRanges(repoDescriptions map[api.RepoID]string, descriptionPatterns []string) map[api.RepoID][]result.Range {
	res := make(map[api.RepoID][]result.Range, len(repoDescriptions))

	for repoID, repoDescription := range repoDescriptions {
		for i, dp := range descriptionPatterns {
			rg, err := regexp.Compile(`(?i)` + dp)
			if err != nil {
				// `dp` is invalid regex, skip this pattern and remove it from descriptionPatterns so we don't check it again
				descriptionPatterns[i] = descriptionPatterns[len(descriptionPatterns)-1]
				descriptionPatterns = descriptionPatterns[:len(descriptionPatterns)-1]
				continue
			}

			submatches := rg.FindAllStringSubmatchIndex(repoDescription, -1)
			if len(submatches) > 0 {
				for _, sm := range submatches {
					res[repoID] = append(res[repoID], result.Range{
						Start: result.Location{
							Offset: sm[0],
							Line:   0, // TODO: what happens if description contains a newline?
							Column: sm[0],
						},
						End: result.Location{
							Offset: sm[1],
							Line:   0,
							Column: sm[1],
						},
					})
				}
			}
		}
	}

	return res
}

func (*RepoSearchJob) Name() string {
	return "RepoSearchJob"
}

func (s *RepoSearchJob) Fields(v job.Verbosity) (res []log.Field) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Scoped("repoOpts", s.RepoOpts.Tags()...),
		)
	}
	return res
}

func (s *RepoSearchJob) Children() []job.Describer       { return nil }
func (s *RepoSearchJob) MapChildren(job.MapFunc) job.Job { return s }

func repoRevsToRepoMatches(repos []*search.RepositoryRevisions, descriptionMatches map[api.RepoID][]result.Range) []result.Match {
	matches := make([]result.Match, 0, len(repos))
	for _, r := range repos {
		for _, rev := range r.Revs {
			rm := &result.RepoMatch{
				Name: r.Repo.Name,
				ID:   r.Repo.ID,
				Rev:  rev,
			}
			if len(descriptionMatches) > 0 {
				if dms, ok := descriptionMatches[r.Repo.ID]; ok {
					rm.DescriptionMatches = dms
				}
			}

			matches = append(matches, rm)
		}
	}
	return matches
}
