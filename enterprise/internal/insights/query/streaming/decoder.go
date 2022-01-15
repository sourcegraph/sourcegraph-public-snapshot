package streaming

import "github.com/inconshreveable/log15"

// jsonDecoder streams results as JSON to w.
func tabulationDecoder() decoder {

	return decoder{
		OnProgress: func(progress *Progress) {
			if !progress.Done {
				return
			}
		},
		OnMatches: func(matches []EventMatch) {
			for _, match := range matches {
				switch match := match.(type) {
				case *EventContentMatch:
					count := 0
					for _, lineMatch := range match.LineMatches {
						count += len(lineMatch.OffsetAndLengths)
					}
					// searchMatches = append(searchMatches, SearchMatch{
					// 	repositoryID:   match.RepositoryID,
					// 	repositoryName: match.Repository,
					// 	matchCount:     count,
					// })
					log15.Info("EventContentMatch", "count", count)
				case *EventPathMatch:
					log15.Info("EventPathMatch", "count", 1)

				case *EventRepoMatch:

				case *EventCommitMatch:

				case *EventSymbolMatch:

				}
			}
		},

		OnError: func(eventError *EventError) {

		},
	}
}

type SearchMatch struct {
	repositoryID   int32
	repositoryName string
	matchCount     int32
}
