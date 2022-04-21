package streaming

import (
	"fmt"

	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// TabulationDecoder will tabulate the result counts per repository.
func TabulationDecoder() (streamhttp.FrontendStreamDecoder, *StreamingResult) {
	var sr = &StreamingResult{
		RepoCounts: make(map[string]*SearchMatch),
	}

	addCount := func(repo string, repoId int32, count int) {
		if forRepo, ok := sr.RepoCounts[repo]; !ok {
			sr.RepoCounts[repo] = &SearchMatch{
				RepositoryID:   repoId,
				RepositoryName: repo,
				MatchCount:     count,
			}
			return
		} else {
			forRepo.MatchCount += count
		}
	}

	return streamhttp.FrontendStreamDecoder{
		OnProgress: func(progress *streamapi.Progress) {
			if !progress.Done {
				return
			}
			// Skipped elements are built progressively for a Progress update until it is Done, so
			// we want to register its contents only once it is done.
			for _, skipped := range progress.Skipped {
				sr.SkippedReasons = append(sr.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
			}
		},
		OnMatches: func(matches []streamhttp.EventMatch) {
			for _, match := range matches {
				switch match := match.(type) {
				case *streamhttp.EventContentMatch:
					count := 0
					for _, lineMatch := range match.LineMatches {
						count += len(lineMatch.OffsetAndLengths)
					}
					sr.TotalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				case *streamhttp.EventPathMatch:
					sr.TotalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventRepoMatch:
					sr.TotalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventCommitMatch:
					sr.TotalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventSymbolMatch:
					count := len(match.Symbols)
					sr.TotalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				}
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			sr.Errors = append(sr.Errors, eventError.Message)
		},
	}, sr
}

type StreamingResult struct {
	RepoCounts     map[string]*SearchMatch
	TotalCount     int
	SkippedReasons []string
	Errors         []string
}

type SearchMatch struct {
	RepositoryID   int32
	RepositoryName string
	MatchCount     int
}
