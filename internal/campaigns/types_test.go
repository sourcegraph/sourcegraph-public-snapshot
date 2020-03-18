package campaigns

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

func TestChangesetMetadata(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)

	githubActor := github.Actor{
		AvatarURL: "https://avatars2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}

	githubPR := &github.PullRequest{
		ID:           "FOOBARID",
		Title:        "Fix a bunch of bugs",
		Body:         "This fixes a bunch of bugs",
		URL:          "https://github.com/sourcegraph/sourcegraph/pull/12345",
		Number:       12345,
		State:        "MERGED",
		Author:       githubActor,
		Participants: []github.Actor{githubActor},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	changeset := &Changeset{
		RepoID:              42,
		CreatedAt:           now,
		UpdatedAt:           now,
		Metadata:            githubPR,
		CampaignIDs:         []int64{},
		ExternalID:          "12345",
		ExternalServiceType: "github",
	}

	title, err := changeset.Title()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := githubPR.Title, title; want != have {
		t.Errorf("changeset title wrong. want=%q, have=%q", want, have)
	}

	body, err := changeset.Body()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := githubPR.Body, body; want != have {
		t.Errorf("changeset body wrong. want=%q, have=%q", want, have)
	}

	state, err := changeset.state()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := ChangesetStateMerged, state; want != have {
		t.Errorf("changeset state wrong. want=%q, have=%q", want, have)
	}

	url, err := changeset.URL()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := githubPR.URL, url; want != have {
		t.Errorf("changeset url wrong. want=%q, have=%q", want, have)
	}
}

func TestChangesetEvents(t *testing.T) {
	type testCase struct {
		name      string
		changeset Changeset
		events    []*ChangesetEvent
	}

	var cases []testCase

	{ // Github

		now := time.Now().UTC()

		reviewComments := []*github.PullRequestReviewComment{
			{DatabaseID: 1, Body: "foo"},
			{DatabaseID: 2, Body: "bar"},
			{DatabaseID: 3, Body: "baz"},
		}

		actor := github.Actor{Login: "john-doe"}

		assignedEvent := &github.AssignedEvent{
			Actor:     actor,
			Assignee:  actor,
			CreatedAt: now,
		}

		unassignedEvent := &github.UnassignedEvent{
			Actor:     actor,
			Assignee:  actor,
			CreatedAt: now,
		}

		closedEvent := &github.ClosedEvent{
			Actor:     actor,
			CreatedAt: now,
		}

		commit := &github.PullRequestCommit{
			Commit: github.Commit{
				OID: "123",
			},
		}

		cases = append(cases, testCase{"github",
			Changeset{
				ID: 23,
				Metadata: &github.PullRequest{
					TimelineItems: []github.TimelineItem{
						{Type: "AssignedEvent", Item: assignedEvent},
						{Type: "PullRequestReviewThread", Item: &github.PullRequestReviewThread{
							Comments: reviewComments[:2],
						}},
						{Type: "UnassignedEvent", Item: unassignedEvent},
						{Type: "PullRequestReviewThread", Item: &github.PullRequestReviewThread{
							Comments: reviewComments[2:],
						}},
						{Type: "ClosedEvent", Item: closedEvent},
						{Type: "PullRequestCommit", Item: commit},
					},
				},
			},
			[]*ChangesetEvent{{
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubAssigned,
				Key:         assignedEvent.Key(),
				Metadata:    assignedEvent,
			}, {
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubReviewCommented,
				Key:         reviewComments[0].Key(),
				Metadata:    reviewComments[0],
			}, {
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubReviewCommented,
				Key:         reviewComments[1].Key(),
				Metadata:    reviewComments[1],
			}, {
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubUnassigned,
				Key:         unassignedEvent.Key(),
				Metadata:    unassignedEvent,
			}, {
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubReviewCommented,
				Key:         reviewComments[2].Key(),
				Metadata:    reviewComments[2],
			}, {
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubClosed,
				Key:         closedEvent.Key(),
				Metadata:    closedEvent,
			}, {
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubCommit,
				Key:         commit.Key(),
				Metadata:    commit,
			}},
		})
	}

	{ // Bitbucket Server

		user := bitbucketserver.User{Name: "john-doe"}
		reviewer := bitbucketserver.User{Name: "jane-doe"}

		activities := []*bitbucketserver.Activity{{
			ID:     1,
			User:   user,
			Action: bitbucketserver.OpenedActivityAction,
		}, {
			ID:     2,
			User:   reviewer,
			Action: bitbucketserver.ReviewedActivityAction,
		}, {
			ID:     3,
			User:   reviewer,
			Action: bitbucketserver.DeclinedActivityAction,
		}, {
			ID:     4,
			User:   user,
			Action: bitbucketserver.ReopenedActivityAction,
		}, {
			ID:     5,
			User:   user,
			Action: bitbucketserver.MergedActivityAction,
		}}

		cases = append(cases, testCase{"bitbucketserver",
			Changeset{
				ID: 24,
				Metadata: &bitbucketserver.PullRequest{
					Activities: activities,
				},
			},
			[]*ChangesetEvent{{
				ChangesetID: 24,
				Kind:        ChangesetEventKindBitbucketServerOpened,
				Key:         activities[0].Key(),
				Metadata:    activities[0],
			}, {
				ChangesetID: 24,
				Kind:        ChangesetEventKindBitbucketServerReviewed,
				Key:         activities[1].Key(),
				Metadata:    activities[1],
			}, {
				ChangesetID: 24,
				Kind:        ChangesetEventKindBitbucketServerDeclined,
				Key:         activities[2].Key(),
				Metadata:    activities[2],
			}, {
				ChangesetID: 24,
				Kind:        ChangesetEventKindBitbucketServerReopened,
				Key:         activities[3].Key(),
				Metadata:    activities[3],
			}, {
				ChangesetID: 24,
				Kind:        ChangesetEventKindBitbucketServerMerged,
				Key:         activities[4].Key(),
				Metadata:    activities[4],
			}},
		})
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			have := tc.changeset.Events()
			want := tc.events

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestChangesetEventsReviewState(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }
	ghReview := func(t time.Time, login, state string) *ChangesetEvent {
		return &ChangesetEvent{
			Kind: ChangesetEventKindGitHubReviewed,
			Metadata: &github.PullRequestReview{
				UpdatedAt: t,
				State:     state,
				Author: github.Actor{
					Login: login,
				},
			},
		}
	}

	ghReviewDismissed := func(t time.Time, login, reviewer string) *ChangesetEvent {
		return &ChangesetEvent{
			Kind: ChangesetEventKindGitHubReviewDismissed,
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

	bbsActivity := func(t time.Time, login string, kind ChangesetEventKind) *ChangesetEvent {
		return &ChangesetEvent{
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
		want   ChangesetReviewState
	}{
		{
			events: ChangesetEvents{
				ghReview(daysAgo(0), "user1", "APPROVED"),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "APPROVED"),
				ghReview(daysAgo(0), "user1", "COMMENTED"),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user1", "COMMENTED"),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "APPROVED"),
				ghReview(daysAgo(0), "user1", "PENDING"),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user1", "PENDING"),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "APPROVED"),
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(1), "user1", "APPROVED"),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(0), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user2", "APPROVED"),
				ghReview(daysAgo(0), "user3", "APPROVED"),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(3), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(2), "user2", "APPROVED"),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(3), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(2), "user2", "APPROVED"),
				ghReview(daysAgo(0), "user1", "APPROVED"),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReviewDismissed(daysAgo(0), "user2", "user1"),
			},
			want: ChangesetReviewStatePending,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReviewDismissed(daysAgo(1), "user2", "user1"),
				ghReview(daysAgo(0), "user1", "CHANGES_REQUESTED"),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReviewDismissed(daysAgo(1), "user2", "user1"),
				ghReview(daysAgo(0), "user3", "APPROVED"),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(0), "user1", "DISMISSED"),
			},
			want: ChangesetReviewStatePending,
		},
		{
			events: ChangesetEvents{
				ghReview(daysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(daysAgo(1), "user1", "DISMISSED"),
				ghReview(daysAgo(0), "user3", "APPROVED"),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", ChangesetEventKindBitbucketServerApproved),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", ChangesetEventKindBitbucketServerReviewed),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user2", ChangesetEventKindBitbucketServerReviewed),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user2", ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(daysAgo(0), "user3", ChangesetEventKindBitbucketServerApproved),
			},
			want: ChangesetReviewStateChangesRequested,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user2", ChangesetEventKindBitbucketServerReviewed),
				bbsActivity(daysAgo(0), "user2", ChangesetEventKindBitbucketServerApproved),
			},
			want: ChangesetReviewStateApproved,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user1", ChangesetEventKindBitbucketServerUnapproved),
			},
			want: ChangesetReviewStatePending,
		},
		{
			events: ChangesetEvents{
				bbsActivity(daysAgo(2), "user1", ChangesetEventKindBitbucketServerApproved),
				bbsActivity(daysAgo(1), "user1", ChangesetEventKindBitbucketServerUnapproved),
				bbsActivity(daysAgo(0), "user1", ChangesetEventKindBitbucketServerReviewed),
			},
			want: ChangesetReviewStateChangesRequested,
		},
	}

	for i, tc := range tests {
		have, err := tc.events.reviewState()
		if err != nil {
			t.Fatalf("got error: %s", err)
		}

		if have, want := have, tc.want; have != want {
			t.Errorf("%d: wrong reviewstate. have=%s, want=%s", i, have, want)
		}
	}
}

func TestComputeGithubCheckState(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	commitEvent := func(minutesSinceSync int, context, state string) *ChangesetEvent {
		commit := &github.CommitStatus{
			Context:    context,
			State:      state,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		event := &ChangesetEvent{
			Kind:     ChangesetEventKindCommitStatus,
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
	checkSuiteEvent := func(minutesSinceSync int, id, status, conclusion string, runs ...github.CheckRun) *ChangesetEvent {
		suite := &github.CheckSuite{
			ID:         id,
			Status:     status,
			Conclusion: conclusion,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		suite.CheckRuns.Nodes = runs
		event := &ChangesetEvent{
			Kind:     ChangesetEventKindCheckSuite,
			Metadata: suite,
		}
		return event
	}

	lastSynced := now.Add(-1 * time.Minute)
	pr := &github.PullRequest{}

	tests := []struct {
		name   string
		events []*ChangesetEvent
		want   ChangesetCheckState
	}{
		{
			name:   "empty slice",
			events: nil,
			want:   ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: ChangesetCheckStatePassed,
		},
		{
			name: "success status and suite",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "COMPLETED", "SUCCESS", checkRun("cr1", "COMPLETED", "SUCCESS")),
			},
			want: ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
			},
			want: ChangesetCheckStatePending,
		},
		{
			name: "single error",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
			},
			want: ChangesetCheckStateFailed,
		},
		{
			name: "pending + error",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: ChangesetCheckStatePending,
		},
		{
			name: "success + error",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: ChangesetCheckStateFailed,
		},
		{
			name: "success x2",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: ChangesetCheckStatePassed,
		},
		{
			name: "later events have precedence",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: ChangesetCheckStatePassed,
		},
		{
			name: "suites with zero runs should be ignored",
			events: []*ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "QUEUED", ""),
			},
			want: ChangesetCheckStatePassed,
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

func TestChangesetEventsLabels(t *testing.T) {
	now := time.Now()
	labelEvent := func(name string, kind ChangesetEventKind, when time.Time) *ChangesetEvent {
		removed := kind == ChangesetEventKindGitHubUnlabeled
		return &ChangesetEvent{
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
	changeset := func(names []string, updated time.Time) *Changeset {
		meta := &github.PullRequest{}
		for _, name := range names {
			meta.Labels.Nodes = append(meta.Labels.Nodes, github.Label{
				Name: name,
			})
		}
		return &Changeset{
			UpdatedAt: updated,
			Metadata:  meta,
		}
	}
	labels := func(names ...string) []ChangesetLabel {
		var ls []ChangesetLabel
		for _, name := range names {
			ls = append(ls, ChangesetLabel{Name: name})
		}
		return ls
	}

	tests := []struct {
		name      string
		changeset *Changeset
		events    ChangesetEvents
		want      []ChangesetLabel
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
				labelEvent("label1", ChangesetEventKindGitHubUnlabeled, now),
			},
			want: []ChangesetLabel{},
		},
		{
			name:      "add event",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label2", ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1", "label2"),
		},
		{
			name:      "old add event",
			changeset: changeset([]string{"label1"}, now.Add(5*time.Minute)),
			events: ChangesetEvents{
				labelEvent("label2", ChangesetEventKindGitHubLabeled, now),
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

func timeToUnixMilli(t time.Time) int {
	return int(t.UnixNano()) / int(time.Millisecond)
}
