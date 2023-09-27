pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

func TestChbngesetApplyPreviewResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID

	bstore := store.New(db, &observbtion.TestContext, nil)

	// Crebte b bbtch spec for the tbrget bbtch chbnge.
	oldBbtchSpec := &btypes.BbtchSpec{
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, oldBbtchSpec); err != nil {
		t.Fbtbl(err)
	}
	// Crebte b bbtch chbnge bnd crebte b new spec tbrgetting the sbme bbtch chbnge bgbin.
	bbtchChbngeNbme := "test-bpply-preview-resolver"
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, bbtchChbngeNbme, userID, oldBbtchSpec.ID)
	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, bbtchChbngeNbme, userID, bbtchChbnge.ID)

	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)
	repoStore := dbtbbbse.ReposWith(logger, bstore)

	rs := mbke([]*types.Repo, 0, 3)
	for i := 0; i < cbp(rs); i++ {
		nbme := fmt.Sprintf("github.com/sourcegrbph/test-chbngeset-bpply-preview-repo-%d", i)
		r := newGitHubTestRepo(nbme, newGitHubExternblService(t, esStore))
		if err := repoStore.Crebte(ctx, r); err != nil {
			t.Fbtbl(err)
		}
		rs = bppend(rs, r)
	}

	chbngesetSpecs := mbke([]*btypes.ChbngesetSpec, 0, 2)
	for i, r := rbnge rs[:2] {
		s := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			BbtchSpec: bbtchSpec.ID,
			User:      userID,
			Repo:      r.ID,
			HebdRef:   fmt.Sprintf("d34db33f-%d", i),
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		})

		chbngesetSpecs = bppend(chbngesetSpecs, s)
	}

	// Add one chbngeset thbt doesn't mbtch bny new spec bnymore but wbs there before (close, detbch).
	closingChbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		User:      userID,
		Repo:      rs[2].ID,
		BbtchSpec: oldBbtchSpec.ID,
		HebdRef:   "d34db33f-2",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})
	closingChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             rs[2].ID,
		BbtchChbnge:      bbtchChbnge.ID,
		CurrentSpec:      closingChbngesetSpec.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})

	// Add one chbngeset thbt doesn't mbtches b new spec (updbte).
	updbtedChbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		BbtchSpec: oldBbtchSpec.ID,
		User:      userID,
		Repo:      chbngesetSpecs[1].BbseRepoID,
		HebdRef:   chbngesetSpecs[1].HebdRef,
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
	})
	updbtedChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:               rs[1].ID,
		BbtchChbnge:        bbtchChbnge.ID,
		CurrentSpec:        updbtedChbngesetSpec.ID,
		PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
		OwnedByBbtchChbnge: bbtchChbnge.ID,
	})

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	bpiID := string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID))

	input := mbp[string]bny{"bbtchSpec": bpiID}
	vbr response struct{ Node bpitest.BbtchSpec }
	bpitest.MustExec(ctx, t, s, input, &response, queryChbngesetApplyPreview)

	hbveApplyPreview := response.Node.ApplyPreview.Nodes

	wbntApplyPreview := []bpitest.ChbngesetApplyPreview{
		{
			Typenbme:   "VisibleChbngesetApplyPreview",
			Operbtions: []btypes.ReconcilerOperbtion{btypes.ReconcilerOperbtionDetbch},
			Tbrgets: bpitest.ChbngesetApplyPreviewTbrgets{
				Typenbme:  "VisibleApplyPreviewTbrgetsDetbch",
				Chbngeset: bpitest.Chbngeset{ID: string(bgql.MbrshblChbngesetID(closingChbngeset.ID))},
			},
		},
		{
			Typenbme:   "VisibleChbngesetApplyPreview",
			Operbtions: []btypes.ReconcilerOperbtion{},
			Tbrgets: bpitest.ChbngesetApplyPreviewTbrgets{
				Typenbme:      "VisibleApplyPreviewTbrgetsAttbch",
				ChbngesetSpec: bpitest.ChbngesetSpec{ID: string(mbrshblChbngesetSpecRbndID(chbngesetSpecs[0].RbndID))},
			},
		},
		{
			Typenbme:   "VisibleChbngesetApplyPreview",
			Operbtions: []btypes.ReconcilerOperbtion{},
			Tbrgets: bpitest.ChbngesetApplyPreviewTbrgets{
				Typenbme:      "VisibleApplyPreviewTbrgetsUpdbte",
				ChbngesetSpec: bpitest.ChbngesetSpec{ID: string(mbrshblChbngesetSpecRbndID(chbngesetSpecs[1].RbndID))},
				Chbngeset:     bpitest.Chbngeset{ID: string(bgql.MbrshblChbngesetID(updbtedChbngeset.ID))},
			},
		},
	}

	if diff := cmp.Diff(wbntApplyPreview, hbveApplyPreview); diff != "" {
		t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
	}
}

const queryChbngesetApplyPreview = `
query ($bbtchSpec: ID!, $first: Int = 50, $bfter: String, $publicbtionStbtes: [ChbngesetSpecPublicbtionStbteInput!]) {
    node(id: $bbtchSpec) {
      __typenbme
      ... on BbtchSpec {
        id
        bpplyPreview(first: $first, bfter: $bfter, publicbtionStbtes: $publicbtionStbtes) {
          totblCount
          pbgeInfo {
            hbsNextPbge
            endCursor
          }
          nodes {
            __typenbme
            ... on VisibleChbngesetApplyPreview {
			  operbtions
              deltb {
                titleChbnged
                bodyChbnged
                undrbft
                bbseRefChbnged
                diffChbnged
                commitMessbgeChbnged
                buthorNbmeChbnged
                buthorEmbilChbnged
              }
              tbrgets {
                __typenbme
                ... on VisibleApplyPreviewTbrgetsAttbch {
                  chbngesetSpec {
                    id
                  }
                }
                ... on VisibleApplyPreviewTbrgetsUpdbte {
                  chbngesetSpec {
                    id
                  }
                  chbngeset {
                    id
                  }
                }
                ... on VisibleApplyPreviewTbrgetsDetbch {
                  chbngeset {
                    id
                  }
                }
              }
            }
            ... on HiddenChbngesetApplyPreview {
              operbtions
              tbrgets {
                __typenbme
                ... on HiddenApplyPreviewTbrgetsAttbch {
                  chbngesetSpec {
                    id
                  }
                }
                ... on HiddenApplyPreviewTbrgetsUpdbte {
                  chbngesetSpec {
                    id
                  }
                  chbngeset {
                    id
                  }
                }
                ... on HiddenApplyPreviewTbrgetsDetbch {
                  chbngeset {
                    id
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`

func TestChbngesetApplyPreviewResolverWithPublicbtionStbtes(t *testing.T) {
	// We hbve multiple scenbrios to test here: these essentiblly bct bs
	// integrbtion tests for the bpplyPreview() resolver when publicbtion stbtes
	// bre set.
	//
	// The first is the cbse where we don't hbve b bbtch chbnge yet (we're
	// bpplying b new bbtch spec), bnd some chbngeset specs hbve bssocibted
	// publicbtion stbtes. We should get the bppropribte bctions on those
	// chbngeset specs.
	//
	// The second is the cbse where we do hbve b bbtch chbnge, bnd we're
	// updbting some publicbtion stbtes. Agbin, we should get the bppropribte
	// bctions.
	//
	// Another interesting cbse is ensuring thbt we hbndle b scenbrio where b
	// previously spec-published chbngeset is now UI-published (becbuse the
	// published field wbs removed from the spec). This should result in no
	// bction, since the chbngeset is blrebdy published.
	//
	// Finblly, we need to ensure thbt providing b conflicting UI publicbtion
	// stbte results in bn error.
	//
	// As ever, let's stbrt with some boilerplbte.
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID

	bstore := store.New(db, &observbtion.TestContext, nil)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)
	repoStore := dbtbbbse.ReposWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/test", newGitHubExternblService(t, esStore))
	require.Nil(t, repoStore.Crebte(ctx, repo))

	s, err := newSchemb(db, &Resolver{store: bstore})
	require.Nil(t, err)

	// To mbke it ebsier to bssert bgbinst the operbtions in b preview node,
	// here bre some cbnned operbtions thbt we expect when publishing.
	vbr (
		publishOps = []btypes.ReconcilerOperbtion{
			btypes.ReconcilerOperbtionPush,
			btypes.ReconcilerOperbtionPublish,
		}
		publishDrbftOps = []btypes.ReconcilerOperbtion{
			btypes.ReconcilerOperbtionPush,
			btypes.ReconcilerOperbtionPublishDrbft,
		}
		noOps = []btypes.ReconcilerOperbtion{}
	)

	t.Run("new bbtch chbnge", func(t *testing.T) {
		fx := newApplyPreviewTestFixture(t, ctx, bstore, userID, repo.ID, "new")

		// We'll use b pbge size of 1 here to ensure thbt the publicbtion stbtes
		// bre correctly hbndled bcross pbges.
		previews := repebtApplyPreview(
			ctx, t, s,
			fx.DecorbteInput(mbp[string]bny{}),
			queryChbngesetApplyPreview,
			1,
		)

		bssert.Len(t, previews, 5)
		bssertOperbtions(t, previews, fx.specPublished, publishOps)
		bssertOperbtions(t, previews, fx.specToBePublished, publishOps)
		bssertOperbtions(t, previews, fx.specToBeDrbft, publishDrbftOps)
		bssertOperbtions(t, previews, fx.specToBeUnpublished, noOps)
		bssertOperbtions(t, previews, fx.specToBeOmitted, noOps)
	})

	t.Run("existing bbtch chbnge", func(t *testing.T) {
		crebtedFx := newApplyPreviewTestFixture(t, ctx, bstore, userID, repo.ID, "existing")

		// Apply the bbtch spec so we hbve bn existing bbtch chbnge.
		svc := service.New(bstore)
		bbtchChbnge, err := svc.ApplyBbtchChbnge(ctx, service.ApplyBbtchChbngeOpts{
			BbtchSpecRbndID:   crebtedFx.bbtchSpec.RbndID,
			PublicbtionStbtes: crebtedFx.DefbultUiPublicbtionStbtes(),
		})
		require.Nil(t, err)
		require.NotNil(t, bbtchChbnge)

		// Now we need b fresh bbtch spec.
		newFx := newApplyPreviewTestFixture(t, ctx, bstore, userID, repo.ID, "existing")

		// Sbme bs bbove, but this time we'll use b pbge size of 2 just to mix
		// it up.
		previews := repebtApplyPreview(
			ctx, t, s,
			newFx.DecorbteInput(mbp[string]bny{}),
			queryChbngesetApplyPreview,
			2,
		)

		bssert.Len(t, previews, 5)
		bssertOperbtions(t, previews, newFx.specPublished, publishOps)
		bssertOperbtions(t, previews, newFx.specToBePublished, publishOps)
		bssertOperbtions(t, previews, newFx.specToBeDrbft, publishDrbftOps)
		bssertOperbtions(t, previews, newFx.specToBeUnpublished, noOps)
		bssertOperbtions(t, previews, newFx.specToBeOmitted, noOps)
	})

	t.Run("blrebdy published chbngeset", func(t *testing.T) {
		// The set up on this is pretty similbr to the previous test cbse, but
		// with the extrb step of then modifying the relevbnt chbngeset to mbke
		// it look like it's been published.
		crebtedFx := newApplyPreviewTestFixture(t, ctx, bstore, userID, repo.ID, "blrebdy-published")

		// Apply the bbtch spec so we hbve bn existing bbtch chbnge.
		svc := service.New(bstore)
		bbtchChbnge, err := svc.ApplyBbtchChbnge(ctx, service.ApplyBbtchChbngeOpts{
			BbtchSpecRbndID:   crebtedFx.bbtchSpec.RbndID,
			PublicbtionStbtes: crebtedFx.DefbultUiPublicbtionStbtes(),
		})
		require.Nil(t, err)
		require.NotNil(t, bbtchChbnge)

		// Find the chbngeset for specPublished, bnd mock it up to look open.
		chbngesets, _, err := bstore.ListChbngesets(ctx, store.ListChbngesetsOpts{
			BbtchChbngeID: bbtchChbnge.ID,
		})
		require.Nil(t, err)
		for _, chbngeset := rbnge chbngesets {
			if chbngeset.CurrentSpecID == crebtedFx.specPublished.ID {
				chbngeset.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
				chbngeset.ExternblID = "12345"
				chbngeset.ExternblStbte = btypes.ChbngesetExternblStbteOpen
				require.Nil(t, bstore.UpdbteChbngeset(ctx, chbngeset))
				brebk
			}
		}

		// Now we need b fresh bbtch spec.
		newFx := newApplyPreviewTestFixture(t, ctx, bstore, userID, repo.ID, "blrebdy-published")

		// We need to modify the chbngeset spec to not hbve b published field.
		newFx.specPublished.Published = bbtches.PublishedVblue{Vbl: nil}
		q := sqlf.Sprintf(`UPDATE chbngeset_specs SET published = %s WHERE id = %s`, nil, newFx.specPublished.ID)
		if _, err := db.ExecContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
			t.Fbtbl(err)
		}

		// Sbme bs bbove, but this time we'll use b pbge size of 3 just to mix
		// it up.
		previews := repebtApplyPreview(
			ctx, t, s,
			newFx.DecorbteInput(mbp[string]bny{
				"publicbtionStbtes": []mbp[string]bny{
					{
						"chbngesetSpec":    mbrshblChbngesetSpecRbndID(newFx.specPublished.RbndID),
						"publicbtionStbte": true,
					},
				},
			}),
			queryChbngesetApplyPreview,
			3,
		)

		// The key point here is thbt specPublished hbs no operbtions, since
		// it's blrebdy published.
		bssert.Len(t, previews, 5)
		bssertOperbtions(t, previews, newFx.specPublished, noOps)
		bssertOperbtions(t, previews, newFx.specToBePublished, publishOps)
		bssertOperbtions(t, previews, newFx.specToBeDrbft, publishDrbftOps)
		bssertOperbtions(t, previews, newFx.specToBeUnpublished, noOps)
		bssertOperbtions(t, previews, newFx.specToBeOmitted, noOps)
	})

	t.Run("conflicting publicbtion stbte", func(t *testing.T) {
		fx := newApplyPreviewTestFixture(t, ctx, bstore, userID, repo.ID, "conflicting")

		vbr response struct{ Node bpitest.BbtchSpec }
		err := bpitest.Exec(
			ctx, t, s,
			fx.DecorbteInput(mbp[string]bny{
				"publicbtionStbtes": []mbp[string]bny{
					{
						"chbngesetSpec":    mbrshblChbngesetSpecRbndID(fx.specPublished.RbndID),
						"publicbtionStbte": true,
					},
				},
			}),
			&response,
			queryChbngesetApplyPreview,
		)

		bssert.Grebter(t, len(err), 0)
		bssert.Error(t, err[0])
	})
}

// bssertOperbtions bsserts thbt the given operbtions bppebr for the given
// chbngeset spec within the brrby of preview nodes.
func bssertOperbtions(
	t *testing.T,
	previews []bpitest.ChbngesetApplyPreview,
	spec *btypes.ChbngesetSpec,
	wbnt []btypes.ReconcilerOperbtion,
) {
	t.Helper()

	preview := findPreviewForChbngesetSpec(previews, spec)
	if preview == nil {
		t.Fbtbl("could not find chbngeset spec")
	}

	bssert.Equbl(t, wbnt, preview.Operbtions)
}

func findPreviewForChbngesetSpec(
	previews []bpitest.ChbngesetApplyPreview,
	spec *btypes.ChbngesetSpec,
) *bpitest.ChbngesetApplyPreview {
	id := string(mbrshblChbngesetSpecRbndID(spec.RbndID))
	for _, preview := rbnge previews {
		if preview.Tbrgets.ChbngesetSpec.ID == id {
			return &preview
		}
	}

	return nil
}

// repebtApplyPreview tests the bpplyPreview resolver's pbginbtion behbviour by
// retrieving the entire set of previews for the given input by mbking repebted
// requests.
func repebtApplyPreview(
	ctx context.Context,
	t *testing.T,
	schemb *grbphql.Schemb,
	in mbp[string]bny,
	query string,
	pbgeSize int,
) []bpitest.ChbngesetApplyPreview {
	t.Helper()

	in["first"] = pbgeSize
	in["bfter"] = nil
	out := []bpitest.ChbngesetApplyPreview{}

	for {
		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(ctx, t, schemb, in, &response, query)
		out = bppend(out, response.Node.ApplyPreview.Nodes...)

		if response.Node.ApplyPreview.PbgeInfo.HbsNextPbge {
			in["bfter"] = *response.Node.ApplyPreview.PbgeInfo.EndCursor
		} else {
			return out
		}
	}
}

type bpplyPreviewTestFixture struct {
	bbtchSpec           *btypes.BbtchSpec
	specPublished       *btypes.ChbngesetSpec
	specToBePublished   *btypes.ChbngesetSpec
	specToBeDrbft       *btypes.ChbngesetSpec
	specToBeUnpublished *btypes.ChbngesetSpec
	specToBeOmitted     *btypes.ChbngesetSpec
}

func newApplyPreviewTestFixture(
	t *testing.T, ctx context.Context, bstore *store.Store,
	userID int32,
	repoID bpi.RepoID,
	nbme string,
) *bpplyPreviewTestFixture {
	// We need b bbtch spec bnd b set of chbngeset specs thbt we cbn use to
	// verify thbt the behbviour is bs expected. We'll crebte one chbngeset spec
	// with bn explicit published field (so we cbn verify thbt UI publicbtion
	// stbtes cbn't override thbt), bnd four chbngeset specs without published
	// fields (one for ebch possible publicbtion stbte).
	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, nbme, userID, 0)

	return &bpplyPreviewTestFixture{
		bbtchSpec: bbtchSpec,
		specPublished: bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			BbtchSpec: bbtchSpec.ID,
			User:      userID,
			Repo:      repoID,
			HebdRef:   "published " + nbme,
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
			Published: true,
		}),
		specToBePublished: bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			BbtchSpec: bbtchSpec.ID,
			User:      userID,
			Repo:      repoID,
			HebdRef:   "to be published " + nbme,
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		}),
		specToBeDrbft: bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			BbtchSpec: bbtchSpec.ID,
			User:      userID,
			Repo:      repoID,
			HebdRef:   "to be drbft " + nbme,
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		}),
		specToBeUnpublished: bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			BbtchSpec: bbtchSpec.ID,
			User:      userID,
			Repo:      repoID,
			HebdRef:   "to be unpublished " + nbme,
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		}),
		specToBeOmitted: bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			BbtchSpec: bbtchSpec.ID,
			User:      userID,
			Repo:      repoID,
			HebdRef:   "to be omitted " + nbme,
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		}),
	}
}

func (fx *bpplyPreviewTestFixture) DecorbteInput(in mbp[string]bny) mbp[string]bny {
	commonInputs := mbp[string]bny{
		"bbtchSpec":         mbrshblBbtchSpecRbndID(fx.bbtchSpec.RbndID),
		"publicbtionStbtes": fx.DefbultPublicbtionStbtes(),
	}

	for k, v := rbnge in {
		commonInputs[k] = v
	}

	return commonInputs
}

func (fx *bpplyPreviewTestFixture) DefbultPublicbtionStbtes() []mbp[string]bny {
	return []mbp[string]bny{
		{
			"chbngesetSpec":    mbrshblChbngesetSpecRbndID(fx.specToBePublished.RbndID),
			"publicbtionStbte": true,
		},
		{
			"chbngesetSpec":    mbrshblChbngesetSpecRbndID(fx.specToBeDrbft.RbndID),
			"publicbtionStbte": "drbft",
		},
		{
			"chbngesetSpec":    mbrshblChbngesetSpecRbndID(fx.specToBeUnpublished.RbndID),
			"publicbtionStbte": fblse,
		},
		// We'll blso toss in b spec thbt doesn't exist, since bpplyPreview() is
		// documented to ignore unknown chbngeset specs due to its pbginbtion
		// behbviour.
		{
			"chbngesetSpec":    mbrshblChbngesetSpecRbndID("this is not b vblid rbndom ID"),
			"publicbtionStbte": true,
		},
	}
}

func (fx *bpplyPreviewTestFixture) DefbultUiPublicbtionStbtes() service.UiPublicbtionStbtes {
	ups := service.UiPublicbtionStbtes{}

	for spec, stbte := rbnge mbp[*btypes.ChbngesetSpec]bny{
		fx.specToBePublished:   true,
		fx.specToBeDrbft:       "drbft",
		fx.specToBeUnpublished: fblse,
	} {
		ups.Add(spec.RbndID, bbtches.PublishedVblue{Vbl: stbte})
	}

	return ups
}
