package streaming

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-enry/go-enry/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

const LuckySearchAlert = "lucky search triggered additional search queries"

type AggregationMatchResult struct {
	Key   MatchKey
	Count int
}

type AggregationTabulator func(*AggregationMatchResult, error)
type OnMatches func(matches []streamhttp.EventMatch)

// AggregationDecoder will tabulate the result using the passed in tabulator
func AggregationDecoder(onMatches OnMatches) (streamhttp.FrontendStreamDecoder, *StreamDecoderEvents) {
	decoderEvents := &StreamDecoderEvents{}

	return streamhttp.FrontendStreamDecoder{
		OnProgress: func(progress *streamapi.Progress) {
			if !progress.Done {
				return
			}
			// Skipped elements are built progressively for a Progress update until it is Done, so
			// we want to register its contents only once it is done.
			for _, skipped := range progress.Skipped {
				// ShardTimeout is a specific skipped event that we want to retry on. Currently
				// we only retry on Alert events so this is why we add it there. This behaviour will
				// be uniformised eventually.
				if skipped.Reason == streamapi.ShardTimeout {
					decoderEvents.Alerts = append(decoderEvents.Alerts, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				} else {
					decoderEvents.SkippedReasons = append(decoderEvents.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				}
			}
		},
		OnMatches: onMatches,
		OnAlert: func(ea *streamhttp.EventAlert) {
			if ea.Title == "No repositories found" {
				// If we hit a case where we don't find a repository we don't want to error, just
				// complete our search.
			} else {
				decoderEvents.Alerts = append(decoderEvents.Alerts, fmt.Sprintf("%s: %s", ea.Title, ea.Description))
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

	var modeCountFunc countFunc

	switch mode {
	case types.REPO_AGGREGATION_MODE:
		modeCountFunc = countRepo
	case types.PATH_AGGREGATION_MODE:
		modeCountFunc = countPath
	case types.AUTHOR_AGGREGATION_MODE:
		modeCountFunc = countAuthor
	default:
		return nil, errors.New("unsupported search aggregation mode")
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
		return &eventMatch{
			Repo:        match.Repository,
			RepoID:      match.RepositoryID,
			Path:        match.Path,
			Lang:        lang,
			ResultCount: len(match.ChunkMatches),
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
			ResultCount: 1,
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
