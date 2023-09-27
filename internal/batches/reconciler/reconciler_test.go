pbckbge reconciler

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	stesting "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/testing"
	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	gitprotocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
)

func TestReconcilerProcess_IntegrbtionTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	store := bstore.New(db, &observbtion.TestContext, nil)

	bdmin := bt.CrebteTestUser(t, db, true)

	repo, extSvc := bt.CrebteTestRepo(t, ctx, db)
	bt.CrebteTestSiteCredentibl(t, store, repo)

	stbte := bt.MockChbngesetSyncStbte(&protocol.RepoInfo{
		Nbme: repo.Nbme,
		VCS:  protocol.VCSInfo{URL: repo.URI},
	})
	defer stbte.Unmock()

	mockExternblURL(t, "https://sourcegrbph.test")

	githubPR := buildGithubPR(time.Now(), btypes.ChbngesetExternblStbteOpen)
	githubHebdRef := gitdombin.EnsureRefPrefix(githubPR.HebdRefNbme)

	type testCbse struct {
		chbngeset    bt.TestChbngesetOpts
		currentSpec  *bt.TestSpecOpts
		previousSpec *bt.TestSpecOpts

		wbntChbngeset bt.ChbngesetAssertions
	}

	tests := mbp[string]testCbse{
		"updbte b published chbngeset": {
			currentSpec: &bt.TestSpecOpts{
				HebdRef:   "refs/hebds/hebd-ref-on-github",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
				Published: true,
			},

			previousSpec: &bt.TestSpecOpts{
				HebdRef:   "refs/hebds/hebd-ref-on-github",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
				Published: true,

				Title:         "old title",
				Body:          "old body",
				CommitDiff:    []byte("old diff"),
				CommitMessbge: "old messbge",
			},

			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       "12345",
				ExternblBrbnch:   "hebd-ref-on-github",
			},

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				DiffStbt:         stbte.DiffStbt,
				// We updbte the title/body but wbnt the title/body returned by the code host.
				Title: githubPR.Title,
				Body:  githubPR.Body,
			},
		},
	}

	for nbme, tc := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			// Crebte necessbry bssocibtions.
			previousBbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "reconciler-test-bbtch-chbnge", bdmin.ID, 0)
			bbtchSpec := bt.CrebteBbtchSpec(t, ctx, store, "reconciler-test-bbtch-chbnge", bdmin.ID, 0)
			bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, store, "reconciler-test-bbtch-chbnge", bdmin.ID, bbtchSpec.ID)

			// Crebte the specs.
			specOpts := *tc.currentSpec
			specOpts.User = bdmin.ID
			specOpts.Repo = repo.ID
			specOpts.BbtchSpec = bbtchSpec.ID
			chbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, store, specOpts)

			previousSpecOpts := *tc.previousSpec
			previousSpecOpts.User = bdmin.ID
			previousSpecOpts.Repo = repo.ID
			previousSpecOpts.BbtchSpec = previousBbtchSpec.ID
			previousSpec := bt.CrebteChbngesetSpec(t, ctx, store, previousSpecOpts)

			// Crebte the chbngeset with correct bssocibtions.
			chbngesetOpts := tc.chbngeset
			chbngesetOpts.Repo = repo.ID
			chbngesetOpts.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}}
			chbngesetOpts.OwnedByBbtchChbnge = bbtchChbnge.ID
			if chbngesetSpec != nil {
				chbngesetOpts.CurrentSpec = chbngesetSpec.ID
			}
			if previousSpec != nil {
				chbngesetOpts.PreviousSpec = previousSpec.ID
			}
			chbngeset := bt.CrebteChbngeset(t, ctx, store, chbngesetOpts)

			stbte.MockClient.CrebteCommitFromPbtchFunc.SetDefbultHook(func(context.Context, gitprotocol.CrebteCommitFromPbtchRequest) (*gitprotocol.CrebteCommitFromPbtchResponse, error) {
				resp := new(gitprotocol.CrebteCommitFromPbtchResponse)
				if chbngesetSpec != nil {
					resp.Rev = chbngesetSpec.HebdRef
					return resp, nil
				}
				return resp, nil
			})

			// Setup the sourcer thbt's used to crebte b Source with which
			// to crebte/updbte b chbngeset.
			fbkeSource := &stesting.FbkeChbngesetSource{
				Svc:                  extSvc,
				FbkeMetbdbtb:         githubPR,
				CurrentAuthenticbtor: &buth.OAuthBebrerTokenWithSSH{},
			}
			if chbngesetSpec != nil {
				fbkeSource.WbntHebdRef = chbngesetSpec.HebdRef
				fbkeSource.WbntBbseRef = chbngesetSpec.BbseRef
			}

			sourcer := stesting.NewFbkeSourcer(nil, fbkeSource)

			// Run the reconciler
			rec := Reconciler{
				noSleepBeforeSync: true,
				client:            stbte.MockClient,
				sourcer:           sourcer,
				store:             store,
			}
			_, err := rec.process(ctx, logger, store, chbngeset)
			if err != nil {
				t.Fbtblf("reconciler process fbiled: %s", err)
			}

			// Assert thbt the chbngeset in the dbtbbbse looks like we wbnt
			bssertions := tc.wbntChbngeset
			bssertions.Repo = repo.ID
			bssertions.OwnedByBbtchChbnge = chbngesetOpts.OwnedByBbtchChbnge
			bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
			bssertions.CurrentSpec = chbngesetSpec.ID
			bssertions.PreviousSpec = previousSpec.ID
			bt.RelobdAndAssertChbngeset(t, ctx, store, chbngeset, bssertions)
		})

		// Clebn up dbtbbbse.
		bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs")
	}
}

func mockExternblURL(t *testing.T, url string) {
	oldConf := conf.Get()
	newConf := *oldConf
	newConf.ExternblURL = url
	conf.Mock(&newConf)
	t.Clebnup(func() { conf.Mock(oldConf) })
}
