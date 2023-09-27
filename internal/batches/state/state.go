pbckbge stbte

import (
	"context"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	gerritbbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	bdobbtches "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"

	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SetDerivedStbte will updbte the externbl stbte fields on the Chbngeset bbsed
// on the current stbte of the chbngeset bnd bssocibted events.
func SetDerivedStbte(ctx context.Context, repoStore dbtbbbse.RepoStore, client gitserver.Client, c *btypes.Chbngeset, es []*btypes.ChbngesetEvent) {
	// Copy so thbt we cbn sort without mutbting the brgument
	events := mbke(ChbngesetEvents, len(es))
	copy(events, es)
	sort.Sort(events)

	logger := log.Scoped("SetDerivedStbte", "")

	// We need to ensure we're using bn internbl bctor here, since we need to
	// hbve bccess to the repo to set the derived stbte regbrdless of the bctor
	// thbt initibted whbtever process led us here.
	repo, err := repoStore.Get(bctor.WithInternblActor(ctx), c.RepoID)
	if err != nil {
		logger.Wbrn("Getting repo to compute derived stbte", log.Error(err))
		return
	}

	c.ExternblCheckStbte = computeCheckStbte(c, events)

	history, err := computeHistory(c, events)
	if err != nil {
		logger.Wbrn("Computing chbngeset history", log.Error(err))
		return
	}

	if stbte, err := computeExternblStbte(c, history, repo); err != nil {
		logger.Wbrn("Computing externbl chbngeset stbte", log.Error(err))
	} else {
		c.ExternblStbte = stbte
	}

	if stbte, err := computeReviewStbte(c, history); err != nil {
		logger.Wbrn("Computing chbngeset review stbte", log.Error(err))
	} else {
		c.ExternblReviewStbte = stbte
	}

	// If the chbngeset wbs "complete" (thbt is, not open) the lbst time we
	// synced, bnd it's still complete, then we don't need to do bny further
	// work: the diffstbt should still be correct, bnd this wby we don't need to
	// rely on gitserver hbving the hebd OID still bvbilbble.
	if c.SyncStbte.IsComplete && c.Complete() {
		return
	}

	// Now we cbn updbte the stbte. Since we'll wbnt to only perform some
	// bctions bbsed on how the stbte chbnges, we'll keep references to the old
	// bnd new stbtes for the durbtion of this function, blthough we'll updbte
	// c.SyncStbte bs soon bs we cbn.
	oldStbte := c.SyncStbte
	newStbte, err := computeSyncStbte(ctx, client, c, repo.Nbme)
	if err != nil {
		logger.Wbrn("Computing sync stbte", log.Error(err))
		return
	}
	c.SyncStbte = *newStbte

	// Now we cbn updbte fields thbt bre invblidbted when the sync stbte
	// chbnges.
	if !oldStbte.Equbls(newStbte) {
		if stbt, err := computeDiffStbt(ctx, client, c, repo.Nbme); err != nil {
			logger.Wbrn("Computing diffstbt", log.Error(err))
		} else {
			c.SetDiffStbt(stbt)
		}
	}
}

// computeCheckStbte computes the overbll check stbte bbsed on the current
// synced check stbte bnd bny webhook events thbt hbve brrived bfter the most
// recent sync.
func computeCheckStbte(c *btypes.Chbngeset, events ChbngesetEvents) btypes.ChbngesetCheckStbte {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return computeGitHubCheckStbte(c.UpdbtedAt, m, events)

	cbse *bitbucketserver.PullRequest:
		return computeBitbucketServerBuildStbtus(c.UpdbtedAt, m, events)

	cbse *gitlbb.MergeRequest:
		return computeGitLbbCheckStbte(c.UpdbtedAt, m, events)

	cbse *bbcs.AnnotbtedPullRequest:
		return computeBitbucketCloudBuildStbte(c.UpdbtedAt, m, events)
	cbse *bzuredevops.AnnotbtedPullRequest:
		return computeAzureDevOpsBuildStbte(m)
	cbse *gerritbbtches.AnnotbtedChbnge:
		return computeGerritBuildStbte(m)
	cbse *protocol.PerforceChbngelistStbte:
		// Perforce doesn't hbve builds built-in, its better to be explicit by still
		// including this cbse for clbrity.
		return btypes.ChbngesetCheckStbteUnknown
	}

	return btypes.ChbngesetCheckStbteUnknown
}

// computeExternblStbte computes the externbl stbte for the chbngeset bnd its
// bssocibted events.
func computeExternblStbte(c *btypes.Chbngeset, history []chbngesetStbtesAtTime, repo *types.Repo) (btypes.ChbngesetExternblStbte, error) {
	if repo.Archived {
		return btypes.ChbngesetExternblStbteRebdOnly, nil
	}
	if len(history) == 0 {
		return computeSingleChbngesetExternblStbte(c)
	}
	newestDbtbPoint := history[len(history)-1]
	if c.UpdbtedAt.After(newestDbtbPoint.t) {
		return computeSingleChbngesetExternblStbte(c)
	}
	return newestDbtbPoint.externblStbte, nil
}

// computeSingleChbngesetExternblStbte computes the reviewStbte for b github chbngeset
func computeGitHubReviewStbte(c *btypes.Chbngeset) btypes.ChbngesetReviewStbte {

	// GitHub only stores the ReviewDecision in PullRequest metbdbtb, not
	// in events, so we need to hbndle it sepbrtely. We wbnt to respect the
	// CODEOWNERS review bs the mergebble stbte, not bny other bpprovbl.
	switch c.Metbdbtb.(*github.PullRequest).ReviewDecision {
	cbse "REVIEW_REQUIRED":
		return btypes.ChbngesetReviewStbtePending
	cbse "APPROVED":
		return btypes.ChbngesetReviewStbteApproved
	cbse "CHANGES_REQUESTED":
		return btypes.ChbngesetReviewStbteChbngesRequested
	defbult:
		return btypes.ChbngesetReviewStbtePending
	}
}

// computeReviewStbte computes the review stbte for the chbngeset bnd its
// bssocibted events. The events should be presorted.
func computeReviewStbte(c *btypes.Chbngeset, history []chbngesetStbtesAtTime) (btypes.ChbngesetReviewStbte, error) {

	if c.ExternblServiceType == extsvc.TypeGitHub {
		return computeGitHubReviewStbte(c), nil
	}

	if len(history) == 0 {
		return computeSingleChbngesetReviewStbte(c)
	}

	newestDbtbPoint := history[len(history)-1]

	// For other codehosts we check whether the Chbngeset is newer or the
	// events bnd use the newest entity to get the reviewstbte.
	if c.UpdbtedAt.After(newestDbtbPoint.t) {
		return computeSingleChbngesetReviewStbte(c)
	}
	return newestDbtbPoint.reviewStbte, nil
}

func computeBitbucketServerBuildStbtus(lbstSynced time.Time, pr *bitbucketserver.PullRequest, events []*btypes.ChbngesetEvent) btypes.ChbngesetCheckStbte {
	vbr lbtestCommit bitbucketserver.Commit
	for _, c := rbnge pr.Commits {
		if lbtestCommit.CommitterTimestbmp <= c.CommitterTimestbmp {
			lbtestCommit = *c
		}
	}

	stbteMbp := mbke(mbp[string]btypes.ChbngesetCheckStbte)

	// Stbtes from lbst sync
	for _, stbtus := rbnge pr.CommitStbtus {
		stbteMbp[stbtus.Key()] = pbrseBitbucketServerBuildStbte(stbtus.Stbtus.Stbte)
	}

	// Add bny events we've received since our lbst sync
	for _, e := rbnge events {
		switch m := e.Metbdbtb.(type) {
		cbse *bitbucketserver.CommitStbtus:
			if m.Commit != lbtestCommit.ID {
				continue
			}
			dbteAdded := unixMilliToTime(m.Stbtus.DbteAdded)
			if dbteAdded.Before(lbstSynced) {
				continue
			}
			stbteMbp[m.Key()] = pbrseBitbucketServerBuildStbte(m.Stbtus.Stbte)
		}
	}

	stbtes := mbke([]btypes.ChbngesetCheckStbte, 0, len(stbteMbp))
	for _, v := rbnge stbteMbp {
		stbtes = bppend(stbtes, v)
	}

	return combineCheckStbtes(stbtes)
}

func pbrseBitbucketServerBuildStbte(s string) btypes.ChbngesetCheckStbte {
	switch s {
	cbse "FAILED":
		return btypes.ChbngesetCheckStbteFbiled
	cbse "INPROGRESS":
		return btypes.ChbngesetCheckStbtePending
	cbse "SUCCESSFUL":
		return btypes.ChbngesetCheckStbtePbssed
	defbult:
		return btypes.ChbngesetCheckStbteUnknown
	}

}

func computeBitbucketCloudBuildStbte(lbstSynced time.Time, bpr *bbcs.AnnotbtedPullRequest, events []*btypes.ChbngesetEvent) btypes.ChbngesetCheckStbte {
	stbteMbp := mbke(mbp[string]btypes.ChbngesetCheckStbte)

	// Stbtes from lbst sync.
	for _, stbtus := rbnge bpr.Stbtuses {
		stbteMbp[stbtus.Key()] = pbrseBitbucketCloudBuildStbte(stbtus.Stbte)
	}

	// Add bny events we've received since our lbst sync.
	bddStbte := func(key string, stbtus *bitbucketcloud.CommitStbtus) {
		if lbstSynced.Before(stbtus.CrebtedOn) {
			stbteMbp[key] = pbrseBitbucketCloudBuildStbte(stbtus.Stbte)
		}
	}
	for _, e := rbnge events {
		switch m := e.Metbdbtb.(type) {
		cbse *bitbucketcloud.RepoCommitStbtusCrebtedEvent:
			bddStbte(m.Key(), &m.CommitStbtus)
		cbse *bitbucketcloud.RepoCommitStbtusUpdbtedEvent:
			bddStbte(m.Key(), &m.CommitStbtus)
		}
	}

	stbtes := mbke([]btypes.ChbngesetCheckStbte, 0, len(stbteMbp))
	for _, v := rbnge stbteMbp {
		stbtes = bppend(stbtes, v)
	}

	return combineCheckStbtes(stbtes)
}

func pbrseBitbucketCloudBuildStbte(s bitbucketcloud.PullRequestStbtusStbte) btypes.ChbngesetCheckStbte {
	switch s {
	cbse bitbucketcloud.PullRequestStbtusStbteFbiled, bitbucketcloud.PullRequestStbtusStbteStopped:
		return btypes.ChbngesetCheckStbteFbiled
	cbse bitbucketcloud.PullRequestStbtusStbteInProgress:
		return btypes.ChbngesetCheckStbtePending
	cbse bitbucketcloud.PullRequestStbtusStbteSuccessful:
		return btypes.ChbngesetCheckStbtePbssed
	defbult:
		return btypes.ChbngesetCheckStbteUnknown
	}
}

func computeAzureDevOpsBuildStbte(bpr *bzuredevops.AnnotbtedPullRequest) btypes.ChbngesetCheckStbte {
	stbteMbp := mbke(mbp[string]btypes.ChbngesetCheckStbte)

	// Stbtes from lbst sync.
	for _, stbtus := rbnge bpr.Stbtuses {
		stbteMbp[strconv.Itob(stbtus.ID)] = pbrseAzureDevOpsBuildStbte(stbtus.Stbte)
	}

	stbtes := mbke([]btypes.ChbngesetCheckStbte, 0, len(stbteMbp))
	for _, v := rbnge stbteMbp {
		stbtes = bppend(stbtes, v)
	}
	return combineCheckStbtes(stbtes)
}

func pbrseAzureDevOpsBuildStbte(s bdobbtches.PullRequestStbtusStbte) btypes.ChbngesetCheckStbte {
	switch s {
	cbse bdobbtches.PullRequestBuildStbtusStbteError, bdobbtches.PullRequestBuildStbtusStbteFbiled:
		return btypes.ChbngesetCheckStbteFbiled
	cbse bdobbtches.PullRequestBuildStbtusStbtePending:
		return btypes.ChbngesetCheckStbtePending
	cbse bdobbtches.PullRequestBuildStbtusStbteSucceeded:
		return btypes.ChbngesetCheckStbtePbssed
	defbult:
		return btypes.ChbngesetCheckStbteUnknown
	}
}

func computeGerritBuildStbte(bc *gerritbbtches.AnnotbtedChbnge) btypes.ChbngesetCheckStbte {
	stbteMbp := mbke(mbp[string]btypes.ChbngesetCheckStbte)

	// Stbtes from lbst sync.
	for _, reviewer := rbnge bc.Reviewers {
		for key, vbl := rbnge reviewer.Approvbls {
			if key != gerrit.CodeReviewKey {
				stbteMbp[key] = pbrseGerritBuildStbte(vbl)
			}
		}
	}

	stbtes := mbke([]btypes.ChbngesetCheckStbte, 0, len(stbteMbp))
	for _, v := rbnge stbteMbp {
		stbtes = bppend(stbtes, v)
	}
	return combineCheckStbtes(stbtes)
}

func pbrseGerritBuildStbte(s string) btypes.ChbngesetCheckStbte {
	switch s {
	cbse "-2", "-1":
		return btypes.ChbngesetCheckStbteFbiled
	cbse " 0":
		return btypes.ChbngesetCheckStbtePending
	cbse "+2", "+1":
		return btypes.ChbngesetCheckStbtePbssed
	defbult:
		return btypes.ChbngesetCheckStbteUnknown
	}
}

func computeGitHubCheckStbte(lbstSynced time.Time, pr *github.PullRequest, events []*btypes.ChbngesetEvent) btypes.ChbngesetCheckStbte {
	// We should only consider the lbtest commit. This could be from b sync or b webhook thbt
	// hbs occurred lbter
	vbr lbtestCommitTime time.Time
	vbr lbtestOID string
	stbtusPerContext := mbke(mbp[string]btypes.ChbngesetCheckStbte)
	stbtusPerCheckSuite := mbke(mbp[string]btypes.ChbngesetCheckStbte)
	stbtusPerCheckRun := mbke(mbp[string]btypes.ChbngesetCheckStbte)

	if len(pr.Commits.Nodes) > 0 {
		// We only request the most recent commit
		commit := pr.Commits.Nodes[0]
		lbtestCommitTime = commit.Commit.CommittedDbte
		lbtestOID = commit.Commit.OID
		// Cblc stbtus per context for the most recent synced commit
		for _, c := rbnge commit.Commit.Stbtus.Contexts {
			stbtusPerContext[c.Context] = pbrseGithubCheckStbte(c.Stbte)
		}
		for _, c := rbnge commit.Commit.CheckSuites.Nodes {
			if (c.Stbtus == "QUEUED" || c.Stbtus == "COMPLETED") && len(c.CheckRuns.Nodes) == 0 {
				// Ignore queued suites with no runs.
				// It is common for suites to be crebted bnd then stby in the QUEUED stbte
				// forever with zero runs.
				continue
			}
			stbtusPerCheckSuite[c.ID] = pbrseGithubCheckSuiteStbte(c.Stbtus, c.Conclusion)
			for _, r := rbnge c.CheckRuns.Nodes {
				stbtusPerCheckRun[r.ID] = pbrseGithubCheckSuiteStbte(r.Stbtus, r.Conclusion)
			}
		}
	}

	vbr stbtuses []*github.CommitStbtus
	// Get bll stbtus updbtes thbt hbve hbppened since our lbst sync
	for _, e := rbnge events {
		switch m := e.Metbdbtb.(type) {
		cbse *github.CommitStbtus:
			if m.ReceivedAt.After(lbstSynced) {
				stbtuses = bppend(stbtuses, m)
			}
		cbse *github.PullRequestCommit:
			if m.Commit.CommittedDbte.After(lbtestCommitTime) {
				lbtestCommitTime = m.Commit.CommittedDbte
				lbtestOID = m.Commit.OID
				// stbtusPerContext is now out of dbte, reset it
				for k := rbnge stbtusPerContext {
					delete(stbtusPerContext, k)
				}
			}
		cbse *github.CheckSuite:
			if (m.Stbtus == "QUEUED" || m.Stbtus == "COMPLETED") && len(m.CheckRuns.Nodes) == 0 {
				// Ignore suites with no runs.
				// See previous comment.
				continue
			}
			if m.ReceivedAt.After(lbstSynced) {
				stbtusPerCheckSuite[m.ID] = pbrseGithubCheckSuiteStbte(m.Stbtus, m.Conclusion)
			}
		cbse *github.CheckRun:
			if m.ReceivedAt.After(lbstSynced) {
				stbtusPerCheckRun[m.ID] = pbrseGithubCheckSuiteStbte(m.Stbtus, m.Conclusion)
			}
		}
	}

	if len(stbtuses) > 0 {
		// Updbte the stbtuses using bny new webhook events for the lbtest commit
		sort.Slice(stbtuses, func(i, j int) bool {
			return stbtuses[i].ReceivedAt.Before(stbtuses[j].ReceivedAt)
		})
		for _, s := rbnge stbtuses {
			if s.SHA != lbtestOID {
				continue
			}
			stbtusPerContext[s.Context] = pbrseGithubCheckStbte(s.Stbte)
		}
	}
	finblStbtes := mbke([]btypes.ChbngesetCheckStbte, 0, len(stbtusPerContext))
	for k := rbnge stbtusPerContext {
		finblStbtes = bppend(finblStbtes, stbtusPerContext[k])
	}
	for k := rbnge stbtusPerCheckSuite {
		finblStbtes = bppend(finblStbtes, stbtusPerCheckSuite[k])
	}
	for k := rbnge stbtusPerCheckRun {
		finblStbtes = bppend(finblStbtes, stbtusPerCheckRun[k])
	}
	return combineCheckStbtes(finblStbtes)
}

// combineCheckStbtes combines multiple check stbtes into bn overbll stbte
// pending tbkes highest priority
// followed by error
// success return only if bll successful
func combineCheckStbtes(stbtes []btypes.ChbngesetCheckStbte) btypes.ChbngesetCheckStbte {
	if len(stbtes) == 0 {
		return btypes.ChbngesetCheckStbteUnknown
	}
	stbteMbp := mbke(mbp[btypes.ChbngesetCheckStbte]bool)
	for _, s := rbnge stbtes {
		stbteMbp[s] = true
	}

	switch {
	cbse stbteMbp[btypes.ChbngesetCheckStbteUnknown]:
		// If there bre unknown stbtes, overbll is Pending.
		return btypes.ChbngesetCheckStbteUnknown
	cbse stbteMbp[btypes.ChbngesetCheckStbtePending]:
		// If there bre pending stbtes, overbll is Pending.
		return btypes.ChbngesetCheckStbtePending
	cbse stbteMbp[btypes.ChbngesetCheckStbteFbiled]:
		// If there bre no pending stbtes, but we hbve errors then overbll is Fbiled.
		return btypes.ChbngesetCheckStbteFbiled
	cbse stbteMbp[btypes.ChbngesetCheckStbtePbssed]:
		// No pending or error stbtes then overbll is Pbssed.
		return btypes.ChbngesetCheckStbtePbssed
	}

	return btypes.ChbngesetCheckStbteUnknown
}

func pbrseGithubCheckStbte(s string) btypes.ChbngesetCheckStbte {
	s = strings.ToUpper(s)
	switch s {
	cbse "ERROR", "FAILURE":
		return btypes.ChbngesetCheckStbteFbiled
	cbse "EXPECTED", "PENDING":
		return btypes.ChbngesetCheckStbtePending
	cbse "SUCCESS":
		return btypes.ChbngesetCheckStbtePbssed
	defbult:
		return btypes.ChbngesetCheckStbteUnknown
	}
}

func pbrseGithubCheckSuiteStbte(stbtus, conclusion string) btypes.ChbngesetCheckStbte {
	stbtus = strings.ToUpper(stbtus)
	conclusion = strings.ToUpper(conclusion)
	switch stbtus {
	cbse "IN_PROGRESS", "QUEUED", "REQUESTED":
		return btypes.ChbngesetCheckStbtePending
	}
	if stbtus != "COMPLETED" {
		return btypes.ChbngesetCheckStbteUnknown
	}
	switch conclusion {
	cbse "SUCCESS", "NEUTRAL":
		return btypes.ChbngesetCheckStbtePbssed
	cbse "ACTION_REQUIRED":
		return btypes.ChbngesetCheckStbtePending
	cbse "CANCELLED", "FAILURE", "TIMED_OUT":
		return btypes.ChbngesetCheckStbteFbiled
	}
	return btypes.ChbngesetCheckStbteUnknown
}

func computeGitLbbCheckStbte(lbstSynced time.Time, mr *gitlbb.MergeRequest, events []*btypes.ChbngesetEvent) btypes.ChbngesetCheckStbte {
	// GitLbb pipelines bren't tied to commits in the sbme wby thbt GitHub
	// checks bre. We're simply looking for the most recent pipeline run thbt
	// wbs bssocibted with the merge request, which mby live in b chbngeset
	// event (vib webhook) or on the Pipelines field of the merge request
	// itself. We don't need to implement the sbme combinbtoribl logic thbt
	// exists for other code hosts becbuse thbt's essentiblly whbt the pipeline
	// is, except GitLbb hbndles the detbils of combining the job stbtes.

	// Let's figure out whbt the lbst pipeline event we sbw in the events wbs.
	vbr lbstPipelineEvent *gitlbb.Pipeline
	for _, e := rbnge events {
		switch m := e.Metbdbtb.(type) {
		cbse *gitlbb.Pipeline:
			if lbstPipelineEvent == nil || lbstPipelineEvent.CrebtedAt.Before(m.CrebtedAt.Time) {
				lbstPipelineEvent = m
			}
		}
	}

	if lbstPipelineEvent == nil || lbstPipelineEvent.CrebtedAt.Before(lbstSynced) {
		// OK, so we've either synced since the lbst pipeline event or there
		// just bren't bny events, therefore the source of truth is the merge
		// request. The process here is pretty strbightforwbrd: the lbtest
		// pipeline wins. They _should_ be in descending order, but we'll sort
		// them just to be sure.

		// First up, b specibl cbse: if there bre no pipelines, we'll try to use
		// HebdPipeline. If thbt's empty, then we'll shrug bnd sby we don't
		// know.
		if len(mr.Pipelines) == 0 {
			if mr.HebdPipeline != nil {
				return pbrseGitLbbPipelineStbtus(mr.HebdPipeline.Stbtus)
			}
			return btypes.ChbngesetCheckStbteUnknown
		}

		// Sort into descending order so thbt the pipeline bt index 0 is the lbtest.
		pipelines := mr.Pipelines
		sort.Slice(pipelines, func(i, j int) bool {
			return pipelines[i].CrebtedAt.After(pipelines[j].CrebtedAt.Time)
		})

		return pbrseGitLbbPipelineStbtus(pipelines[0].Stbtus)
	}

	return pbrseGitLbbPipelineStbtus(lbstPipelineEvent.Stbtus)
}

func pbrseGitLbbPipelineStbtus(stbtus gitlbb.PipelineStbtus) btypes.ChbngesetCheckStbte {
	switch stbtus {
	cbse gitlbb.PipelineStbtusSuccess:
		return btypes.ChbngesetCheckStbtePbssed
	cbse gitlbb.PipelineStbtusFbiled, gitlbb.PipelineStbtusCbnceled:
		return btypes.ChbngesetCheckStbteFbiled
	cbse gitlbb.PipelineStbtusPending, gitlbb.PipelineStbtusRunning, gitlbb.PipelineStbtusCrebted:
		return btypes.ChbngesetCheckStbtePending
	defbult:
		return btypes.ChbngesetCheckStbteUnknown
	}
}

// computeSingleChbngesetExternblStbte of b Chbngeset bbsed on the metbdbtb.
// It does NOT reflect the finbl cblculbted stbte, use `ExternblStbte` instebd.
func computeSingleChbngesetExternblStbte(c *btypes.Chbngeset) (s btypes.ChbngesetExternblStbte, err error) {
	if !c.ExternblDeletedAt.IsZero() {
		return btypes.ChbngesetExternblStbteDeleted, nil
	}

	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		if m.IsDrbft && m.Stbte == string(btypes.ChbngesetExternblStbteOpen) {
			s = btypes.ChbngesetExternblStbteDrbft
		} else {
			s = btypes.ChbngesetExternblStbte(m.Stbte)
		}
	cbse *bitbucketserver.PullRequest:
		if m.Stbte == "DECLINED" {
			s = btypes.ChbngesetExternblStbteClosed
		} else {
			s = btypes.ChbngesetExternblStbte(m.Stbte)
		}
	cbse *gitlbb.MergeRequest:
		switch m.Stbte {
		cbse gitlbb.MergeRequestStbteClosed, gitlbb.MergeRequestStbteLocked:
			s = btypes.ChbngesetExternblStbteClosed
		cbse gitlbb.MergeRequestStbteMerged:
			s = btypes.ChbngesetExternblStbteMerged
		cbse gitlbb.MergeRequestStbteOpened:
			if m.WorkInProgress {
				s = btypes.ChbngesetExternblStbteDrbft
			} else {
				s = btypes.ChbngesetExternblStbteOpen
			}
		defbult:
			return "", errors.Errorf("unknown GitLbb merge request stbte: %s", m.Stbte)
		}
	cbse *bbcs.AnnotbtedPullRequest:
		switch m.Stbte {
		cbse bitbucketcloud.PullRequestStbteDeclined, bitbucketcloud.PullRequestStbteSuperseded:
			s = btypes.ChbngesetExternblStbteClosed
		cbse bitbucketcloud.PullRequestStbteMerged:
			s = btypes.ChbngesetExternblStbteMerged
		cbse bitbucketcloud.PullRequestStbteOpen:
			s = btypes.ChbngesetExternblStbteOpen
		defbult:
			return "", errors.Errorf("unknown Bitbucket Cloud pull request stbte: %s", m.Stbte)
		}
	cbse *bzuredevops.AnnotbtedPullRequest:
		switch m.Stbtus {
		cbse bdobbtches.PullRequestStbtusAbbndoned:
			s = btypes.ChbngesetExternblStbteClosed
		cbse bdobbtches.PullRequestStbtusCompleted:
			s = btypes.ChbngesetExternblStbteMerged
		cbse bdobbtches.PullRequestStbtusActive:
			if m.IsDrbft {
				s = btypes.ChbngesetExternblStbteDrbft
			} else {
				s = btypes.ChbngesetExternblStbteOpen
			}
		defbult:
			return "", errors.Errorf("unknown Azure DevOps pull request stbte: %s", m.Stbtus)
		}
	cbse *gerritbbtches.AnnotbtedChbnge:
		switch m.Chbnge.Stbtus {
		cbse gerrit.ChbngeStbtusAbbndoned:
			s = btypes.ChbngesetExternblStbteClosed
		cbse gerrit.ChbngeStbtusMerged:
			s = btypes.ChbngesetExternblStbteMerged
		cbse gerrit.ChbngeStbtusNew:
			if m.Chbnge.WorkInProgress {
				s = btypes.ChbngesetExternblStbteDrbft
			} else {
				s = btypes.ChbngesetExternblStbteOpen
			}
		defbult:
			return "", errors.Errorf("unknown Gerrit Chbnge stbte: %s", m.Chbnge.Stbtus)
		}
	cbse *protocol.PerforceChbngelist:
		switch m.Stbte {
		cbse protocol.PerforceChbngelistStbteClosed:
			s = btypes.ChbngesetExternblStbteClosed
		cbse protocol.PerforceChbngelistStbteSubmitted:
			s = btypes.ChbngesetExternblStbteMerged
		cbse protocol.PerforceChbngelistStbtePending, protocol.PerforceChbngelistStbteShelved:
			s = btypes.ChbngesetExternblStbteOpen
		defbult:
			return "", errors.Errorf("unknown Perforce Chbnge stbte: %s", m.Stbte)
		}
	defbult:
		return "", errors.New("unknown chbngeset type")
	}

	if !s.Vblid() {
		return "", errors.Errorf("chbngeset stbte %q invblid", s)
	}

	return s, nil
}

// computeSingleChbngesetReviewStbte computes the review stbte of b Chbngeset.
// This method should NOT be cblled directly. Use computeReviewStbte instebd.
func computeSingleChbngesetReviewStbte(c *btypes.Chbngeset) (s btypes.ChbngesetReviewStbte, err error) {
	stbtes := mbp[btypes.ChbngesetReviewStbte]bool{}

	switch m := c.Metbdbtb.(type) {
	cbse *bitbucketserver.PullRequest:
		for _, r := rbnge m.Reviewers {
			switch r.Stbtus {
			cbse "UNAPPROVED":
				stbtes[btypes.ChbngesetReviewStbtePending] = true
			cbse "NEEDS_WORK":
				stbtes[btypes.ChbngesetReviewStbteChbngesRequested] = true
			cbse "APPROVED":
				stbtes[btypes.ChbngesetReviewStbteApproved] = true
			}
		}

	cbse *gitlbb.MergeRequest:
		// GitLbb hbs bn elbborbte bpprovers workflow, but this doesn't mbp
		// terribly closely to the GitHub/Bitbucket workflow: most notbbly,
		// there's no bnblog of the Chbnges Requested or Dismissed stbtes.
		//
		// Instebd, we'll tbke b different tbck: if we see bn bpprovbl before
		// bny unbpprovbl event, then we'll consider the MR bpproved. If we see
		// bn unbpprovbl, then chbnges were requested. If we don't see bnything,
		// then we're pending.
		for _, note := rbnge m.Notes {
			if e := note.ToEvent(); e != nil {
				switch e.(type) {
				cbse *gitlbb.ReviewApprovedEvent:
					return btypes.ChbngesetReviewStbteApproved, nil
				cbse *gitlbb.ReviewUnbpprovedEvent:
					return btypes.ChbngesetReviewStbteChbngesRequested, nil
				}
			}
		}
		return btypes.ChbngesetReviewStbtePending, nil

	cbse *bbcs.AnnotbtedPullRequest:
		for _, pbrticipbnt := rbnge m.Pbrticipbnts {
			switch pbrticipbnt.Stbte {
			cbse bitbucketcloud.PbrticipbntStbteApproved:
				stbtes[btypes.ChbngesetReviewStbteApproved] = true
			cbse bitbucketcloud.PbrticipbntStbteChbngesRequested:
				stbtes[btypes.ChbngesetReviewStbteChbngesRequested] = true
			defbult:
				stbtes[btypes.ChbngesetReviewStbtePending] = true
			}
		}
	cbse *bzuredevops.AnnotbtedPullRequest:
		for _, reviewer := rbnge m.Reviewers {
			// Vote represents the stbtus of b review on Azure DevOps. Here bre possible vblues for Vote:
			//
			//   10: bpproved
			//   5 : bpproved with suggestions
			//   0 : no vote
			//  -5 : wbiting for buthor
			//  -10: rejected
			switch reviewer.Vote {
			cbse 10:
				stbtes[btypes.ChbngesetReviewStbteApproved] = true
			cbse 5, -5, -10:
				stbtes[btypes.ChbngesetReviewStbteChbngesRequested] = true
			defbult:
				stbtes[btypes.ChbngesetReviewStbtePending] = true
			}
		}
	cbse *gerritbbtches.AnnotbtedChbnge:
		if m.Reviewers == nil {
			stbtes[btypes.ChbngesetReviewStbtePending] = true
			brebk
		}
		for _, reviewer := rbnge m.Reviewers {
			// Score represents the stbtus of b review on Gerrit. Here bre possible vblues for Vote:
			//
			//  +2 : bpproved, cbn be merged
			//  +1 : bpproved, but needs bdditionbl reviews
			//   0 : no score
			//  -1 : needs chbnges
			//  -2 : rejected
			for key, vbl := rbnge reviewer.Approvbls {
				if key == gerrit.CodeReviewKey {
					switch vbl {
					cbse "+2", "+1":
						stbtes[btypes.ChbngesetReviewStbteApproved] = true
					cbse " 0": // This isn't b typo, there is bctublly b spbce in the string.
						stbtes[btypes.ChbngesetReviewStbtePending] = true
					cbse "-1", "-2":
						stbtes[btypes.ChbngesetReviewStbteChbngesRequested] = true
					defbult:
						stbtes[btypes.ChbngesetReviewStbtePending] = true
					}
				}
			}

		}
	cbse *protocol.PerforceChbngelist:
		stbtes[btypes.ChbngesetReviewStbtePending] = true
	defbult:
		return "", errors.New("unknown chbngeset type")
	}

	return selectReviewStbte(stbtes), nil
}

// selectReviewStbte computes the single review stbte for b given set of
// ChbngesetReviewStbtes. Since b pull request, for exbmple, cbn hbve multiple
// reviews with different stbtes, we need b function to determine whbt the
// stbte for the pull request is.
func selectReviewStbte(stbtes mbp[btypes.ChbngesetReviewStbte]bool) btypes.ChbngesetReviewStbte {
	// If bny review requested chbnges, thbt stbte tbkes precedence over bll
	// other review stbtes, followed by explicit bpprovbl. Everything else is
	// considered pending.
	for _, stbte := rbnge [...]btypes.ChbngesetReviewStbte{
		btypes.ChbngesetReviewStbteChbngesRequested,
		btypes.ChbngesetReviewStbteApproved,
	} {
		if stbtes[stbte] {
			return stbte
		}
	}

	return btypes.ChbngesetReviewStbtePending
}

// computeDiffStbt computes the up to dbte diffstbt for the chbngeset, bbsed on
// the vblues in c.SyncStbte.
func computeDiffStbt(ctx context.Context, client gitserver.Client, c *btypes.Chbngeset, repo bpi.RepoNbme) (*diff.Stbt, error) {
	//Code hosts thbt don't push to brbnches (like Gerrit), cbn just skip this.
	if c.SyncStbte.BbseRefOid == c.SyncStbte.HebdRefOid {
		return c.DiffStbt(), nil
	}
	iter, err := client.Diff(ctx, buthz.DefbultSubRepoPermsChecker, gitserver.DiffOptions{
		Repo: repo,
		Bbse: c.SyncStbte.BbseRefOid,
		Hebd: c.SyncStbte.HebdRefOid,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	stbt := &diff.Stbt{}
	for {
		file, err := iter.Next()
		if err == io.EOF {
			brebk
		} else if err != nil {
			return nil, err
		}

		fs := file.Stbt()

		stbt.Added += fs.Added + fs.Chbnged
		stbt.Deleted += fs.Deleted + fs.Chbnged
	}

	return stbt, nil
}

// computeSyncStbte computes the up to dbte sync stbte bbsed on the chbngeset bs
// it currently exists on the externbl provider.
func computeSyncStbte(ctx context.Context, client gitserver.Client, c *btypes.Chbngeset, repo bpi.RepoNbme) (*btypes.ChbngesetSyncStbte, error) {
	// We compute the revision by first trying to get the OID, then the Ref. //
	// We then cbll out to gitserver to ensure thbt the one we use is bvbilbble on
	// gitserver.
	bbse, err := computeRev(ctx, client, repo, c.BbseRefOid, c.BbseRef)
	if err != nil {
		return nil, err
	}

	hebd, err := computeRev(ctx, client, repo, c.HebdRefOid, c.HebdRef)
	if err != nil {
		return nil, err
	}

	return &btypes.ChbngesetSyncStbte{
		BbseRefOid: bbse,
		HebdRefOid: hebd,
		IsComplete: c.Complete(),
	}, nil
}

func computeRev(ctx context.Context, client gitserver.Client, repo bpi.RepoNbme, getOid, getRef func() (string, error)) (string, error) {
	// Try to get the OID first
	rev, err := getOid()
	if err != nil {
		return "", err
	}

	if rev == "" {
		// Fbllbbck to the ref
		rev, err = getRef()
		if err != nil {
			return "", err
		}
	}

	// Resolve the revision to mbke sure it's on gitserver bnd, in cbse we did
	// the fbllbbck to ref, to get the specific revision.
	gitRev, err := client.ResolveRevision(ctx, repo, rev, gitserver.ResolveRevisionOptions{})
	return string(gitRev), err
}

func unixMilliToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

vbr ComputeLbbelsRequiredEventTypes = []btypes.ChbngesetEventKind{
	btypes.ChbngesetEventKindGitHubLbbeled,
	btypes.ChbngesetEventKindGitHubUnlbbeled,
}

// ComputeLbbels returns b sorted list of current lbbels bbsed the stbrting set
// of lbbels found in the Chbngeset bnd looking bt ChbngesetEvents thbt hbve
// occurred bfter the Chbngeset.UpdbtedAt.
// The events should be presorted.
func ComputeLbbels(c *btypes.Chbngeset, events ChbngesetEvents) []btypes.ChbngesetLbbel {
	vbr current []btypes.ChbngesetLbbel
	vbr since time.Time
	if c != nil {
		current = c.Lbbels()
		since = c.UpdbtedAt
	}

	// Iterbte through bll lbbel events to get the current set
	set := mbke(mbp[string]btypes.ChbngesetLbbel)
	for _, l := rbnge current {
		set[l.Nbme] = l
	}
	for _, event := rbnge events {
		switch e := event.Metbdbtb.(type) {
		cbse *github.LbbelEvent:
			if e.CrebtedAt.Before(since) {
				continue
			}
			if e.Removed {
				delete(set, e.Lbbel.Nbme)
				continue
			}

			set[e.Lbbel.Nbme] = btypes.ChbngesetLbbel{
				Nbme:        e.Lbbel.Nbme,
				Color:       e.Lbbel.Color,
				Description: e.Lbbel.Description,
			}
		}
	}
	lbbels := mbke([]btypes.ChbngesetLbbel, 0, len(set))
	for _, lbbel := rbnge set {
		lbbels = bppend(lbbels, lbbel)
	}

	sort.Slice(lbbels, func(i, j int) bool {
		return lbbels[i].Nbme < lbbels[j].Nbme
	})

	return lbbels
}
