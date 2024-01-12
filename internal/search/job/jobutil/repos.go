package jobutil

import (
	"context"
	"unicode/utf8"

	"github.com/grafana/regexp"
	"go.opentelemetry.io/otel/attribute"

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
	RepoOpts            search.RepoOptions
	DescriptionPatterns []*regexp.Regexp
	RepoNamePatterns    []*regexp.Regexp // used for getting repo name match ranges
}

func (s *RepoSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	repos := searchrepos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SearcherURLs, clients.SearcherGRPCConnectionCache, clients.Zoekt)
	it := repos.Iterator(ctx, s.RepoOpts)

	for it.Next() {
		page := it.Current()
		tr.SetAttributes(attribute.Int("resolved.len", len(page.RepoRevs)))
		page.MaybeSendStats(stream)

		descriptionMatches := make(map[api.RepoID][]result.Range)
		if len(s.DescriptionPatterns) > 0 {
			repoDescriptionsSet, err := s.repoDescriptions(ctx, clients.DB, page.RepoRevs)
			if err != nil {
				return nil, err
			}
			descriptionMatches = s.descriptionMatchRanges(repoDescriptionsSet)
		}

		stream.Send(streaming.SearchEvent{
			Results: repoRevsToRepoMatches(page.RepoRevs, s.RepoNamePatterns, descriptionMatches),
		})
	}

	// Do not error with no results for repo search. For text search, this is an
	// actionable error, but for repo search, it is not.
	err = errors.Ignore(it.Err(), errors.IsPred(searchrepos.ErrNoResolvedRepos))
	return nil, err
}

// repoDescriptions gets the repo ID and repo description from the database for each of the repos in repoRevs, and returns
// a map of repo ID to repo description.
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

// descriptionMatchRanges takes a map of repo IDs to their descriptions, and a list of patterns to match against those repo descriptions.
// It returns a map of repo IDs to []result.Range. The []result.Range value contains the match ranges
// for repos with a description that matches at least one of the patterns in descriptionPatterns.
func (s *RepoSearchJob) descriptionMatchRanges(repoDescriptions map[api.RepoID]string) map[api.RepoID][]result.Range {
	res := make(map[api.RepoID][]result.Range)

	for repoID, repoDescription := range repoDescriptions {
		for _, re := range s.DescriptionPatterns {
			submatches := re.FindAllStringSubmatchIndex(repoDescription, -1)
			for _, sm := range submatches {
				res[repoID] = append(res[repoID], result.Range{
					Start: result.Location{
						Offset: sm[0],
						Line:   0,
						Column: utf8.RuneCountInString(repoDescription[:sm[0]]),
					},
					End: result.Location{
						Offset: sm[1],
						Line:   0,
						Column: utf8.RuneCountInString(repoDescription[:sm[1]]),
					},
				})
			}
		}
	}

	return res
}

func (*RepoSearchJob) Name() string {
	return "RepoSearchJob"
}

func (s *RepoSearchJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res, trace.Scoped("repoOpts", s.RepoOpts.Attributes()...)...)
		res = append(res, trace.Stringers("repoNamePatterns", s.RepoNamePatterns))
	}
	return res
}

func (s *RepoSearchJob) Children() []job.Describer       { return nil }
func (s *RepoSearchJob) MapChildren(job.MapFunc) job.Job { return s }

func repoRevsToRepoMatches(repos []*search.RepositoryRevisions, repoNameRegexps []*regexp.Regexp, descriptionMatches map[api.RepoID][]result.Range) []result.Match {
	matches := make([]result.Match, 0, len(repos))

	for _, r := range repos {
		// Get repo name matches once per repo
		repoNameMatches := repoMatchRanges(string(r.Repo.Name), repoNameRegexps)

		for _, rev := range r.Revs {
			rm := result.RepoMatch{
				Name:            r.Repo.Name,
				ID:              r.Repo.ID,
				Rev:             rev,
				RepoNameMatches: repoNameMatches,
			}
			if ranges, ok := descriptionMatches[r.Repo.ID]; ok {
				rm.DescriptionMatches = ranges
			}
			matches = append(matches, &rm)
		}
	}
	return matches
}

func repoMatchRanges(repoName string, repoNameRegexps []*regexp.Regexp) (res []result.Range) {
	for _, repoNameRe := range repoNameRegexps {
		submatches := repoNameRe.FindAllStringSubmatchIndex(repoName, -1)
		for _, sm := range submatches {
			res = append(res, result.Range{
				Start: result.Location{
					Offset: sm[0],
					Line:   0, // we can treat repo names as single-line
					Column: utf8.RuneCountInString(repoName[:sm[0]]),
				},
				End: result.Location{
					Offset: sm[1],
					Line:   0,
					Column: utf8.RuneCountInString(repoName[:sm[1]]),
				},
			})
		}
	}

	return res
}
