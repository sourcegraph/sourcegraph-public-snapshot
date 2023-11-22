package streaming

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/compute/client"

	"github.com/sourcegraph/sourcegraph/internal/api"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
)

type StreamDecoderEvents struct {
	SkippedReasons []string
	Errors         []string
	Alerts         []string
	DidTimeout     bool
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

type RepoResult struct {
	StreamDecoderEvents
	Repos []itypes.MinimalRepo
}

// onProgress is the common FrontendStreamDecoder.OnProgress handler.
func (s *StreamDecoderEvents) onProgress(progress *streamapi.Progress) {
	if !progress.Done {
		return
	}
	// Skipped elements are built progressively for a Progress update until it is Done, so
	// we want to register its contents only once it is done.
	for _, skipped := range progress.Skipped {
		switch skipped.Reason {
		case streamapi.ShardTimeout:
			// ShardTimeout is a specific skipped event that we want to retry on. Currently
			// we only retry on Alert events so this is why we add it there. This behaviour will
			// be uniformised eventually.
			s.Alerts = append(s.Alerts, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
			s.DidTimeout = true

		case streamapi.BackendMissing:
			// BackendMissing means we may be missing results due to
			// Zoekt rolling out. We add an alert to cause a retry.
			s.Alerts = append(s.Alerts, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))

		default:
			s.SkippedReasons = append(s.SkippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
		}
	}
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
		OnProgress: tr.onProgress,
		OnMatches: func(matches []streamhttp.EventMatch) {
			for _, match := range matches {
				switch match := match.(type) {
				case *streamhttp.EventContentMatch:
					count := 0
					for _, chunkMatch := range match.ChunkMatches {
						count += len(chunkMatch.Ranges)
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
	TotalCount int
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
		OnProgress: ctr.onProgress,
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
						ctr.TotalCount += 1
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
		OnProgress: ctr.onProgress,
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

func RepoDecoder() (streamhttp.FrontendStreamDecoder, *RepoResult) {
	repoResult := &RepoResult{
		Repos: []itypes.MinimalRepo{},
	}

	return streamhttp.FrontendStreamDecoder{
		OnProgress: repoResult.onProgress,
		OnMatches: func(matches []streamhttp.EventMatch) {
			for _, match := range matches {
				switch match := match.(type) {
				case *streamhttp.EventRepoMatch:
					repoResult.Repos = append(repoResult.Repos, itypes.MinimalRepo{ID: api.RepoID(match.RepositoryID), Name: api.RepoName(match.Repository)})
				}
			}
		},
		OnAlert: func(ea *streamhttp.EventAlert) {
			if ea.Title == "No repositories found" {
				// If we hit a case where we don't find a repository we don't want to error, just
				// complete our search.
			} else {
				repoResult.Alerts = append(repoResult.Alerts, fmt.Sprintf("%s: %s", ea.Title, ea.Description))
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			repoResult.Errors = append(repoResult.Errors, eventError.Message)
		},
	}, repoResult
}
