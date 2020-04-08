package campaigns

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestChangesetEventsReviewState(t *testing.T) {
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
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReviewDismissed(daysAgo(0), "user2", "user1"),
			},
			want: cmpgn.ChangesetReviewStatePending,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReviewDismissed(daysAgo(1), "user2", "user1"),
				ghReview(daysAgo(0), "user1", "CHANGES_REQUESTED"),
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
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
		have, err := tc.events.reviewState()
		if err != nil {
			t.Fatalf("got error: %s", err)
		}

		if have, want := have, tc.want; have != want {
			t.Errorf("%d: wrong reviewstate. have=%s, want=%s", i, have, want)
		}
	}
}

func TestChangesetEventsState(t *testing.T) {
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
				{Kind: cmpgn.ChangesetEventKindGitHubClosed},
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindBitbucketServerDeclined},
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindGitHubClosed},
				{Kind: cmpgn.ChangesetEventKindGitHubReopened},
			},
			want: cmpgn.ChangesetStateOpen,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindBitbucketServerDeclined},
				{Kind: cmpgn.ChangesetEventKindBitbucketServerReopened},
			},
			want: cmpgn.ChangesetStateOpen,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindGitHubClosed},
				{Kind: cmpgn.ChangesetEventKindGitHubReopened},
				{Kind: cmpgn.ChangesetEventKindGitHubClosed},
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindBitbucketServerDeclined},
				{Kind: cmpgn.ChangesetEventKindBitbucketServerReopened},
				{Kind: cmpgn.ChangesetEventKindBitbucketServerDeclined},
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindGitHubMerged},
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindBitbucketServerMerged},
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindGitHubMerged},
				// Merged is a final state. Events after should be ignored.
				{Kind: cmpgn.ChangesetEventKindGitHubClosed},
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				{Kind: cmpgn.ChangesetEventKindBitbucketServerMerged},
				// Merged is a final state. Events after should be ignored.
				{Kind: cmpgn.ChangesetEventKindBitbucketServerDeclined},
			},
			want: cmpgn.ChangesetStateMerged,
		},
		{
			sortedEvents: ChangesetEvents{
				// GitHub emits Closed and Merged events at the same time.
				// We want to report Merged.
				{Kind: cmpgn.ChangesetEventKindGitHubClosed},
				{Kind: cmpgn.ChangesetEventKindGitHubMerged},
			},
			want: cmpgn.ChangesetStateMerged,
		},
	}

	for i, tc := range tests {
		if have, want := tc.sortedEvents.State(), tc.want; have != want {
			t.Errorf("%d: wrong state. have=%s, want=%s", i, have, want)
		}
	}
}

func TestChangesetEventsLabels(t *testing.T) {
	now := time.Now()
	labelEvent := func(name string, kind cmpgn.ChangesetEventKind, when time.Time) *cmpgn.ChangesetEvent {
		removed := kind == cmpgn.ChangesetEventKindGitHubUnlabeled
		return &cmpgn.ChangesetEvent{
			Kind:      kind,
			UpdatedAt: when,
			Metadata: &github.LabelEvent{
				Actor: github.Actor{},
				Label: github.Label{
					Name: name,
				},
				CreatedAt: when,
				Removed:   removed,
			},
		}
	}
	changeset := func(names []string, updated time.Time) *cmpgn.Changeset {
		meta := &github.PullRequest{}
		for _, name := range names {
			meta.Labels.Nodes = append(meta.Labels.Nodes, github.Label{
				Name: name,
			})
		}
		return &cmpgn.Changeset{
			UpdatedAt: updated,
			Metadata:  meta,
		}
	}
	labels := func(names ...string) []cmpgn.ChangesetLabel {
		var ls []cmpgn.ChangesetLabel
		for _, name := range names {
			ls = append(ls, cmpgn.ChangesetLabel{Name: name})
		}
		return ls
	}

	tests := []struct {
		name      string
		changeset *cmpgn.Changeset
		events    ChangesetEvents
		want      []cmpgn.ChangesetLabel
	}{
		{
			name: "zero values",
		},
		{
			name:      "no events",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events:    ChangesetEvents{},
			want:      labels("label1"),
		},
		{
			name:      "remove event",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label1", cmpgn.ChangesetEventKindGitHubUnlabeled, now),
			},
			want: []cmpgn.ChangesetLabel{},
		},
		{
			name:      "add event",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label2", cmpgn.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1", "label2"),
		},
		{
			name:      "old add event",
			changeset: changeset([]string{"label1"}, now.Add(5*time.Minute)),
			events: ChangesetEvents{
				labelEvent("label2", cmpgn.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			have := tc.events.UpdateLabelsSince(tc.changeset)
			want := tc.want
			sort.Slice(have, func(i, j int) bool { return have[i].Name < have[j].Name })
			sort.Slice(want, func(i, j int) bool { return want[i].Name < want[j].Name })
			if diff := cmp.Diff(have, want, cmpopts.EquateEmpty()); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
