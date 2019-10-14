package a8n

import (
	"testing"
	"time"

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
