package aggregation

import (
	"regexp"
	"time"

	"github.com/go-enry/go-enry/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
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
	Repo         string
	RepoID       int32
	Path         string
	Commit       string
	Author       string
	Date         time.Time
	Lang         string
	ResultCount  int
	Content      string
	ChunkMatches result.ChunkMatches
}

// NewEventEnvironment maps event matches into a consistent type
func newEventMatch(event result.Match) *eventMatch {
	switch match := event.(type) {
	case *result.FileMatch:
		lang, _ := enry.GetLanguageByExtension(match.Path)
		return &eventMatch{
			Repo:         string(match.Repo.Name),
			RepoID:       int32(match.Repo.ID),
			Path:         match.Path,
			Lang:         lang,
			ResultCount:  match.ResultCount(),
			ChunkMatches: match.ChunkMatches,
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

type AggregationCountFunc func(result.Match) (map[MatchKey]int, error)
type MatchKey struct {
	Repo   string
	RepoID int32
	Group  string
}

func countRepo(r result.Match) (map[MatchKey]int, error) {
	match := newEventMatch(r)
	if match.Repo != "" {
		return map[MatchKey]int{{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Repo,
		}: match.ResultCount}, nil
	}
	return nil, nil
}

func countLang(r result.Match) (map[MatchKey]int, error) {
	match := newEventMatch(r)
	if match.Lang != "" {
		return map[MatchKey]int{{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Lang,
		}: match.ResultCount}, nil
	}
	return nil, nil
}

func countPath(r result.Match) (map[MatchKey]int, error) {
	match := newEventMatch(r)
	if match.Path != "" {
		return map[MatchKey]int{{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Path,
		}: match.ResultCount}, nil
	}
	return nil, nil
}

func countAuthor(r result.Match) (map[MatchKey]int, error) {
	match := newEventMatch(r)
	if match.Author != "" {
		return map[MatchKey]int{{
			RepoID: match.RepoID,
			Repo:   match.Repo,
			Group:  match.Author,
		}: match.ResultCount}, nil
	}
	return nil, nil
}

func countCaptureGroupsFunc(querystring string) (AggregationCountFunc, error) {
	pattern, err := getCasedPattern(querystring)
	if err != nil {
		return nil, errors.Wrap(err, "getCasedPattern")
	}
	regexp, err := regexp.Compile(pattern.String())
	if err != nil {
		return nil, errors.Wrap(err, "Could not compile regexp")
	}

	return func(r result.Match) (map[MatchKey]int, error) {
		match := newEventMatch(r)
		if len(match.ChunkMatches) != 0 {
			matches := map[MatchKey]int{}
			for _, cm := range match.ChunkMatches {
				for _, range_ := range cm.Ranges {
					content := chunkContent(cm, range_)
					for _, submatches := range regexp.FindAllStringSubmatchIndex(content, -1) {
						chunkMatches := fromRegexpMatches(submatches, regexp.SubexpNames(), content, range_)
						for value, count := range chunkMatches {
							key := MatchKey{Repo: string(r.RepoName().Name), RepoID: int32(r.RepoName().ID), Group: value}
							current := matches[key]
							matches[key] = current + count
						}
					}
				}
			}
			return matches, nil
		}
		return nil, nil
	}, nil
}

func GetCountFuncForMode(query, patternType string, mode types.SearchAggregationMode) (AggregationCountFunc, error) {
	modeCountTypes := map[types.SearchAggregationMode]AggregationCountFunc{
		types.REPO_AGGREGATION_MODE:   countRepo,
		types.PATH_AGGREGATION_MODE:   countPath,
		types.AUTHOR_AGGREGATION_MODE: countAuthor,
	}

	if mode == types.CAPTURE_GROUP_AGGREGATION_MODE {
		captureGroupsCount, err := countCaptureGroupsFunc(query)
		if err != nil {
			return nil, err
		}
		modeCountTypes[types.CAPTURE_GROUP_AGGREGATION_MODE] = captureGroupsCount
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
		groups, err := r.countFunc(match)
		for groupKey, count := range groups {
			// delegate error handling to the passed in tabulator
			if err != nil {
				r.tabulator(nil, err)
				continue
			}
			if groups == nil {
				continue
			}
			current, _ := combined[groupKey]
			combined[groupKey] = current + count
		}
	}
	for key, count := range combined {
		r.tabulator(&AggregationMatchResult{Key: key, Count: count}, nil)
	}
}

// Pulls the pattern out of the querystring
// If the query contains a case:no field, we need to wrap the pattern in some additional regex.
func getCasedPattern(querystring string) (MatchPattern, error) {
	query, err := querybuilder.ParseQuery(querystring, "regexp")
	if err != nil {
		return nil, errors.Wrap(err, "ParseQuery")
	}
	q := query.ToQ()

	if len(query) != 1 {
		// Not sure when we would run into this; calling it out to help during testing.
		return nil, errors.New("Pipeline generated plan with multiple steps.")
	}
	basic := query[0]

	pattern, err := extractPattern(&basic)
	if err != nil {
		return nil, err
	}
	patternValue := pattern.Value
	if !q.IsCaseSensitive() {
		patternValue = "(?i:" + pattern.Value + ")"
	}
	casedPattern, err := toRegexpPattern(patternValue)
	if err != nil {
		return nil, err
	}
	return casedPattern, nil
}
