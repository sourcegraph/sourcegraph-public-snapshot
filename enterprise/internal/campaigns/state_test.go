package campaigns

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
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

func TestComputeGitLabCheckState(t *testing.T) {
	for name, tc := range map[string]struct {
		mr   *gitlab.MergeRequest
		want cmpgn.ChangesetCheckState
	}{
		"no pipelines at all": {
			mr:   &gitlab.MergeRequest{},
			want: cmpgn.ChangesetCheckStateUnknown,
		},
		"only a head pipeline": {
			mr: &gitlab.MergeRequest{
				HeadPipeline: &gitlab.Pipeline{
					Status: gitlab.PipelineStatusPending,
				},
			},
			want: cmpgn.ChangesetCheckStatePending,
		},
		"one pipeline only": {
			mr: &gitlab.MergeRequest{
				HeadPipeline: &gitlab.Pipeline{
					Status: gitlab.PipelineStatusPending,
				},
				Pipelines: []*gitlab.Pipeline{
					{
						CreatedAt: time.Unix(10, 0),
						Status:    gitlab.PipelineStatusFailed,
					},
				},
			},
			want: cmpgn.ChangesetCheckStateFailed,
		},
		"two pipelines in the expected order": {
			mr: &gitlab.MergeRequest{
				HeadPipeline: &gitlab.Pipeline{
					Status: gitlab.PipelineStatusPending,
				},
				Pipelines: []*gitlab.Pipeline{
					{
						CreatedAt: time.Unix(10, 0),
						Status:    gitlab.PipelineStatusFailed,
					},
					{
						CreatedAt: time.Unix(5, 0),
						Status:    gitlab.PipelineStatusSuccess,
					},
				},
			},
			want: cmpgn.ChangesetCheckStateFailed,
		},
		"two pipelines in an unexpected order": {
			mr: &gitlab.MergeRequest{
				HeadPipeline: &gitlab.Pipeline{
					Status: gitlab.PipelineStatusPending,
				},
				Pipelines: []*gitlab.Pipeline{
					{
						CreatedAt: time.Unix(5, 0),
						Status:    gitlab.PipelineStatusFailed,
					},
					{
						CreatedAt: time.Unix(10, 0),
						Status:    gitlab.PipelineStatusSuccess,
					},
				},
			},
			want: cmpgn.ChangesetCheckStatePassed,
		},
	} {
		t.Run(name, func(t *testing.T) {
			have := computeGitLabCheckState(tc.mr)
			if have != tc.want {
				t.Errorf("unexpected check state: have %s; want %s", have, tc.want)
			}
		})
	}
}

func TestComputeReviewState(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	tests := []struct {
		name      string
		changeset *campaigns.Changeset
		history   []changesetStatesAtTime
		want      cmpgn.ChangesetReviewState
	}{
		{
			name:      "github - no events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history:   []changesetStatesAtTime{},
			want:      cmpgn.ChangesetReviewStatePending,
		},
		{
			name:      "github - changeset older than events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			name:      "github - changeset newer than events",
			changeset: githubChangeset(daysAgo(0), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},
		{
			name:      "bitbucketserver - no events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []changesetStatesAtTime{},
			want:      cmpgn.ChangesetReviewStateChangesRequested,
		},

		{
			name:      "bitbucketserver - changeset older than events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: cmpgn.ChangesetReviewStateApproved,
		},

		{
			name:      "bitbucketserver - changeset newer than events",
			changeset: bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), reviewState: campaigns.ChangesetReviewStateApproved},
			},
			want: cmpgn.ChangesetReviewStateChangesRequested,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changeset := tc.changeset

			have, err := ComputeReviewState(changeset, tc.history)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}

			if have, want := have, tc.want; have != want {
				t.Errorf("%d: wrong reviewstate. have=%s, want=%s", i, have, want)
			}
		})
	}
}

func TestComputeChangesetState(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	daysAgo := func(days int) time.Time { return now.AddDate(0, 0, -days) }

	tests := []struct {
		name      string
		changeset *campaigns.Changeset
		history   []changesetStatesAtTime
		want      cmpgn.ChangesetState
	}{
		{
			name:      "github - no events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history:   []changesetStatesAtTime{},
			want:      cmpgn.ChangesetStateOpen,
		},
		{
			name:      "github - changeset older than events",
			changeset: githubChangeset(daysAgo(10), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), state: campaigns.ChangesetStateClosed},
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			name:      "github - changeset newer than events",
			changeset: githubChangeset(daysAgo(0), "OPEN"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), state: campaigns.ChangesetStateClosed},
			},
			want: cmpgn.ChangesetStateOpen,
		},
		{
			name:      "github - changeset newer and deleted",
			changeset: setDeletedAt(githubChangeset(daysAgo(0), "OPEN"), daysAgo(0)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), state: campaigns.ChangesetStateClosed},
			},
			want: cmpgn.ChangesetStateDeleted,
		},
		{
			name:      "bitbucketserver - no events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []changesetStatesAtTime{},
			want:      cmpgn.ChangesetStateOpen,
		},
		{
			name:      "bitbucketserver - changeset older than events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), state: campaigns.ChangesetStateClosed},
			},
			want: cmpgn.ChangesetStateClosed,
		},
		{
			name:      "bitbucketserver - changeset newer than events",
			changeset: bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), state: campaigns.ChangesetStateClosed},
			},
			want: cmpgn.ChangesetStateOpen,
		},
		{
			name:      "bitbucketserver - changeset newer and deleted",
			changeset: setDeletedAt(bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"), daysAgo(0)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), state: campaigns.ChangesetStateClosed},
			},
			want: cmpgn.ChangesetStateDeleted,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changeset := tc.changeset

			have, err := ComputeChangesetState(changeset, tc.history)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}

			if have, want := have, tc.want; have != want {
				t.Errorf("%d: wrong changeset state. have=%s, want=%s", i, have, want)
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
			// TODO: Reviewers should be its own struct
			Reviewers: []struct {
				User               *bitbucketserver.User `json:"user"`
				LastReviewedCommit string                `json:"lastReviewedCommit"`
				Role               string                `json:"role"`
				Approved           bool                  `json:"approved"`
				Status             string                `json:"status"`
			}{
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

func setDeletedAt(c *campaigns.Changeset, deletedAt time.Time) *campaigns.Changeset {
	c.ExternalDeletedAt = deletedAt
	return c
}
