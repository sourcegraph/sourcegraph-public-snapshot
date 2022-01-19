package streaming

import (
	"github.com/inconshreveable/log15"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// jsonDecoder streams results as JSON to w.
func TabulationDecoder() (streamhttp.FrontendStreamDecoder, *int, map[string]*SearchMatch, []string) {
	var totalCount int
	repoCounts := make(map[string]*SearchMatch)
	addCount := func(repo string, repoId int32, count int) {
		if forRepo, ok := repoCounts[repo]; !ok {
			repoCounts[repo] = &SearchMatch{
				RepositoryID:   repoId,
				RepositoryName: repo,
				MatchCount:     count,
			}
			return
		} else {
			forRepo.MatchCount += count
		}
	}
	var errors []string

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
					log15.Debug("EventContentMatch", "count", count)
					totalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				case *streamhttp.EventPathMatch:
					log15.Debug("EventPathMatch", "count", 1)
					totalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventRepoMatch:
					log15.Debug("EventRepoMatch", "count", 1)
					totalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventCommitMatch:
					log15.Debug("EventCommitMatch", "count", 1)
					totalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventSymbolMatch:
					count := len(match.Symbols)
					log15.Debug("EventSymbolMatch", "count", count)
					totalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				}
			}
		},
		OnAlert: func(alert *streamhttp.EventAlert) {
			log15.Debug("stream alert", "title", alert.Title)
		},

		OnError: func(eventError *streamhttp.EventError) {
			errors = append(errors, eventError.Message)
		},
	}, &totalCount, repoCounts, errors
}

type SearchMatch struct {
	RepositoryID   int32
	RepositoryName string
	MatchCount     int
}
