package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// monitor periodically polls Sourcegraph via the GraphQL API for the status of each
// given repo, as well as the status of each given upload. When there is a change of
// state for a repository, it is printed. The state changes that can occur are:
//
// - An upload fails to process (returns an error)
// - An upload completes processing
// - The last upload for a repository completes processing, but the
//   containing repo has a stale commit graph
// - A repository with no pending uploads has a fresh commit graph
func monitor(ctx context.Context, repoNames []string, uploads []uploadMeta) error {
	var oldState map[string]repoState
	waitMessageDisplayed := make(map[string]struct{}, len(repoNames))
	finishedMessageDisplayed := make(map[string]struct{}, len(repoNames))

	fmt.Printf("[%5s] %s Waiting for uploads to finish processing\n", internal.TimeSince(start), internal.EmojiLightbulb)

	for {
		state, err := queryRepoState(ctx, repoNames, uploads)
		if err != nil {
			return err
		}

		if verbose {
			parts := make([]string, 0, len(repoNames))
			for _, repoName := range repoNames {
				states := make([]string, 0, len(state[repoName].uploadStates))
				for _, uploadState := range state[repoName].uploadStates {
					states = append(states, fmt.Sprintf("%s=%-10s", uploadState.upload.commit[:7], uploadState.state))
				}
				sort.Strings(states)

				parts = append(parts, fmt.Sprintf("%s\tstale=%v\t%s", repoName, state[repoName].stale, strings.Join(states, "\t")))
			}

			fmt.Printf("[%5s] %s\n", internal.TimeSince(start), strings.Join(parts, "\n\t"))
		}

		numReposCompleted := 0

		for repoName, data := range state {
			oldData := oldState[repoName]

			numUploadsCompleted := 0
			for _, uploadState := range data.uploadStates {
				if uploadState.state == "ERRORED" {
					return errors.Newf("failed to process (%s)", uploadState.failure)
				}

				if uploadState.state == "COMPLETED" {
					numUploadsCompleted++

					var oldState string
					for _, oldUploadState := range oldData.uploadStates {
						if oldUploadState.upload.id == uploadState.upload.id {
							oldState = oldUploadState.state
						}
					}

					if oldState != "COMPLETED" {
						fmt.Printf("[%5s] %s Finished processing index for %s@%s\n", internal.TimeSince(start), internal.EmojiSuccess, repoName, uploadState.upload.commit[:7])
					}
				} else if uploadState.state != "QUEUED" && uploadState.state != "PROCESSING" {
					return errors.Newf("unexpected state '%s' for %s@%s", uploadState.state, uploadState.upload.repoName, uploadState.upload.commit[:7])
				}
			}

			if numUploadsCompleted == len(data.uploadStates) {
				if !data.stale {
					numReposCompleted++

					if _, ok := finishedMessageDisplayed[repoName]; !ok {
						finishedMessageDisplayed[repoName] = struct{}{}
						fmt.Printf("[%5s] %s Commit graph refreshed for %s\n", internal.TimeSince(start), internal.EmojiSuccess, repoName)
					}
				} else if _, ok := waitMessageDisplayed[repoName]; !ok {
					waitMessageDisplayed[repoName] = struct{}{}
					fmt.Printf("[%5s] %s Waiting for commit graph to refresh for %s\n", internal.TimeSince(start), internal.EmojiLightbulb, repoName)
				}
			}
		}

		if numReposCompleted == len(repoNames) {
			break
		}

		oldState = state

		select {
		case <-time.After(pollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	fmt.Printf("[%5s] %s All uploads processed\n", internal.TimeSince(start), internal.EmojiSuccess)
	return nil
}

type repoState struct {
	stale        bool
	uploadStates []uploadState
}

type uploadState struct {
	upload  uploadMeta
	state   string
	failure string
}

// queryRepoState makes a GraphQL request for the given repositories and uploads and
// returns a map from repository names to the state of that repository. Each repository
// state has a flag indicating whether or not its commit graph is stale, and an entry
// for each upload belonging to that repository including that upload's state.
func queryRepoState(_ context.Context, repoNames []string, uploads []uploadMeta) (map[string]repoState, error) {
	uploadIDs := make([]string, 0, len(uploads))
	for _, upload := range uploads {
		uploadIDs = append(uploadIDs, upload.id)
	}
	sort.Strings(uploadIDs)

	var payload struct{ Data map[string]jsonUploadResult }
	if err := internal.GraphQLClient().GraphQL(internal.SourcegraphAccessToken, makeRepoStateQuery(repoNames, uploadIDs), nil, &payload); err != nil {
		return nil, err
	}

	state := make(map[string]repoState, len(repoNames))
	for name, data := range payload.Data {
		if name[0] == 'r' {
			index, _ := strconv.Atoi(name[1:])
			repoName := repoNames[index]

			state[repoName] = repoState{
				stale:        data.CommitGraph.Stale,
				uploadStates: []uploadState{},
			}
		}
	}

	for name, data := range payload.Data {
		if name[0] == 'u' {
			index, _ := strconv.Atoi(name[1:])
			upload := uploads[index]

			state[upload.repoName] = repoState{
				stale: state[upload.repoName].stale,
				uploadStates: append(state[upload.repoName].uploadStates, uploadState{
					upload:  upload,
					state:   data.State,
					failure: data.Failure,
				}),
			}
		}
	}

	return state, nil
}

// makeRepoStateQuery constructs a GraphQL query for use by queryRepoState.
func makeRepoStateQuery(repoNames, uploadIDs []string) string {
	fragments := make([]string, 0, len(repoNames)+len(uploadIDs))
	for i, repoName := range repoNames {
		fragments = append(fragments, fmt.Sprintf(repositoryQueryFragment, i, internal.MakeTestRepoName(repoName)))
	}
	for i, id := range uploadIDs {
		fragments = append(fragments, fmt.Sprintf(uploadQueryFragment, i, id))
	}

	return fmt.Sprintf("query CodeIntelQA_Upload_RepositoryState {%s}", strings.Join(fragments, "\n"))
}

const repositoryQueryFragment = `
	r%d: repository(name: "%s") {
		codeIntelligenceCommitGraph {
			stale
		}
	}
`

const uploadQueryFragment = `
	u%d: node(id: "%s") {
		... on LSIFUpload {
			state
			failure
		}
	}
`

type jsonUploadResult struct {
	State       string                `json:"state"`
	Failure     string                `json:"failure"`
	CommitGraph jsonCommitGraphResult `json:"codeIntelligenceCommitGraph"`
}

type jsonCommitGraphResult struct {
	Stale bool `json:"stale"`
}
