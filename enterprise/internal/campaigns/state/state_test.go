package state

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestComputeGithubCheckState(t *testing.T) {
	t.Parallel()

	now := timeutil.Now()
	commitEvent := func(minutesSinceSync int, context, state string) *campaigns.ChangesetEvent {
		commit := &github.CommitStatus{
			Context:    context,
			State:      state,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		event := &campaigns.ChangesetEvent{
			Kind:     campaigns.ChangesetEventKindCommitStatus,
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
	checkSuiteEvent := func(minutesSinceSync int, id, status, conclusion string, runs ...github.CheckRun) *campaigns.ChangesetEvent {
		suite := &github.CheckSuite{
			ID:         id,
			Status:     status,
			Conclusion: conclusion,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		suite.CheckRuns.Nodes = runs
		event := &campaigns.ChangesetEvent{
			Kind:     campaigns.ChangesetEventKindCheckSuite,
			Metadata: suite,
		}
		return event
	}

	lastSynced := now.Add(-1 * time.Minute)
	pr := &github.PullRequest{}

	tests := []struct {
		name   string
		events []*campaigns.ChangesetEvent
		want   campaigns.ChangesetCheckState
	}{
		{
			name:   "empty slice",
			events: nil,
			want:   campaigns.ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: campaigns.ChangesetCheckStatePassed,
		},
		{
			name: "success status and suite",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "COMPLETED", "SUCCESS", checkRun("cr1", "COMPLETED", "SUCCESS")),
			},
			want: campaigns.ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
			},
			want: campaigns.ChangesetCheckStatePending,
		},
		{
			name: "single error",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
			},
			want: campaigns.ChangesetCheckStateFailed,
		},
		{
			name: "pending + error",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: campaigns.ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: campaigns.ChangesetCheckStatePending,
		},
		{
			name: "success + error",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: campaigns.ChangesetCheckStateFailed,
		},
		{
			name: "success x2",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: campaigns.ChangesetCheckStatePassed,
		},
		{
			name: "later events have precedence",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: campaigns.ChangesetCheckStatePassed,
		},
		{
			name: "queued suites with zero runs should be ignored",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "QUEUED", ""),
			},
			want: campaigns.ChangesetCheckStatePassed,
		},
		{
			name: "completed suites with zero runs should be ignored",
			events: []*campaigns.ChangesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
				checkSuiteEvent(1, "cs1", "COMPLETED", ""),
			},
			want: campaigns.ChangesetCheckStateFailed,
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
	t.Parallel()

	now := timeutil.Now()
	sha := "abcdef"
	statusEvent := func(minutesSinceSync int, key, state string) *campaigns.ChangesetEvent {
		commit := &bitbucketserver.CommitStatus{
			Commit: sha,
			Status: bitbucketserver.BuildStatus{
				State:     state,
				Key:       key,
				DateAdded: now.Add(1*time.Second).Unix() * 1000,
			},
		}
		event := &campaigns.ChangesetEvent{
			Kind:     campaigns.ChangesetEventKindBitbucketServerCommitStatus,
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
		events []*campaigns.ChangesetEvent
		want   campaigns.ChangesetCheckState
	}{
		{
			name:   "empty slice",
			events: nil,
			want:   campaigns.ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "SUCCESSFUL"),
			},
			want: campaigns.ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
			},
			want: campaigns.ChangesetCheckStatePending,
		},
		{
			name: "single error",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "FAILED"),
			},
			want: campaigns.ChangesetCheckStateFailed,
		},
		{
			name: "pending + error",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
				statusEvent(1, "ctx2", "FAILED"),
			},
			want: campaigns.ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
				statusEvent(1, "ctx2", "SUCCESSFUL"),
			},
			want: campaigns.ChangesetCheckStatePending,
		},
		{
			name: "success + error",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "SUCCESSFUL"),
				statusEvent(1, "ctx2", "FAILED"),
			},
			want: campaigns.ChangesetCheckStateFailed,
		},
		{
			name: "success x2",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "SUCCESSFUL"),
				statusEvent(1, "ctx2", "SUCCESSFUL"),
			},
			want: campaigns.ChangesetCheckStatePassed,
		},
		{
			name: "later events have precedence",
			events: []*campaigns.ChangesetEvent{
				statusEvent(1, "ctx1", "INPROGRESS"),
				statusEvent(1, "ctx1", "SUCCESSFUL"),
			},
			want: campaigns.ChangesetCheckStatePassed,
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

func TestComputeGitLabCheckState(t *testing.T) {
	t.Parallel()

	t.Run("no events", func(t *testing.T) {
		for name, tc := range map[string]struct {
			mr   *gitlab.MergeRequest
			want campaigns.ChangesetCheckState
		}{
			"no pipelines at all": {
				mr:   &gitlab.MergeRequest{},
				want: campaigns.ChangesetCheckStateUnknown,
			},
			"only a head pipeline": {
				mr: &gitlab.MergeRequest{
					HeadPipeline: &gitlab.Pipeline{
						Status: gitlab.PipelineStatusPending,
					},
				},
				want: campaigns.ChangesetCheckStatePending,
			},
			"one pipeline only": {
				mr: &gitlab.MergeRequest{
					HeadPipeline: &gitlab.Pipeline{
						Status: gitlab.PipelineStatusPending,
					},
					Pipelines: []*gitlab.Pipeline{
						{
							CreatedAt: gitlab.Time{Time: time.Unix(10, 0)},
							Status:    gitlab.PipelineStatusFailed,
						},
					},
				},
				want: campaigns.ChangesetCheckStateFailed,
			},
			"two pipelines in the expected order": {
				mr: &gitlab.MergeRequest{
					HeadPipeline: &gitlab.Pipeline{
						Status: gitlab.PipelineStatusPending,
					},
					Pipelines: []*gitlab.Pipeline{
						{
							CreatedAt: gitlab.Time{Time: time.Unix(10, 0)},
							Status:    gitlab.PipelineStatusFailed,
						},
						{
							CreatedAt: gitlab.Time{Time: time.Unix(5, 0)},
							Status:    gitlab.PipelineStatusSuccess,
						},
					},
				},
				want: campaigns.ChangesetCheckStateFailed,
			},
			"two pipelines in an unexpected order": {
				mr: &gitlab.MergeRequest{
					HeadPipeline: &gitlab.Pipeline{
						Status: gitlab.PipelineStatusPending,
					},
					Pipelines: []*gitlab.Pipeline{
						{
							CreatedAt: gitlab.Time{Time: time.Unix(5, 0)},
							Status:    gitlab.PipelineStatusFailed,
						},
						{
							CreatedAt: gitlab.Time{Time: time.Unix(10, 0)},
							Status:    gitlab.PipelineStatusSuccess,
						},
					},
				},
				want: campaigns.ChangesetCheckStatePassed,
			},
		} {
			t.Run(name, func(t *testing.T) {
				have := computeGitLabCheckState(time.Unix(0, 0), tc.mr, nil)
				if have != tc.want {
					t.Errorf("unexpected check state: have %s; want %s", have, tc.want)
				}
			})
		}
	})

	t.Run("with events", func(t *testing.T) {
		mr := &gitlab.MergeRequest{
			HeadPipeline: &gitlab.Pipeline{
				Status: gitlab.PipelineStatusPending,
			},
		}

		events := []*campaigns.ChangesetEvent{
			{
				Kind: campaigns.ChangesetEventKindGitLabPipeline,
				Metadata: &gitlab.Pipeline{
					CreatedAt: gitlab.Time{Time: time.Unix(5, 0)},
					Status:    gitlab.PipelineStatusSuccess,
				},
			},
			{
				Kind: campaigns.ChangesetEventKindGitLabPipeline,
				Metadata: &gitlab.Pipeline{
					CreatedAt: gitlab.Time{Time: time.Unix(4, 0)},
					Status:    gitlab.PipelineStatusFailed,
				},
			},
		}

		for name, tc := range map[string]struct {
			events     []*campaigns.ChangesetEvent
			lastSynced time.Time
			want       campaigns.ChangesetCheckState
		}{
			"older events only": {
				events:     events,
				lastSynced: time.Unix(10, 0),
				want:       campaigns.ChangesetCheckStatePending,
			},
			"newer events only": {
				events:     events,
				lastSynced: time.Unix(3, 0),
				want:       campaigns.ChangesetCheckStatePassed,
			},
		} {
			t.Run(name, func(t *testing.T) {
				have := computeGitLabCheckState(tc.lastSynced, mr, tc.events)
				if have != tc.want {
					t.Errorf("unexpected check state: have %s; want %s", have, tc.want)
				}
			})
		}
	})
}

func TestComputeReviewState(t *testing.T) {
	t.Parallel()

	now := timeutil.Now()
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	tests := []struct {
		name      string
		changeset *campaigns.Changeset
		history   []changesetStatesAtTime
		want      campaigns.ChangesetReviewState
	}{
		{
			name:      "github - no events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetReviewStatePending,
		},
		{
			name:      "github - changeset older than events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: campaigns.ChangesetReviewStateApproved,
		},
		{
			name:      "github - changeset newer than events",
			changeset: githubChangeset(daysAgo(0), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: campaigns.ChangesetReviewStateApproved,
		},
		{
			name:      "bitbucketserver - no events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetReviewStateChangesRequested,
		},

		{
			name:      "bitbucketserver - changeset older than events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: campaigns.ChangesetReviewStateApproved,
		},

		{
			name:      "bitbucketserver - changeset newer than events",
			changeset: bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: campaigns.ChangesetReviewStateChangesRequested,
		},
		{
			name:      "gitlab - no events, no approvals",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, []*gitlab.Note{}),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetReviewStatePending,
		},
		{
			name: "gitlab - no events, one approval",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, []*gitlab.Note{
				{
					System: true,
					Body:   "approved this merge request",
				},
			}),
			history: []changesetStatesAtTime{},
			want:    campaigns.ChangesetReviewStateApproved,
		},
		{
			name: "gitlab - no events, one unapproval",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, []*gitlab.Note{
				{
					System: true,
					Body:   "unapproved this merge request",
				},
			}),
			history: []changesetStatesAtTime{},
			want:    campaigns.ChangesetReviewStateChangesRequested,
		},
		{
			name: "gitlab - no events, several notes",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, []*gitlab.Note{
				{Body: "this is a user note"},
				{
					System: true,
					Body:   "unapproved this merge request",
				},
				{Body: "this is a user note"},
				{
					System: true,
					Body:   "approved this merge request",
				},
			}),
			history: []changesetStatesAtTime{},
			want:    campaigns.ChangesetReviewStateChangesRequested,
		},
		{
			name: "gitlab - changeset older than events",
			changeset: gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateOpened, []*gitlab.Note{
				{
					System: true,
					Body:   "unapproved this merge request",
				},
			}),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: campaigns.ChangesetReviewStateApproved,
		},
		{
			name: "gitlab - changeset newer than events",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, []*gitlab.Note{
				{
					System: true,
					Body:   "unapproved this merge request",
				},
			}),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: campaigns.ChangesetReviewStateChangesRequested,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changeset := tc.changeset

			have, err := computeReviewState(changeset, tc.history)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}

			if have, want := have, tc.want; have != want {
				t.Errorf("%d: wrong reviewstate. have=%s, want=%s", i, have, want)
			}
		})
	}
}

func TestComputeExternalState(t *testing.T) {
	t.Parallel()

	now := timeutil.Now()
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	tests := []struct {
		name      string
		changeset *campaigns.Changeset
		history   []changesetStatesAtTime
		want      campaigns.ChangesetExternalState
	}{
		{
			name:      "github - no events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateOpen,
		},
		{
			name:      "github - changeset older than events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateClosed,
		},
		{
			name:      "github - changeset newer than events",
			changeset: githubChangeset(daysAgo(0), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateOpen,
		},
		{
			name:      "github - changeset newer and deleted",
			changeset: setDeletedAt(githubChangeset(daysAgo(0), "OPEN"), daysAgo(0)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateDeleted,
		},
		{
			name:      "github draft - no events",
			changeset: setDraft(githubChangeset(daysAgo(10), "OPEN")),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateDraft,
		},
		{
			name:      "github draft - changeset older than events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: campaigns.ChangesetExternalStateDraft},
			},
			want: campaigns.ChangesetExternalStateDraft,
		},
		{
			name:      "github draft - changeset newer than events",
			changeset: setDraft(githubChangeset(daysAgo(0), "OPEN")),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateDraft,
		},
		{
			name:      "github draft closed",
			changeset: setDraft(githubChangeset(daysAgo(1), "CLOSED")),
			history:   []changesetStatesAtTime{{t: daysAgo(2), externalState: campaigns.ChangesetExternalStateClosed}},
			want:      campaigns.ChangesetExternalStateClosed,
		},
		{
			name:      "bitbucketserver - no events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateOpen,
		},
		{
			name:      "bitbucketserver - changeset older than events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateClosed,
		},
		{
			name:      "bitbucketserver - changeset newer than events",
			changeset: bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateOpen,
		},
		{
			name:      "bitbucketserver - changeset newer and deleted",
			changeset: setDeletedAt(bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"), daysAgo(0)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateDeleted,
		},
		{
			name:      "gitlab - no events, opened",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, nil),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateOpen,
		},
		{
			name:      "gitlab - no events, closed",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateClosed, nil),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateClosed,
		},
		{
			name:      "gitlab - no events, locked",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateLocked, nil),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateClosed,
		},
		{
			name:      "gitlab - no events, merged",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateMerged, nil),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateMerged,
		},
		{
			name:      "gitlab - changeset older than events",
			changeset: gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateMerged, nil),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateClosed,
		},
		{
			name:      "gitlab - changeset newer than events",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateMerged, nil),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateMerged,
		},
		{
			name:      "gitlab draft - no events",
			changeset: setDraft(gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateOpened, nil)),
			history:   []changesetStatesAtTime{},
			want:      campaigns.ChangesetExternalStateDraft,
		},
		{
			name:      "gitlab draft - changeset older than events",
			changeset: gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateOpened, nil),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: campaigns.ChangesetExternalStateDraft},
			},
			want: campaigns.ChangesetExternalStateDraft,
		},
		{
			name:      "gitlab draft - changeset newer than events",
			changeset: setDraft(gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, nil)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: campaigns.ChangesetExternalStateClosed},
			},
			want: campaigns.ChangesetExternalStateDraft,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changeset := tc.changeset

			have, err := computeExternalState(changeset, tc.history)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}

			if have, want := have, tc.want; have != want {
				t.Errorf("%d: wrong external state. have=%s, want=%s", i, have, want)
			}
		})
	}
}

func TestComputeLabels(t *testing.T) {
	t.Parallel()

	now := time.Now()
	labelEvent := func(name string, kind campaigns.ChangesetEventKind, when time.Time) *campaigns.ChangesetEvent {
		removed := kind == campaigns.ChangesetEventKindGitHubUnlabeled
		return &campaigns.ChangesetEvent{
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
	changeset := func(names []string, updated time.Time) *campaigns.Changeset {
		meta := &github.PullRequest{}
		for _, name := range names {
			meta.Labels.Nodes = append(meta.Labels.Nodes, github.Label{
				Name: name,
			})
		}
		return &campaigns.Changeset{
			UpdatedAt: updated,
			Metadata:  meta,
		}
	}
	labels := func(names ...string) []campaigns.ChangesetLabel {
		var ls []campaigns.ChangesetLabel
		for _, name := range names {
			ls = append(ls, campaigns.ChangesetLabel{Name: name})
		}
		return ls
	}

	tests := []struct {
		name      string
		changeset *campaigns.Changeset
		events    ChangesetEvents
		want      []campaigns.ChangesetLabel
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
				labelEvent("label1", campaigns.ChangesetEventKindGitHubUnlabeled, now),
			},
			want: []campaigns.ChangesetLabel{},
		},
		{
			name:      "add event",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label2", campaigns.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1", "label2"),
		},
		{
			name:      "old add event",
			changeset: changeset([]string{"label1"}, now.Add(5*time.Minute)),
			events: ChangesetEvents{
				labelEvent("label2", campaigns.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1"),
		},
		{
			name:      "sorting",
			changeset: changeset([]string{"label4", "label3"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label2", campaigns.ChangesetEventKindGitHubLabeled, now),
				labelEvent("label1", campaigns.ChangesetEventKindGitHubLabeled, now.Add(5*time.Minute)),
			},
			want: labels("label1", "label2", "label3", "label4"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			have := ComputeLabels(tc.changeset, tc.events)
			want := tc.want
			if diff := cmp.Diff(have, want, cmpopts.EquateEmpty()); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func bitbucketChangeset(updatedAt time.Time, state, reviewStatus string) *campaigns.Changeset {
	return &campaigns.Changeset{
		ExternalServiceType: extsvc.TypeBitbucketServer,
		UpdatedAt:           updatedAt,
		Metadata: &bitbucketserver.PullRequest{
			State: state,
			Reviewers: []bitbucketserver.Reviewer{
				{Status: reviewStatus},
			},
		},
	}
}

func githubChangeset(updatedAt time.Time, state string) *campaigns.Changeset {
	return &campaigns.Changeset{
		ExternalServiceType: extsvc.TypeGitHub,
		UpdatedAt:           updatedAt,
		Metadata:            &github.PullRequest{State: state},
	}
}

func gitLabChangeset(updatedAt time.Time, state gitlab.MergeRequestState, notes []*gitlab.Note) *campaigns.Changeset {
	return &campaigns.Changeset{
		ExternalServiceType: extsvc.TypeGitLab,
		UpdatedAt:           updatedAt,
		Metadata: &gitlab.MergeRequest{
			Notes: notes,
			State: state,
		},
	}
}

func setDeletedAt(c *campaigns.Changeset, deletedAt time.Time) *campaigns.Changeset {
	c.ExternalDeletedAt = deletedAt
	return c
}
