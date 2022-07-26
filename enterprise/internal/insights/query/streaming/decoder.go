package streaming

import (
	"fmt"
	"strings"

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
	LineMatches    []streamhttp.EventLineMatch
	Path           string
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
				// ShardTimeout is a specific skipped event that we want to retry on. Currently
				// we only retry on Alert events so this is why we add it there. This behaviour will
				// be uniformised eventually.
				if skipped.Reason == streamapi.ShardTimeout {
					tr.Alerts = append(tr.Alerts, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				} else {
					tr.SkippedReasons = append(tr.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				}
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

// MetadataResult contains information about matches like line matches, paths.
type MetadataResult struct {
	StreamDecoderEvents
	Matches []*SearchMatch
}

// MetadataDecoder will tabulate metadata for a query. This is to return useful information for
// related insights.
func MetadataDecoder() (streamhttp.FrontendStreamDecoder, *MetadataResult) {
	mr := &MetadataResult{}

	return streamhttp.FrontendStreamDecoder{
		OnMatches: func(matches []streamhttp.EventMatch) {
			for _, match := range matches {
				switch match := match.(type) {
				// Right now we only care about inline matches.
				// Should be extended when we care about repo and file results.
				case *streamhttp.EventContentMatch:
					mr.Matches = append(mr.Matches, &SearchMatch{LineMatches: match.LineMatches})
				case *streamhttp.EventPathMatch:
					mr.Matches = append(mr.Matches, &SearchMatch{Path: match.Path})
				}
			}
		},
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
					mr.Alerts = append(mr.Alerts, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				} else {
					mr.SkippedReasons = append(mr.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				}
			}
		},
		OnAlert: func(ea *streamhttp.EventAlert) {
			if ea.Title == "No repositories found" {
				// If we hit a case where we don't find a repository we don't want to error, just
				// complete our search.
			} else {
				mr.Alerts = append(mr.Alerts, fmt.Sprintf("%s: %s", ea.Title, ea.Description))
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			mr.Errors = append(mr.Errors, eventError.Message)
		},
	}, mr
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

func MatchContextComputeDecoder() (client.ComputeMatchContextStreamDecoder, *ComputeTabulationResult) {
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
					ctr.Alerts = append(ctr.Alerts, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				} else {
					ctr.SkippedReasons = append(ctr.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				}
			}
		},
		OnResult: func(results []compute.MatchContext) {
			for _, result := range results {
				current := getRepoCounts(result)
				for _, match := range result.Matches {
					for _, data := range match.Environment {
						value := data.Value
						if value == "" {
							continue // a bug in upstream compute processing means we need to check for empty replacements (https://github.com/sourcegraph/sourcegraph/issues/37972)
						}
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

func ComputeTextDecoder() (client.ComputeTextExtraStreamDecoder, *ComputeTabulationResult) {
	ctr := &ComputeTabulationResult{
		RepoCounts: make(map[string]*ComputeMatch),
	}
	getRepoCounts := func(matchContext compute.TextExtra) *ComputeMatch {
		var v *ComputeMatch
		if got, ok := ctr.RepoCounts[matchContext.Repository]; ok {
			return got
		}
		v = newComputeMatch(matchContext.Repository, matchContext.RepositoryID)
		ctr.RepoCounts[matchContext.Repository] = v
		return v
	}

	return client.ComputeTextExtraStreamDecoder{
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
					ctr.Alerts = append(ctr.Alerts, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				} else {
					ctr.SkippedReasons = append(ctr.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
				}
			}
		},
		OnResult: func(results []compute.TextExtra) {
			for _, result := range results {
				vals := strings.Split(result.Value, "\n")
				for _, val := range vals {
					if val == "" {
						continue // a bug in upstream compute processing means we need to check for empty replacements (https://github.com/sourcegraph/sourcegraph/issues/37972)
					}
					current := getRepoCounts(result)
					value := val
					if len(value) > capturedValueMaxLength {
						value = value[:capturedValueMaxLength]
					}
					current.ValueCounts[value] += 1
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
