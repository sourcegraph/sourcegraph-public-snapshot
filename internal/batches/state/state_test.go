package state

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	azuredevops2 "github.com/sourcegraph/sourcegraph/internal/batches/sources/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestComputeGithubCheckState(t *testing.T) {
	t.Parallel()

	now := timeutil.Now()
	commitEvent := func(minutesSinceSync int, context, state string) *btypes.ChangesetEvent {
		commit := &github.CommitStatus{
			Context:    context,
			State:      state,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		event := &btypes.ChangesetEvent{
			Kind:     btypes.ChangesetEventKindCommitStatus,
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
	checkSuiteEvent := func(minutesSinceSync int, id, status, conclusion string, runs ...github.CheckRun) *btypes.ChangesetEvent {
		suite := &github.CheckSuite{
			ID:         id,
			Status:     status,
			Conclusion: conclusion,
			ReceivedAt: now.Add(time.Duration(minutesSinceSync) * time.Minute),
		}
		suite.CheckRuns.Nodes = runs
		event := &btypes.ChangesetEvent{
			Kind:     btypes.ChangesetEventKindCheckSuite,
			Metadata: suite,
		}
		return event
	}

	lastSynced := now.Add(-1 * time.Minute)
	pr := &github.PullRequest{}

	tests := []struct {
		name   string
		events []*btypes.ChangesetEvent
		want   btypes.ChangesetCheckState
	}{
		{
			name:   "empty slice",
			events: nil,
			want:   btypes.ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "success status and suite",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "COMPLETED", "SUCCESS", checkRun("cr1", "COMPLETED", "SUCCESS")),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "single error",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
			},
			want: btypes.ChangesetCheckStateFailed,
		},
		{
			name: "pending + error",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "success + error",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			want: btypes.ChangesetCheckStateFailed,
		},
		{
			name: "success x2",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "later events have precedence",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "queued suites with zero runs should be ignored",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "QUEUED", ""),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "completed suites with zero runs should be ignored",
			events: []*btypes.ChangesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
				checkSuiteEvent(1, "cs1", "COMPLETED", ""),
			},
			want: btypes.ChangesetCheckStateFailed,
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
	statusEvent := func(key, state string) *btypes.ChangesetEvent {
		commit := &bitbucketserver.CommitStatus{
			Commit: sha,
			Status: bitbucketserver.BuildStatus{
				State:     state,
				Key:       key,
				DateAdded: now.Add(1*time.Second).Unix() * 1000,
			},
		}
		event := &btypes.ChangesetEvent{
			Kind:     btypes.ChangesetEventKindBitbucketServerCommitStatus,
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
		events []*btypes.ChangesetEvent
		want   btypes.ChangesetCheckState
	}{
		{
			name:   "empty slice",
			events: nil,
			want:   btypes.ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "SUCCESSFUL"),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "INPROGRESS"),
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "single error",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "FAILED"),
			},
			want: btypes.ChangesetCheckStateFailed,
		},
		{
			name: "pending + error",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "INPROGRESS"),
				statusEvent("ctx2", "FAILED"),
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "INPROGRESS"),
				statusEvent("ctx2", "SUCCESSFUL"),
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "success + error",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "SUCCESSFUL"),
				statusEvent("ctx2", "FAILED"),
			},
			want: btypes.ChangesetCheckStateFailed,
		},
		{
			name: "success x2",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "SUCCESSFUL"),
				statusEvent("ctx2", "SUCCESSFUL"),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "later events have precedence",
			events: []*btypes.ChangesetEvent{
				statusEvent("ctx1", "INPROGRESS"),
				statusEvent("ctx1", "SUCCESSFUL"),
			},
			want: btypes.ChangesetCheckStatePassed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			have := computeBitbucketServerBuildStatus(lastSynced, pr, tc.events)
			if diff := cmp.Diff(tc.want, have); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestComputeAzureDevOpsBuildStatus(t *testing.T) {
	t.Parallel()

	pr := &azuredevops2.AnnotatedPullRequest{
		PullRequest: &azuredevops.PullRequest{},
	}

	tests := []struct {
		name     string
		statuses []*azuredevops.PullRequestBuildStatus
		want     btypes.ChangesetCheckState
	}{
		{
			name:     "empty slice",
			statuses: []*azuredevops.PullRequestBuildStatus{},
			want:     btypes.ChangesetCheckStateUnknown,
		},
		{
			name: "single success",
			statuses: []*azuredevops.PullRequestBuildStatus{
				{
					ID:    1,
					State: azuredevops.PullRequestBuildStatusStateSucceeded,
				},
			},
			want: btypes.ChangesetCheckStatePassed,
		},
		{
			name: "single pending",
			statuses: []*azuredevops.PullRequestBuildStatus{
				{
					ID:    1,
					State: azuredevops.PullRequestBuildStatusStatePending,
				},
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "single error",
			statuses: []*azuredevops.PullRequestBuildStatus{
				{
					ID:    1,
					State: azuredevops.PullRequestBuildStatusStateError,
				},
			},
			want: btypes.ChangesetCheckStateFailed,
		},
		{
			name: "single failed",
			statuses: []*azuredevops.PullRequestBuildStatus{
				{
					ID:    1,
					State: azuredevops.PullRequestBuildStatusStateFailed,
				},
			},
			want: btypes.ChangesetCheckStateFailed,
		},
		{
			name: "failed + pending",
			statuses: []*azuredevops.PullRequestBuildStatus{
				{
					ID:    1,
					State: azuredevops.PullRequestBuildStatusStateFailed,
				},
				{
					ID:    2,
					State: azuredevops.PullRequestBuildStatusStatePending,
				},
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "pending + success",
			statuses: []*azuredevops.PullRequestBuildStatus{
				{
					ID:    1,
					State: azuredevops.PullRequestBuildStatusStateSucceeded,
				},
				{
					ID:    2,
					State: azuredevops.PullRequestBuildStatusStatePending,
				},
			},
			want: btypes.ChangesetCheckStatePending,
		},
		{
			name: "failed + success",
			statuses: []*azuredevops.PullRequestBuildStatus{
				{
					ID:    1,
					State: azuredevops.PullRequestBuildStatusStateFailed,
				},
				{
					ID:    2,
					State: azuredevops.PullRequestBuildStatusStateSucceeded,
				},
			},
			want: btypes.ChangesetCheckStateFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pr.Statuses = tc.statuses
			have := computeAzureDevOpsBuildState(pr)
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
			want btypes.ChangesetCheckState
		}{
			"no pipelines at all": {
				mr:   &gitlab.MergeRequest{},
				want: btypes.ChangesetCheckStateUnknown,
			},
			"only a head pipeline": {
				mr: &gitlab.MergeRequest{
					HeadPipeline: &gitlab.Pipeline{
						Status: gitlab.PipelineStatusPending,
					},
				},
				want: btypes.ChangesetCheckStatePending,
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
				want: btypes.ChangesetCheckStateFailed,
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
				want: btypes.ChangesetCheckStateFailed,
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
				want: btypes.ChangesetCheckStatePassed,
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

		events := []*btypes.ChangesetEvent{
			{
				Kind: btypes.ChangesetEventKindGitLabPipeline,
				Metadata: &gitlab.Pipeline{
					CreatedAt: gitlab.Time{Time: time.Unix(5, 0)},
					Status:    gitlab.PipelineStatusSuccess,
				},
			},
			{
				Kind: btypes.ChangesetEventKindGitLabPipeline,
				Metadata: &gitlab.Pipeline{
					CreatedAt: gitlab.Time{Time: time.Unix(4, 0)},
					Status:    gitlab.PipelineStatusFailed,
				},
			},
		}

		for name, tc := range map[string]struct {
			events     []*btypes.ChangesetEvent
			lastSynced time.Time
			want       btypes.ChangesetCheckState
		}{
			"older events only": {
				events:     events,
				lastSynced: time.Unix(10, 0),
				want:       btypes.ChangesetCheckStatePending,
			},
			"newer events only": {
				events:     events,
				lastSynced: time.Unix(3, 0),
				want:       btypes.ChangesetCheckStatePassed,
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
		changeset *btypes.Changeset
		history   []changesetStatesAtTime
		want      btypes.ChangesetReviewState
	}{
		{
			name:      "github - review required",
			changeset: githubChangeset(daysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetReviewStatePending,
		},
		{
			name:      "github - approved by codeowner",
			changeset: githubChangeset(daysAgo(10), "OPEN", "APPROVED"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), reviewState: btypes.ChangesetReviewStateApproved},
			},
			want: btypes.ChangesetReviewStateApproved,
		},
		{
			name:      "github - requires changes",
			changeset: githubChangeset(daysAgo(0), "OPEN", "CHANGES_REQUESTED"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), reviewState: btypes.ChangesetReviewStateChangesRequested},
			},
			want: btypes.ChangesetReviewStateChangesRequested,
		},
		{
			name:      "bitbucketserver - no events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetReviewStateChangesRequested,
		},

		{
			name:      "bitbucketserver - changeset older than events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), reviewState: btypes.ChangesetReviewStateApproved},
			},
			want: btypes.ChangesetReviewStateApproved,
		},

		{
			name:      "bitbucketserver - changeset newer than events",
			changeset: bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), reviewState: btypes.ChangesetReviewStateApproved},
			},
			want: btypes.ChangesetReviewStateChangesRequested,
		},
		{
			name:      "gitlab - no events, no approvals",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, []*gitlab.Note{}),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetReviewStatePending,
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
			want:    btypes.ChangesetReviewStateApproved,
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
			want:    btypes.ChangesetReviewStateChangesRequested,
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
			want:    btypes.ChangesetReviewStateChangesRequested,
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
				{t: daysAgo(0), reviewState: btypes.ChangesetReviewStateApproved},
			},
			want: btypes.ChangesetReviewStateApproved,
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
				{t: daysAgo(10), reviewState: btypes.ChangesetReviewStateApproved},
			},
			want: btypes.ChangesetReviewStateChangesRequested,
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

	archivedRepo := &types.Repo{Archived: true}

	tests := []struct {
		name      string
		changeset *btypes.Changeset
		repo      *types.Repo
		history   []changesetStatesAtTime
		want      btypes.ChangesetExternalState
		wantErr   error
	}{
		{
			name:      "github - no events",
			changeset: githubChangeset(daysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateOpen,
		},
		{
			name:      "github - changeset older than events",
			changeset: githubChangeset(daysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateClosed,
		},
		{
			name:      "github - changeset newer than events",
			changeset: githubChangeset(daysAgo(0), "OPEN", "REVIEW_REQUIRED"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateOpen,
		},
		{
			name:      "github - changeset newer and deleted",
			changeset: setDeletedAt(githubChangeset(daysAgo(0), "OPEN", "REVIEW_REQUIRED"), daysAgo(0)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateDeleted,
		},
		{
			name:      "github draft - no events",
			changeset: setDraft(githubChangeset(daysAgo(10), "OPEN", "REVIEW_REQUIRED")),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateDraft,
		},
		{
			name:      "github draft - changeset older than events",
			changeset: githubChangeset(daysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: btypes.ChangesetExternalStateDraft},
			},
			want: btypes.ChangesetExternalStateDraft,
		},
		{
			name:      "github draft - changeset newer than events",
			changeset: setDraft(githubChangeset(daysAgo(0), "OPEN", "REVIEW_REQUIRED")),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateDraft,
		},
		{
			name:      "github draft closed",
			changeset: setDraft(githubChangeset(daysAgo(1), "CLOSED", "REVIEW_REQUIRED")),
			history:   []changesetStatesAtTime{{t: daysAgo(2), externalState: btypes.ChangesetExternalStateClosed}},
			want:      btypes.ChangesetExternalStateClosed,
		},
		{
			name:      "bitbucketserver - no events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateOpen,
		},
		{
			name:      "bitbucketserver - changeset older than events",
			changeset: bitbucketChangeset(daysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateClosed,
		},
		{
			name:      "bitbucketserver - changeset newer than events",
			changeset: bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateOpen,
		},
		{
			name:      "bitbucketserver - changeset newer and deleted",
			changeset: setDeletedAt(bitbucketChangeset(daysAgo(0), "OPEN", "NEEDS_WORK"), daysAgo(0)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateDeleted,
		},
		{
			name:      "gitlab - no events, opened",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, nil),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateOpen,
		},
		{
			name:      "gitlab - no events, closed",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateClosed, nil),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateClosed,
		},
		{
			name:      "gitlab - no events, locked",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateLocked, nil),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateClosed,
		},
		{
			name:      "gitlab - no events, merged",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateMerged, nil),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateMerged,
		},
		{
			name:      "gitlab - changeset older than events",
			changeset: gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateMerged, nil),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateClosed,
		},
		{
			name:      "gitlab - changeset newer than events",
			changeset: gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateMerged, nil),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateMerged,
		},
		{
			name:      "gitlab draft - no events",
			changeset: setDraft(gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateOpened, nil)),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateDraft,
		},
		{
			name:      "gitlab draft - changeset older than events",
			changeset: gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateOpened, nil),
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: btypes.ChangesetExternalStateDraft},
			},
			want: btypes.ChangesetExternalStateDraft,
		},
		{
			name:      "gitlab draft - changeset newer than events",
			changeset: setDraft(gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, nil)),
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateDraft,
		},
		{
			name:      "gitlab read-only - no events",
			changeset: setDraft(gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateOpened, nil)),
			repo:      archivedRepo,
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateReadOnly,
		},
		{
			name:      "gitlab read-only - changeset older than events",
			changeset: gitLabChangeset(daysAgo(10), gitlab.MergeRequestStateOpened, nil),
			repo:      archivedRepo,
			history: []changesetStatesAtTime{
				{t: daysAgo(0), externalState: btypes.ChangesetExternalStateDraft},
			},
			want: btypes.ChangesetExternalStateReadOnly,
		},
		{
			name:      "gitlab read-only - changeset newer than events",
			changeset: setDraft(gitLabChangeset(daysAgo(0), gitlab.MergeRequestStateOpened, nil)),
			repo:      archivedRepo,
			history: []changesetStatesAtTime{
				{t: daysAgo(10), externalState: btypes.ChangesetExternalStateClosed},
			},
			want: btypes.ChangesetExternalStateReadOnly,
		},
		{
			name:      "perforce closed - no events",
			changeset: perforceChangeset(daysAgo(10), perforce.ChangelistStateClosed),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateClosed,
		},
		{
			name:      "perforce submitted - no events",
			changeset: perforceChangeset(daysAgo(10), perforce.ChangelistStateSubmitted),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateMerged,
		},
		{
			name:      "perforce pending - no events",
			changeset: perforceChangeset(daysAgo(10), perforce.ChangelistStatePending),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateOpen,
		},
		{
			name:      "perforce shelved - no events",
			changeset: perforceChangeset(daysAgo(10), perforce.ChangelistStateShelved),
			history:   []changesetStatesAtTime{},
			want:      btypes.ChangesetExternalStateOpen,
		},
		{
			name:      "perforce unknown state",
			changeset: perforceChangeset(daysAgo(10), "foobar"),
			history:   []changesetStatesAtTime{},
			wantErr:   errors.New("unknown Perforce Change state: foobar"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changeset := tc.changeset

			repo := tc.repo
			if repo == nil {
				repo = &types.Repo{Archived: false}
			}

			have, err := computeExternalState(changeset, tc.history, repo)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				assert.Empty(t, have)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, have)
			}
		})
	}
}

func TestComputeLabels(t *testing.T) {
	t.Parallel()

	now := timeutil.Now()
	labelEvent := func(name string, kind btypes.ChangesetEventKind, when time.Time) *btypes.ChangesetEvent {
		removed := kind == btypes.ChangesetEventKindGitHubUnlabeled
		return &btypes.ChangesetEvent{
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
	changeset := func(names []string, updated time.Time) *btypes.Changeset {
		meta := &github.PullRequest{}
		for _, name := range names {
			meta.Labels.Nodes = append(meta.Labels.Nodes, github.Label{
				Name: name,
			})
		}
		return &btypes.Changeset{
			UpdatedAt: updated,
			Metadata:  meta,
		}
	}
	labels := func(names ...string) []btypes.ChangesetLabel {
		var ls []btypes.ChangesetLabel
		for _, name := range names {
			ls = append(ls, btypes.ChangesetLabel{Name: name})
		}
		return ls
	}

	tests := []struct {
		name      string
		changeset *btypes.Changeset
		events    ChangesetEvents
		want      []btypes.ChangesetLabel
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
				labelEvent("label1", btypes.ChangesetEventKindGitHubUnlabeled, now),
			},
			want: []btypes.ChangesetLabel{},
		},
		{
			name:      "add event",
			changeset: changeset([]string{"label1"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label2", btypes.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1", "label2"),
		},
		{
			name:      "old add event",
			changeset: changeset([]string{"label1"}, now.Add(5*time.Minute)),
			events: ChangesetEvents{
				labelEvent("label2", btypes.ChangesetEventKindGitHubLabeled, now),
			},
			want: labels("label1"),
		},
		{
			name:      "sorting",
			changeset: changeset([]string{"label4", "label3"}, time.Time{}),
			events: ChangesetEvents{
				labelEvent("label2", btypes.ChangesetEventKindGitHubLabeled, now),
				labelEvent("label1", btypes.ChangesetEventKindGitHubLabeled, now.Add(5*time.Minute)),
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

func bitbucketChangeset(updatedAt time.Time, state, reviewStatus string) *btypes.Changeset {
	return &btypes.Changeset{
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

func githubChangeset(updatedAt time.Time, state string, reviewDecision string) *btypes.Changeset {
	return &btypes.Changeset{
		ExternalServiceType: extsvc.TypeGitHub,
		UpdatedAt:           updatedAt,
		Metadata:            &github.PullRequest{State: state, ReviewDecision: reviewDecision},
	}
}

func gitLabChangeset(updatedAt time.Time, state gitlab.MergeRequestState, notes []*gitlab.Note) *btypes.Changeset {
	return &btypes.Changeset{
		ExternalServiceType: extsvc.TypeGitLab,
		UpdatedAt:           updatedAt,
		Metadata: &gitlab.MergeRequest{
			Notes: notes,
			State: state,
		},
	}
}

func perforceChangeset(updatedAt time.Time, state perforce.ChangelistState) *btypes.Changeset {
	return &btypes.Changeset{
		ExternalServiceType: extsvc.TypePerforce,
		UpdatedAt:           updatedAt,
		Metadata: &perforce.Changelist{
			State: state,
		},
	}
}

func setDeletedAt(c *btypes.Changeset, deletedAt time.Time) *btypes.Changeset {
	c.ExternalDeletedAt = deletedAt
	return c
}
