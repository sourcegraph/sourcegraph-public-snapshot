package campaigns

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
