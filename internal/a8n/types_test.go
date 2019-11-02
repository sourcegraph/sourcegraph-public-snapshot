package a8n

import (
	"testing"
	"time"

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

	state, err := changeset.State()
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

		cases = append(cases, testCase{"github",
			Changeset{
				ID: 23,
				Metadata: &github.PullRequest{
					TimelineItems: []github.TimelineItem{
						{"AssignedEvent", assignedEvent},
						{"PullRequestReviewThread", &github.PullRequestReviewThread{
							Comments: reviewComments[:2],
						}},
						{"UnassignedEvent", unassignedEvent},
						{"PullRequestReviewThread", &github.PullRequestReviewThread{
							Comments: reviewComments[2:],
						}},
						{"ClosedEvent", closedEvent},
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
			}},
		})
	}

	{ // Bitbucket Server

		user := bitbucketserver.User{Name: "john-doe"}
		reviewer := bitbucketserver.User{Name: "jane-doe"}

		activities := []bitbucketserver.Activity{{
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
	}

	for _, tc := range tests {
		have, err := tc.events.ReviewState()
		if err != nil {
			t.Fatalf("got error: %s", err)
		}

		if have, want := have, tc.want; have != want {
			t.Errorf("wrong reviewstate. have=%s, want=%s", have, want)
		}
	}
}
