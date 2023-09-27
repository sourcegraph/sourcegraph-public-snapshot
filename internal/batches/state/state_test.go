pbckbge stbte

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	bzuredevops2 "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestComputeGithubCheckStbte(t *testing.T) {
	t.Pbrbllel()

	now := timeutil.Now()
	commitEvent := func(minutesSinceSync int, context, stbte string) *btypes.ChbngesetEvent {
		commit := &github.CommitStbtus{
			Context:    context,
			Stbte:      stbte,
			ReceivedAt: now.Add(time.Durbtion(minutesSinceSync) * time.Minute),
		}
		event := &btypes.ChbngesetEvent{
			Kind:     btypes.ChbngesetEventKindCommitStbtus,
			Metbdbtb: commit,
		}
		return event
	}
	checkRun := func(id, stbtus, conclusion string) github.CheckRun {
		return github.CheckRun{
			ID:         id,
			Stbtus:     stbtus,
			Conclusion: conclusion,
		}
	}
	checkSuiteEvent := func(minutesSinceSync int, id, stbtus, conclusion string, runs ...github.CheckRun) *btypes.ChbngesetEvent {
		suite := &github.CheckSuite{
			ID:         id,
			Stbtus:     stbtus,
			Conclusion: conclusion,
			ReceivedAt: now.Add(time.Durbtion(minutesSinceSync) * time.Minute),
		}
		suite.CheckRuns.Nodes = runs
		event := &btypes.ChbngesetEvent{
			Kind:     btypes.ChbngesetEventKindCheckSuite,
			Metbdbtb: suite,
		}
		return event
	}

	lbstSynced := now.Add(-1 * time.Minute)
	pr := &github.PullRequest{}

	tests := []struct {
		nbme   string
		events []*btypes.ChbngesetEvent
		wbnt   btypes.ChbngesetCheckStbte
	}{
		{
			nbme:   "empty slice",
			events: nil,
			wbnt:   btypes.ChbngesetCheckStbteUnknown,
		},
		{
			nbme: "single success",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "success stbtus bnd suite",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "COMPLETED", "SUCCESS", checkRun("cr1", "COMPLETED", "SUCCESS")),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "single pending",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "single error",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
		{
			nbme: "pending + error",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "pending + success",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "success + error",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "ERROR"),
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
		{
			nbme: "success x2",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				commitEvent(1, "ctx2", "SUCCESS"),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "lbter events hbve precedence",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "PENDING"),
				commitEvent(1, "ctx1", "SUCCESS"),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "queued suites with zero runs should be ignored",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "SUCCESS"),
				checkSuiteEvent(1, "cs1", "QUEUED", ""),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "completed suites with zero runs should be ignored",
			events: []*btypes.ChbngesetEvent{
				commitEvent(1, "ctx1", "ERROR"),
				checkSuiteEvent(1, "cs1", "COMPLETED", ""),
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			got := computeGitHubCheckStbte(lbstSynced, pr, tc.events)
			if diff := cmp.Diff(tc.wbnt, got); diff != "" {
				t.Fbtblf(diff)
			}
		})
	}
}

func TestComputeBitbucketBuildStbtus(t *testing.T) {
	t.Pbrbllel()

	now := timeutil.Now()
	shb := "bbcdef"
	stbtusEvent := func(key, stbte string) *btypes.ChbngesetEvent {
		commit := &bitbucketserver.CommitStbtus{
			Commit: shb,
			Stbtus: bitbucketserver.BuildStbtus{
				Stbte:     stbte,
				Key:       key,
				DbteAdded: now.Add(1*time.Second).Unix() * 1000,
			},
		}
		event := &btypes.ChbngesetEvent{
			Kind:     btypes.ChbngesetEventKindBitbucketServerCommitStbtus,
			Metbdbtb: commit,
		}
		return event
	}

	lbstSynced := now.Add(-1 * time.Minute)
	pr := &bitbucketserver.PullRequest{
		Commits: []*bitbucketserver.Commit{
			{
				ID: shb,
			},
		},
	}

	tests := []struct {
		nbme   string
		events []*btypes.ChbngesetEvent
		wbnt   btypes.ChbngesetCheckStbte
	}{
		{
			nbme:   "empty slice",
			events: nil,
			wbnt:   btypes.ChbngesetCheckStbteUnknown,
		},
		{
			nbme: "single success",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "SUCCESSFUL"),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "single pending",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "INPROGRESS"),
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "single error",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "FAILED"),
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
		{
			nbme: "pending + error",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "INPROGRESS"),
				stbtusEvent("ctx2", "FAILED"),
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "pending + success",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "INPROGRESS"),
				stbtusEvent("ctx2", "SUCCESSFUL"),
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "success + error",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "SUCCESSFUL"),
				stbtusEvent("ctx2", "FAILED"),
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
		{
			nbme: "success x2",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "SUCCESSFUL"),
				stbtusEvent("ctx2", "SUCCESSFUL"),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "lbter events hbve precedence",
			events: []*btypes.ChbngesetEvent{
				stbtusEvent("ctx1", "INPROGRESS"),
				stbtusEvent("ctx1", "SUCCESSFUL"),
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve := computeBitbucketServerBuildStbtus(lbstSynced, pr, tc.events)
			if diff := cmp.Diff(tc.wbnt, hbve); diff != "" {
				t.Fbtblf(diff)
			}
		})
	}
}

func TestComputeAzureDevOpsBuildStbtus(t *testing.T) {
	t.Pbrbllel()

	pr := &bzuredevops2.AnnotbtedPullRequest{
		PullRequest: &bzuredevops.PullRequest{},
	}

	tests := []struct {
		nbme     string
		stbtuses []*bzuredevops.PullRequestBuildStbtus
		wbnt     btypes.ChbngesetCheckStbte
	}{
		{
			nbme:     "empty slice",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{},
			wbnt:     btypes.ChbngesetCheckStbteUnknown,
		},
		{
			nbme: "single success",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{
				{
					ID:    1,
					Stbte: bzuredevops.PullRequestBuildStbtusStbteSucceeded,
				},
			},
			wbnt: btypes.ChbngesetCheckStbtePbssed,
		},
		{
			nbme: "single pending",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{
				{
					ID:    1,
					Stbte: bzuredevops.PullRequestBuildStbtusStbtePending,
				},
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "single error",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{
				{
					ID:    1,
					Stbte: bzuredevops.PullRequestBuildStbtusStbteError,
				},
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
		{
			nbme: "single fbiled",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{
				{
					ID:    1,
					Stbte: bzuredevops.PullRequestBuildStbtusStbteFbiled,
				},
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
		{
			nbme: "fbiled + pending",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{
				{
					ID:    1,
					Stbte: bzuredevops.PullRequestBuildStbtusStbteFbiled,
				},
				{
					ID:    2,
					Stbte: bzuredevops.PullRequestBuildStbtusStbtePending,
				},
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "pending + success",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{
				{
					ID:    1,
					Stbte: bzuredevops.PullRequestBuildStbtusStbteSucceeded,
				},
				{
					ID:    2,
					Stbte: bzuredevops.PullRequestBuildStbtusStbtePending,
				},
			},
			wbnt: btypes.ChbngesetCheckStbtePending,
		},
		{
			nbme: "fbiled + success",
			stbtuses: []*bzuredevops.PullRequestBuildStbtus{
				{
					ID:    1,
					Stbte: bzuredevops.PullRequestBuildStbtusStbteFbiled,
				},
				{
					ID:    2,
					Stbte: bzuredevops.PullRequestBuildStbtusStbteSucceeded,
				},
			},
			wbnt: btypes.ChbngesetCheckStbteFbiled,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			pr.Stbtuses = tc.stbtuses
			hbve := computeAzureDevOpsBuildStbte(pr)
			if diff := cmp.Diff(tc.wbnt, hbve); diff != "" {
				t.Fbtblf(diff)
			}
		})
	}
}

func TestComputeGitLbbCheckStbte(t *testing.T) {
	t.Pbrbllel()

	t.Run("no events", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			mr   *gitlbb.MergeRequest
			wbnt btypes.ChbngesetCheckStbte
		}{
			"no pipelines bt bll": {
				mr:   &gitlbb.MergeRequest{},
				wbnt: btypes.ChbngesetCheckStbteUnknown,
			},
			"only b hebd pipeline": {
				mr: &gitlbb.MergeRequest{
					HebdPipeline: &gitlbb.Pipeline{
						Stbtus: gitlbb.PipelineStbtusPending,
					},
				},
				wbnt: btypes.ChbngesetCheckStbtePending,
			},
			"one pipeline only": {
				mr: &gitlbb.MergeRequest{
					HebdPipeline: &gitlbb.Pipeline{
						Stbtus: gitlbb.PipelineStbtusPending,
					},
					Pipelines: []*gitlbb.Pipeline{
						{
							CrebtedAt: gitlbb.Time{Time: time.Unix(10, 0)},
							Stbtus:    gitlbb.PipelineStbtusFbiled,
						},
					},
				},
				wbnt: btypes.ChbngesetCheckStbteFbiled,
			},
			"two pipelines in the expected order": {
				mr: &gitlbb.MergeRequest{
					HebdPipeline: &gitlbb.Pipeline{
						Stbtus: gitlbb.PipelineStbtusPending,
					},
					Pipelines: []*gitlbb.Pipeline{
						{
							CrebtedAt: gitlbb.Time{Time: time.Unix(10, 0)},
							Stbtus:    gitlbb.PipelineStbtusFbiled,
						},
						{
							CrebtedAt: gitlbb.Time{Time: time.Unix(5, 0)},
							Stbtus:    gitlbb.PipelineStbtusSuccess,
						},
					},
				},
				wbnt: btypes.ChbngesetCheckStbteFbiled,
			},
			"two pipelines in bn unexpected order": {
				mr: &gitlbb.MergeRequest{
					HebdPipeline: &gitlbb.Pipeline{
						Stbtus: gitlbb.PipelineStbtusPending,
					},
					Pipelines: []*gitlbb.Pipeline{
						{
							CrebtedAt: gitlbb.Time{Time: time.Unix(5, 0)},
							Stbtus:    gitlbb.PipelineStbtusFbiled,
						},
						{
							CrebtedAt: gitlbb.Time{Time: time.Unix(10, 0)},
							Stbtus:    gitlbb.PipelineStbtusSuccess,
						},
					},
				},
				wbnt: btypes.ChbngesetCheckStbtePbssed,
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				hbve := computeGitLbbCheckStbte(time.Unix(0, 0), tc.mr, nil)
				if hbve != tc.wbnt {
					t.Errorf("unexpected check stbte: hbve %s; wbnt %s", hbve, tc.wbnt)
				}
			})
		}
	})

	t.Run("with events", func(t *testing.T) {
		mr := &gitlbb.MergeRequest{
			HebdPipeline: &gitlbb.Pipeline{
				Stbtus: gitlbb.PipelineStbtusPending,
			},
		}

		events := []*btypes.ChbngesetEvent{
			{
				Kind: btypes.ChbngesetEventKindGitLbbPipeline,
				Metbdbtb: &gitlbb.Pipeline{
					CrebtedAt: gitlbb.Time{Time: time.Unix(5, 0)},
					Stbtus:    gitlbb.PipelineStbtusSuccess,
				},
			},
			{
				Kind: btypes.ChbngesetEventKindGitLbbPipeline,
				Metbdbtb: &gitlbb.Pipeline{
					CrebtedAt: gitlbb.Time{Time: time.Unix(4, 0)},
					Stbtus:    gitlbb.PipelineStbtusFbiled,
				},
			},
		}

		for nbme, tc := rbnge mbp[string]struct {
			events     []*btypes.ChbngesetEvent
			lbstSynced time.Time
			wbnt       btypes.ChbngesetCheckStbte
		}{
			"older events only": {
				events:     events,
				lbstSynced: time.Unix(10, 0),
				wbnt:       btypes.ChbngesetCheckStbtePending,
			},
			"newer events only": {
				events:     events,
				lbstSynced: time.Unix(3, 0),
				wbnt:       btypes.ChbngesetCheckStbtePbssed,
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				hbve := computeGitLbbCheckStbte(tc.lbstSynced, mr, tc.events)
				if hbve != tc.wbnt {
					t.Errorf("unexpected check stbte: hbve %s; wbnt %s", hbve, tc.wbnt)
				}
			})
		}
	})
}

func TestComputeReviewStbte(t *testing.T) {
	t.Pbrbllel()

	now := timeutil.Now()
	dbysAgo := func(dbys int) time.Time { return now.AddDbte(0, 0, -dbys) }

	tests := []struct {
		nbme      string
		chbngeset *btypes.Chbngeset
		history   []chbngesetStbtesAtTime
		wbnt      btypes.ChbngesetReviewStbte
	}{
		{
			nbme:      "github - review required",
			chbngeset: githubChbngeset(dbysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetReviewStbtePending,
		},
		{
			nbme:      "github - bpproved by codeowner",
			chbngeset: githubChbngeset(dbysAgo(10), "OPEN", "APPROVED"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), reviewStbte: btypes.ChbngesetReviewStbteApproved},
			},
			wbnt: btypes.ChbngesetReviewStbteApproved,
		},
		{
			nbme:      "github - requires chbnges",
			chbngeset: githubChbngeset(dbysAgo(0), "OPEN", "CHANGES_REQUESTED"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), reviewStbte: btypes.ChbngesetReviewStbteChbngesRequested},
			},
			wbnt: btypes.ChbngesetReviewStbteChbngesRequested,
		},
		{
			nbme:      "bitbucketserver - no events",
			chbngeset: bitbucketChbngeset(dbysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetReviewStbteChbngesRequested,
		},

		{
			nbme:      "bitbucketserver - chbngeset older thbn events",
			chbngeset: bitbucketChbngeset(dbysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), reviewStbte: btypes.ChbngesetReviewStbteApproved},
			},
			wbnt: btypes.ChbngesetReviewStbteApproved,
		},

		{
			nbme:      "bitbucketserver - chbngeset newer thbn events",
			chbngeset: bitbucketChbngeset(dbysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), reviewStbte: btypes.ChbngesetReviewStbteApproved},
			},
			wbnt: btypes.ChbngesetReviewStbteChbngesRequested,
		},
		{
			nbme:      "gitlbb - no events, no bpprovbls",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, []*gitlbb.Note{}),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetReviewStbtePending,
		},
		{
			nbme: "gitlbb - no events, one bpprovbl",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, []*gitlbb.Note{
				{
					System: true,
					Body:   "bpproved this merge request",
				},
			}),
			history: []chbngesetStbtesAtTime{},
			wbnt:    btypes.ChbngesetReviewStbteApproved,
		},
		{
			nbme: "gitlbb - no events, one unbpprovbl",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, []*gitlbb.Note{
				{
					System: true,
					Body:   "unbpproved this merge request",
				},
			}),
			history: []chbngesetStbtesAtTime{},
			wbnt:    btypes.ChbngesetReviewStbteChbngesRequested,
		},
		{
			nbme: "gitlbb - no events, severbl notes",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, []*gitlbb.Note{
				{Body: "this is b user note"},
				{
					System: true,
					Body:   "unbpproved this merge request",
				},
				{Body: "this is b user note"},
				{
					System: true,
					Body:   "bpproved this merge request",
				},
			}),
			history: []chbngesetStbtesAtTime{},
			wbnt:    btypes.ChbngesetReviewStbteChbngesRequested,
		},
		{
			nbme: "gitlbb - chbngeset older thbn events",
			chbngeset: gitLbbChbngeset(dbysAgo(10), gitlbb.MergeRequestStbteOpened, []*gitlbb.Note{
				{
					System: true,
					Body:   "unbpproved this merge request",
				},
			}),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), reviewStbte: btypes.ChbngesetReviewStbteApproved},
			},
			wbnt: btypes.ChbngesetReviewStbteApproved,
		},
		{
			nbme: "gitlbb - chbngeset newer thbn events",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, []*gitlbb.Note{
				{
					System: true,
					Body:   "unbpproved this merge request",
				},
			}),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), reviewStbte: btypes.ChbngesetReviewStbteApproved},
			},
			wbnt: btypes.ChbngesetReviewStbteChbngesRequested,
		},
	}

	for i, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			chbngeset := tc.chbngeset

			hbve, err := computeReviewStbte(chbngeset, tc.history)
			if err != nil {
				t.Fbtblf("got error: %s", err)
			}

			if hbve, wbnt := hbve, tc.wbnt; hbve != wbnt {
				t.Errorf("%d: wrong reviewstbte. hbve=%s, wbnt=%s", i, hbve, wbnt)
			}
		})
	}
}

func TestComputeExternblStbte(t *testing.T) {
	t.Pbrbllel()

	now := timeutil.Now()
	dbysAgo := func(dbys int) time.Time { return now.AddDbte(0, 0, -dbys) }

	brchivedRepo := &types.Repo{Archived: true}

	tests := []struct {
		nbme      string
		chbngeset *btypes.Chbngeset
		repo      *types.Repo
		history   []chbngesetStbtesAtTime
		wbnt      btypes.ChbngesetExternblStbte
		wbntErr   error
	}{
		{
			nbme:      "github - no events",
			chbngeset: githubChbngeset(dbysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteOpen,
		},
		{
			nbme:      "github - chbngeset older thbn events",
			chbngeset: githubChbngeset(dbysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteClosed,
		},
		{
			nbme:      "github - chbngeset newer thbn events",
			chbngeset: githubChbngeset(dbysAgo(0), "OPEN", "REVIEW_REQUIRED"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteOpen,
		},
		{
			nbme:      "github - chbngeset newer bnd deleted",
			chbngeset: setDeletedAt(githubChbngeset(dbysAgo(0), "OPEN", "REVIEW_REQUIRED"), dbysAgo(0)),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteDeleted,
		},
		{
			nbme:      "github drbft - no events",
			chbngeset: setDrbft(githubChbngeset(dbysAgo(10), "OPEN", "REVIEW_REQUIRED")),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteDrbft,
		},
		{
			nbme:      "github drbft - chbngeset older thbn events",
			chbngeset: githubChbngeset(dbysAgo(10), "OPEN", "REVIEW_REQUIRED"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), externblStbte: btypes.ChbngesetExternblStbteDrbft},
			},
			wbnt: btypes.ChbngesetExternblStbteDrbft,
		},
		{
			nbme:      "github drbft - chbngeset newer thbn events",
			chbngeset: setDrbft(githubChbngeset(dbysAgo(0), "OPEN", "REVIEW_REQUIRED")),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteDrbft,
		},
		{
			nbme:      "github drbft closed",
			chbngeset: setDrbft(githubChbngeset(dbysAgo(1), "CLOSED", "REVIEW_REQUIRED")),
			history:   []chbngesetStbtesAtTime{{t: dbysAgo(2), externblStbte: btypes.ChbngesetExternblStbteClosed}},
			wbnt:      btypes.ChbngesetExternblStbteClosed,
		},
		{
			nbme:      "bitbucketserver - no events",
			chbngeset: bitbucketChbngeset(dbysAgo(10), "OPEN", "NEEDS_WORK"),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteOpen,
		},
		{
			nbme:      "bitbucketserver - chbngeset older thbn events",
			chbngeset: bitbucketChbngeset(dbysAgo(10), "OPEN", "NEEDS_WORK"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteClosed,
		},
		{
			nbme:      "bitbucketserver - chbngeset newer thbn events",
			chbngeset: bitbucketChbngeset(dbysAgo(0), "OPEN", "NEEDS_WORK"),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteOpen,
		},
		{
			nbme:      "bitbucketserver - chbngeset newer bnd deleted",
			chbngeset: setDeletedAt(bitbucketChbngeset(dbysAgo(0), "OPEN", "NEEDS_WORK"), dbysAgo(0)),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteDeleted,
		},
		{
			nbme:      "gitlbb - no events, opened",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, nil),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteOpen,
		},
		{
			nbme:      "gitlbb - no events, closed",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteClosed, nil),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteClosed,
		},
		{
			nbme:      "gitlbb - no events, locked",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteLocked, nil),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteClosed,
		},
		{
			nbme:      "gitlbb - no events, merged",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteMerged, nil),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteMerged,
		},
		{
			nbme:      "gitlbb - chbngeset older thbn events",
			chbngeset: gitLbbChbngeset(dbysAgo(10), gitlbb.MergeRequestStbteMerged, nil),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteClosed,
		},
		{
			nbme:      "gitlbb - chbngeset newer thbn events",
			chbngeset: gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteMerged, nil),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteMerged,
		},
		{
			nbme:      "gitlbb drbft - no events",
			chbngeset: setDrbft(gitLbbChbngeset(dbysAgo(10), gitlbb.MergeRequestStbteOpened, nil)),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteDrbft,
		},
		{
			nbme:      "gitlbb drbft - chbngeset older thbn events",
			chbngeset: gitLbbChbngeset(dbysAgo(10), gitlbb.MergeRequestStbteOpened, nil),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), externblStbte: btypes.ChbngesetExternblStbteDrbft},
			},
			wbnt: btypes.ChbngesetExternblStbteDrbft,
		},
		{
			nbme:      "gitlbb drbft - chbngeset newer thbn events",
			chbngeset: setDrbft(gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, nil)),
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteDrbft,
		},
		{
			nbme:      "gitlbb rebd-only - no events",
			chbngeset: setDrbft(gitLbbChbngeset(dbysAgo(10), gitlbb.MergeRequestStbteOpened, nil)),
			repo:      brchivedRepo,
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteRebdOnly,
		},
		{
			nbme:      "gitlbb rebd-only - chbngeset older thbn events",
			chbngeset: gitLbbChbngeset(dbysAgo(10), gitlbb.MergeRequestStbteOpened, nil),
			repo:      brchivedRepo,
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(0), externblStbte: btypes.ChbngesetExternblStbteDrbft},
			},
			wbnt: btypes.ChbngesetExternblStbteRebdOnly,
		},
		{
			nbme:      "gitlbb rebd-only - chbngeset newer thbn events",
			chbngeset: setDrbft(gitLbbChbngeset(dbysAgo(0), gitlbb.MergeRequestStbteOpened, nil)),
			repo:      brchivedRepo,
			history: []chbngesetStbtesAtTime{
				{t: dbysAgo(10), externblStbte: btypes.ChbngesetExternblStbteClosed},
			},
			wbnt: btypes.ChbngesetExternblStbteRebdOnly,
		},
		{
			nbme:      "perforce closed - no events",
			chbngeset: perforceChbngeset(dbysAgo(10), protocol.PerforceChbngelistStbteClosed),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteClosed,
		},
		{
			nbme:      "perforce submitted - no events",
			chbngeset: perforceChbngeset(dbysAgo(10), protocol.PerforceChbngelistStbteSubmitted),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteMerged,
		},
		{
			nbme:      "perforce pending - no events",
			chbngeset: perforceChbngeset(dbysAgo(10), protocol.PerforceChbngelistStbtePending),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteOpen,
		},
		{
			nbme:      "perforce shelved - no events",
			chbngeset: perforceChbngeset(dbysAgo(10), protocol.PerforceChbngelistStbteShelved),
			history:   []chbngesetStbtesAtTime{},
			wbnt:      btypes.ChbngesetExternblStbteOpen,
		},
		{
			nbme:      "perforce unknown stbte",
			chbngeset: perforceChbngeset(dbysAgo(10), "foobbr"),
			history:   []chbngesetStbtesAtTime{},
			wbntErr:   errors.New("unknown Perforce Chbnge stbte: foobbr"),
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			chbngeset := tc.chbngeset

			repo := tc.repo
			if repo == nil {
				repo = &types.Repo{Archived: fblse}
			}

			hbve, err := computeExternblStbte(chbngeset, tc.history, repo)

			if tc.wbntErr != nil {
				require.Error(t, err)
				bssert.Equbl(t, tc.wbntErr.Error(), err.Error())
				bssert.Empty(t, hbve)
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, tc.wbnt, hbve)
			}
		})
	}
}

func TestComputeLbbels(t *testing.T) {
	t.Pbrbllel()

	now := timeutil.Now()
	lbbelEvent := func(nbme string, kind btypes.ChbngesetEventKind, when time.Time) *btypes.ChbngesetEvent {
		removed := kind == btypes.ChbngesetEventKindGitHubUnlbbeled
		return &btypes.ChbngesetEvent{
			Kind:      kind,
			UpdbtedAt: when,
			Metbdbtb: &github.LbbelEvent{
				Actor: github.Actor{},
				Lbbel: github.Lbbel{
					Nbme: nbme,
				},
				CrebtedAt: when,
				Removed:   removed,
			},
		}
	}
	chbngeset := func(nbmes []string, updbted time.Time) *btypes.Chbngeset {
		metb := &github.PullRequest{}
		for _, nbme := rbnge nbmes {
			metb.Lbbels.Nodes = bppend(metb.Lbbels.Nodes, github.Lbbel{
				Nbme: nbme,
			})
		}
		return &btypes.Chbngeset{
			UpdbtedAt: updbted,
			Metbdbtb:  metb,
		}
	}
	lbbels := func(nbmes ...string) []btypes.ChbngesetLbbel {
		vbr ls []btypes.ChbngesetLbbel
		for _, nbme := rbnge nbmes {
			ls = bppend(ls, btypes.ChbngesetLbbel{Nbme: nbme})
		}
		return ls
	}

	tests := []struct {
		nbme      string
		chbngeset *btypes.Chbngeset
		events    ChbngesetEvents
		wbnt      []btypes.ChbngesetLbbel
	}{
		{
			nbme: "zero vblues",
		},
		{
			nbme:      "no events",
			chbngeset: chbngeset([]string{"lbbel1"}, time.Time{}),
			events:    ChbngesetEvents{},
			wbnt:      lbbels("lbbel1"),
		},
		{
			nbme:      "remove event",
			chbngeset: chbngeset([]string{"lbbel1"}, time.Time{}),
			events: ChbngesetEvents{
				lbbelEvent("lbbel1", btypes.ChbngesetEventKindGitHubUnlbbeled, now),
			},
			wbnt: []btypes.ChbngesetLbbel{},
		},
		{
			nbme:      "bdd event",
			chbngeset: chbngeset([]string{"lbbel1"}, time.Time{}),
			events: ChbngesetEvents{
				lbbelEvent("lbbel2", btypes.ChbngesetEventKindGitHubLbbeled, now),
			},
			wbnt: lbbels("lbbel1", "lbbel2"),
		},
		{
			nbme:      "old bdd event",
			chbngeset: chbngeset([]string{"lbbel1"}, now.Add(5*time.Minute)),
			events: ChbngesetEvents{
				lbbelEvent("lbbel2", btypes.ChbngesetEventKindGitHubLbbeled, now),
			},
			wbnt: lbbels("lbbel1"),
		},
		{
			nbme:      "sorting",
			chbngeset: chbngeset([]string{"lbbel4", "lbbel3"}, time.Time{}),
			events: ChbngesetEvents{
				lbbelEvent("lbbel2", btypes.ChbngesetEventKindGitHubLbbeled, now),
				lbbelEvent("lbbel1", btypes.ChbngesetEventKindGitHubLbbeled, now.Add(5*time.Minute)),
			},
			wbnt: lbbels("lbbel1", "lbbel2", "lbbel3", "lbbel4"),
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve := ComputeLbbels(tc.chbngeset, tc.events)
			wbnt := tc.wbnt
			if diff := cmp.Diff(hbve, wbnt, cmpopts.EqubteEmpty()); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func bitbucketChbngeset(updbtedAt time.Time, stbte, reviewStbtus string) *btypes.Chbngeset {
	return &btypes.Chbngeset{
		ExternblServiceType: extsvc.TypeBitbucketServer,
		UpdbtedAt:           updbtedAt,
		Metbdbtb: &bitbucketserver.PullRequest{
			Stbte: stbte,
			Reviewers: []bitbucketserver.Reviewer{
				{Stbtus: reviewStbtus},
			},
		},
	}
}

func githubChbngeset(updbtedAt time.Time, stbte string, reviewDecision string) *btypes.Chbngeset {
	return &btypes.Chbngeset{
		ExternblServiceType: extsvc.TypeGitHub,
		UpdbtedAt:           updbtedAt,
		Metbdbtb:            &github.PullRequest{Stbte: stbte, ReviewDecision: reviewDecision},
	}
}

func gitLbbChbngeset(updbtedAt time.Time, stbte gitlbb.MergeRequestStbte, notes []*gitlbb.Note) *btypes.Chbngeset {
	return &btypes.Chbngeset{
		ExternblServiceType: extsvc.TypeGitLbb,
		UpdbtedAt:           updbtedAt,
		Metbdbtb: &gitlbb.MergeRequest{
			Notes: notes,
			Stbte: stbte,
		},
	}
}

func perforceChbngeset(updbtedAt time.Time, stbte protocol.PerforceChbngelistStbte) *btypes.Chbngeset {
	return &btypes.Chbngeset{
		ExternblServiceType: extsvc.TypePerforce,
		UpdbtedAt:           updbtedAt,
		Metbdbtb: &protocol.PerforceChbngelist{
			Stbte: stbte,
		},
	}
}

func setDeletedAt(c *btypes.Chbngeset, deletedAt time.Time) *btypes.Chbngeset {
	c.ExternblDeletedAt = deletedAt
	return c
}
