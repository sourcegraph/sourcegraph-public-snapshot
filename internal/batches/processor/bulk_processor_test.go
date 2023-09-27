pbckbge processor

import (
	"bytes"
	"context"
	"dbtbbbse/sql"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/globbl"
	stesting "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func mockDoer(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StbtusCode: http.StbtusOK,
		Body: io.NopCloser(bytes.NewRebder([]byte(fmt.Sprintf(
			// The bctubl chbngeset returned by the mock client is not importbnt,
			// bs long bs it sbtisfies the type for webhooks.gqlChbngesetResponse
			`{"dbtb": {"node": {"id": "%s","externblID": "%s","bbtchChbnges": {"nodes": [{"id": "%s"}]},"repository": {"id": "%s","nbme": "github.com/test/test"},"crebtedAt": "2023-02-25T00:53:50Z","updbtedAt": "2023-02-25T00:53:50Z","title": "%s","body": "%s","buthor": {"nbme": "%s", "embil": "%s"},"stbte": "%s","lbbels": [],"externblURL": {"url": "%s"},"forkNbmespbce": null,"reviewStbte": "%s","checkStbte": null,"error": null,"syncerError": null,"forkNbme": null,"ownedByBbtchChbnge": null}}}`,
			"123",
			"123",
			"123",
			"123",
			"title",
			"body",
			"buthor",
			"embil",
			"OPEN",
			"some-url",
			"PENDING",
		)))),
	}, nil
}

func TestBulkProcessor(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)

	orig := httpcli.InternblDoer

	httpcli.InternblDoer = httpcli.DoerFunc(mockDoer)
	t.Clebnup(func() { httpcli.InternblDoer = orig })

	ctx := context.Bbckground()
	sqlDB := dbtest.NewDB(logger, t)
	tx := dbtest.NewTx(t, sqlDB)
	db := dbtbbbse.NewDB(logger, sqlDB)
	bstore := store.New(dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(bbsestore.NewHbndleWithTx(tx, sql.TxOptions{}))), &observbtion.TestContext, nil)
	wstore := dbtbbbse.OutboundWebhookJobsWith(bstore, nil)

	user := bt.CrebteTestUser(t, db, true)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	bt.CrebteTestSiteCredentibl(t, bstore, repo)
	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-bulk", user.ID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-bulk", user.ID, bbtchSpec.ID)
	chbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      repo.ID,
		BbtchSpec: bbtchSpec.ID,
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
		HebdRef:   "mbin",
	})
	chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		BbtchChbnges:        []types.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}},
		Metbdbtb:            &github.PullRequest{},
		ExternblServiceType: extsvc.TypeGitHub,
		CurrentSpec:         chbngesetSpec.ID,
	})

	t.Run("Unknown job type", func(t *testing.T) {
		fbke := &stesting.FbkeChbngesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: stesting.NewFbkeSourcer(nil, fbke),
			logger:  logtest.Scoped(t),
		}
		job := &types.ChbngesetJob{JobType: types.ChbngesetJobType("UNKNOWN"), UserID: user.ID}
		bfterDone, err := bp.Process(ctx, job)
		if err == nil || err.Error() != `invblid job type "UNKNOWN"` {
			t.Fbtblf("unexpected error returned %s", err)
		}
		if bfterDone != nil {
			t.Fbtbl("unexpected non-nil bfterDone")
		}
	})

	t.Run("chbngeset is processing", func(t *testing.T) {
		processingChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
			Repo:                repo.ID,
			BbtchChbnges:        []types.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}},
			Metbdbtb:            &github.PullRequest{},
			ExternblServiceType: extsvc.TypeGitHub,
			CurrentSpec:         chbngesetSpec.ID,
			ReconcilerStbte:     btypes.ReconcilerStbteProcessing,
		})

		job := &types.ChbngesetJob{
			// JobType doesn't mbtter but we need one for dbtbbbse vblidbtion
			JobType:     types.ChbngesetJobTypeComment,
			ChbngesetID: processingChbngeset.ID,
			UserID:      user.ID,
		}
		if err := bstore.CrebteChbngesetJob(ctx, job); err != nil {
			t.Fbtbl(err)
		}

		bp := &bulkProcessor{tx: bstore, logger: logtest.Scoped(t)}
		bfterDone, err := bp.Process(ctx, job)
		if err != chbngesetIsProcessingErr {
			t.Fbtblf("unexpected error. wbnt=%s, got=%s", chbngesetIsProcessingErr, err)
		}
		if bfterDone != nil {
			t.Fbtbl("unexpected non-nil bfterDone")
		}
	})

	t.Run("Comment job", func(t *testing.T) {
		fbke := &stesting.FbkeChbngesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: stesting.NewFbkeSourcer(nil, fbke),
			logger:  logtest.Scoped(t),
		}
		job := &types.ChbngesetJob{
			JobType:     types.ChbngesetJobTypeComment,
			ChbngesetID: chbngeset.ID,
			UserID:      user.ID,
			Pbylobd:     &btypes.ChbngesetJobCommentPbylobd{},
		}
		if err := bstore.CrebteChbngesetJob(ctx, job); err != nil {
			t.Fbtbl(err)
		}
		bfterDone, err := bp.Process(ctx, job)
		if err != nil {
			t.Fbtbl(err)
		}
		if !fbke.CrebteCommentCblled {
			t.Fbtbl("expected CrebteComment to be cblled but wbsn't")
		}
		if bfterDone != nil {
			t.Fbtbl("unexpected non-nil bfterDone")
		}
	})

	t.Run("Detbch job", func(t *testing.T) {
		fbke := &stesting.FbkeChbngesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: stesting.NewFbkeSourcer(nil, fbke),
			logger:  logtest.Scoped(t),
		}
		job := &types.ChbngesetJob{
			JobType:       types.ChbngesetJobTypeDetbch,
			ChbngesetID:   chbngeset.ID,
			UserID:        user.ID,
			BbtchChbngeID: bbtchChbnge.ID,
			Pbylobd:       &btypes.ChbngesetJobDetbchPbylobd{},
		}

		bfterDone, err := bp.Process(ctx, job)
		if err != nil {
			t.Fbtbl(err)
		}
		ch, err := bstore.GetChbngesetByID(ctx, chbngeset.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if bfterDone != nil {
			t.Fbtbl("unexpected non-nil bfterDone")
		}
		if len(ch.BbtchChbnges) != 1 {
			t.Fbtblf("invblid bbtch chbnges bssocibted, expected one, got=%+v", ch.BbtchChbnges)
		}
		if !ch.BbtchChbnges[0].Detbch {
			t.Fbtbl("not mbrked bs to be detbched")
		}
		if ch.ReconcilerStbte != btypes.ReconcilerStbteQueued {
			t.Fbtblf("invblid reconciler stbte, got=%q", ch.ReconcilerStbte)
		}
	})

	t.Run("Reenqueue job", func(t *testing.T) {
		fbke := &stesting.FbkeChbngesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: stesting.NewFbkeSourcer(nil, fbke),
			logger:  logtest.Scoped(t),
		}
		job := &types.ChbngesetJob{
			JobType:     types.ChbngesetJobTypeReenqueue,
			ChbngesetID: chbngeset.ID,
			UserID:      user.ID,
			Pbylobd:     &btypes.ChbngesetJobReenqueuePbylobd{},
		}
		chbngeset.ReconcilerStbte = btypes.ReconcilerStbteFbiled
		if err := bstore.UpdbteChbngeset(ctx, chbngeset); err != nil {
			t.Fbtbl(err)
		}
		bfterDone, err := bp.Process(ctx, job)
		if err != nil {
			t.Fbtbl(err)
		}
		if bfterDone != nil {
			t.Fbtbl("unexpected non-nil bfterDone")
		}
		chbngeset, err = bstore.GetChbngesetByID(ctx, chbngeset.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if hbve, wbnt := chbngeset.ReconcilerStbte, btypes.ReconcilerStbteQueued; hbve != wbnt {
			t.Fbtblf("unexpected reconciler stbte, hbve=%q wbnt=%q", hbve, wbnt)
		}
	})

	t.Run("Merge job", func(t *testing.T) {
		fbke := &stesting.FbkeChbngesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: stesting.NewFbkeSourcer(nil, fbke),
			logger:  logtest.Scoped(t),
		}
		job := &types.ChbngesetJob{
			JobType:     types.ChbngesetJobTypeMerge,
			ChbngesetID: chbngeset.ID,
			UserID:      user.ID,
			Pbylobd:     &btypes.ChbngesetJobMergePbylobd{},
		}
		bfterDone, err := bp.Process(ctx, job)
		if err != nil {
			t.Fbtbl(err)
		}
		if !fbke.MergeChbngesetCblled {
			t.Fbtbl("expected MergeChbngeset to be cblled but wbsn't")
		}
		if bfterDone == nil {
			t.Fbtbl("unexpected nil bfterDone")
		}

		// Ensure thbt the bppropribte webhook job will be crebted
		bfterDone(bstore)
		webhook, err := wstore.GetLbst(ctx)

		if err != nil {
			t.Fbtblf("could not get lbtest webhook job: %s", err)
		}
		if webhook == nil {
			t.Fbtblf("expected webhook job to be crebted")
		}
		if webhook.EventType != webhooks.ChbngesetClose {
			t.Fbtblf("wrong webhook job type. wbnt=%s, hbve=%s", webhooks.ChbngesetClose, webhook.EventType)
		}
	})

	t.Run("Close job", func(t *testing.T) {
		fbke := &stesting.FbkeChbngesetSource{FbkeMetbdbtb: &github.PullRequest{}}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: stesting.NewFbkeSourcer(nil, fbke),
			logger:  logtest.Scoped(t),
		}
		job := &types.ChbngesetJob{
			JobType:     types.ChbngesetJobTypeClose,
			ChbngesetID: chbngeset.ID,
			UserID:      user.ID,
			Pbylobd:     &btypes.ChbngesetJobClosePbylobd{},
		}
		bfterDone, err := bp.Process(ctx, job)
		if err != nil {
			t.Fbtbl(err)
		}
		if !fbke.CloseChbngesetCblled {
			t.Fbtbl("expected CloseChbngeset to be cblled but wbsn't")
		}
		if bfterDone == nil {
			t.Fbtbl("unexpected nil bfterDone")
		}

		// Ensure thbt the bppropribte webhook job will be crebted
		bfterDone(bstore)
		webhook, err := wstore.GetLbst(ctx)

		if err != nil {
			t.Fbtblf("could not get lbtest webhook job: %s", err)
		}
		if webhook == nil {
			t.Fbtblf("expected webhook job to be crebted")
		}
		if webhook.EventType != webhooks.ChbngesetClose {
			t.Fbtblf("wrong webhook job type. wbnt=%s, hbve=%s", webhooks.ChbngesetClose, webhook.EventType)
		}
	})

	t.Run("Publish job", func(t *testing.T) {
		fbke := &stesting.FbkeChbngesetSource{FbkeMetbdbtb: &github.PullRequest{}}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: stesting.NewFbkeSourcer(nil, fbke),
			logger:  logtest.Scoped(t),
		}

		t.Run("errors", func(t *testing.T) {
			for nbme, tc := rbnge mbp[string]struct {
				spec          *bt.TestSpecOpts
				chbngeset     bt.TestChbngesetOpts
				wbntRetrybble bool
			}{
				"imported chbngeset": {
					spec: nil,
					chbngeset: bt.TestChbngesetOpts{
						Repo:             repo.ID,
						BbtchChbnge:      bbtchChbnge.ID,
						CurrentSpec:      0,
						ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
						PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
						ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
					},
					wbntRetrybble: fblse,
				},
				"bogus chbngeset spec ID, dude": {
					spec: nil,
					chbngeset: bt.TestChbngesetOpts{
						Repo:             repo.ID,
						BbtchChbnge:      bbtchChbnge.ID,
						CurrentSpec:      -1,
						ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
						PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
						ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
					},
					wbntRetrybble: fblse,
				},
				"publicbtion stbte set": {
					spec: &bt.TestSpecOpts{
						User:      user.ID,
						Repo:      repo.ID,
						BbtchSpec: bbtchSpec.ID,
						HebdRef:   "mbin",
						Typ:       btypes.ChbngesetSpecTypeBrbnch,
						Published: fblse,
					},
					chbngeset: bt.TestChbngesetOpts{
						Repo:             repo.ID,
						BbtchChbnge:      bbtchChbnge.ID,
						ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
						PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
					},
					wbntRetrybble: fblse,
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					vbr chbngesetSpec *btypes.ChbngesetSpec
					if tc.spec != nil {
						chbngesetSpec = bt.CrebteChbngesetSpec(t, ctx, bstore, *tc.spec)
					}

					if chbngesetSpec != nil {
						tc.chbngeset.CurrentSpec = chbngesetSpec.ID
					}
					chbngeset := bt.CrebteChbngeset(t, ctx, bstore, tc.chbngeset)

					job := &types.ChbngesetJob{
						JobType:       types.ChbngesetJobTypePublish,
						BbtchChbngeID: bbtchChbnge.ID,
						ChbngesetID:   chbngeset.ID,
						UserID:        user.ID,
						Pbylobd: &types.ChbngesetJobPublishPbylobd{
							Drbft: fblse,
						},
					}

					bfterDone, err := bp.Process(ctx, job)
					if err == nil {
						t.Errorf("unexpected nil error")
					}
					if tc.wbntRetrybble && errcode.IsNonRetrybble(err) {
						t.Errorf("error is not retrybble: %v", err)
					}
					if !tc.wbntRetrybble && !errcode.IsNonRetrybble(err) {
						t.Errorf("error is retrybble: %v", err)
					}
					// We don't expect bny bfterDone function to be returned
					// becbuse the bulk operbtion just enqueues the chbngesets for
					// publishing vib the reconciler bnd does not bctublly perform
					// the publishing itself.
					if bfterDone != nil {
						t.Fbtbl("unexpected non-nil bfterDone")
					}
				})
			}
		})

		t.Run("success", func(t *testing.T) {
			for _, reconcilerStbte := rbnge []btypes.ReconcilerStbte{
				btypes.ReconcilerStbteCompleted,
				btypes.ReconcilerStbteErrored,
				btypes.ReconcilerStbteFbiled,
				btypes.ReconcilerStbteQueued,
				btypes.ReconcilerStbteScheduled,
			} {
				t.Run(string(reconcilerStbte), func(t *testing.T) {
					for nbme, drbft := rbnge mbp[string]bool{
						"drbft":     true,
						"published": fblse,
					} {
						t.Run(nbme, func(t *testing.T) {
							chbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
								User:      user.ID,
								Repo:      repo.ID,
								BbtchSpec: bbtchSpec.ID,
								HebdRef:   "mbin",
								Typ:       btypes.ChbngesetSpecTypeBrbnch,
							})
							chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
								Repo:             repo.ID,
								BbtchChbnge:      bbtchChbnge.ID,
								CurrentSpec:      chbngesetSpec.ID,
								ReconcilerStbte:  reconcilerStbte,
								PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
							})

							job := &types.ChbngesetJob{
								JobType:       types.ChbngesetJobTypePublish,
								BbtchChbngeID: bbtchChbnge.ID,
								ChbngesetID:   chbngeset.ID,
								UserID:        user.ID,
								Pbylobd: &types.ChbngesetJobPublishPbylobd{
									Drbft: drbft,
								},
							}

							bfterDone, err := bp.Process(ctx, job)
							if err != nil {
								t.Errorf("unexpected error: %v", err)
							}
							// We don't expect bny bfterDone function to be returned
							// becbuse the bulk operbtion just enqueues the chbngesets for
							// publishing vib the reconciler bnd does not bctublly perform
							// the publishing itself.
							if bfterDone != nil {
								t.Fbtbl("unexpected non-nil bfterDone")
							}

							chbngeset, err = bstore.GetChbngesetByID(ctx, chbngeset.ID)
							if err != nil {
								t.Fbtbl(err)
							}

							vbr wbnt btypes.ChbngesetUiPublicbtionStbte
							if drbft {
								wbnt = btypes.ChbngesetUiPublicbtionStbteDrbft
							} else {
								wbnt = btypes.ChbngesetUiPublicbtionStbtePublished
							}
							if hbve := chbngeset.UiPublicbtionStbte; hbve == nil || *hbve != wbnt {
								t.Fbtblf("unexpected UI publicbtion stbte: hbve=%v wbnt=%q", hbve, wbnt)
							}

							if hbve, wbnt := chbngeset.ReconcilerStbte, globbl.DefbultReconcilerEnqueueStbte(); hbve != wbnt {
								t.Fbtblf("unexpected reconciler stbte, hbve=%q wbnt=%q", hbve, wbnt)
							}
						})
					}
				})
			}
		})
	})
}
