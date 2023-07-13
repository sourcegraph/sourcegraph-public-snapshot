package aggregation

import (
	"context"
	"sync"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	sApi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/client"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	sTypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AggregationMatchResult struct {
	Key   MatchKey
	Count int
}

type SearchResultsAggregator interface {
	streaming.Sender
	ShardTimeoutOccurred() bool
	ResultLimitHit(limit int) bool
}

type AggregationTabulator func(*AggregationMatchResult, error)
type OnMatches func(matches []result.Match)

type AggregationCountFunc func(result.Match, *sTypes.Repo) (map[MatchKey]int, error)
type MatchKey struct {
	Repo   string
	RepoID int32
	Group  string
}

func countRepo(r result.Match, _ *sTypes.Repo) (map[MatchKey]int, error) {
	if r.RepoName().Name != "" {
		return map[MatchKey]int{{
			RepoID: int32(r.RepoName().ID),
			Repo:   string(r.RepoName().Name),
			Group:  string(r.RepoName().Name),
		}: r.ResultCount()}, nil
	}
	return nil, nil
}

func countPath(r result.Match, _ *sTypes.Repo) (map[MatchKey]int, error) {
	var path string
	switch match := r.(type) {
	case *result.FileMatch:
		path = match.Path
	default:
	}
	if path != "" {
		return map[MatchKey]int{{
			RepoID: int32(r.RepoName().ID),
			Repo:   string(r.RepoName().Name),
			Group:  path,
		}: r.ResultCount()}, nil
	}
	return nil, nil
}

func countAuthor(r result.Match, _ *sTypes.Repo) (map[MatchKey]int, error) {
	var author string
	switch match := r.(type) {
	case *result.CommitMatch:
		author = match.Commit.Author.Name
	default:
	}
	if author != "" {
		return map[MatchKey]int{{
			RepoID: int32(r.RepoName().ID),
			Repo:   string(r.RepoName().Name),
			Group:  author,
		}: r.ResultCount()}, nil
	}
	return nil, nil
}

func countCaptureGroupsFunc(querystring string) (AggregationCountFunc, error) {
	pattern, err := getCasedPattern(querystring)
	if err != nil {
		return nil, errors.Wrap(err, "getCasedPattern")
	}
	regex, err := regexp.Compile(pattern.String())
	if err != nil {
		return nil, errors.Wrap(err, "Could not compile regexp")
	}

	return func(r result.Match, _ *sTypes.Repo) (map[MatchKey]int, error) {
		content := matchContent(r)
		if len(content) != 0 {
			matches := map[MatchKey]int{}
			for _, contentPiece := range content {
				for _, submatches := range regex.FindAllStringSubmatchIndex(contentPiece, -1) {
					contentMatches := fromRegexpMatches(submatches, contentPiece)
					for value, count := range contentMatches {
						key := MatchKey{Repo: string(r.RepoName().Name), RepoID: int32(r.RepoName().ID), Group: value}
						if len(key.Group) > 100 {
							key.Group = key.Group[:100]
						}
						current := matches[key]
						matches[key] = current + count
					}
				}
			}
			return matches, nil
		}
		return nil, nil
	}, nil
}

func matchContent(event result.Match) []string {
	switch match := event.(type) {
	case *result.FileMatch:
		capacity := len(match.ChunkMatches)
		var content = make([]string, 0, capacity)
		if len(match.ChunkMatches) > 0 { // This File match with the subtype of text results
			for _, cm := range match.ChunkMatches {
				for _, range_ := range cm.Ranges {
					content = append(content, chunkContent(cm, range_))
				}
			}
			return content
		} else if len(match.Symbols) > 0 { // This File match with the subtype of symbol results
			return nil
		} else { // This is a File match representing a whole file
			return []string{match.Path}
		}
	case *result.RepoMatch:
		return []string{string(match.RepoName().Name)}
	case *result.CommitMatch:
		if match.DiffPreview != nil { // signals this is a Diff match
			return nil
		} else {
			return []string{string(match.Commit.Message)}
		}
	default:
		return nil
	}
}

func countRepoMetadata(r result.Match, repo *sTypes.Repo) (map[MatchKey]int, error) {
	metadata := map[string]*string{types.NO_REPO_METADATA_TEXT: nil}
	if repo != nil && repo.KeyValuePairs != nil {
		metadata = repo.KeyValuePairs
	}
	matches := map[MatchKey]int{}
	for key, value := range metadata {
		group := key
		if value != nil && *value != "" {
			group += ":" + *value
		}
		matchKey := MatchKey{Repo: string(r.RepoName().Name), RepoID: int32(r.RepoName().ID), Group: group}
		matches[matchKey] = r.ResultCount()
	}
	return matches, nil
}

func GetCountFuncForMode(query, patternType string, mode types.SearchAggregationMode) (AggregationCountFunc, error) {
	modeCountTypes := map[types.SearchAggregationMode]AggregationCountFunc{
		types.REPO_AGGREGATION_MODE:          countRepo,
		types.PATH_AGGREGATION_MODE:          countPath,
		types.AUTHOR_AGGREGATION_MODE:        countAuthor,
		types.REPO_METADATA_AGGREGATION_MODE: countRepoMetadata,
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

func NewSearchResultsAggregatorWithContext(ctx context.Context, tabulator AggregationTabulator, countFunc AggregationCountFunc, db database.DB, mode types.SearchAggregationMode) SearchResultsAggregator {
	return &searchAggregationResults{
		db:        db,
		ctx:       ctx,
		mode:      mode,
		tabulator: tabulator,
		countFunc: countFunc,
		progress: client.ProgressAggregator{
			Start:     time.Now(),
			RepoNamer: client.RepoNamer(ctx, db),
			Trace:     trace.URL(trace.ID(ctx), conf.DefaultClient()),
		},
	}
}

type searchAggregationResults struct {
	db          database.DB
	ctx         context.Context
	mode        types.SearchAggregationMode
	tabulator   AggregationTabulator
	countFunc   AggregationCountFunc
	progress    client.ProgressAggregator
	resultCount int

	mu sync.Mutex
}

func (r *searchAggregationResults) ShardTimeoutOccurred() bool {
	for _, skip := range r.progress.Current().Skipped {
		if skip.Reason == sApi.ShardTimeout {
			return true
		}
	}

	return false
}

func (r *searchAggregationResults) ResultLimitHit(limit int) bool {

	return limit <= r.resultCount
}

func (r *searchAggregationResults) repos(matches result.Matches) (map[api.RepoID]*sTypes.Repo, error) {
	repoIDs := collections.NewSet[api.RepoID]()
	for _, r := range matches {
		repoIDs.Add(r.RepoName().ID)
	}

	res, err := r.db.Repos().List(r.ctx, database.ReposListOptions{IDs: repoIDs.Values()})
	repos := make(map[api.RepoID]*sTypes.Repo, len(res))
	if err != nil {
		return nil, err
	}
	for _, repo := range res {
		repos[repo.ID] = repo
	}
	return repos, nil
}

func (r *searchAggregationResults) Send(event streaming.SearchEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.progress.Update(event)
	r.resultCount += event.Results.ResultCount()
	combined := map[MatchKey]int{}
	repos := make(map[api.RepoID]*sTypes.Repo, 0)
	// initialize repos if we are in repo metadata aggregation mode
	// other modes currently don't use the repo parameter
	if r.mode == types.REPO_METADATA_AGGREGATION_MODE {
		res, err := r.repos(event.Results)
		if err != nil {
			r.tabulator(nil, err)
			return
		}
		repos = res
	}
	for _, match := range event.Results {
		select {
		case <-r.ctx.Done():
			// let the tabulator an error occured.
			err := errors.Wrap(r.ctx.Err(), "tabulation terminated context is done")
			r.tabulator(nil, err)
			return
		default:
			groups, err := r.countFunc(match, repos[match.RepoName().ID])
			for groupKey, count := range groups {
				// delegate error handling to the passed in tabulator
				if err != nil {
					r.tabulator(nil, err)
					continue
				}
				current := combined[groupKey]
				combined[groupKey] = current + count
			}
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
