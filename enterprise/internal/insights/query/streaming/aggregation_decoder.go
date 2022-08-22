package streaming

import (
	"time"

	"github.com/go-enry/go-enry/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AggregationMatchResult struct {
	Key   MatchKey
	Count int
}

const ShardTimeoutSkippedReason = streamapi.ShardTimeout
const LuckySearchAlertKind = "lucky-search-queries"

type SearchSkipped struct {
	Reason   string
	Severity string
}

type SearchAlert struct {
	Title string
	Kind  string
}

type AggregationDecoderEvents struct {
	Skipped []SearchSkipped
	Errors  []string
	Alerts  []SearchAlert
}

type AggregationTabulator func(*AggregationMatchResult, error)
type OnMatches func(matches []streamhttp.EventMatch)

// AggregationDecoder will tabulate the result using the passed in tabulator
func AggregationDecoder(onMatches OnMatches) (streamhttp.FrontendStreamDecoder, *AggregationDecoderEvents) {
	decoderEvents := &AggregationDecoderEvents{}

	return streamhttp.FrontendStreamDecoder{
		OnProgress: func(progress *streamapi.Progress) {
			if !progress.Done {
				return
			}
			for _, skipped := range progress.Skipped {
				decoderEvents.Skipped = append(decoderEvents.Skipped, SearchSkipped{Reason: string(skipped.Reason), Severity: string(skipped.Severity)})
			}
		},
		OnMatches: onMatches,
		OnAlert: func(ea *streamhttp.EventAlert) {
			if ea.Title == "No repositories found" {
				// If we hit a case where we don't find a repository we don't want to error, just
				// complete our search.
			} else {
				decoderEvents.Alerts = append(decoderEvents.Alerts, SearchAlert{Title: ea.Title, Kind: ea.Kind})
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			decoderEvents.Errors = append(decoderEvents.Errors, eventError.Message)
		},
	}, decoderEvents
}

type eventMatch struct {
	Repo        string
	RepoID      int32
	Path        string
	Commit      string
	Author      string
	Date        time.Time
	Lang        string
	ResultCount int
}

func TabulateAggregationMatches(tabulator AggregationTabulator, mode types.SearchAggregationMode) (OnMatches, error) {

	modeCountTypes := map[types.SearchAggregationMode]countFunc{
		types.REPO_AGGREGATION_MODE:   countRepo,
		types.PATH_AGGREGATION_MODE:   countPath,
		types.AUTHOR_AGGREGATION_MODE: countAuthor,
	}

	modeCountFunc, ok := modeCountTypes[mode]
	if !ok {
		return nil, errors.Newf("unsupported search aggregation mode: %s", mode)
	}

	return func(matches []streamhttp.EventMatch) {
		combined := map[MatchKey]int{}
		for _, match := range matches {
			key, count, err := modeCountFunc(match)
			// delegate error handling to the passed in tabulator
			if err != nil {
				tabulator(nil, err)
				continue
			}
			if count == 0 {
				continue
			}
			current, _ := combined[key]
			combined[key] = current + count
		}
		for key, count := range combined {
			tabulator(&AggregationMatchResult{Key: key, Count: count}, nil)
		}
	}, nil
}

// NewEventEnvironment maps event matches into a consistent type
func newEventMatch(event streamhttp.EventMatch) *eventMatch {
	switch match := event.(type) {
	case *streamhttp.EventContentMatch:
		lang, _ := enry.GetLanguageByExtension(match.Path)
		resultCount := 0
		for _, lineMatch := range match.LineMatches {
			resultCount += len(lineMatch.OffsetAndLengths)
		}
		return &eventMatch{
			Repo:        match.Repository,
			RepoID:      match.RepositoryID,
			Path:        match.Path,
			Lang:        lang,
			ResultCount: resultCount,
		}
	case *streamhttp.EventPathMatch:
		lang, _ := enry.GetLanguageByExtension(match.Path)
		return &eventMatch{
			Repo:        match.Repository,
			RepoID:      match.RepositoryID,
			Path:        match.Path,
			Lang:        lang,
			ResultCount: 1,
		}
	case *streamhttp.EventRepoMatch:
		return &eventMatch{
			Repo:        match.Repository,
			RepoID:      match.RepositoryID,
			ResultCount: 1,
		}
	case *streamhttp.EventCommitMatch:
		return &eventMatch{
			Repo:        match.Repository,
			RepoID:      match.RepositoryID,
			Author:      match.AuthorName,
			Date:        match.AuthorDate,
			ResultCount: 1, //TODO(chwarwick): Verify that we want to count commits not matches in the commit
		}
	case *streamhttp.EventSymbolMatch:
		lang, _ := enry.GetLanguageByExtension(match.Path)
		return &eventMatch{
			Repo:        match.Repository,
			RepoID:      match.RepositoryID,
			Path:        match.Path,
			Lang:        lang,
			ResultCount: len(match.Symbols),
		}
	default:
		return &eventMatch{}
	}
}

type countFunc func(streamhttp.EventMatch) (MatchKey, int, error)
type MatchKey struct {
	Repo   string
	RepoID int32
	Group  string
}

func countRepo(r streamhttp.EventMatch) (MatchKey, int, error) {
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

func countLang(r streamhttp.EventMatch) (MatchKey, int, error) {
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

func countPath(r streamhttp.EventMatch) (MatchKey, int, error) {
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

func countAuthor(r streamhttp.EventMatch) (MatchKey, int, error) {
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
