package streaming

import (
	"fmt"

	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// TabulationDecoder will tabulate the result counts per repository.
func TabulationDecoder() (streamhttp.FrontendStreamDecoder, *TabulationResult) {
	var tr = &TabulationResult{
		RepoCounts: make(map[string]*SearchMatch),
	}

	addCount := func(repo string, repoId int32, count int) {
		if forRepo, ok := tr.RepoCounts[repo]; !ok {
			tr.RepoCounts[repo] = &SearchMatch{
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
				tr.SkippedReasons = append(tr.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
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
					tr.TotalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				case *streamhttp.EventPathMatch:
					tr.TotalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventRepoMatch:
					tr.TotalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventCommitMatch:
					tr.TotalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventSymbolMatch:
					count := len(match.Symbols)
					tr.TotalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				}
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			tr.Errors = append(tr.Errors, eventError.Message)
		},
	}, tr
}

type TabulationResult struct {
	StreamDecoderEvents
	RepoCounts map[string]*SearchMatch
	TotalCount int
}

type StreamDecoderEvents struct {
	SkippedReasons []string
	Errors         []string
}

type SearchMatch struct {
	RepositoryID   int32
	RepositoryName string
	MatchCount     int
}
