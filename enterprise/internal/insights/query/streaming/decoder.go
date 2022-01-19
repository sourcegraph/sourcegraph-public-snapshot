package streaming

import (
	"github.com/inconshreveable/log15"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// jsonDecoder streams results as JSON to w.
func TabulationDecoder() (streamhttp.FrontendStreamDecoder, *int, map[string]int) {
	var totalCount int
	repoCounts := make(map[string]int)

	addCount := func(repo string, count int) {
		repoCounts[repo] += count
	}

	return streamhttp.FrontendStreamDecoder{
		OnProgress: func(progress *streamapi.Progress) {
			if !progress.Done {
				return
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
					log15.Info("EventContentMatch", "count", count)
					totalCount += count
					addCount(match.Repository, count)
				case *streamhttp.EventPathMatch:
					log15.Info("EventPathMatch", "count", 1)
					totalCount += 1
					addCount(match.Repository, 1)
				case *streamhttp.EventRepoMatch:
					log15.Info("EventRepoMatch", "count", 1)
					totalCount += 1
					addCount(match.Repository, 1)
				case *streamhttp.EventCommitMatch:
					log15.Info("EventCommitMatch", "count", 1)
					totalCount += 1
					addCount(match.Repository, 1)
				case *streamhttp.EventSymbolMatch:
					count := len(match.Symbols)
					log15.Info("EventSymbolMatch", "count", count)
					totalCount += count
					addCount(match.Repository, count)
				}
			}
		},

		OnError: func(eventError *streamhttp.EventError) {

		},
	}, &totalCount, repoCounts
}

type SearchMatch struct {
	repositoryID   int32
	repositoryName string
	matchCount     int32
}
