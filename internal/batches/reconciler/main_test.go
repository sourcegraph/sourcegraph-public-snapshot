package reconciler

import (
	"time"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func buildGithubPR(now time.Time, externalState btypes.ChangesetExternalState) *github.PullRequest {
	state := string(externalState)

	pr := &github.PullRequest{
		ID:          "12345",
		Number:      12345,
		Title:       state + " GitHub PR",
		Body:        state + " GitHub PR",
		State:       state,
		HeadRefName: gitdomain.AbbreviateRef("head-ref-on-github"),
		TimelineItems: []github.TimelineItem{
			{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
				Commit: github.Commit{
					OID:           "new-f00bar",
					PushedDate:    now,
					CommittedDate: now,
				},
			}},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if externalState == btypes.ChangesetExternalStateDraft {
		pr.State = "OPEN"
		pr.IsDraft = true
	}

	if externalState == btypes.ChangesetExternalStateClosed {
		// We add a "ClosedEvent" so that the SyncChangesets call that happens after closing
		// the PR has the "correct" state to set the ExternalState
		pr.TimelineItems = append(pr.TimelineItems, github.TimelineItem{
			Type: "ClosedEvent",
			Item: &github.ClosedEvent{CreatedAt: now.Add(1 * time.Hour)},
		})
		pr.UpdatedAt = now.Add(1 * time.Hour)
	}

	return pr
}
