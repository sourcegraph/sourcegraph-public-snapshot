package streaming

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute/client"

	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

type StreamDecoderEvents struct {
	SkippedReasons []string
	Errors         []string
	Alerts         []string
}

type SearchMatch struct {
	RepositoryID   int32
	RepositoryName string
	MatchCount     int
}

type TabulationResult struct {
	StreamDecoderEvents
	RepoCounts map[string]*SearchMatch
	TotalCount int
}

// TabulationDecoder will tabulate the result counts per repository.
func TabulationDecoder() (streamhttp.FrontendStreamDecoder, *TabulationResult) {
	tr := &TabulationResult{
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
		OnAlert: func(ea *streamhttp.EventAlert) {
			if ea.Title == "No repositories found" {
				// If we hit a case where we don't find a repository we don't want to error, just
				// complete our search.
			} else {
				tr.Alerts = append(tr.Alerts, fmt.Sprintf("%s: %s", ea.Title, ea.Description))
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			tr.Errors = append(tr.Errors, eventError.Message)
		},
	}, tr
}

// ComputeMatch is our internal representation of a match retrieved from a Compute Streaming Search.
// It is internally different from the `ComputeMatch` returned by the Compute GraphQL query but they
// serve the same end goal.
type ComputeMatch struct {
	RepositoryID   int32
	RepositoryName string
	ValueCounts    map[string]int
}

func newComputeMatch(repoName string, repoID int32) *ComputeMatch {
	return &ComputeMatch{
		ValueCounts:    make(map[string]int),
		RepositoryID:   repoID,
		RepositoryName: repoName,
	}
}

type ComputeTabulationResult struct {
	StreamDecoderEvents
	RepoCounts map[string]*ComputeMatch
}

const capturedValueMaxLength = 100

func ComputeDecoder() (client.ComputeMatchContextStreamDecoder, *ComputeTabulationResult) {
	ctr := &ComputeTabulationResult{
		RepoCounts: make(map[string]*ComputeMatch),
	}
	getRepoCounts := func(matchContext compute.MatchContext) *ComputeMatch {
		var v *ComputeMatch
		if got, ok := ctr.RepoCounts[matchContext.Repository]; ok {
			return got
		}
		v = newComputeMatch(matchContext.Repository, matchContext.RepositoryID)
		ctr.RepoCounts[matchContext.Repository] = v
		return v
	}

	return client.ComputeMatchContextStreamDecoder{
		OnResult: func(results []compute.MatchContext) {
			for _, result := range results {
				current := getRepoCounts(result)
				for _, match := range result.Matches {
					for _, data := range match.Environment {
						value := data.Value
						if len(value) > capturedValueMaxLength {
							value = value[:capturedValueMaxLength]
						}
						current.ValueCounts[value] += 1
					}
				}
			}
		},
		OnAlert: func(ea *streamhttp.EventAlert) {
			if ea.Title == "No repositories found" {
				// If we hit a case where we don't find a repository we don't want to error, just
				// complete our search.
			} else {
				ctr.Alerts = append(ctr.Alerts, fmt.Sprintf("%s: %s", ea.Title, ea.Description))
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			ctr.Errors = append(ctr.Errors, eventError.Message)
		},
	}, ctr
}
