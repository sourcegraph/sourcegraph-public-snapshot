package campaigns

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestComputeGithubCheckState(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	commitEvent := func(minutesSinceSync int, context, state string) *cmpgn.ChangesetEvent {
		commit := &github.CommitStatus{
			Context:    context,
			State:      state,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		event := &cmpgn.ChangesetEvent{
			Kind:     cmpgn.ChangesetEventKindCommitStatus,
			Metadata: commit,
		}
		return event
	}
	checkRun := func(id, status, conclusion string) github.CheckRun {
		return github.CheckRun{
			ID:         id,
			Status:     status,
			Conclusion: conclusion,
		}
	}
	checkSuiteEvent := func(minutesSinceSync int, id, status, conclusion string, runs ...github.CheckRun) *cmpgn.ChangesetEvent {
		suite := &github.CheckSuite{
			ID:         id,
			Status:     status,
			Conclusion: conclusion,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		suite.CheckRuns.Nodes = runs
		event := &cmpgn.ChangesetEvent{
			Kind:     cmpgn.ChangesetEventKindCheckSuite,
			Metadata: suite,
		}
		return event
	}

	lastSynced := now.Add(-1 * time.Minute)
	pr := &github.PullRequest{}

	tests := []struct {
		name   string
		events []*cmpgn.ChangesetEvent
		want   cmpgn.ChangesetCheckState
	}{
		{
			name:   "empty slice",
			events: nil,
			want:   cmpgn.ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
		{
			name: "success status and suite",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "COMPLETED", "SUCCESS", checkRun("cr1", "COMPLETED", "SUCCESS")),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
			},
			want: cmpgn.ChangesetCheckStatePending,
		},
		{
			name: "single error",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
			},
			want: cmpgn.ChangesetCheckStateFailed,
		},
		{
			name: "pending + error",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: cmpgn.ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: cmpgn.ChangesetCheckStatePending,
		},
		{
			name: "success + error",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: cmpgn.ChangesetCheckStateFailed,
		},
		{
			name: "success x2",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
		{
			name: "later events have precedence",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
		{
			name: "suites with zero runs should be ignored",
			events: []*cmpgn.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "QUEUED", ""),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeGitHubCheckState(lastSynced, pr, tc.events)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestComputeBitbucketBuildStatus(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	sha := "abcdef"
	statusEvent := func(minutesSinceSync int, key, state string) *cmpgn.ChangesetEvent {
		commit := &bitbucketserver.CommitStatus{
			Commit: sha,
			Status: bitbucketserver.BuildStatus{
				State:     state,
				Key:       key,
				DateAdded: now.Add(1*time.Second).Unix() * 1000,
			},
		}
		event := &cmpgn.ChangesetEvent{
			Kind:     cmpgn.ChangesetEventKindBitbucketServerCommitStatus,
			Metadata: commit,
		}
		return event
	}

	lastSynced := now.Add(-1 * time.Minute)
	pr := &bitbucketserver.PullRequest{
		Commits: []*bitbucketserver.Commit{
			{
				ID: sha,
			},
		},
	}

	tests := []struct {
		name   string
		events []*cmpgn.ChangesetEvent
		want   cmpgn.ChangesetCheckState
	}{
		{
			name:   "empty slice",
			events: nil,
			want:   cmpgn.ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "SUCCESSFUL"),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
			},
			want: cmpgn.ChangesetCheckStatePending,
		},
		{
			name: "single error",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "FAILED"),
			},
			want: cmpgn.ChangesetCheckStateFailed,
		},
		{
			name: "pending + error",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
				statusEvent(1, "ctx2", "FAILED"),
			},
			want: cmpgn.ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
				statusEvent(1, "ctx2", "SUCCESSFUL"),
			},
			want: cmpgn.ChangesetCheckStatePending,
		},
		{
			name: "success + error",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "SUCCESSFUL"),
				statusEvent(1, "ctx2", "FAILED"),
			},
			want: cmpgn.ChangesetCheckStateFailed,
		},
		{
			name: "success x2",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "SUCCESSFUL"),
				statusEvent(1, "ctx2", "SUCCESSFUL"),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
		{
			name: "later events have precedence",
			events: []*cmpgn.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
				statusEvent(1, "ctx1", "SUCCESSFUL"),
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			have := computeBitbucketBuildStatus(lastSynced, pr, tc.events)
			if diff := cmp.Diff(tc.want, have); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestComputeReviewState(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }
	ghReview := func(t time.Time, login, state string) *cmpgn.ChangesetEvent {
		return &cmpgn.ChangesetEvent{
			Kind: cmpgn.ChangesetEventKindGitHubReviewed,
			Metadata: &github.PullRequestReview{
				UpdatedAt: t,
				State:     state,
				Author: github.Actor{
					Login: login,
				},
			},
		}
	}

	ghReviewDismissed := func(t time.Time, login, reviewer string) *cmpgn.ChangesetEvent {
		return &cmpgn.ChangesetEvent{
			Kind: cmpgn.ChangesetEventKindGitHubReviewDismissed,
			Metadata: &github.ReviewDismissedEvent{
				CreatedAt: t,
				Actor:     github.Actor{Login: login},
				Review: github.PullRequestReview{
					Author: github.Actor{
						Login: reviewer,
					},
				},
			},
		}
	}

	bbsActivity := func(t time.Time, login string, kind cmpgn.ChangesetEventKind) *cmpgn.ChangesetEvent {
		return &cmpgn.ChangesetEvent{
			Kind: kind,
			Metadata: &bitbucketserver.Activity{
				CreatedDate: timeToUnixMilli(t),
				User: bitbucketserver.User{
					Name: login,
				},
			},
		}
	}

	tests := []struct {
		events ChangesetEvents
		want   cmpgn.ChangesetReviewState
	}{
		{
			events: ChangesetEvents{
				ghReview(daysAgo(0), "user1", "APPROVED"),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "APPROVED"),
				ghReview(daysAgo(0), "user1", "COMMENTED"),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user1", "COMMENTED"),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "APPROVED"),
				ghReview(daysAgo(0), "user1", "PENDING"),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user1", "PENDING"),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "APPROVED"),
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(1), "user1", "APPROVED"),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(0), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user2", "APPROVED"),
				ghReview(daysAgo(0), "user3", "APPROVED"),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(3), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(2), "user2", "APPROVED"),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(3), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(2), "user2", "APPROVED"),
				ghReview(daysAgo(0), "user1", "APPROVED"),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				// GitHub updates the state of the reviews when they're dismissed
				ghReview(daysAgo(1), "user1", "DISMISSED"),
				ghReviewDismissed(daysAgo(0), "user2", "user1"),
			},
			want: cmpgn.ChangesetReviewStatePending,
		},
		{
			events: ChangesetEvents{
				// GitHub updates the state of the reviews when they're dismissed
				ghReview(daysAgo(2), "user1", "DISMISSED"),
				ghReviewDismissed(daysAgo(1), "user2", "user1"),
				ghReview(daysAgo(0), "user1", "CHANGES_REQUESTED"),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				// GitHub updates the state of the reviews when they're dismissed
				ghReview(daysAgo(2), "user1", "DISMISSED"),
				ghReviewDismissed(daysAgo(1), "user2", "user1"),
				ghReview(daysAgo(0), "user3", "APPROVED"),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user1", "DISMISSED"),
			},
			want: cmpgn.ChangesetReviewStatePending,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(1), "user1", "DISMISSED"),
				ghReview(daysAgo(0), "user3", "APPROVED"),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", cmpgn.ChangesetEventKindBitbucketServerApproved),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", cmpgn.ChangesetEventKindBitbucketServerReviewed),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", cmpgn.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user2", cmpgn.ChangesetEventKindBitbucketServerReviewed),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", cmpgn.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user2", cmpgn.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(daysAgo(0), "user3", cmpgn.ChangesetEventKindBitbucketServerApproved),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", cmpgn.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user2", cmpgn.ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(daysAgo(0), "user2", cmpgn.ChangesetEventKindBitbucketServerApproved),
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", cmpgn.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user1", cmpgn.ChangesetEventKindBitbucketServerUnapproved),
			},
			want: cmpgn.ChangesetReviewStatePending,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", cmpgn.ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user1", cmpgn.ChangesetEventKindBitbucketServerUnapproved),
				bbsActivity(daysAgo(0), "user1", cmpgn.ChangesetEventKindBitbucketServerReviewed),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
	}

	for i, tc := range tests {
		sort.Sort(tc.events)
		changeset := &campaigns.Changeset{Metadata: &github.PullRequest{CreatedAt: daysAgo(10)}}

		history, err := computeHistory(changeset, tc.events)
		if err != nil {
			t.Fatalf("computing history failed: %s", err)
		}

		have, err := ComputeReviewState(changeset, history)
		if err != nil {
			t.Fatalf("got error: %s", err)
		}

		if have, want := have, tc.want; have != want {
			t.Errorf("%d: wrong reviewstate. have=%s, want=%s", i, have, want)
		}
	}
}

func TestComputeChangesetState(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	tests := []struct {
		sortedEvents ChangesetEvents
		want         cmpgn.ChangesetState
	}{
		{
			sortedEvents: ChangesetEvents{},
			want:         cmpgn.ChangesetStateOpen,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(1), campaigns.ChangesetEventKindGitHubClosed, 1),
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(2), cmpgn.ChangesetEventKindBitbucketServerDeclined, 1),
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(2), cmpgn.ChangesetEventKindGitHubClosed, 1),
				event(t, daysAgo(1), cmpgn.ChangesetEventKindGitHubReopened, 1),
			},
			want: cmpgn.ChangesetStateOpen,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(2), cmpgn.ChangesetEventKindBitbucketServerDeclined, 1),
				event(t, daysAgo(1), cmpgn.ChangesetEventKindBitbucketServerReopened, 1),
			},
			want: cmpgn.ChangesetStateOpen,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(3), cmpgn.ChangesetEventKindGitHubClosed, 1),
				event(t, daysAgo(2), cmpgn.ChangesetEventKindGitHubReopened, 1),
				event(t, daysAgo(1), cmpgn.ChangesetEventKindGitHubClosed, 1),
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(3), cmpgn.ChangesetEventKindBitbucketServerDeclined, 1),
				event(t, daysAgo(2), cmpgn.ChangesetEventKindBitbucketServerReopened, 1),
				event(t, daysAgo(1), cmpgn.ChangesetEventKindBitbucketServerDeclined, 1),
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(2), cmpgn.ChangesetEventKindGitHubMerged, 1),
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(2), cmpgn.ChangesetEventKindBitbucketServerMerged, 1),
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(3), cmpgn.ChangesetEventKindGitHubMerged, 1),
				// Merged is a final state. Events after should be ignored.
				event(t, daysAgo(1), cmpgn.ChangesetEventKindGitHubClosed, 1),
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				event(t, daysAgo(3), cmpgn.ChangesetEventKindBitbucketServerMerged, 1),
				// Merged is a final state. Events after should be ignored.
				event(t, daysAgo(1), cmpgn.ChangesetEventKindBitbucketServerDeclined, 1),
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				// GitHub emits Closed and Merged events at the same time.
				// We want to report Merged.
				event(t, daysAgo(3), cmpgn.ChangesetEventKindGitHubClosed, 1),
				event(t, daysAgo(1), cmpgn.ChangesetEventKindGitHubMerged, 1),
			},
			want: cmpgn.ChangesetStateMerged,
		},
	}

	for i, tc := range tests {
		changeset := &campaigns.Changeset{Metadata: &github.PullRequest{CreatedAt: daysAgo(10)}}

		history, err := computeHistory(changeset, tc.sortedEvents)
		if err != nil {
			t.Fatalf("computing history failed: %s", err)
		}

		have, err := ComputeChangesetState(changeset, history)
		if err != nil {
			t.Fatalf("got error: %s", err)
		}

		if have, want := have, tc.want; have != want {
			t.Errorf("%d: wrong state. have=%s, want=%s", i, have, want)
		}
	}
}
