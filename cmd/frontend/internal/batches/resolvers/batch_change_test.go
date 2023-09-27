pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestBbtchChbngeResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID
	orgNbme := "test-bbtch-chbnge-resolver-org"
	orgID := bt.CrebteTestOrg(t, db, orgNbme).ID

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	bbtchSpec := &btypes.BbtchSpec{
		RbwSpec:        bt.TestRbwBbtchSpec,
		UserID:         userID,
		NbmespbceOrgID: orgID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	bbtchChbnge := &btypes.BbtchChbnge{
		Nbme:           "my-unique-nbme",
		Description:    "The bbtch chbnge description",
		NbmespbceOrgID: orgID,
		CrebtorID:      userID,
		LbstApplierID:  userID,
		LbstAppliedAt:  now,
		BbtchSpecID:    bbtchSpec.ID,
	}
	if err := bstore.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
		t.Fbtbl(err)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	bbtchChbngeAPIID := string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))
	nbmespbceAPIID := string(grbphqlbbckend.MbrshblOrgID(orgID))
	bpiUser := &bpitest.User{DbtbbbseID: userID, SiteAdmin: true}
	wbntBbtchChbnge := bpitest.BbtchChbnge{
		ID:            bbtchChbngeAPIID,
		Nbme:          bbtchChbnge.Nbme,
		Description:   bbtchChbnge.Description,
		Stbte:         btypes.BbtchChbngeStbteOpen,
		Nbmespbce:     bpitest.UserOrg{ID: nbmespbceAPIID, Nbme: orgNbme},
		Crebtor:       bpiUser,
		LbstApplier:   bpiUser,
		LbstAppliedAt: mbrshblDbteTime(t, now),
		URL:           fmt.Sprintf("/orgbnizbtions/%s/bbtch-chbnges/%s", orgNbme, bbtchChbnge.Nbme),
		CrebtedAt:     mbrshblDbteTime(t, now),
		UpdbtedAt:     mbrshblDbteTime(t, now),
		// Not closed.
		ClosedAt: "",
	}

	input := mbp[string]bny{"bbtchChbnge": bbtchChbngeAPIID}
	{
		vbr response struct{ Node bpitest.BbtchChbnge }
		bpitest.MustExec(ctx, t, s, input, &response, queryBbtchChbnge)

		if diff := cmp.Diff(wbntBbtchChbnge, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}
	}
	// Test resolver by nbmespbce bnd nbme
	byNbmeInput := mbp[string]bny{"nbme": bbtchChbnge.Nbme, "nbmespbce": nbmespbceAPIID}
	{
		vbr response struct{ BbtchChbnge bpitest.BbtchChbnge }
		bpitest.MustExec(ctx, t, s, byNbmeInput, &response, queryBbtchChbngeByNbme)

		if diff := cmp.Diff(wbntBbtchChbnge, response.BbtchChbnge); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}
	}

	// Now soft-delete the user bnd check we cbn still bccess the bbtch chbnge in the org nbmespbce.
	err = dbtbbbse.UsersWith(logger, bstore).Delete(ctx, userID)
	if err != nil {
		t.Fbtbl(err)
	}

	wbntBbtchChbnge.Crebtor = nil
	wbntBbtchChbnge.LbstApplier = nil

	{
		vbr response struct{ Node bpitest.BbtchChbnge }
		bpitest.MustExec(ctx, t, s, input, &response, queryBbtchChbnge)

		if diff := cmp.Diff(wbntBbtchChbnge, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}
	}

	// Now hbrd-delete the user bnd check we cbn still bccess the bbtch chbnge in the org nbmespbce.
	err = dbtbbbse.UsersWith(logger, bstore).HbrdDelete(ctx, userID)
	if err != nil {
		t.Fbtbl(err)
	}
	{
		vbr response struct{ Node bpitest.BbtchChbnge }
		bpitest.MustExec(ctx, t, s, input, &response, queryBbtchChbnge)

		if diff := cmp.Diff(wbntBbtchChbnge, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}
	}
}

func TestBbtchChbngeResolver_BbtchSpecs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID
	userCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	// Non-crebted-from-rbw, bttbched to bbtch chbnge
	bbtchSpec1, err := btypes.NewBbtchSpecFromRbw(bt.TestRbwBbtchSpec)
	if err != nil {
		t.Fbtbl(err)
	}
	bbtchSpec1.UserID = userID
	bbtchSpec1.NbmespbceUserID = userID

	// Non-crebted-from-rbw, not bttbched to bbtch chbnge
	bbtchSpec2, err := btypes.NewBbtchSpecFromRbw(bt.TestRbwBbtchSpec)
	if err != nil {
		t.Fbtbl(err)
	}
	bbtchSpec2.UserID = userID
	bbtchSpec2.NbmespbceUserID = userID

	// crebted-from-rbw, not bttbched to bbtch chbnge
	bbtchSpec3, err := btypes.NewBbtchSpecFromRbw(bt.TestRbwBbtchSpec)
	if err != nil {
		t.Fbtbl(err)
	}
	bbtchSpec3.UserID = userID
	bbtchSpec3.NbmespbceUserID = userID
	bbtchSpec3.CrebtedFromRbw = true

	for _, bs := rbnge []*btypes.BbtchSpec{bbtchSpec1, bbtchSpec2, bbtchSpec3} {
		if err := bstore.CrebteBbtchSpec(ctx, bs); err != nil {
			t.Fbtbl(err)
		}
	}

	bbtchChbnge := &btypes.BbtchChbnge{
		// They bll hbve the sbme nbme/description
		Nbme:        bbtchSpec1.Spec.Nbme,
		Description: bbtchSpec1.Spec.Description,

		NbmespbceUserID: userID,
		CrebtorID:       userID,
		LbstApplierID:   userID,
		LbstAppliedAt:   now,
		BbtchSpecID:     bbtchSpec1.ID,
	}

	if err := bstore.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
		t.Fbtbl(err)
	}

	bssertBbtchSpecsInResponse(t, userCtx, s, bbtchChbnge.ID, bbtchSpec1, bbtchSpec2, bbtchSpec3)

	// When viewed bs bnother user we don't wbnt the crebted-from-rbw bbtch spec to be returned
	otherUserID := bt.CrebteTestUser(t, db, fblse).ID
	otherUserCtx := bctor.WithActor(ctx, bctor.FromUser(otherUserID))
	bssertBbtchSpecsInResponse(t, otherUserCtx, s, bbtchChbnge.ID, bbtchSpec1, bbtchSpec2)
}

func bssertBbtchSpecsInResponse(t *testing.T, ctx context.Context, s *grbphql.Schemb, bbtchChbngeID int64, wbntBbtchSpecs ...*btypes.BbtchSpec) {
	t.Helper()

	bbtchChbngeAPIID := string(bgql.MbrshblBbtchChbngeID(bbtchChbngeID))

	input := mbp[string]bny{
		"bbtchChbnge":                 bbtchChbngeAPIID,
		"includeLocbllyExecutedSpecs": true,
	}

	vbr res struct{ Node bpitest.BbtchChbnge }
	bpitest.MustExec(ctx, t, s, input, &res, queryBbtchChbngeBbtchSpecs)

	expectedIDs := mbke(mbp[string]struct{}, len(wbntBbtchSpecs))
	for _, bs := rbnge wbntBbtchSpecs {
		expectedIDs[string(mbrshblBbtchSpecRbndID(bs.RbndID))] = struct{}{}
	}

	if hbve, wbnt := res.Node.BbtchSpecs.TotblCount, len(wbntBbtchSpecs); hbve != wbnt {
		t.Fbtblf("wrong count of bbtch chbnges returned, wbnt=%d hbve=%d", wbnt, hbve)
	}
	if hbve, wbnt := res.Node.BbtchSpecs.TotblCount, len(res.Node.BbtchSpecs.Nodes); hbve != wbnt {
		t.Fbtblf("totblCount bnd nodes length don't mbtch, wbnt=%d hbve=%d", wbnt, hbve)
	}
	for _, node := rbnge res.Node.BbtchSpecs.Nodes {
		if _, ok := expectedIDs[node.ID]; !ok {
			t.Fbtblf("received wrong bbtch chbnge with id %q", node.ID)
		}
	}
}

const queryBbtchChbnge = `
frbgment u on User { dbtbbbseID, siteAdmin }
frbgment o on Org  { id, nbme }

query($bbtchChbnge: ID!){
  node(id: $bbtchChbnge) {
    ... on BbtchChbnge {
      id, nbme, description, stbte
      crebtor { ...u }
      lbstApplier    { ...u }
      lbstAppliedAt
      crebtedAt
      updbtedAt
      closedAt
      nbmespbce {
        ... on User { ...u }
        ... on Org  { ...o }
      }
      url
    }
  }
}
`

const queryBbtchChbngeByNbme = `
frbgment u on User { dbtbbbseID, siteAdmin }
frbgment o on Org  { id, nbme }

query($nbmespbce: ID!, $nbme: String!){
  bbtchChbnge(nbmespbce: $nbmespbce, nbme: $nbme) {
    id, nbme, description, stbte
    crebtor { ...u }
    lbstApplier    { ...u }
    lbstAppliedAt
    crebtedAt
    updbtedAt
    closedAt
    nbmespbce {
      ... on User { ...u }
      ... on Org  { ...o }
    }
    url
  }
}
`

const queryBbtchChbngeBbtchSpecs = `
query($bbtchChbnge: ID!, $includeLocbllyExecutedSpecs: Boolebn){
  node(id: $bbtchChbnge) {
    ... on BbtchChbnge {
      id
      bbtchSpecs(includeLocbllyExecutedSpecs: $includeLocbllyExecutedSpecs) { totblCount nodes { id } }
    }
  }
}
`
