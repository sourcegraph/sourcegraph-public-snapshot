package aggregation

import (
	"time"

	"github.com/go-enry/go-enry/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/client"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AggregationMatchResult struct {
	Key   MatchKey
	Count int
}

type SearchResultsAggregator interface {
	streaming.Sender
	ShardTimeoutOccurred() bool
}

type AggregationTabulator func(*AggregationMatchResult, error)
type OnMatches func(matches []result.Match)

type eventMatch struct {
	Repo        string
	RepoID      int32
	Path        string
	Commit      string
	Author      string
	Date        time.Time
	Lang        string
	ResultCount int
	Content     string
}

// NewEventEnvironment maps event matches into a consistent type
func newEventMatch(event result.Match) *eventMatch {
	switch match := event.(type) {
	case *result.FileMatch:
		lang, _ := enry.GetLanguageByExtension(match.Path)
		return &eventMatch{
			Repo:        string(match.Repo.Name),
			RepoID:      int32(match.Repo.ID),
			Path:        match.Path,
			Lang:        lang,
			ResultCount: match.ResultCount(),
		}
	case *result.RepoMatch:
		return &eventMatch{
			Repo:        string(match.RepoName().Name),
			RepoID:      int32(match.RepoName().ID),
			ResultCount: 1,
		}
	case *result.CommitMatch:
		return &eventMatch{
			Repo:        string(match.Repo.Name),
			RepoID:      int32(match.Repo.ID),
			Author:      match.Commit.Author.Name,
			Date:        match.Commit.Author.Date,
			ResultCount: 1, //TODO(chwarwick): Verify that we want to count commits not matches in the commit
		}
	case *result.CommitDiffMatch:
		return &eventMatch{
			Repo:        string(match.Repo.Name),
			RepoID:      int32(match.Repo.ID),
			Author:      match.Commit.Author.Name,
			Date:        match.Commit.Author.Date,
			ResultCount: 1, //TODO(chwarwick): Verify that we want to count commits not matches in the commit
		}

	default:
		return &eventMatch{}
	}
}

type AggregationCountFunc func(result.Match) (MatchKey, int, error)
type MatchKey struct {
	Repo   string
	RepoID int32
	Group  string
}

func countRepo(r result.Match) (MatchKey, int, error) {
	match := newEventMatch(r)
	if match.Repo != "" {
		return MatchKey{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Repo,
		}, match.ResultCount, nil
	}
	return MatchKey{}, 0, nil
}

func countLang(r result.Match) (MatchKey, int, error) {
	match := newEventMatch(r)
	if match.Lang != "" {
		return MatchKey{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Lang,
		}, match.ResultCount, nil
	}
	return MatchKey{}, 0, nil
}

func countPath(r result.Match) (MatchKey, int, error) {
	match := newEventMatch(r)
	if match.Path != "" {
		return MatchKey{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Path,
		}, match.ResultCount, nil
	}
	return MatchKey{}, 0, nil
}

func countAuthor(r result.Match) (MatchKey, int, error) {
	match := newEventMatch(r)
	if match.Author != "" {
		return MatchKey{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Author,
		}, match.ResultCount, nil
	}
	return MatchKey{}, 0, nil
}

func GetCountFuncForMode(query, patternType string, mode types.SearchAggregationMode) (AggregationCountFunc, error) {
	modeCountTypes := map[types.SearchAggregationMode]AggregationCountFunc{
		types.REPO_AGGREGATION_MODE:   countRepo,
		types.PATH_AGGREGATION_MODE:   countPath,
		types.AUTHOR_AGGREGATION_MODE: countAuthor,
	}

	modeCountFunc, ok := modeCountTypes[mode]
	if !ok {
		return nil, errors.Newf("unsupported aggregation mode: %s for query", mode)
	}
	return modeCountFunc, nil
}

func NewSearchResultsAggregator(tabulator AggregationTabulator, countFunc AggregationCountFunc) SearchResultsAggregator {
	return &searchAggregationResults{
		tabulator: tabulator,
		countFunc: countFunc,
	}
}

type searchAggregationResults struct {
	tabulator AggregationTabulator
	countFunc AggregationCountFunc
	progress  client.ProgressAggregator
}

func (r *searchAggregationResults) ShardTimeoutOccurred() bool {
	for _, skip := range r.progress.Current().Skipped {
		if skip.Reason == api.ShardTimeout {
			return true
		}
	}

	return false
}

func (r *searchAggregationResults) Send(event streaming.SearchEvent) {
	r.progress.Update(event)
	combined := map[MatchKey]int{}
	for _, match := range event.Results {
		key, count, err := r.countFunc(match)
		// delegate error handling to the passed in tabulator
		if err != nil {
			r.tabulator(nil, err)
			continue
		}
		if count == 0 {
			continue
		}
		current, _ := combined[key]
		combined[key] = current + count
	}
	for key, count := range combined {
		r.tabulator(&AggregationMatchResult{Key: key, Count: count}, nil)
	}
}
