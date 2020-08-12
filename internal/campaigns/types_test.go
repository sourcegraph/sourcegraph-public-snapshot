package campaigns

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
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
		ExternalServiceType: extsvc.TypeGitHub,
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

	state, err := changeset.externalState()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := ChangesetExternalStateMerged, state; want != have {
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

		reviewRequestedActorEvent := &github.ReviewRequestedEvent{
			RequestedReviewer: github.Actor{Login: "the-great-tortellini"},
			Actor:             actor,
			CreatedAt:         now,
		}
		reviewRequestedTeamEvent := &github.ReviewRequestedEvent{
			RequestedTeam: github.Team{Name: "the-belgian-waffles"},
			Actor:         actor,
			CreatedAt:     now,
		}

		cases = append(cases, testCase{"github-blank-review-requested",
			Changeset{
				ID: 23,
				Metadata: &github.PullRequest{
					TimelineItems: []github.TimelineItem{
						{Type: "ReviewRequestedEvent", Item: reviewRequestedActorEvent},
						{Type: "ReviewRequestedEvent", Item: reviewRequestedTeamEvent},
						{Type: "ReviewRequestedEvent", Item: &github.ReviewRequestedEvent{
							// Both Team and Reviewer are blank.
							Actor:     actor,
							CreatedAt: now,
						}},
					},
				},
			},
			[]*ChangesetEvent{{
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubReviewRequested,
				Key:         reviewRequestedActorEvent.Key(),
				Metadata:    reviewRequestedActorEvent,
			}, {
				ChangesetID: 23,
				Kind:        ChangesetEventKindGitHubReviewRequested,
				Key:         reviewRequestedTeamEvent.Key(),
				Metadata:    reviewRequestedTeamEvent,
			}},
		})
	}

	{ // bitbucketserver

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

	{ // GitLab
		notes := []*gitlab.Note{
			{ID: 11, System: false, Body: "this is a user note"},
			{ID: 12, System: true, Body: "approved this merge request"},
			{ID: 13, System: true, Body: "unapproved this merge request"},
		}

		pipelines := []*gitlab.Pipeline{
			{ID: 21},
			{ID: 22},
		}

		mr := &gitlab.MergeRequest{
			Notes:     notes,
			Pipelines: pipelines,
		}

		cases = append(cases, testCase{
			name: "gitlab",
			changeset: Changeset{
				ID:       1234,
				Metadata: mr,
			},
			events: []*ChangesetEvent{
				{
					ChangesetID: 1234,
					Kind:        ChangesetEventKindGitLabApproved,
					Key:         notes[1].Key(),
					Metadata:    notes[1].ToReview(),
				},
				{
					ChangesetID: 1234,
					Kind:        ChangesetEventKindGitLabUnapproved,
					Key:         notes[2].Key(),
					Metadata:    notes[2].ToReview(),
				},
				{
					ChangesetID: 1234,
					Kind:        ChangesetEventKindGitLabPipeline,
					Key:         pipelines[0].Key(),
					Metadata:    pipelines[0],
				},
				{
					ChangesetID: 1234,
					Kind:        ChangesetEventKindGitLabPipeline,
					Key:         pipelines[1].Key(),
					Metadata:    pipelines[1],
				},
			},
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

func TestChangesetDiffStat(t *testing.T) {
	var (
		added   int32 = 77
		changed int32 = 88
		deleted int32 = 99
	)

	for name, tc := range map[string]struct {
		c    Changeset
		want *diff.Stat
	}{
		"added missing": {
			c: Changeset{
				DiffStatAdded:   nil,
				DiffStatChanged: &changed,
				DiffStatDeleted: &deleted,
			},
			want: nil,
		},
		"changed missing": {
			c: Changeset{
				DiffStatAdded:   &added,
				DiffStatChanged: nil,
				DiffStatDeleted: &deleted,
			},
			want: nil,
		},
		"deleted missing": {
			c: Changeset{
				DiffStatAdded:   &added,
				DiffStatChanged: &changed,
				DiffStatDeleted: nil,
			},
			want: nil,
		},
		"all present": {
			c: Changeset{
				DiffStatAdded:   &added,
				DiffStatChanged: &changed,
				DiffStatDeleted: &deleted,
			},
			want: &diff.Stat{
				Added:   added,
				Changed: changed,
				Deleted: deleted,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			have := tc.c.DiffStat()
			if (tc.want == nil && have != nil) || (tc.want != nil && have == nil) {
				t.Errorf("mismatched nils in diff stats: have %+v; want %+v", have, tc.want)
			} else if tc.want != nil && have != nil {
				if d := cmp.Diff(*have, *tc.want); d != "" {
					t.Errorf("incorrect diff stat: %s", d)
				}
			}
		})
	}
}

type changesetSyncStateTestCase struct {
	state [2]ChangesetSyncState
	want  bool
}

func TestChangesetSyncStateEquals(t *testing.T) {
	testCases := make(map[string]changesetSyncStateTestCase)

	for baseName, basePairs := range map[string][2]string{
		"base equal":     {"abc", "abc"},
		"base different": {"abc", "def"},
	} {
		for headName, headPairs := range map[string][2]string{
			"head equal":     {"abc", "abc"},
			"head different": {"abc", "def"},
		} {
			for completeName, completePairs := range map[string][2]bool{
				"complete both true":  {true, true},
				"complete both false": {false, false},
				"complete different":  {true, false},
			} {
				key := fmt.Sprintf("%s; %s; %s", baseName, headName, completeName)

				testCases[key] = changesetSyncStateTestCase{
					state: [2]ChangesetSyncState{
						{
							BaseRefOid: basePairs[0],
							HeadRefOid: headPairs[0],
							IsComplete: completePairs[0],
						},
						{
							BaseRefOid: basePairs[1],
							HeadRefOid: headPairs[1],
							IsComplete: completePairs[1],
						},
					},
					// This is icky, but works, and means we're not just
					// repeating the implementation of Equals().
					want: strings.HasPrefix(key, "base equal; head equal; complete both"),
				}
			}
		}
	}

	for name, tc := range testCases {
		if have := tc.state[0].Equals(&tc.state[1]); have != tc.want {
			t.Errorf("%s: unexpected Equals result: have %v; want %v", name, have, tc.want)
		}
	}
}

func TestChangeset_SetMetadata(t *testing.T) {
	for name, tc := range map[string]struct {
		meta interface{}
		want *Changeset
	}{
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{
				ID:          12345,
				FromRef:     bitbucketserver.Ref{ID: "refs/heads/branch"},
				UpdatedDate: 10 * 1000,
			},
			want: &Changeset{
				ExternalID:          "12345",
				ExternalServiceType: extsvc.TypeBitbucketServer,
				ExternalBranch:      "branch",
				ExternalUpdatedAt:   time.Unix(10, 0),
			},
		},
		"GitHub": {
			meta: &github.PullRequest{
				Number:      12345,
				HeadRefName: "branch",
				UpdatedAt:   time.Unix(10, 0),
			},
			want: &Changeset{
				ExternalID:          "12345",
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalBranch:      "branch",
				ExternalUpdatedAt:   time.Unix(10, 0),
			},
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				IID:          12345,
				SourceBranch: "branch",
				UpdatedAt:    gitlab.Time{Time: time.Unix(10, 0)},
			},
			want: &Changeset{
				ExternalID:          "12345",
				ExternalServiceType: extsvc.TypeGitLab,
				ExternalBranch:      "branch",
				ExternalUpdatedAt:   time.Unix(10, 0),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			have := &Changeset{}
			want := tc.want
			want.Metadata = tc.meta

			if err := have.SetMetadata(tc.meta); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if d := cmp.Diff(have, want); d != "" {
				t.Errorf("metadata not updated as expected: %s", d)
			}
		})
	}
}

func TestChangeset_Title(t *testing.T) {
	want := "foo"
	for name, meta := range map[string]interface{}{
		"bitbucketserver": &bitbucketserver.PullRequest{
			Title: want,
		},
		"GitHub": &github.PullRequest{
			Title: want,
		},
		"GitLab": &gitlab.MergeRequest{
			Title: want,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta}
			have, err := c.Title()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != want {
				t.Errorf("unexpected title: have %s; want %s", have, want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.Title(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_ExternalCreatedAt(t *testing.T) {
	want := time.Unix(10, 0)
	for name, meta := range map[string]interface{}{
		"bitbucketserver": &bitbucketserver.PullRequest{
			CreatedDate: 10 * 1000,
		},
		"GitHub": &github.PullRequest{
			CreatedAt: want,
		},
		"GitLab": &gitlab.MergeRequest{
			CreatedAt: gitlab.Time{Time: want},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta}
			if have := c.ExternalCreatedAt(); have != want {
				t.Errorf("unexpected external creation date: have %+v; want %+v", have, want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		want := time.Time{}
		if have := c.ExternalCreatedAt(); have != want {
			t.Errorf("unexpected external creation date: have %+v; want %+v", have, want)
		}
	})
}

func TestChangeset_Body(t *testing.T) {
	want := "foo"
	for name, meta := range map[string]interface{}{
		"bitbucketserver": &bitbucketserver.PullRequest{
			Description: want,
		},
		"GitHub": &github.PullRequest{
			Body: want,
		},
		"GitLab": &gitlab.MergeRequest{
			Description: want,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta}
			have, err := c.Body()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != want {
				t.Errorf("unexpected body: have %s; want %s", have, want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.Body(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_externalState(t *testing.T) {
	for name, tc := range map[string]struct {
		meta interface{}
		want ChangesetExternalState
	}{
		"bitbucketserver: declined": {
			meta: &bitbucketserver.PullRequest{
				State: "DECLINED",
			},
			want: ChangesetExternalStateClosed,
		},
		"bitbucketserver: open": {
			meta: &bitbucketserver.PullRequest{
				State: "OPEN",
			},
			want: ChangesetExternalStateOpen,
		},
		"GitHub: open": {
			meta: &github.PullRequest{
				State: "OPEN",
			},
			want: ChangesetExternalStateOpen,
		},
		"GitLab: opened": {
			meta: &gitlab.MergeRequest{
				State: gitlab.MergeRequestStateOpened,
			},
			want: ChangesetExternalStateOpen,
		},
		"GitLab: closed": {
			meta: &gitlab.MergeRequest{
				State: gitlab.MergeRequestStateClosed,
			},
			want: ChangesetExternalStateClosed,
		},
		"GitLab: locked": {
			meta: &gitlab.MergeRequest{
				State: gitlab.MergeRequestStateLocked,
			},
			want: ChangesetExternalStateClosed,
		},
		"GitLab: merged": {
			meta: &gitlab.MergeRequest{
				State: gitlab.MergeRequestStateMerged,
			},
			want: ChangesetExternalStateMerged,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.externalState()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected state: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("deleted", func(t *testing.T) {
		c := &Changeset{ExternalDeletedAt: time.Unix(10, 0)}
		have, err := c.externalState()
		if err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
		if want := ChangesetExternalStateDeleted; have != want {
			t.Errorf("unexpected state: have %s; want %s", have, want)
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		c := &Changeset{Metadata: &github.PullRequest{
			State: "FOO",
		}}
		if _, err := c.externalState(); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.externalState(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_URL(t *testing.T) {
	want := "foo"
	for name, meta := range map[string]interface{}{
		"bitbucketserver": &bitbucketserver.PullRequest{
			Links: struct {
				Self []struct {
					Href string `json:"href"`
				} `json:"self"`
			}{
				Self: []struct {
					Href string `json:"href"`
				}{{Href: want}},
			},
		},
		"GitHub": &github.PullRequest{
			URL: want,
		},
		"GitLab": &gitlab.MergeRequest{
			WebURL: want,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta}
			have, err := c.URL()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != want {
				t.Errorf("unexpected URL: have %s; want %s", have, want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.URL(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_HeadRefOid(t *testing.T) {
	for name, tc := range map[string]struct {
		meta interface{}
		want string
	}{
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{},
			want: "",
		},
		"GitHub": {
			meta: &github.PullRequest{HeadRefOid: "foo"},
			want: "foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				DiffRefs: gitlab.DiffRefs{HeadSHA: "foo"},
			},
			want: "foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.HeadRefOid()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected head ref OID: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.HeadRefOid(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_HeadRef(t *testing.T) {
	for name, tc := range map[string]struct {
		meta interface{}
		want string
	}{
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{
				FromRef: bitbucketserver.Ref{ID: "foo"},
			},
			want: "foo",
		},
		"GitHub": {
			meta: &github.PullRequest{HeadRefName: "foo"},
			want: "refs/heads/foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				SourceBranch: "foo",
			},
			want: "refs/heads/foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.HeadRef()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected head ref: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.HeadRef(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_BaseRefOid(t *testing.T) {
	for name, tc := range map[string]struct {
		meta interface{}
		want string
	}{
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{},
			want: "",
		},
		"GitHub": {
			meta: &github.PullRequest{BaseRefOid: "foo"},
			want: "foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				DiffRefs: gitlab.DiffRefs{BaseSHA: "foo"},
			},
			want: "foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.BaseRefOid()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected base ref OID: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.BaseRefOid(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_BaseRef(t *testing.T) {
	for name, tc := range map[string]struct {
		meta interface{}
		want string
	}{
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{
				ToRef: bitbucketserver.Ref{ID: "foo"},
			},
			want: "foo",
		},
		"GitHub": {
			meta: &github.PullRequest{BaseRefName: "foo"},
			want: "refs/heads/foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				TargetBranch: "foo",
			},
			want: "refs/heads/foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.BaseRef()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected base ref: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.BaseRef(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_Labels(t *testing.T) {
	for name, tc := range map[string]struct {
		meta interface{}
		want []ChangesetLabel
	}{
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{},
			want: []ChangesetLabel{},
		},
		"GitHub": {
			meta: &github.PullRequest{
				Labels: struct{ Nodes []github.Label }{
					Nodes: []github.Label{
						{
							Name:        "red door",
							Color:       "black",
							Description: "paint it black",
						},
						{
							Name:        "grün",
							Color:       "green",
							Description: "groan",
						},
					},
				},
			},
			want: []ChangesetLabel{
				{
					Name:        "red door",
					Color:       "black",
					Description: "paint it black",
				},
				{
					Name:        "grün",
					Color:       "green",
					Description: "groan",
				},
			},
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				Labels: []string{"black", "green"},
			},
			want: []ChangesetLabel{
				{Name: "black", Color: "000000"},
				{Name: "green", Color: "000000"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			if d := cmp.Diff(c.Labels(), tc.want); d != "" {
				t.Errorf("unexpected labels: %s", d)
			}
		})
	}
}

func TestChangesetSpecUnmarshalValidate(t *testing.T) {
	tests := []struct {
		name    string
		rawSpec string
		err     string
	}{
		{
			name: "valid ExistingChangesetReference",
			rawSpec: `{
				"baseRepository": "graphql-id",
				"externalID": "1234"
			}`,
		},
		{
			name: "valid GitBranchChangesetDescription",
			rawSpec: `{
				"baseRepository": "graphql-id",
				"baseRef": "refs/heads/master",
				"baseRev": "d34db33f",
				"headRef": "refs/heads/my-branch",
				"headRepository": "graphql-id",
				"title": "my title",
				"body": "my body",
				"published": false,
				"commits": [{
				  "message": "commit message",
				  "diff": "the diff"
				}]
			}`,
		},
		{
			name: "missing fields in GitBranchChangesetDescription",
			rawSpec: `{
				"baseRepository": "graphql-id",
				"baseRef": "refs/heads/master",
				"headRef": "refs/heads/my-branch",
				"headRepository": "graphql-id",
				"title": "my title",
				"published": false,
				"commits": [{
				  "diff": "the diff"
				}]
			}`,
			err: "4 errors occurred:\n\t* Must validate one and only one schema (oneOf)\n\t* baseRev is required\n\t* body is required\n\t* commits.0: message is required\n\n",
		},
		{
			name: "missing fields in ExistingChangesetReference",
			rawSpec: `{
				"baseRepository": "graphql-id"
			}`,
			err: "2 errors occurred:\n\t* Must validate one and only one schema (oneOf)\n\t* externalID is required\n\n",
		},
		{
			name: "headRepository in GitBranchChangesetDescription does not match baseRepository",
			rawSpec: `{
				"baseRepository": "graphql-id",
				"baseRef": "refs/heads/master",
				"baseRev": "d34db33f",
				"headRef": "refs/heads/my-branch",
				"headRepository": "graphql-id999999",
				"title": "my title",
				"body": "my body",
				"published": false,
				"commits": [{
				  "message": "commit message",
				  "diff": "the diff"
				}]
			}`,
			err: ErrHeadBaseMismatch.Error(),
		},
		{
			name: "too many commits in GitBranchChangesetDescription",
			rawSpec: `{
				"baseRepository": "graphql-id",
				"baseRef": "refs/heads/master",
				"baseRev": "d34db33f",
				"headRef": "refs/heads/my-branch",
				"headRepository": "graphql-id",
				"title": "my title",
				"body": "my body",
				"published": false,
				"commits": [
				  {
				    "message": "commit message",
				    "diff": "the diff"
				  },
                  {
				    "message": "commit message2",
				    "diff": "the diff2"
				  }
				]
			}`,
			err: "2 errors occurred:\n\t* Must validate one and only one schema (oneOf)\n\t* commits: Array must have at most 1 items\n\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := &ChangesetSpec{RawSpec: tc.rawSpec}
			haveErr := fmt.Sprintf("%v", spec.UnmarshalValidate())
			if haveErr == "<nil>" {
				haveErr = ""
			}
			if diff := cmp.Diff(tc.err, haveErr); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCampaignSpecUnmarshalValidate(t *testing.T) {
	tests := []struct {
		name    string
		rawSpec string
		err     string
	}{
		{
			name: "valid",
			rawSpec: `{
				"name": "my-unique-name",
				"description": "My description",
				"on": [
				    {"repositoriesMatchingQuery": "lang:go func main"},
					{"repository": "github.com/sourcegraph/src-cli"}
				],
				"steps": [
				{
					"run": "echo 'foobar'",
					"container": "alpine",
					"env": {
						"PATH": "/work/foobar:$PATH"
					}
				}
				],
				"changesetTemplate": {
					"title": "Hello World",
					"body": "My first campaign!",
					"branch": "hello-world",
					"commit": {
						"message": "Append Hello World to all README.md files"
					},
					"published": false
				}
			}`,
		},
		{
			name: "valid YAML",
			rawSpec: `
name: my-unique-name
description: My description
on:
- repositoriesMatchingQuery: lang:go func main
- repository: github.com/sourcegraph/src-cli
steps:
- run: echo 'foobar'
  container: alpine
  env:
    PATH: "/work/foobar:$PATH"
changesetTemplate:
  title: Hello World
  body: My first campaign!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`,
		},
		{
			name: "invalid name",
			rawSpec: `{
				"name": "this contains spaces",
				"description": "My description",
				"on": [
				    {"repositoriesMatchingQuery": "lang:go func main"},
					{"repository": "github.com/sourcegraph/src-cli"}
				],
				"steps": [
				{
					"run": "echo 'foobar'",
					"container": "alpine",
					"env": {
						"PATH": "/work/foobar:$PATH"
					}
				}
				],
				"changesetTemplate": {
					"title": "Hello World",
					"body": "My first campaign!",
					"branch": "hello-world",
					"commit": {
						"message": "Append Hello World to all README.md files"
					},
					"published": false
				}
			}`,
			err: "1 error occurred:\n\t* name: Does not match pattern '^[\\w.-]+$'\n\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := &CampaignSpec{RawSpec: tc.rawSpec}
			haveErr := fmt.Sprintf("%v", spec.UnmarshalValidate())
			if haveErr == "<nil>" {
				haveErr = ""
			}
			if diff := cmp.Diff(tc.err, haveErr); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		})
	}
}
