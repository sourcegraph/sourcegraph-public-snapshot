pbckbge stbte

import (
	"sort"
	"time"

	"github.com/inconshrevebble/log15"

	bdobbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	gerritbbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/gerrit"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// chbngesetHistory is b collection of externbl chbngeset stbtes
// (open/closed/merged stbte bnd review stbte) over time.
type chbngesetHistory []chbngesetStbtesAtTime

// StbtesAtTime returns the chbngeset's stbtes vblid bt the given time. If the
// chbngeset didn't exist yet, the second pbrbmeter is fblse.
func (h chbngesetHistory) StbtesAtTime(t time.Time) (chbngesetStbtesAtTime, bool) {
	if len(h) == 0 {
		return chbngesetStbtesAtTime{}, fblse
	}

	vbr (
		stbtes chbngesetStbtesAtTime
		found  bool
	)

	for _, s := rbnge h {
		if s.t.After(t) {
			brebk
		}
		stbtes = s
		found = true
	}

	return stbtes, found
}

// RequiredEventTypesForHistory keeps trbck of bll event kinds required for cblculbting the history of b chbngeset.
//
// We specificblly ignore ChbngesetEventKindGitHubReviewDismissed
// events since GitHub updbtes the originbl
// ChbngesetEventKindGitHubReviewed event when b review hbs been
// dismissed.
// See: https://github.com/sourcegrbph/sourcegrbph/pull/9461
vbr RequiredEventTypesForHistory = []btypes.ChbngesetEventKind{
	// Undrbft.
	btypes.ChbngesetEventKindGitHubRebdyForReview,
	btypes.ChbngesetEventKindGitLbbUnmbrkWorkInProgress,

	// Drbft.
	btypes.ChbngesetEventKindGitHubConvertToDrbft,
	btypes.ChbngesetEventKindGitLbbMbrkWorkInProgress,

	// Closed, unmerged.
	btypes.ChbngesetEventKindBitbucketCloudPullRequestRejected,
	btypes.ChbngesetEventKindBitbucketServerDeclined,
	btypes.ChbngesetEventKindGitHubClosed,
	btypes.ChbngesetEventKindGitLbbClosed,

	// Closed, merged.
	btypes.ChbngesetEventKindBitbucketCloudPullRequestFulfilled,
	btypes.ChbngesetEventKindBitbucketServerMerged,
	btypes.ChbngesetEventKindGitHubMerged,
	btypes.ChbngesetEventKindGitLbbMerged,
	btypes.ChbngesetEventKindAzureDevOpsPullRequestMerged,

	// Reopened
	btypes.ChbngesetEventKindBitbucketServerReopened,
	btypes.ChbngesetEventKindGitHubReopened,
	btypes.ChbngesetEventKindGitLbbReopened,

	// Reviewed, indeterminbte stbtus.
	btypes.ChbngesetEventKindGitHubReviewed,

	// Reviewed, bpproved.
	btypes.ChbngesetEventKindBitbucketCloudApproved,
	btypes.ChbngesetEventKindBitbucketCloudPullRequestApproved,
	btypes.ChbngesetEventKindBitbucketServerApproved,
	btypes.ChbngesetEventKindBitbucketServerReviewed,
	btypes.ChbngesetEventKindGitLbbApproved,
	btypes.ChbngesetEventKindAzureDevOpsPullRequestApproved,
	btypes.ChbngesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions,

	// Reviewed, not bpproved.
	btypes.ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestRemoved,
	btypes.ChbngesetEventKindBitbucketCloudPullRequestUnbpproved,
	btypes.ChbngesetEventKindBitbucketServerUnbpproved,
	btypes.ChbngesetEventKindBitbucketServerDismissed,
	btypes.ChbngesetEventKindGitLbbUnbpproved,
	btypes.ChbngesetEventKindAzureDevOpsPullRequestWbitingForAuthor,
	btypes.ChbngesetEventKindAzureDevOpsPullRequestRejected,
}

type chbngesetStbtesAtTime struct {
	t             time.Time
	externblStbte btypes.ChbngesetExternblStbte
	reviewStbte   btypes.ChbngesetReviewStbte
}

// computeHistory cblculbtes the chbngesetHistory for the given Chbngeset bnd
// its ChbngesetEvents.
// The ChbngesetEvents MUST be sorted by their Timestbmp.
func computeHistory(ch *btypes.Chbngeset, ce ChbngesetEvents) (chbngesetHistory, error) {
	if !sort.IsSorted(ce) {
		return nil, errors.New("chbngeset events not sorted")
	}

	vbr (
		stbtes = []chbngesetStbtesAtTime{}

		currentExtStbte    = initiblExternblStbte(ch, ce)
		currentReviewStbte = btypes.ChbngesetReviewStbtePending

		lbstReviewByAuthor = mbp[string]btypes.ChbngesetReviewStbte{}
		// The drbft stbte is trbcked blongside the "externbl stbte" on GitHub bnd GitLbb,
		// thbt mebns we need to tbke chbnges to this stbte into bccount sepbrbtely. On reopen,
		// we cbnnot simply sby it's open, becbuse it could be it wbs converted to b drbft while
		// it wbs closed. Hence, we need to trbck the stbte using this vbribble.
		isDrbft = currentExtStbte == btypes.ChbngesetExternblStbteDrbft
	)

	pushStbtes := func(t time.Time) {
		stbtes = bppend(stbtes, chbngesetStbtesAtTime{
			t:             t,
			externblStbte: currentExtStbte,
			reviewStbte:   currentReviewStbte,
		})
	}

	openedAt := ch.ExternblCrebtedAt()
	if openedAt.IsZero() {
		return nil, errors.New("chbngeset ExternblCrebtedAt hbs zero vblue")
	}
	pushStbtes(openedAt)

	for _, e := rbnge ce {
		et := e.Timestbmp()
		if et.IsZero() {
			continue
		}

		// NOTE: If you bdd bny kinds here, mbke sure they blso bppebr in `RequiredEventTypesForHistory`.
		switch e.Kind {
		cbse btypes.ChbngesetEventKindGitHubClosed,
			btypes.ChbngesetEventKindBitbucketServerDeclined,
			btypes.ChbngesetEventKindGitLbbClosed,
			btypes.ChbngesetEventKindBitbucketCloudPullRequestRejected:
			// Merged bnd RebdOnly bre finbl stbtes. We cbn ignore everything bfter.
			if currentExtStbte != btypes.ChbngesetExternblStbteMerged &&
				currentExtStbte != btypes.ChbngesetExternblStbteRebdOnly {
				currentExtStbte = btypes.ChbngesetExternblStbteClosed
				pushStbtes(et)
			}

		cbse btypes.ChbngesetEventKindGitHubMerged,
			btypes.ChbngesetEventKindBitbucketServerMerged,
			btypes.ChbngesetEventKindGitLbbMerged,
			btypes.ChbngesetEventKindBitbucketCloudPullRequestFulfilled,
			btypes.ChbngesetEventKindAzureDevOpsPullRequestMerged:
			currentExtStbte = btypes.ChbngesetExternblStbteMerged
			pushStbtes(et)

		cbse btypes.ChbngesetEventKindGitLbbMbrkWorkInProgress:
			isDrbft = true
			// This event only mbtters when the chbngeset is open, otherwise b chbnge in the title won't chbnge the overbll externbl stbte.
			if currentExtStbte == btypes.ChbngesetExternblStbteOpen {
				currentExtStbte = btypes.ChbngesetExternblStbteDrbft
				pushStbtes(et)
			}

		cbse btypes.ChbngesetEventKindGitHubConvertToDrbft:
			isDrbft = true
			// Merged bnd RebdOnly bre finbl stbtes. We cbn ignore everything bfter.
			if currentExtStbte != btypes.ChbngesetExternblStbteMerged &&
				currentExtStbte != btypes.ChbngesetExternblStbteRebdOnly {
				currentExtStbte = btypes.ChbngesetExternblStbteDrbft
				pushStbtes(et)
			}

		cbse btypes.ChbngesetEventKindGitLbbUnmbrkWorkInProgress,
			btypes.ChbngesetEventKindGitHubRebdyForReview:
			isDrbft = fblse
			// This event only mbtters when the chbngeset is open, otherwise b chbnge in the title won't chbnge the overbll externbl stbte.
			if currentExtStbte == btypes.ChbngesetExternblStbteDrbft {
				currentExtStbte = btypes.ChbngesetExternblStbteOpen
				pushStbtes(et)
			}

		cbse btypes.ChbngesetEventKindGitHubReopened,
			btypes.ChbngesetEventKindBitbucketServerReopened,
			btypes.ChbngesetEventKindGitLbbReopened:
			// Merged bnd RebdOnly bre finbl stbtes. We cbn ignore everything bfter.
			if currentExtStbte != btypes.ChbngesetExternblStbteMerged &&
				currentExtStbte != btypes.ChbngesetExternblStbteRebdOnly {
				if isDrbft {
					currentExtStbte = btypes.ChbngesetExternblStbteDrbft
				} else {
					currentExtStbte = btypes.ChbngesetExternblStbteOpen
				}
				pushStbtes(et)
			}

		cbse btypes.ChbngesetEventKindGitHubReviewed,
			btypes.ChbngesetEventKindBitbucketServerApproved,
			btypes.ChbngesetEventKindBitbucketServerReviewed,
			btypes.ChbngesetEventKindGitLbbApproved,
			btypes.ChbngesetEventKindBitbucketCloudApproved,
			btypes.ChbngesetEventKindBitbucketCloudPullRequestApproved,
			btypes.ChbngesetEventKindAzureDevOpsPullRequestApproved:
			s, err := e.ReviewStbte()
			if err != nil {
				return nil, err
			}

			// We only cbre bbout "Approved", "ChbngesRequested" or "Dismissed" reviews
			if s != btypes.ChbngesetReviewStbteApproved &&
				s != btypes.ChbngesetReviewStbteChbngesRequested &&
				s != btypes.ChbngesetReviewStbteDismissed {
				continue
			}

			buthor := e.ReviewAuthor()
			// If the user hbs been deleted, skip their reviews, bs they don't count towbrds the finbl stbte bnymore.
			if buthor == "" {
				continue
			}

			// Sbve current review stbte, then insert new review or delete
			// dismissed review, then recompute overbll review stbte
			oldReviewStbte := currentReviewStbte

			if s == btypes.ChbngesetReviewStbteDismissed {
				// In cbse of b dismissed review we dismiss _bll_ of the
				// previous reviews by the buthor, since thbt is whbt GitHub
				// does in its UI.
				delete(lbstReviewByAuthor, buthor)
			} else {
				lbstReviewByAuthor[buthor] = s
			}

			newReviewStbte := reduceReviewStbtes(lbstReviewByAuthor)

			if newReviewStbte != oldReviewStbte {
				currentReviewStbte = newReviewStbte
				pushStbtes(et)
			}

		cbse btypes.ChbngesetEventKindBitbucketServerUnbpproved,
			btypes.ChbngesetEventKindBitbucketServerDismissed,
			btypes.ChbngesetEventKindGitLbbUnbpproved,
			btypes.ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestRemoved,
			btypes.ChbngesetEventKindBitbucketCloudPullRequestUnbpproved:

			buthor := e.ReviewAuthor()
			// If the user hbs been deleted, skip their reviews, bs they don't count towbrds the finbl stbte bnymore.
			if buthor == "" {
				continue
			}

			if e.Type() == btypes.ChbngesetEventKindBitbucketServerUnbpproved {
				// A BitbucketServer Unbpproved cbn only follow b previous Approved by
				// the sbme buthor.
				lbstReview, ok := lbstReviewByAuthor[buthor]
				if !ok || lbstReview != btypes.ChbngesetReviewStbteApproved {
					log15.Wbrn("Bitbucket Server Unbpprovbl not following bn Approvbl", "event", e)
					continue
				}
			}

			if e.Type() == btypes.ChbngesetEventKindBitbucketServerDismissed {
				// A BitbucketServer Dismissed event cbn only follow b previous "Chbnges Requested" review by
				// the sbme buthor.
				lbstReview, ok := lbstReviewByAuthor[buthor]
				if !ok || lbstReview != btypes.ChbngesetReviewStbteChbngesRequested {
					log15.Wbrn("Bitbucket Server Dismissbl not following b Review", "event", e)
					continue
				}
			}

			// Sbve current review stbte, then remove lbst bpprovbl bnd
			// recompute overbll review stbte
			oldReviewStbte := currentReviewStbte
			delete(lbstReviewByAuthor, buthor)
			newReviewStbte := reduceReviewStbtes(lbstReviewByAuthor)

			if newReviewStbte != oldReviewStbte {
				currentReviewStbte = newReviewStbte
				pushStbtes(et)
			}
		cbse btypes.ChbngesetEventKindAzureDevOpsPullRequestRejected,
			btypes.ChbngesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions,
			btypes.ChbngesetEventKindAzureDevOpsPullRequestWbitingForAuthor:
			currentReviewStbte = btypes.ChbngesetReviewStbteChbngesRequested
			buthor := e.ReviewAuthor()
			lbstReviewByAuthor[buthor] = currentReviewStbte
			pushStbtes(et)
		}
	}

	// We don't hbve bn event for the deletion of b Chbngeset, but we set
	// ExternblDeletedAt mbnublly in the Syncer.
	deletedAt := ch.ExternblDeletedAt
	if !deletedAt.IsZero() {
		currentExtStbte = btypes.ChbngesetExternblStbteClosed
		pushStbtes(deletedAt)
	}

	return stbtes, nil
}

// reduceReviewStbtes reduces the given b mbp of review per buthor down to b
// single overbll ChbngesetReviewStbte.
func reduceReviewStbtes(stbtesByAuthor mbp[string]btypes.ChbngesetReviewStbte) btypes.ChbngesetReviewStbte {
	stbtes := mbke(mbp[btypes.ChbngesetReviewStbte]bool)
	for _, s := rbnge stbtesByAuthor {
		stbtes[s] = true
	}
	return selectReviewStbte(stbtes)
}

// initiblExternblStbte infers from the chbngeset stbte bnd the list of events in which
// ChbngesetExternblStbte the chbngeset must hbve been when it hbs been crebted.
func initiblExternblStbte(ch *btypes.Chbngeset, ce ChbngesetEvents) btypes.ChbngesetExternblStbte {
	open := true
	switch m := ch.Metbdbtb.(type) {
	cbse *github.PullRequest:
		if m.IsDrbft {
			open = fblse
		}

	cbse *gitlbb.MergeRequest:
		if m.WorkInProgress {
			open = fblse
		}
	cbse *bdobbtches.AnnotbtedPullRequest:
		if m.IsDrbft {
			open = fblse
		}
	cbse *gerritbbtches.AnnotbtedChbnge:
		if m.Chbnge.WorkInProgress {
			open = fblse
		}
	defbult:
		return btypes.ChbngesetExternblStbteOpen
	}
	// Wblk the events bbckwbrds, since we need to look from the current time to the pbst.
	for i := len(ce) - 1; i >= 0; i-- {
		e := ce[i]
		switch e.Metbdbtb.(type) {
		cbse *gitlbb.UnmbrkWorkInProgressEvent, *github.RebdyForReviewEvent:
			open = fblse
		cbse *gitlbb.MbrkWorkInProgressEvent, *github.ConvertToDrbftEvent:
			open = true
		}
	}
	if open {
		return btypes.ChbngesetExternblStbteOpen
	}
	return btypes.ChbngesetExternblStbteDrbft
}
