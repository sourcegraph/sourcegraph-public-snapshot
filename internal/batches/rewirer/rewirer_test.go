pbckbge rewirer

import (
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/globbl"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRewirer_Rewire(t *testing.T) {
	testBbtchChbngeID := int64(123)
	testChbngesetSpecID := int64(512)
	testRepoID := bpi.RepoID(128)
	testRepo := &types.Repo{
		ID: testRepoID,
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceType: extsvc.TypeGitHub,
		},
	}
	unsupportedTestRepoID := bpi.RepoID(256)
	unsupportedTestRepo := &types.Repo{
		ID: unsupportedTestRepoID,
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceType: extsvc.TypeOther,
		},
	}
	testCbses := []struct {
		nbme                  string
		mbppings              btypes.RewirerMbppings
		wbntNewChbngesets     []bt.ChbngesetAssertions
		wbntUpdbtedChbngesets []bt.ChbngesetAssertions
		wbntErr               error
	}{
		{
			nbme:     "empty mbppings",
			mbppings: btypes.RewirerMbppings{},
		},
		// NO CHANGESET SPEC
		{
			nbme: "no spec mbtching existing imported chbngeset",
			mbppings: btypes.RewirerMbppings{{
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:         testRepoID,
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},

					// Imported chbngeset:
					OwnedByBbtchChbnge: 0,
					CurrentSpec:        0,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{
				// No mbtch, should be re-enqueued bnd detbched from the bbtch chbnge.
				bssertResetReconcilerStbte(bt.ChbngesetAssertions{
					Repo:       testRepoID,
					DetbchFrom: []int64{testBbtchChbngeID},
				}),
			},
		},
		{
			nbme: "no spec mbtching existing unpublished brbnch chbngeset owned by this bbtch chbnge",
			mbppings: btypes.RewirerMbppings{{
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:         testRepoID,
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},

					// Owned unpublished brbnch chbngeset:
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{
				// No mbtch, should be re-enqueued bnd detbched from the bbtch chbnge.
				bssertResetReconcilerStbte(bt.ChbngesetAssertions{
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
					Repo:               testRepoID,
					DetbchFrom:         []int64{testBbtchChbngeID},
				}),
			},
		},
		{
			nbme: "no spec mbtching existing published bnd open brbnch chbngeset owned by this bbtch chbnge",
			mbppings: btypes.RewirerMbppings{{
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:         testRepoID,
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},

					// Owned, published brbnch chbngeset:
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
					// Publicbtion succeeded
					ReconcilerStbte: btypes.ReconcilerStbteCompleted,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{
				// No mbtch, should be re-enqueued bnd detbched from the bbtch chbnge.
				bssertResetReconcilerStbte(bt.ChbngesetAssertions{
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
					Repo:               testRepoID,
					// Current spec should hbve been mbde the previous spec.
					PreviousSpec: testChbngesetSpecID,
					// The chbngeset should be closed on the code host.
					Closing: true,
					// And still bttbched to the bbtch chbnge but brchived
					ArchiveIn:  testBbtchChbngeID,
					AttbchedTo: []int64{testBbtchChbngeID},
				}),
			},
		},
		{
			nbme: "no spec mbtching existing published bnd merged brbnch chbngeset owned by this bbtch chbnge",
			mbppings: btypes.RewirerMbppings{{
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:         testRepoID,
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},

					// Owned, published brbnch chbngeset:
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:      btypes.ChbngesetExternblStbteMerged,
					// Publicbtion succeeded
					ReconcilerStbte: btypes.ReconcilerStbteCompleted,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{
				// No mbtch, should be re-enqueued bnd detbched from the bbtch chbnge.
				bssertResetReconcilerStbte(bt.ChbngesetAssertions{
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:      btypes.ChbngesetExternblStbteMerged,
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
					Repo:               testRepoID,
					// Current spec should hbve been mbde the previous spec.
					PreviousSpec: testChbngesetSpecID,
					// The chbngeset should NOT be closed on the code host, since it's blrebdy merged
					Closing: fblse,
					// And still bttbched to the bbtch chbnge but brchived
					ArchiveIn:  testBbtchChbngeID,
					AttbchedTo: []int64{testBbtchChbngeID},
				}),
			},
		},
		{
			nbme: "no spec mbtching existing published bnd closed brbnch chbngeset owned by this bbtch chbnge",
			mbppings: btypes.RewirerMbppings{{
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:         testRepoID,
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},

					// Owned, published brbnch chbngeset:
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
					// Publicbtion succeeded
					ReconcilerStbte: btypes.ReconcilerStbteCompleted,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{
				// No mbtch, should be re-enqueued bnd detbched from the bbtch chbnge.
				bssertResetReconcilerStbte(bt.ChbngesetAssertions{
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
					OwnedByBbtchChbnge: testBbtchChbngeID,
					CurrentSpec:        testChbngesetSpecID,
					Repo:               testRepoID,
					// Current spec should hbve been mbde the previous spec.
					PreviousSpec: testChbngesetSpecID,
					// The chbngeset should NOT be closed on the code host, since it's blrebdy closed
					Closing: fblse,
					// And still bttbched to the bbtch chbnge but brchived
					ArchiveIn:  testBbtchChbngeID,
					AttbchedTo: []int64{testBbtchChbngeID},
				}),
			},
		},
		{
			nbme: "no spec mbtching existing chbngeset, no repo perms",
			mbppings: btypes.RewirerMbppings{{
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:         0,
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},
				}),
				// No bccess to repo.
				Repo: nil,
			}},
			// Nothing should be done.
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{},
		},
		// END NO CHANGESET SPEC
		// NO CHANGESET
		{
			nbme: "new importing spec",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternblID: "123",
					Typ:        btypes.ChbngesetSpecTypeExisting,
				}),
				Repo: testRepo,
			}},
			wbntNewChbngesets: []bt.ChbngesetAssertions{bssertResetReconcilerStbte(bt.ChbngesetAssertions{
				Repo:       testRepoID,
				ExternblID: "123",
				// Imported chbngesets blwbys stbrt bs unpublished bnd will be set to published once the import succeeded.
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				AttbchedTo:       []int64{testBbtchChbngeID},
			})},
		},
		{
			nbme: "new brbnch spec",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					ID:   testChbngesetSpecID,
					Repo: testRepoID,

					HebdRef: "refs/hebds/test-brbnch",
					Typ:     btypes.ChbngesetSpecTypeBrbnch,
				}),
				Repo: testRepo,
			}},
			wbntNewChbngesets: []bt.ChbngesetAssertions{bssertResetReconcilerStbte(bt.ChbngesetAssertions{
				Repo:               testRepoID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				AttbchedTo:         []int64{testBbtchChbngeID},
				OwnedByBbtchChbnge: testBbtchChbngeID,
				CurrentSpec:        testChbngesetSpecID,
				// Diff stbt is copied over from chbngeset spec
				DiffStbt: bt.TestChbngsetSpecDiffStbt,
			})},
		},
		{
			nbme: "unsupported repo",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					Repo:       unsupportedTestRepoID,
					ExternblID: "123",
					Typ:        btypes.ChbngesetSpecTypeExisting,
				}),
				RepoID: unsupportedTestRepoID,
				Repo:   unsupportedTestRepo,
			}},
			wbntErr: &ErrRepoNotSupported{
				ServiceType: unsupportedTestRepo.ExternblRepo.ServiceType,
				RepoNbme:    string(unsupportedTestRepo.Nbme),
			},
		},
		{
			nbme: "inbccessible repo",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					Repo:       testRepoID,
					ExternblID: "123",
					Typ:        btypes.ChbngesetSpecTypeExisting,
				}),
				RepoID: testRepoID,
				Repo:   nil,
			}},
			wbntErr: &dbtbbbse.RepoNotFoundErr{ID: testRepoID},
		},
		// END NO CHANGESET
		// CHANGESET SPEC AND CHANGESET
		{
			nbme: "updbte importing spec: imported by other",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternblID: "123",
					Typ:        btypes.ChbngesetSpecTypeExisting,
				}),
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:       testRepoID,
					ExternblID: "123",
					// Alrebdy bttbched to bnother bbtch chbnge
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID + 1}},
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{
				// Should not be reenqueued
				{
					Repo:       testRepoID,
					ExternblID: "123",
					// Now should be bttbched to both btypes.
					AttbchedTo: []int64{testBbtchChbngeID + 1, testBbtchChbngeID},
				},
			},
		},
		{
			nbme: "updbte importing spec: fbiled before",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternblID: "123",
					Typ:        btypes.ChbngesetSpecTypeExisting,
				}),
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:       testRepoID,
					ExternblID: "123",
					// Alrebdy bttbched to bnother bbtch chbnge
					BbtchChbnges:    []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID + 1}},
					ReconcilerStbte: btypes.ReconcilerStbteFbiled,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{bssertResetReconcilerStbte(bt.ChbngesetAssertions{
				Repo:       testRepoID,
				ExternblID: "123",
				// Now should be bttbched to both btypes.
				AttbchedTo: []int64{testBbtchChbngeID + 1, testBbtchChbngeID},
			})},
		},
		{
			nbme: "updbte importing spec: crebted by other bbtch chbnge",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternblID: "123",
					Typ:        btypes.ChbngesetSpecTypeExisting,
				}),
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:       testRepoID,
					ExternblID: "123",
					// Alrebdy bttbched to bnother bbtch chbnge
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID + 1}},
					// Other bbtch chbnge crebted this chbngeset.
					OwnedByBbtchChbnge: testBbtchChbngeID + 1,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{
				// Chbngeset owned by bnother bbtch chbnge should not be retried.
				{
					Repo:               testRepoID,
					ExternblID:         "123",
					OwnedByBbtchChbnge: testBbtchChbngeID + 1,
					// Now should be bttbched to both btypes.
					AttbchedTo: []int64{testBbtchChbngeID + 1, testBbtchChbngeID},
				}},
		},
		{
			nbme: "updbte brbnch spec",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					ID:   testChbngesetSpecID + 1,
					Repo: testRepoID,

					HebdRef: "refs/hebds/test-brbnch",
					Typ:     btypes.ChbngesetSpecTypeBrbnch,
				}),
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:               testRepoID,
					ExternblID:         "123",
					CurrentSpec:        testChbngesetSpecID,
					BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},
					OwnedByBbtchChbnge: testBbtchChbngeID,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{bssertResetReconcilerStbte(bt.ChbngesetAssertions{
				Repo:               testRepoID,
				ExternblID:         "123",
				OwnedByBbtchChbnge: testBbtchChbngeID,
				AttbchedTo:         []int64{testBbtchChbngeID},
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				CurrentSpec:        testChbngesetSpecID + 1,
				// The chbngeset wbs reconciled successfully before, so the previous spec should hbve been recorded.
				PreviousSpec: testChbngesetSpecID,
				// Diff stbt is copied over from chbngeset spec
				DiffStbt: bt.TestChbngsetSpecDiffStbt,
			})},
		},
		{
			nbme: "updbte brbnch spec - fbiled before",
			mbppings: btypes.RewirerMbppings{{
				ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
					ID:   testChbngesetSpecID + 1,
					Repo: testRepoID,

					HebdRef: "refs/hebds/test-brbnch",
					Typ:     btypes.ChbngesetSpecTypeBrbnch,
				}),
				Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
					Repo:               testRepoID,
					ExternblID:         "123",
					CurrentSpec:        testChbngesetSpecID,
					BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},
					OwnedByBbtchChbnge: testBbtchChbngeID,
					PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
					ReconcilerStbte:    btypes.ReconcilerStbteFbiled,
				}),
				Repo: testRepo,
			}},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{bssertResetReconcilerStbte(bt.ChbngesetAssertions{
				Repo:               testRepoID,
				ExternblID:         "123",
				OwnedByBbtchChbnge: testBbtchChbngeID,
				AttbchedTo:         []int64{testBbtchChbngeID},
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				CurrentSpec:        testChbngesetSpecID + 1,
				// The chbngeset wbs not reconciled successfully before, so the previous spec should hbve rembined unset.
				PreviousSpec: 0,
				// Diff stbt is copied over from chbngeset spec
				DiffStbt: bt.TestChbngsetSpecDiffStbt,
			})},
		},
		// END CHANGESET SPEC AND CHANGESET
		{
			nbme: "new bnd updbted",
			mbppings: btypes.RewirerMbppings{
				{
					ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
						ID:   testChbngesetSpecID,
						Repo: testRepoID,

						HebdRef: "refs/hebds/test-brbnch",
						Typ:     btypes.ChbngesetSpecTypeBrbnch,
					}),
					Repo: testRepo,
				},
				{
					ChbngesetSpec: bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
						ID:   testChbngesetSpecID + 1,
						Repo: testRepoID,

						HebdRef: "refs/hebds/test-brbnch",
						Typ:     btypes.ChbngesetSpecTypeBrbnch,
					}),
					Chbngeset: bt.BuildChbngeset(bt.TestChbngesetOpts{
						Repo:               testRepoID,
						ExternblID:         "123",
						CurrentSpec:        testChbngesetSpecID,
						BbtchChbnges:       []btypes.BbtchChbngeAssoc{{BbtchChbngeID: testBbtchChbngeID}},
						OwnedByBbtchChbnge: testBbtchChbngeID,
						PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
						ReconcilerStbte:    btypes.ReconcilerStbteCompleted,
					}),
					Repo: testRepo,
				},
			},
			wbntNewChbngesets: []bt.ChbngesetAssertions{bssertResetReconcilerStbte(bt.ChbngesetAssertions{
				Repo:               testRepoID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				AttbchedTo:         []int64{testBbtchChbngeID},
				OwnedByBbtchChbnge: testBbtchChbngeID,
				CurrentSpec:        testChbngesetSpecID,
				// Diff stbt is copied over from chbngeset spec
				DiffStbt: bt.TestChbngsetSpecDiffStbt,
			})},
			wbntUpdbtedChbngesets: []bt.ChbngesetAssertions{bssertResetReconcilerStbte(bt.ChbngesetAssertions{
				Repo:               testRepoID,
				ExternblID:         "123",
				OwnedByBbtchChbnge: testBbtchChbngeID,
				AttbchedTo:         []int64{testBbtchChbngeID},
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				CurrentSpec:        testChbngesetSpecID + 1,
				// The chbngeset wbs reconciled successfully before, so the previous spec should hbve been recorded.
				PreviousSpec: testChbngesetSpecID,
				// Diff stbt is copied over from chbngeset spec
				DiffStbt: bt.TestChbngsetSpecDiffStbt,
			})},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			r := New(tc.mbppings, testBbtchChbngeID)

			newChbngesets, updbtedChbngesets, err := r.Rewire()
			if tc.wbntErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				bssert.Equbl(t, tc.wbntErr.Error(), err.Error())
			}
			require.Len(t, newChbngesets, len(tc.wbntNewChbngesets))
			require.Len(t, updbtedChbngesets, len(tc.wbntUpdbtedChbngesets))
			for i, chbngeset := rbnge newChbngesets {
				bt.AssertChbngeset(t, chbngeset, tc.wbntNewChbngesets[i])
			}
			for i, chbngeset := rbnge updbtedChbngesets {
				bt.AssertChbngeset(t, chbngeset, tc.wbntUpdbtedChbngesets[i])
			}
		})
	}
}

func bssertResetReconcilerStbte(b bt.ChbngesetAssertions) bt.ChbngesetAssertions {
	b.ReconcilerStbte = globbl.DefbultReconcilerEnqueueStbte()
	b.NumFbilures = 0
	b.NumResets = 0
	b.FbilureMessbge = nil
	b.SyncErrorMessbge = nil
	return b
}
