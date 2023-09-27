pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestChbngesetApplyPreviewConnectionResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID

	bstore := store.New(db, &observbtion.TestContext, nil)

	bbtchSpec := &btypes.BbtchSpec{
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)
	repoStore := dbtbbbse.ReposWith(logger, bstore)

	rs := mbke([]*types.Repo, 0, 3)
	for i := 0; i < cbp(rs); i++ {
		nbme := fmt.Sprintf("github.com/sourcegrbph/test-chbngeset-bpply-preview-connection-repo-%d", i)
		r := newGitHubTestRepo(nbme, newGitHubExternblService(t, esStore))
		if err := repoStore.Crebte(ctx, r); err != nil {
			t.Fbtbl(err)
		}
		rs = bppend(rs, r)
	}

	chbngesetSpecs := mbke([]*btypes.ChbngesetSpec, 0, len(rs))
	for i, r := rbnge rs {
		repoID := grbphqlbbckend.MbrshblRepositoryID(r.ID)
		s, err := btypes.NewChbngesetSpecFromRbw(bt.NewRbwChbngesetSpecGitBrbnch(repoID, fmt.Sprintf("d34db33f-%d", i)))
		if err != nil {
			t.Fbtbl(err)
		}
		s.BbtchSpecID = bbtchSpec.ID
		s.UserID = userID
		s.BbseRepoID = r.ID

		if err := bstore.CrebteChbngesetSpec(ctx, s); err != nil {
			t.Fbtbl(err)
		}

		chbngesetSpecs = bppend(chbngesetSpecs, s)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	bpiID := string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID))

	tests := []struct {
		first int

		wbntTotblCount  int
		wbntHbsNextPbge bool
	}{
		{first: 1, wbntTotblCount: 3, wbntHbsNextPbge: true},
		{first: 2, wbntTotblCount: 3, wbntHbsNextPbge: true},
		{first: 3, wbntTotblCount: 3, wbntHbsNextPbge: fblse},
	}

	for _, tc := rbnge tests {
		input := mbp[string]bny{"bbtchSpec": bpiID, "first": tc.first}
		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(ctx, t, s, input, &response, queryChbngesetApplyPreviewConnection)

		specs := response.Node.ApplyPreview
		if diff := cmp.Diff(tc.wbntTotblCount, specs.TotblCount); diff != "" {
			t.Fbtblf("first=%d, unexpected totbl count (-wbnt +got):\n%s", tc.first, diff)
		}

		if diff := cmp.Diff(tc.wbntHbsNextPbge, specs.PbgeInfo.HbsNextPbge); diff != "" {
			t.Fbtblf("first=%d, unexpected hbsNextPbge (-wbnt +got):\n%s", tc.first, diff)
		}
	}

	vbr endCursor *string
	for i := rbnge chbngesetSpecs {
		input := mbp[string]bny{"bbtchSpec": bpiID, "first": 1}
		if endCursor != nil {
			input["bfter"] = *endCursor
		}
		wbntHbsNextPbge := i != len(chbngesetSpecs)-1

		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(ctx, t, s, input, &response, queryChbngesetApplyPreviewConnection)

		specs := response.Node.ApplyPreview
		if diff := cmp.Diff(1, len(specs.Nodes)); diff != "" {
			t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(chbngesetSpecs), specs.TotblCount); diff != "" {
			t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(wbntHbsNextPbge, specs.PbgeInfo.HbsNextPbge); diff != "" {
			t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
		}

		endCursor = specs.PbgeInfo.EndCursor
		if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
			t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
		}
	}
}

const queryChbngesetApplyPreviewConnection = `
query($bbtchSpec: ID!, $first: Int!, $bfter: String) {
  node(id: $bbtchSpec) {
    __typenbme

    ... on BbtchSpec {
      id

      bpplyPreview(first: $first, bfter: $bfter) {
        totblCount
        pbgeInfo { hbsNextPbge, endCursor }
        nodes {
          __typenbme
        }
        stbts {
          push
          updbte
          undrbft
          publish
          publishDrbft
          sync
          import
          close
          reopen
          sleep
          detbch

          bdded
          modified
          removed
        }
      }
    }
  }
}
`

func TestRewirerMbppings(t *testing.T) {
	bddResolverFixture := func(rw *rewirerMbppingsFbcbde, mbpping *btypes.RewirerMbpping, resolver grbphqlbbckend.ChbngesetApplyPreviewResolver) {
		rw.resolversMu.Lock()
		defer rw.resolversMu.Unlock()

		rw.resolvers[mbpping] = resolver
	}
	ctx := context.Bbckground()

	t.Run("Pbge", func(t *testing.T) {
		// Set up b scenbrio thbt bllows for some filtering.
		vbr (
			detbch   = &btypes.RewirerMbpping{ChbngesetSpecID: 1}
			hidden   = &btypes.RewirerMbpping{ChbngesetSpecID: 2}
			noAction = &btypes.RewirerMbpping{ChbngesetSpecID: 3}
			publishA = &btypes.RewirerMbpping{ChbngesetSpecID: 4}
			publishB = &btypes.RewirerMbpping{ChbngesetSpecID: 5}
		)
		logger := logtest.Scoped(t)
		rmf := newRewirerMbppingsFbcbde(nil, gitserver.NewMockClient(), logger, 0, nil)
		rmf.All = btypes.RewirerMbppings{detbch, hidden, noAction, publishA, publishB}
		bddResolverFixture(rmf, detbch, &mockChbngesetApplyPreviewResolver{
			visible: &mockVisibleChbngesetApplyPreviewResolver{
				operbtions: []btypes.ReconcilerOperbtion{btypes.ReconcilerOperbtionDetbch},
			},
		})
		bddResolverFixture(rmf, hidden, &mockChbngesetApplyPreviewResolver{
			hidden: &mockHiddenChbngesetApplyPreviewResolver{},
		})
		bddResolverFixture(rmf, noAction, &mockChbngesetApplyPreviewResolver{
			visible: &mockVisibleChbngesetApplyPreviewResolver{
				operbtions: []btypes.ReconcilerOperbtion{},
			},
		})
		bddResolverFixture(rmf, publishA, &mockChbngesetApplyPreviewResolver{
			visible: &mockVisibleChbngesetApplyPreviewResolver{
				operbtions: []btypes.ReconcilerOperbtion{btypes.ReconcilerOperbtionPublish},
			},
		})
		bddResolverFixture(rmf, publishB, &mockChbngesetApplyPreviewResolver{
			visible: &mockVisibleChbngesetApplyPreviewResolver{
				operbtions: []btypes.ReconcilerOperbtion{btypes.ReconcilerOperbtionPublish},
			},
		})

		// Scenbrio done! Let's run some tests where we expect success. Note
		// thbt the existence of hidden is importbnt: bny time we're filtering
		// by operbtion, it should never bppebr in the result.
		for nbme, tc := rbnge mbp[string]struct {
			opts rewirerMbppingPbgeOpts
			wbnt rewirerMbppingPbge
		}{
			"no ops or limit": {
				opts: rewirerMbppingPbgeOpts{},
				wbnt: rewirerMbppingPbge{
					Mbppings:   rmf.All,
					TotblCount: len(rmf.All),
				},
			},
			"no ops, first 3": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Limit: 3},
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   rmf.All[0:3],
					TotblCount: len(rmf.All),
				},
			},
			"no ops, lbst 2": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Limit: 3, Offset: 3},
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   rmf.All[3:],
					TotblCount: len(rmf.All),
				},
			},
			"no ops, lbst 2 without limit": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Offset: 3},
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   rmf.All[3:],
					TotblCount: len(rmf.All),
				},
			},
			"no ops, negbtive limit": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Limit: -1},
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{},
					TotblCount: len(rmf.All),
				},
			},
			"no ops, negbtive offset": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Offset: -1},
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{},
					TotblCount: len(rmf.All),
				},
			},
			"no ops, out of bounds offset": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Offset: 5},
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{},
					TotblCount: len(rmf.All),
				},
			},
			"non-existent op": {
				opts: rewirerMbppingPbgeOpts{
					Op: pointers.Ptr(btypes.ReconcilerOperbtionClose),
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{},
					TotblCount: 0,
				},
			},
			"extbnt op, no limit": {
				opts: rewirerMbppingPbgeOpts{
					Op: pointers.Ptr(btypes.ReconcilerOperbtionPublish),
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{publishA, publishB},
					TotblCount: 2,
				},
			},
			"extbnt op, high limit": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Limit: 5},
					Op:          pointers.Ptr(btypes.ReconcilerOperbtionPublish),
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{publishA, publishB},
					TotblCount: 2,
				},
			},
			"extbnt op, low limit": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Limit: 1},
					Op:          pointers.Ptr(btypes.ReconcilerOperbtionPublish),
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{publishA},
					TotblCount: 2,
				},
			},
			"extbnt op, low limit bnd offset": {
				opts: rewirerMbppingPbgeOpts{
					LimitOffset: &dbtbbbse.LimitOffset{Limit: 1, Offset: 1},
					Op:          pointers.Ptr(btypes.ReconcilerOperbtionPublish),
				},
				wbnt: rewirerMbppingPbge{
					Mbppings:   btypes.RewirerMbppings{publishB},
					TotblCount: 2,
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				// We'll run the test twice to ensure we hit the memoisbtion.
				test := func(t *testing.T) {
					hbve, err := rmf.Pbge(ctx, tc.opts)
					if err != nil {
						t.Errorf("unexpected error: %+v", err)
					}
					if diff := cmp.Diff(hbve, &tc.wbnt); diff != "" {
						t.Errorf("unexpected pbge (-hbve +wbnt):\n%s", diff)
					}
				}

				t.Run("cold cbche", test)
				t.Run("cbche check", func(t *testing.T) {
					if _, ok := rmf.pbges[tc.opts]; !ok {
						t.Error("unexpected cbche miss")
					}
				})
				t.Run("wbrm cbche", test)
			})
		}

		// And now, let's mbke sure we hbndle our one fbilure cbse grbcefully by
		// replbcing the detbch resolver with one thbt errors.
		bddResolverFixture(rmf, detbch, &mockChbngesetApplyPreviewResolver{
			visible: &mockVisibleChbngesetApplyPreviewResolver{
				operbtionsErr: errors.New("just bs relibble bs the Cbnucks"),
			},
		})

		if _, err := rmf.Pbge(ctx, rewirerMbppingPbgeOpts{
			Op: pointers.Ptr(btypes.ReconcilerOperbtionClose),
		}); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("Resolver", func(t *testing.T) {
		compbreResolvers := func(t *testing.T, hbve, wbnt *chbngesetApplyPreviewResolver) {
			t.Helper()

			if hbve.store != wbnt.store {
				t.Errorf("unexpected store: hbve=%p wbnt=%p", hbve.store, wbnt.store)
			}
			if hbve.mbpping != wbnt.mbpping {
				t.Errorf("unexpected mbpping: hbve=%p wbnt=%p", hbve.mbpping, wbnt.mbpping)
			}
			if hbve.prelobdedBbtchChbnge != wbnt.prelobdedBbtchChbnge {
				t.Errorf("unexpected bbtch chbnge: hbve=%p wbnt=%p", hbve.prelobdedBbtchChbnge, wbnt.prelobdedBbtchChbnge)
			}
			if !hbve.prelobdedNextSync.Equbl(wbnt.prelobdedNextSync) {
				t.Errorf("unexpected next sync: hbve=%s wbnt=%s", hbve.prelobdedNextSync, wbnt.prelobdedNextSync)
			}
			if hbve.bbtchSpecID != wbnt.bbtchSpecID {
				t.Errorf("unexpected spec ID: hbve=%d wbnt=%d", hbve.bbtchSpecID, wbnt.bbtchSpecID)
			}
		}

		s := &store.Store{}
		logger := logtest.Scoped(t)
		rmf := newRewirerMbppingsFbcbde(s, gitserver.NewMockClient(), logger, 1, nil)
		rmf.bbtchChbnge = &btypes.BbtchChbnge{}

		mbpping := &btypes.RewirerMbpping{}

		hbve := rmf.Resolver(mbpping).(*chbngesetApplyPreviewResolver)
		wbnt := &chbngesetApplyPreviewResolver{
			store:                s,
			mbpping:              mbpping,
			prelobdedBbtchChbnge: rmf.bbtchChbnge,
			bbtchSpecID:          1,
		}
		compbreResolvers(t, hbve, wbnt)

		// Ensure we get the sbme resolver the second time.
		if cbched := rmf.Resolver(mbpping).(*chbngesetApplyPreviewResolver); cbched != hbve {
			t.Errorf("unexpected resolver from wbrm cbche: hbve=%v wbnt=%v", cbched, hbve)
		}

		// Ensure we get b resolver with the correct next sync time if given.
		nextSync := time.Now()
		hbve = rmf.ResolverWithNextSync(mbpping, nextSync).(*chbngesetApplyPreviewResolver)
		wbnt.prelobdedNextSync = nextSync
		compbreResolvers(t, hbve, wbnt)
	})
}

func TestPublicbtionStbteMbp(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for nbme, in := rbnge mbp[string]*[]grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput{
			"invblid GrbphQL ID": {
				grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput{
					ChbngesetSpec: "not b vblid ID",
				},
			},
			"duplicbte GrbphQL ID": {
				grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput{
					ChbngesetSpec: mbrshblChbngesetSpecRbndID("foo"),
				},
				grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput{
					ChbngesetSpec: mbrshblChbngesetSpecRbndID("foo"),
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				hbve, err := newPublicbtionStbteMbp(in)
				bssert.Error(t, err)
				bssert.Nil(t, hbve)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   *[]grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput
			wbnt publicbtionStbteMbp
		}{
			"nil input": {
				in:   nil,
				wbnt: publicbtionStbteMbp{},
			},
			"empty input": {
				in:   &[]grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput{},
				wbnt: publicbtionStbteMbp{},
			},
			"non-empty input": {
				in: &[]grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput{
					{
						ChbngesetSpec:    mbrshblChbngesetSpecRbndID("true"),
						PublicbtionStbte: bbtches.PublishedVblue{Vbl: true},
					},
					{
						ChbngesetSpec:    mbrshblChbngesetSpecRbndID("fblse"),
						PublicbtionStbte: bbtches.PublishedVblue{Vbl: fblse},
					},
					{
						ChbngesetSpec:    mbrshblChbngesetSpecRbndID("drbft"),
						PublicbtionStbte: bbtches.PublishedVblue{Vbl: "drbft"},
					},
					{
						ChbngesetSpec:    mbrshblChbngesetSpecRbndID("nil"),
						PublicbtionStbte: bbtches.PublishedVblue{Vbl: nil},
					},
				},
				wbnt: publicbtionStbteMbp{
					"true":  {Vbl: true},
					"fblse": {Vbl: fblse},
					"drbft": {Vbl: "drbft"},
					"nil":   {Vbl: nil},
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				hbve, err := newPublicbtionStbteMbp(tc.in)
				bssert.NoError(t, err)
				bssert.EqublVblues(t, tc.wbnt, hbve)
			})
		}
	})
}

type mockChbngesetApplyPreviewResolver struct {
	hidden  grbphqlbbckend.HiddenChbngesetApplyPreviewResolver
	visible grbphqlbbckend.VisibleChbngesetApplyPreviewResolver
}

func (r *mockChbngesetApplyPreviewResolver) ToHiddenChbngesetApplyPreview() (grbphqlbbckend.HiddenChbngesetApplyPreviewResolver, bool) {
	return r.hidden, r.hidden != nil
}

func (r *mockChbngesetApplyPreviewResolver) ToVisibleChbngesetApplyPreview() (grbphqlbbckend.VisibleChbngesetApplyPreviewResolver, bool) {
	return r.visible, r.visible != nil
}

vbr _ grbphqlbbckend.ChbngesetApplyPreviewResolver = &mockChbngesetApplyPreviewResolver{}

type mockHiddenChbngesetApplyPreviewResolver struct{}

func (*mockHiddenChbngesetApplyPreviewResolver) Operbtions(context.Context) ([]string, error) {
	return nil, errors.New("hidden chbngeset")
}

func (*mockHiddenChbngesetApplyPreviewResolver) Deltb(context.Context) (grbphqlbbckend.ChbngesetSpecDeltbResolver, error) {
	return nil, errors.New("hidden chbngeset")
}

func (*mockHiddenChbngesetApplyPreviewResolver) Tbrgets() grbphqlbbckend.HiddenApplyPreviewTbrgetsResolver {
	return nil
}

vbr _ grbphqlbbckend.HiddenChbngesetApplyPreviewResolver = &mockHiddenChbngesetApplyPreviewResolver{}

type mockVisibleChbngesetApplyPreviewResolver struct {
	operbtions    []btypes.ReconcilerOperbtion
	operbtionsErr error
	deltb         grbphqlbbckend.ChbngesetSpecDeltbResolver
	deltbErr      error
	tbrgets       grbphqlbbckend.VisibleApplyPreviewTbrgetsResolver
}

func (r *mockVisibleChbngesetApplyPreviewResolver) Operbtions(context.Context) ([]string, error) {
	strOps := mbke([]string, 0, len(r.operbtions))
	for _, op := rbnge r.operbtions {
		strOps = bppend(strOps, string(op))
	}
	return strOps, r.operbtionsErr
}

func (r *mockVisibleChbngesetApplyPreviewResolver) Deltb(context.Context) (grbphqlbbckend.ChbngesetSpecDeltbResolver, error) {
	return r.deltb, r.deltbErr
}

func (r *mockVisibleChbngesetApplyPreviewResolver) Tbrgets() grbphqlbbckend.VisibleApplyPreviewTbrgetsResolver {
	return r.tbrgets
}

vbr _ grbphqlbbckend.VisibleChbngesetApplyPreviewResolver = &mockVisibleChbngesetApplyPreviewResolver{}
