package aggregation

import (
	"regexp"
	"strconv"
	"strings"
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

func countCaptureGroupsFunc(pattern string) (AggregationCountFunc, error) {
	regexp, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.New("Could not compile regexp")
	}
	return func(r result.Match) (MatchKey, int, error) {
		match := newEventMatch(r)
		if len(match.ChunkMatches) != 0 {
			groups := make([]Match, 0, len(match.ChunkMatches))
			for _, cm := range match.ChunkMatches {
				for _, range_ := range cm.Ranges {
					content := chunkContent(cm, range_)
					for _, submatches := range regexp.FindAllStringSubmatchIndex(content, -1) {
						groups = append(groups, fromRegexpMatches(submatches, regexp.SubexpNames(), content, range_))
					}
				}
			}

			// TODO: What if there's more than one capture group per match? This whole thing might need to return an array
			// of these..
			if len(groups) > 0 {
				return MatchKey{
					RepoID: match.RepoID,
					Repo:   match.Repo,
					Group:  groups[0].Value,
				}, match.ResultCount, nil
			}
		}
		return MatchKey{}, 0, nil
	}, nil
}

func GetCountFuncForMode(query, patternType string, mode types.SearchAggregationMode) (AggregationCountFunc, error) {
	captureGroupsCount, err := countCaptureGroupsFunc(query)
	if err != nil {
		return nil, err
	}

	modeCountTypes := map[types.SearchAggregationMode]AggregationCountFunc{
		types.REPO_AGGREGATION_MODE:          countRepo,
		types.PATH_AGGREGATION_MODE:          countPath,
		types.AUTHOR_AGGREGATION_MODE:        countAuthor,
		types.CAPTURE_GROUP_AGGREGATION_MODE: captureGroupsCount,
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

func substituteRegexp(content string, match *regexp.Regexp, replacePattern, separator string) string {
	var b strings.Builder
	for _, submatches := range match.FindAllStringSubmatchIndex(content, -1) {
		b.Write(match.ExpandString([]byte{}, replacePattern, content, submatches))
		b.WriteString(separator)
	}
	return b.String()
}

type Match struct {
	Value       string      `json:"value"`
	Range       Range       `json:"range"`
	Environment Environment `json:"environment"`
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

type Environment map[string]Data

type Location struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Data struct {
	Value string `json:"value"`
	Range Range  `json:"range"`
}

func chunkContent(c result.ChunkMatch, r result.Range) string {
	// Set range relative to the start of the content.
	rr := r.Sub(c.ContentStart)
	return c.Content[rr.Start.Offset:rr.End.Offset]
}

func fromRegexpMatches(submatches []int, namedGroups []string, content string, range_ result.Range) Match {
	env := make(Environment)
	var firstValue string
	var firstRange Range
	// iterate over pairs of offsets. Cf. FindAllStringSubmatchIndex
	// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmatchIndex.
	for j := 0; j < len(submatches); j += 2 {
		start := submatches[j]
		end := submatches[j+1]
		if start == -1 || end == -1 {
			// The entire regexp matched, but a capture
			// group inside it did not. Ignore this entry.
			continue
		}
		value := content[start:end]
		captureRange := newRange(range_.Start.Offset+start, range_.Start.Offset+end)

		if j == 0 {
			// The first submatch is the overall match
			// value. Don’t add this to the Environment
			firstValue = value
			firstRange = captureRange
			continue
		}

		var v string
		if namedGroups[j/2] == "" {
			v = strconv.Itoa(j / 2)
		} else {
			v = namedGroups[j/2]
		}
		env[v] = Data{Value: value, Range: captureRange}
	}
	return Match{Value: firstValue, Range: firstRange, Environment: env}
}

type Regexp struct {
	Value *regexp.Regexp
}

func newRange(startOffset, endOffset int) Range {
	return Range{
		Start: newLocation(-1, -1, startOffset),
		End:   newLocation(-1, -1, endOffset),
	}
}

func newLocation(line, column, offset int) Location {
	return Location{
		Offset: offset,
		Line:   line,
		Column: column,
	}
}
