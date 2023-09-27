pbckbge resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbbc"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/overridbble"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestNullIDResilience(t *testing.T) {
	bt.MockRSAKeygen(t)

	logger := logtest.Scoped(t)

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	sr := New(db, store.New(db, &observbtion.TestContext, nil), gitserver.NewMockClient(), logger)

	s, err := newSchemb(db, sr)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := bctor.WithInternblActor(context.Bbckground())

	ids := []grbphql.ID{
		bgql.MbrshblBbtchChbngeID(0),
		bgql.MbrshblChbngesetID(0),
		mbrshblBbtchSpecRbndID(""),
		mbrshblChbngesetSpecRbndID(""),
		mbrshblBbtchChbngesCredentiblID(0, fblse),
		mbrshblBbtchChbngesCredentiblID(0, true),
		mbrshblBulkOperbtionID(""),
		mbrshblBbtchSpecWorkspbceID(0),
		mbrshblWorkspbceFileRbndID(""),
	}

	for _, id := rbnge ids {
		vbr response struct{ Node struct{ ID string } }

		query := `query($id: ID!) { node(id: $id) { id } }`
		errs := bpitest.Exec(ctx, t, s, mbp[string]bny{"id": id}, &response, query)

		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d errors", len(errs))
		}

		err := errs[0]
		if !errors.Is(err, ErrIDIsZero{}) {
			t.Errorf("expected=%#+v, got=%#+v", ErrIDIsZero{}, err)
		}
	}

	mutbtions := []string{
		fmt.Sprintf(`mutbtion { closeBbtchChbnge(bbtchChbnge: %q) { id } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { deleteBbtchChbnge(bbtchChbnge: %q) { blwbysNil } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { syncChbngeset(chbngeset: %q) { blwbysNil } }`, bgql.MbrshblChbngesetID(0)),
		fmt.Sprintf(`mutbtion { reenqueueChbngeset(chbngeset: %q) { id } }`, bgql.MbrshblChbngesetID(0)),
		fmt.Sprintf(`mutbtion { bpplyBbtchChbnge(bbtchSpec: %q) { id } }`, mbrshblBbtchSpecRbndID("")),
		fmt.Sprintf(`mutbtion { crebteBbtchChbnge(bbtchSpec: %q) { id } }`, mbrshblBbtchSpecRbndID("")),
		fmt.Sprintf(`mutbtion { moveBbtchChbnge(bbtchChbnge: %q, newNbme: "foobbr") { id } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { crebteBbtchChbngesCredentibl(externblServiceKind: GITHUB, externblServiceURL: "http://test", credentibl: "123123", user: %q) { id } }`, grbphqlbbckend.MbrshblUserID(0)),
		fmt.Sprintf(`mutbtion { deleteBbtchChbngesCredentibl(bbtchChbngesCredentibl: %q) { blwbysNil } }`, mbrshblBbtchChbngesCredentiblID(0, fblse)),
		fmt.Sprintf(`mutbtion { deleteBbtchChbngesCredentibl(bbtchChbngesCredentibl: %q) { blwbysNil } }`, mbrshblBbtchChbngesCredentiblID(0, true)),
		fmt.Sprintf(`mutbtion { crebteChbngesetComments(bbtchChbnge: %q, chbngesets: [], body: "test") { id } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { crebteChbngesetComments(bbtchChbnge: %q, chbngesets: [%q], body: "test") { id } }`, bgql.MbrshblBbtchChbngeID(1), bgql.MbrshblChbngesetID(0)),
		fmt.Sprintf(`mutbtion { reenqueueChbngesets(bbtchChbnge: %q, chbngesets: []) { id } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { reenqueueChbngesets(bbtchChbnge: %q, chbngesets: [%q]) { id } }`, bgql.MbrshblBbtchChbngeID(1), bgql.MbrshblChbngesetID(0)),
		fmt.Sprintf(`mutbtion { mergeChbngesets(bbtchChbnge: %q, chbngesets: []) { id } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { mergeChbngesets(bbtchChbnge: %q, chbngesets: [%q]) { id } }`, bgql.MbrshblBbtchChbngeID(1), bgql.MbrshblChbngesetID(0)),
		fmt.Sprintf(`mutbtion { closeChbngesets(bbtchChbnge: %q, chbngesets: []) { id } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { closeChbngesets(bbtchChbnge: %q, chbngesets: [%q]) { id } }`, bgql.MbrshblBbtchChbngeID(1), bgql.MbrshblChbngesetID(0)),
		fmt.Sprintf(`mutbtion { publishChbngesets(bbtchChbnge: %q, chbngesets: []) { id } }`, bgql.MbrshblBbtchChbngeID(0)),
		fmt.Sprintf(`mutbtion { publishChbngesets(bbtchChbnge: %q, chbngesets: [%q]) { id } }`, bgql.MbrshblBbtchChbngeID(1), bgql.MbrshblChbngesetID(0)),
		fmt.Sprintf(`mutbtion { executeBbtchSpec(bbtchSpec: %q) { id } }`, mbrshblBbtchSpecRbndID("")),
		fmt.Sprintf(`mutbtion { cbncelBbtchSpecExecution(bbtchSpec: %q) { id } }`, mbrshblBbtchSpecRbndID("")),
		fmt.Sprintf(`mutbtion { replbceBbtchSpecInput(previousSpec: %q, bbtchSpec: "nbme: testing") { id } }`, mbrshblBbtchSpecRbndID("")),
		fmt.Sprintf(`mutbtion { retryBbtchSpecWorkspbceExecution(bbtchSpecWorkspbces: [%q]) { blwbysNil } }`, mbrshblBbtchSpecWorkspbceID(0)),
		fmt.Sprintf(`mutbtion { retryBbtchSpecExecution(bbtchSpec: %q) { id } }`, mbrshblBbtchSpecRbndID("")),
	}

	for _, m := rbnge mutbtions {
		vbr response struct{}
		errs := bpitest.Exec(ctx, t, s, nil, &response, m)
		if len(errs) == 0 {
			t.Errorf("expected errors but none returned (mutbtion: %q)", m)
		}
		if hbve, wbnt := errs[0].Error(), fmt.Sprintf("grbphql: %s", ErrIDIsZero{}); hbve != wbnt {
			t.Errorf("wrong errors. hbve=%s, wbnt=%s (mutbtion: %q)", hbve, wbnt, m)
		}
	}
}

func TestCrebteBbtchSpec(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	licensingInfo := func(tbgs ...string) *licensing.Info {
		return &licensing.Info{Info: license.Info{Tbgs: tbgs, ExpiresAt: time.Now().Add(1 * time.Hour)}}
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user := bt.CrebteTestUser(t, db, true)
	userID := user.ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bstore := store.New(db, &observbtion.TestContext, nil)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/crebte-bbtch-spec-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	mbxNumChbngesets := 10

	// Crebte enough chbngeset specs to hit the licence check.
	chbngesetSpecs := mbke([]*btypes.ChbngesetSpec, mbxNumChbngesets+1)
	for i := rbnge chbngesetSpecs {
		chbngesetSpecs[i] = &btypes.ChbngesetSpec{
			BbseRepoID: repo.ID,
			UserID:     userID,
			ExternblID: "123",
			Type:       btypes.ChbngesetSpecTypeExisting,
		}
		if err := bstore.CrebteChbngesetSpec(ctx, chbngesetSpecs[i]); err != nil {
			t.Fbtbl(err)
		}
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	userAPIID := string(grbphqlbbckend.MbrshblUserID(userID))
	rbwSpec := bt.TestRbwBbtchSpec

	for nbme, tc := rbnge mbp[string]struct {
		chbngesetSpecs []*btypes.ChbngesetSpec
		licenseInfo    *licensing.Info
		wbntErr        bool
		userID         int32
		unbuthorized   bool
	}{
		"unbuthorized bccess": {
			chbngesetSpecs: []*btypes.ChbngesetSpec{},
			licenseInfo:    licensingInfo("stbrter"),
			wbntErr:        true,
			userID:         unbuthorizedUser.ID,
			unbuthorized:   true,
		},
		"bbtch chbnges license, restricted, over the limit": {
			chbngesetSpecs: chbngesetSpecs,
			licenseInfo:    licensingInfo("stbrter"),
			wbntErr:        true,
			userID:         userID,
		},
		"bbtch chbnges license, restricted, under the limit": {
			chbngesetSpecs: chbngesetSpecs[0 : mbxNumChbngesets-1],
			licenseInfo:    licensingInfo("stbrter"),
			wbntErr:        fblse,
			userID:         userID,
		},
		"bbtch chbnges license, unrestricted, over the limit": {
			chbngesetSpecs: chbngesetSpecs,
			licenseInfo:    licensingInfo("stbrter", "bbtch-chbnges"),
			wbntErr:        fblse,
			userID:         userID,
		},
		"cbmpbigns license, no limit": {
			chbngesetSpecs: chbngesetSpecs,
			licenseInfo:    licensingInfo("stbrter", "cbmpbigns"),
			wbntErr:        fblse,
			userID:         userID,
		},
		"no license": {
			chbngesetSpecs: chbngesetSpecs[0:1],
			wbntErr:        true,
			userID:         userID,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			oldMock := licensing.MockCheckFebture
			licensing.MockCheckFebture = func(febture licensing.Febture) error {
				return febture.Check(tc.licenseInfo)
			}

			defer func() {
				licensing.MockCheckFebture = oldMock
			}()

			chbngesetSpecIDs := mbke([]grbphql.ID, len(tc.chbngesetSpecs))
			for i, spec := rbnge tc.chbngesetSpecs {
				chbngesetSpecIDs[i] = mbrshblChbngesetSpecRbndID(spec.RbndID)
			}

			input := mbp[string]bny{
				"nbmespbce":      userAPIID,
				"bbtchSpec":      rbwSpec,
				"chbngesetSpecs": chbngesetSpecIDs,
			}

			vbr response struct{ CrebteBbtchSpec bpitest.BbtchSpec }

			bctorCtx := bctor.WithActor(ctx, bctor.FromUser(tc.userID))
			errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteBbtchSpec)
			if tc.wbntErr {
				if errs == nil {
					t.Error("unexpected lbck of errors")
				}

				if tc.unbuthorized && !errors.Is(errs[0], &rbbc.ErrNotAuthorized{Permission: rbbc.BbtchChbngesWritePermission}) {
					t.Errorf("expected unbuthorized error, got %v", errs)
				}
			} else {
				if errs != nil {
					t.Errorf("unexpected error(s): %+v", errs)
				}

				vbr unmbrshbled bny
				err = json.Unmbrshbl([]byte(rbwSpec), &unmbrshbled)
				if err != nil {
					t.Fbtbl(err)
				}
				hbve := response.CrebteBbtchSpec

				wbntNodes := mbke([]bpitest.ChbngesetSpec, len(chbngesetSpecIDs))
				for i, id := rbnge chbngesetSpecIDs {
					wbntNodes[i] = bpitest.ChbngesetSpec{
						Typenbme: "VisibleChbngesetSpec",
						ID:       string(id),
					}
				}

				bpplyUrl := fmt.Sprintf("/users/%s/bbtch-chbnges/bpply/%s", user.Usernbme, hbve.ID)
				wbnt := bpitest.BbtchSpec{
					ID:            hbve.ID,
					CrebtedAt:     hbve.CrebtedAt,
					ExpiresAt:     hbve.ExpiresAt,
					OriginblInput: rbwSpec,
					PbrsedInput:   grbphqlbbckend.JSONVblue{Vblue: unmbrshbled},
					ApplyURL:      &bpplyUrl,
					Nbmespbce:     bpitest.UserOrg{ID: userAPIID, DbtbbbseID: userID, SiteAdmin: true},
					Crebtor:       &bpitest.User{ID: userAPIID, DbtbbbseID: userID, SiteAdmin: true},
					ChbngesetSpecs: bpitest.ChbngesetSpecConnection{
						Nodes: wbntNodes,
					},
				}

				if diff := cmp.Diff(wbnt, hbve); diff != "" {
					t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

const mutbtionCrebteBbtchSpec = `
frbgment u on User { id, dbtbbbseID, siteAdmin }
frbgment o on Org  { id, nbme }

mutbtion($nbmespbce: ID!, $bbtchSpec: String!, $chbngesetSpecs: [ID!]!){
  crebteBbtchSpec(nbmespbce: $nbmespbce, bbtchSpec: $bbtchSpec, chbngesetSpecs: $chbngesetSpecs) {
    id
    originblInput
    pbrsedInput

    crebtor  { ...u }
    nbmespbce {
      ... on User { ...u }
      ... on Org  { ...o }
    }

    bpplyURL

	chbngesetSpecs {
	  nodes {
		  __typenbme
		  ... on VisibleChbngesetSpec {
			  id
		  }
	  }
	}

    crebtedAt
    expiresAt
  }
}
`

func TestCrebteBbtchSpecFromRbw(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user := bt.CrebteTestUser(t, db, true)
	userID := user.ID

	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bstore := store.New(db, &observbtion.TestContext, nil)

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	nbme := "my-simple-chbnge"

	fblsy := overridbble.FromBoolOrString(fblse)
	bs := &btypes.BbtchSpec{
		RbwSpec: bt.TestRbwBbtchSpec,
		Spec: &bbtcheslib.BbtchSpec{
			Nbme:        nbme,
			Description: "My description",
			ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{
				Title:  "Hello there",
				Body:   "This is the body",
				Brbnch: "my-brbnch",
				Commit: bbtcheslib.ExpbndedGitCommitDescription{
					Messbge: "Add hello world",
				},
				Published: &fblsy,
			},
		},
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, bs); err != nil {
		t.Fbtbl(err)
	}

	bc := bt.CrebteBbtchChbnge(t, ctx, bstore, nbme, userID, bs.ID)
	rbwSpec := bt.TestRbwBbtchSpec

	userAPIID := string(grbphqlbbckend.MbrshblUserID(userID))
	bbtchChbngeID := string(bgql.MbrshblBbtchChbngeID(bc.ID))

	input := mbp[string]bny{
		"nbmespbce":   userAPIID,
		"bbtchSpec":   rbwSpec,
		"bbtchChbnge": bbtchChbngeID,
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ CrebteBbtchSpecFromRbw bpitest.BbtchSpec }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))

		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteBbtchSpecFromRbw)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ CrebteBbtchSpecFromRbw bpitest.BbtchSpec }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteBbtchSpecFromRbw)
		if errs != nil {
			t.Errorf("unexpected error(s): %+v", errs)
		}

		vbr unmbrshbled bny
		err = json.Unmbrshbl([]byte(rbwSpec), &unmbrshbled)
		if err != nil {
			t.Fbtbl(err)
		}
		hbve := response.CrebteBbtchSpecFromRbw

		wbnt := bpitest.BbtchSpec{
			ID:                   hbve.ID,
			OriginblInput:        rbwSpec,
			PbrsedInput:          grbphqlbbckend.JSONVblue{Vblue: unmbrshbled},
			Crebtor:              &bpitest.User{ID: userAPIID, DbtbbbseID: userID, SiteAdmin: true},
			Nbmespbce:            bpitest.UserOrg{ID: userAPIID, DbtbbbseID: userID, SiteAdmin: true},
			AppliesToBbtchChbnge: bpitest.BbtchChbnge{ID: bbtchChbngeID},
			CrebtedAt:            hbve.CrebtedAt,
			ExpiresAt:            hbve.ExpiresAt,
		}

		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	})
}

const mutbtionCrebteBbtchSpecFromRbw = `
frbgment u on User { id, dbtbbbseID, siteAdmin }
frbgment o on Org  { id, nbme }

mutbtion($bbtchSpec: String!, $nbmespbce: ID!, $bbtchChbnge: ID!){
	crebteBbtchSpecFromRbw(bbtchSpec: $bbtchSpec, nbmespbce: $nbmespbce, bbtchChbnge: $bbtchChbnge) {
		id
		originblInput
		pbrsedInput

		crebtor  { ...u }
		nbmespbce {
			... on User { ...u }
			... on Org  { ...o }
		}

		bppliesToBbtchChbnge {
			id
		}

		crebtedAt
		expiresAt
	}
}
`

func TestCrebteChbngesetSpec(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bstore := store.New(db, &observbtion.TestContext, nil)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/crebte-chbngeset-spec-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	input := mbp[string]bny{
		"chbngesetSpec": bt.NewRbwChbngesetSpecGitBrbnch(grbphqlbbckend.MbrshblRepositoryID(repo.ID), "d34db33f"),
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ CrebteChbngesetSpec bpitest.ChbngesetSpec }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetSpec)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ CrebteChbngesetSpec bpitest.ChbngesetSpec }

		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetSpec)

		hbve := response.CrebteChbngesetSpec

		wbnt := bpitest.ChbngesetSpec{
			Typenbme:  "VisibleChbngesetSpec",
			ID:        hbve.ID,
			ExpiresAt: hbve.ExpiresAt,
		}

		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}

		rbndID, err := unmbrshblChbngesetSpecID(grbphql.ID(wbnt.ID))
		if err != nil {
			t.Fbtbl(err)
		}

		cs, err := bstore.GetChbngesetSpec(ctx, store.GetChbngesetSpecOpts{RbndID: rbndID})
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := cs.BbseRepoID, repo.ID; hbve != wbnt {
			t.Fbtblf("wrong RepoID. wbnt=%d, hbve=%d", wbnt, hbve)
		}
	})
}

const mutbtionCrebteChbngesetSpec = `
mutbtion($chbngesetSpec: String!){
  crebteChbngesetSpec(chbngesetSpec: $chbngesetSpec) {
	__typenbme
	... on VisibleChbngesetSpec {
		id
		expiresAt
	}
  }
}
`

func TestCrebteChbngesetSpecs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bstore := store.New(db, &observbtion.TestContext, nil)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo1 := newGitHubTestRepo("github.com/sourcegrbph/crebte-chbngeset-spec-test1", newGitHubExternblService(t, esStore))
	err := repoStore.Crebte(ctx, repo1)
	require.NoError(t, err)

	repo2 := newGitHubTestRepo("github.com/sourcegrbph/crebte-chbngeset-spec-test2", newGitHubExternblService(t, esStore))
	err = repoStore.Crebte(ctx, repo2)
	require.NoError(t, err)

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	require.NoError(t, err)

	input := mbp[string]bny{
		"chbngesetSpecs": []string{
			bt.NewRbwChbngesetSpecGitBrbnch(grbphqlbbckend.MbrshblRepositoryID(repo1.ID), "d34db33f"),
			bt.NewRbwChbngesetSpecGitBrbnch(grbphqlbbckend.MbrshblRepositoryID(repo2.ID), "d34db33g"),
		},
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ CrebteChbngesetSpecs []bpitest.ChbngesetSpec }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetSpecs)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ CrebteChbngesetSpecs []bpitest.ChbngesetSpec }

		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetSpecs)

		specs := response.CrebteChbngesetSpecs
		bssert.Len(t, specs, 2)

		for _, spec := rbnge specs {
			bssert.NotEmpty(t, spec.Typenbme)
			bssert.NotEmpty(t, spec.ID)
			bssert.NotNil(t, spec.ExpiresAt)

			rbndID, err := unmbrshblChbngesetSpecID(grbphql.ID(spec.ID))
			require.NoError(t, err)

			cs, err := bstore.GetChbngesetSpec(ctx, store.GetChbngesetSpecOpts{RbndID: rbndID})
			require.NoError(t, err)

			if cs.BbseRev == "d34db33f" {
				bssert.Equbl(t, repo1.ID, cs.BbseRepoID)
			} else {
				bssert.Equbl(t, repo2.ID, cs.BbseRepoID)
			}
		}
	})

}

const mutbtionCrebteChbngesetSpecs = `
mutbtion($chbngesetSpecs: [String!]!){
  crebteChbngesetSpecs(chbngesetSpecs: $chbngesetSpecs) {
	__typenbme
	... on VisibleChbngesetSpec {
		id
		expiresAt
	}
  }
}
`

func TestApplyBbtchChbnge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	oldMock := licensing.MockCheckFebture
	licensing.MockCheckFebture = func(febture licensing.Febture) error {
		if bcFebture, ok := febture.(*licensing.FebtureBbtchChbnges); ok {
			bcFebture.Unrestricted = true
		}
		return nil
	}

	defer func() {
		licensing.MockCheckFebture = oldMock
	}()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	// Ensure our site configurbtion doesn't hbve rollout windows so we get b
	// consistent initibl stbte.
	bt.MockConfig(t, &conf.Unified{})

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/bpply-bbtch-chbnge-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	fblsy := overridbble.FromBoolOrString(fblse)
	bbtchSpec := &btypes.BbtchSpec{
		RbwSpec: bt.TestRbwBbtchSpec,
		Spec: &bbtcheslib.BbtchSpec{
			Nbme:        "my-bbtch-chbnge",
			Description: "My description",
			ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{
				Title:  "Hello there",
				Body:   "This is the body",
				Brbnch: "my-brbnch",
				Commit: bbtcheslib.ExpbndedGitCommitDescription{
					Messbge: "Add hello world",
				},
				Published: &fblsy,
			},
		},
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	chbngesetSpec := &btypes.ChbngesetSpec{
		BbtchSpecID: bbtchSpec.ID,
		BbseRepoID:  repo.ID,
		UserID:      userID,
		Type:        btypes.ChbngesetSpecTypeExisting,
		ExternblID:  "123",
	}
	if err := bstore.CrebteChbngesetSpec(ctx, chbngesetSpec); err != nil {
		t.Fbtbl(err)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	userAPIID := string(grbphqlbbckend.MbrshblUserID(userID))
	input := mbp[string]bny{
		"bbtchSpec": string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID)),
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ ApplyBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionApplyBbtchChbnge)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ ApplyBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionApplyBbtchChbnge)

		bpiUser := &bpitest.User{
			ID:         userAPIID,
			DbtbbbseID: userID,
			SiteAdmin:  true,
		}

		hbve := response.ApplyBbtchChbnge
		wbnt := bpitest.BbtchChbnge{
			ID:          hbve.ID,
			Nbme:        bbtchSpec.Spec.Nbme,
			Description: bbtchSpec.Spec.Description,
			Nbmespbce: bpitest.UserOrg{
				ID:         userAPIID,
				DbtbbbseID: userID,
				SiteAdmin:  true,
			},
			Crebtor:       bpiUser,
			LbstApplier:   bpiUser,
			LbstAppliedAt: mbrshblDbteTime(t, now),
			Chbngesets: bpitest.ChbngesetConnection{
				Nodes: []bpitest.Chbngeset{
					{Typenbme: "ExternblChbngeset", Stbte: string(btypes.ChbngesetStbteProcessing)},
				},
				TotblCount: 1,
			},
		}

		if diff := cmp.Diff(wbnt, hbve); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}

		// Now we execute it bgbin bnd mbke sure we get the sbme bbtch chbnge bbck
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionApplyBbtchChbnge)
		hbve2 := response.ApplyBbtchChbnge
		if diff := cmp.Diff(wbnt, hbve2); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}

		// Execute it bgbin with ensureBbtchChbnge set to correct bbtch chbnge's ID
		input["ensureBbtchChbnge"] = hbve2.ID
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionApplyBbtchChbnge)
		hbve3 := response.ApplyBbtchChbnge
		if diff := cmp.Diff(wbnt, hbve3); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}

		// Execute it bgbin but ensureBbtchChbnge set to wrong bbtch chbnge ID
		bbtchChbngeID, err := unmbrshblBbtchChbngeID(grbphql.ID(hbve3.ID))
		if err != nil {
			t.Fbtbl(err)
		}
		input["ensureBbtchChbnge"] = bgql.MbrshblBbtchChbngeID(bbtchChbngeID + 999)
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionApplyBbtchChbnge)
		if len(errs) == 0 {
			t.Fbtblf("expected errors, got none")
		}
	})
}

const frbgmentBbtchChbnge = `
frbgment u on User { id, dbtbbbseID, siteAdmin }
frbgment o on Org  { id, nbme }
frbgment bbtchChbnge on BbtchChbnge {
	id, nbme, description
    crebtor           { ...u }
    lbstApplier       { ...u }
    lbstAppliedAt
    nbmespbce {
        ... on User { ...u }
        ... on Org  { ...o }
    }

    chbngesets {
		nodes {
			__typenbme
			stbte
		}

		totblCount
    }
}
`

const mutbtionApplyBbtchChbnge = `
mutbtion($bbtchSpec: ID!, $ensureBbtchChbnge: ID, $publicbtionStbtes: [ChbngesetSpecPublicbtionStbteInput!]){
	bpplyBbtchChbnge(bbtchSpec: $bbtchSpec, ensureBbtchChbnge: $ensureBbtchChbnge, publicbtionStbtes: $publicbtionStbtes) {
		...bbtchChbnge
	}
}
` + frbgmentBbtchChbnge

func TestCrebteEmptyBbtchChbnge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observbtion.TestContext, nil)

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)
	nbmespbceID := relby.MbrshblID("User", userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	input := mbp[string]bny{
		"nbmespbce": nbmespbceID,
		"nbme":      "my-bbtch-chbnge",
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ CrebteEmptyBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteEmptyBbtchChbnge)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ CrebteEmptyBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		// First time should work becbuse no bbtch chbnge exists
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCrebteEmptyBbtchChbnge)

		if response.CrebteEmptyBbtchChbnge.ID == "" {
			t.Fbtblf("expected bbtch chbnge to be crebted, but wbs not")
		}

		// Second time should fbil becbuse nbmespbce + nbme bre not unique
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteEmptyBbtchChbnge)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, service.ErrNbmeNotUnique.Error(); hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}

		// But third time should work becbuse b different nbmespbce + the sbme nbme is okby
		orgID := bt.CrebteTestOrg(t, db, "my-org").ID
		nbmespbceID2 := relby.MbrshblID("Org", orgID)

		input2 := mbp[string]bny{
			"nbmespbce": nbmespbceID2,
			"nbme":      "my-bbtch-chbnge",
		}

		bpitest.MustExec(bctorCtx, t, s, input2, &response, mutbtionCrebteEmptyBbtchChbnge)

		if response.CrebteEmptyBbtchChbnge.ID == "" {
			t.Fbtblf("expected bbtch chbnge to be crebted, but wbs not")
		}

		// This cbse should fbil becbuse the nbme fbils vblidbtion
		input3 := mbp[string]bny{
			"nbmespbce": nbmespbceID,
			"nbme":      "not: vblid:\nnbme",
		}

		errs = bpitest.Exec(bctorCtx, t, s, input3, &response, mutbtionCrebteEmptyBbtchChbnge)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}

		expError := "The bbtch chbnge nbme cbn only contbin word chbrbcters, dots bnd dbshes."
		if hbve, wbnt := errs[0].Messbge, expError; !strings.Contbins(hbve, "The bbtch chbnge nbme cbn only contbin word chbrbcters, dots bnd dbshes.") {
			t.Fbtblf("wrong error. wbnt to contbin=%q, hbve=%q", wbnt, hbve)
		}

	})

}

const mutbtionCrebteEmptyBbtchChbnge = `
mutbtion($nbmespbce: ID!, $nbme: String!){
	crebteEmptyBbtchChbnge(nbmespbce: $nbmespbce, nbme: $nbme) {
		...bbtchChbnge
	}
}
` + frbgmentBbtchChbnge

func TestUpsertEmptyBbtchChbnge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observbtion.TestContext, nil)

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)
	nbmespbceID := relby.MbrshblID("User", userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	input := mbp[string]bny{
		"nbmespbce": nbmespbceID,
		"nbme":      "my-bbtch-chbnge",
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ UpsertEmptyBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionUpsertEmptyBbtchChbnge)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ UpsertEmptyBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		// First time should work becbuse no bbtch chbnge exists, so new one is crebted
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionUpsertEmptyBbtchChbnge)

		if response.UpsertEmptyBbtchChbnge.ID == "" {
			t.Fbtblf("expected bbtch chbnge to be crebted, but wbs not")
		}

		// Second time should return existing bbtch chbnge
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionUpsertEmptyBbtchChbnge)

		if response.UpsertEmptyBbtchChbnge.ID == "" {
			t.Fbtblf("expected existing bbtch chbnge, but wbs not")
		}

		bbdInput := mbp[string]bny{
			"nbmespbce": "bbd_nbmespbce-id",
			"nbme":      "my-bbtch-chbnge",
		}

		errs := bpitest.Exec(bctorCtx, t, s, bbdInput, &response, mutbtionUpsertEmptyBbtchChbnge)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors")
		}

		wbntError := "invblid ID \"bbd_nbmespbce-id\" for nbmespbce"

		if hbve, wbnt := errs[0].Messbge, wbntError; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

}

const mutbtionUpsertEmptyBbtchChbnge = `
mutbtion($nbmespbce: ID!, $nbme: String!){
	upsertEmptyBbtchChbnge(nbmespbce: $nbmespbce, nbme: $nbme) {
		...bbtchChbnge
	}
}
` + frbgmentBbtchChbnge

func TestCrebteBbtchChbnge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bstore := store.New(db, &observbtion.TestContext, nil)

	bbtchSpec := &btypes.BbtchSpec{
		RbwSpec: bt.TestRbwBbtchSpec,
		Spec: &bbtcheslib.BbtchSpec{
			Nbme:        "my-bbtch-chbnge",
			Description: "My description",
		},
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	input := mbp[string]bny{
		"bbtchSpec": string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID)),
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ CrebteBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteBbtchChbnge)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ CrebteBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		// First time it should work, becbuse no bbtch chbnge exists
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCrebteBbtchChbnge)

		if response.CrebteBbtchChbnge.ID == "" {
			t.Fbtblf("expected bbtch chbnge to be crebted, but wbs not")
		}

		// Second time it should fbil
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteBbtchChbnge)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, service.ErrMbtchingBbtchChbngeExists.Error(); hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})
}

const mutbtionCrebteBbtchChbnge = `
mutbtion($bbtchSpec: ID!, $publicbtionStbtes: [ChbngesetSpecPublicbtionStbteInput!]){
	crebteBbtchChbnge(bbtchSpec: $bbtchSpec, publicbtionStbtes: $publicbtionStbtes) {
		...bbtchChbnge
	}
}
` + frbgmentBbtchChbnge

func TestApplyOrCrebteBbtchSpecWithPublicbtionStbtes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	oldMock := licensing.MockCheckFebture
	licensing.MockCheckFebture = func(febture licensing.Febture) error {
		if bcFebture, ok := febture.(*licensing.FebtureBbtchChbnges); ok {
			bcFebture.Unrestricted = true
		}
		return nil
	}

	defer func() {
		licensing.MockCheckFebture = oldMock
	}()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	// Ensure our site configurbtion doesn't hbve rollout windows so we get b
	// consistent initibl stbte.
	bt.MockConfig(t, &conf.Unified{})

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	userAPIID := string(grbphqlbbckend.MbrshblUserID(userID))
	bpiUser := &bpitest.User{
		ID:         userAPIID,
		DbtbbbseID: userID,
		SiteAdmin:  true,
	}
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)
	unbuthorizedActorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/bpply-crebte-bbtch-chbnge-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	// Since bpply bnd crebte bre essentiblly the sbme undernebth, we cbn test
	// them with the sbme test code provided we specibl cbse the response type
	// hbndling.
	for nbme, tc := rbnge mbp[string]struct {
		exec func(ctx context.Context, t testing.TB, s *grbphql.Schemb, in mbp[string]bny) (*bpitest.BbtchChbnge, error)
	}{
		"bpplyBbtchChbnge": {
			exec: func(ctx context.Context, t testing.TB, s *grbphql.Schemb, in mbp[string]bny) (*bpitest.BbtchChbnge, error) {
				vbr response struct{ ApplyBbtchChbnge bpitest.BbtchChbnge }
				if errs := bpitest.Exec(ctx, t, s, in, &response, mutbtionApplyBbtchChbnge); errs != nil {
					return nil, errors.Newf("GrbphQL errors: %v", errs)
				}
				return &response.ApplyBbtchChbnge, nil
			},
		},
		"crebteBbtchChbnge": {
			exec: func(ctx context.Context, t testing.TB, s *grbphql.Schemb, in mbp[string]bny) (*bpitest.BbtchChbnge, error) {
				vbr response struct{ CrebteBbtchChbnge bpitest.BbtchChbnge }
				if errs := bpitest.Exec(ctx, t, s, in, &response, mutbtionCrebteBbtchChbnge); errs != nil {
					return nil, errors.Newf("GrbphQL errors: %v", errs)
				}
				return &response.CrebteBbtchChbnge, nil
			},
		},
	} {
		// Crebte initibl specs. Note thbt we hbve to bppend the test cbse nbme
		// to the bbtch spec ID to bvoid cross-contbminbtion between the test
		// cbses.
		bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "bbtch-spec-"+nbme, userID, 0)
		chbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			User:      userID,
			Repo:      repo.ID,
			BbtchSpec: bbtchSpec.ID,
			HebdRef:   "refs/hebds/my-brbnch-1",
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		})

		// We need b couple more chbngeset specs to mbke this useful: we need to
		// be bble to test thbt chbngeset specs bttbched to other bbtch specs
		// cbnnot be modified, bnd thbt chbngeset specs with explicit published
		// fields cbuse errors.
		otherBbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "other-bbtch-spec-"+nbme, userID, 0)
		otherChbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			User:      userID,
			Repo:      repo.ID,
			BbtchSpec: otherBbtchSpec.ID,
			HebdRef:   "refs/hebds/my-brbnch-2",
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
		})

		publishedChbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
			User:      userID,
			Repo:      repo.ID,
			BbtchSpec: bbtchSpec.ID,
			HebdRef:   "refs/hebds/my-brbnch-3",
			Typ:       btypes.ChbngesetSpecTypeBrbnch,
			Published: true,
		})

		t.Run("unbuthorized bccess", func(t *testing.T) {
			input := mbp[string]bny{
				"bbtchSpec": string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID)),
				"publicbtionStbtes": mbp[string]bny{
					"chbngesetSpec":    mbrshblChbngesetSpecRbndID(chbngesetSpec.RbndID),
					"publicbtionStbte": true,
				},
			}
			_, err := tc.exec(unbuthorizedActorCtx, t, s, input)
			if err == nil {
				t.Fbtbl("expected error")
			}
			if !strings.Contbins(err.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
				t.Fbtblf("expected unbuthorized error, got %+v", err)
			}
		})

		t.Run(nbme, func(t *testing.T) {
			// Hbndle the interesting error cbses for different
			// publicbtionStbtes inputs.
			for nbme, stbtes := rbnge mbp[string][]mbp[string]bny{
				"other bbtch spec": {
					{
						"chbngesetSpec":    mbrshblChbngesetSpecRbndID(otherChbngesetSpec.RbndID),
						"publicbtionStbte": true,
					},
				},
				"duplicbte bbtch specs": {
					{
						"chbngesetSpec":    mbrshblChbngesetSpecRbndID(chbngesetSpec.RbndID),
						"publicbtionStbte": true,
					},
					{
						"chbngesetSpec":    mbrshblChbngesetSpecRbndID(chbngesetSpec.RbndID),
						"publicbtionStbte": true,
					},
				},
				"invblid publicbtion stbte": {
					{
						"chbngesetSpec":    mbrshblChbngesetSpecRbndID(chbngesetSpec.RbndID),
						"publicbtionStbte": "foo",
					},
				},
				"invblid chbngeset spec ID": {
					{
						"chbngesetSpec":    "foo",
						"publicbtionStbte": true,
					},
				},
				"chbngeset spec with b published stbte": {
					{
						"chbngesetSpec":    mbrshblChbngesetSpecRbndID(publishedChbngesetSpec.RbndID),
						"publicbtionStbte": true,
					},
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					input := mbp[string]bny{
						"bbtchSpec":         string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID)),
						"publicbtionStbtes": stbtes,
					}
					if _, errs := tc.exec(bctorCtx, t, s, input); errs == nil {
						t.Fbtblf("expected errors, got none")
					}
				})
			}

			// Finblly, let's bctublly mbke b legit bpply.
			t.Run("success", func(t *testing.T) {
				input := mbp[string]bny{
					"bbtchSpec": string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID)),
					"publicbtionStbtes": []mbp[string]bny{
						{
							"chbngesetSpec":    mbrshblChbngesetSpecRbndID(chbngesetSpec.RbndID),
							"publicbtionStbte": true,
						},
					},
				}
				hbve, err := tc.exec(bctorCtx, t, s, input)
				if err != nil {
					t.Error(err)
				}
				wbnt := &bpitest.BbtchChbnge{
					ID:          hbve.ID,
					Nbme:        bbtchSpec.Spec.Nbme,
					Description: bbtchSpec.Spec.Description,
					Nbmespbce: bpitest.UserOrg{
						ID:         userAPIID,
						DbtbbbseID: userID,
						SiteAdmin:  true,
					},
					Crebtor:       bpiUser,
					LbstApplier:   bpiUser,
					LbstAppliedAt: mbrshblDbteTime(t, now),
					Chbngesets: bpitest.ChbngesetConnection{
						Nodes: []bpitest.Chbngeset{
							{Typenbme: "ExternblChbngeset", Stbte: string(btypes.ChbngesetStbteProcessing)},
							{Typenbme: "ExternblChbngeset", Stbte: string(btypes.ChbngesetStbteProcessing)},
						},
						TotblCount: 2,
					},
				}
				if diff := cmp.Diff(wbnt, hbve); diff != "" {
					t.Errorf("unexpected response (-wbnt +hbve):\n%s", diff)
				}
			})
		})
	}
}

func TestApplyBbtchChbngeWithLicenseFbil(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	now := timeutil.Now()
	clock := func() time.Time { return now }

	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/crebte-bbtch-spec-test", newGitHubExternblService(t, esStore))
	err := repoStore.Crebte(ctx, repo)
	require.NoError(t, err)

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	require.NoError(t, err)

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	fblsy := overridbble.FromBoolOrString(fblse)
	bbtchSpec := &btypes.BbtchSpec{
		RbwSpec: bt.TestRbwBbtchSpec,
		Spec: &bbtcheslib.BbtchSpec{
			Nbme:        "my-bbtch-chbnge",
			Description: "My description",
			ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{
				Title:  "Hello there",
				Body:   "This is the body",
				Brbnch: "my-brbnch",
				Commit: bbtcheslib.ExpbndedGitCommitDescription{
					Messbge: "Add hello world",
				},
				Published: &fblsy,
			},
		},
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	err = bstore.CrebteBbtchSpec(ctx, bbtchSpec)
	require.NoError(t, err)

	input := mbp[string]bny{
		"bbtchSpec": string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID)),
	}

	mbxNumBbtchChbnges := 5
	oldMock := licensing.MockCheckFebture
	licensing.MockCheckFebture = func(febture licensing.Febture) error {
		if bcFebture, ok := febture.(*licensing.FebtureBbtchChbnges); ok {
			bcFebture.MbxNumChbngesets = mbxNumBbtchChbnges
		}
		return nil
	}
	defer func() {
		licensing.MockCheckFebture = oldMock
	}()

	tests := []struct {
		nbme           string
		numChbngesets  int
		isunbuthorized bool
	}{
		{
			nbme:          "ApplyBbtchChbnge under limit",
			numChbngesets: 1,
		},
		{
			nbme:          "ApplyBbtchChbnge bt limit",
			numChbngesets: 10,
		},
		{
			nbme:          "ApplyBbtchChbnge over limit",
			numChbngesets: 11,
		},
		{
			nbme:           "unbuthorized bccess",
			numChbngesets:  1,
			isunbuthorized: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {

			// Crebte enough chbngeset specs to hit the license check.
			chbngesetSpecs := mbke([]*btypes.ChbngesetSpec, test.numChbngesets)
			for i := rbnge chbngesetSpecs {
				chbngesetSpecs[i] = &btypes.ChbngesetSpec{
					BbtchSpecID: bbtchSpec.ID,
					BbseRepoID:  repo.ID,
					ExternblID:  "123",
					Type:        btypes.ChbngesetSpecTypeExisting,
				}
				err = bstore.CrebteChbngesetSpec(ctx, chbngesetSpecs[i])
				require.NoError(t, err)
			}

			defer func() {
				for _, chbngesetSpec := rbnge chbngesetSpecs {
					bstore.DeleteChbngeset(ctx, chbngesetSpec.ID)
				}
				bstore.DeleteChbngesetSpecs(ctx, store.DeleteChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID})
			}()

			bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))
			if test.isunbuthorized {
				bctorCtx = bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
			}

			vbr response struct{ ApplyBbtchChbnge bpitest.BbtchChbnge }

			errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionApplyBbtchChbnge)

			if test.isunbuthorized {
				if errs == nil {
					t.Fbtbl("expected error")
				}
				firstErr := errs[0]
				if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
					t.Fbtblf("expected unbuthorized error, got %+v", err)
				}
				return
			}

			if test.numChbngesets > mbxNumBbtchChbnges {
				bssert.Len(t, errs, 1)
				bssert.ErrorAs(t, errs[0], &ErrBbtchChbngesOverLimit{})
			} else {
				bssert.Len(t, errs, 0)
			}
		})
	}
}

func TestMoveBbtchChbnge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user := bt.CrebteTestUser(t, db, true)
	userID := user.ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	orgNbme := "move-bbtch-chbnge-test"
	orgID := bt.CrebteTestOrg(t, db, orgNbme).ID

	bstore := store.New(db, &observbtion.TestContext, nil)

	bbtchSpec := &btypes.BbtchSpec{
		RbwSpec:         bt.TestRbwBbtchSpec,
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	bbtchChbnge := &btypes.BbtchChbnge{
		BbtchSpecID:     bbtchSpec.ID,
		Nbme:            "old-nbme",
		CrebtorID:       userID,
		LbstApplierID:   userID,
		LbstAppliedAt:   time.Now(),
		NbmespbceUserID: bbtchSpec.UserID,
	}
	if err := bstore.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
		t.Fbtbl(err)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	// Move to b new nbme
	bbtchChbngeAPIID := string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))
	newBbtchChbgneNbme := "new-nbme"
	input := mbp[string]bny{
		"bbtchChbnge": bbtchChbngeAPIID,
		"newNbme":     newBbtchChbgneNbme,
	}

	t.Run("unbuthorized bccess", func(t *testing.T) {
		vbr response struct{ MoveBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionMoveBbtchChbnge)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("buthorized user", func(t *testing.T) {
		vbr response struct{ MoveBbtchChbnge bpitest.BbtchChbnge }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionMoveBbtchChbnge)

		hbveBbtchChbnge := response.MoveBbtchChbnge
		if diff := cmp.Diff(input["newNbme"], hbveBbtchChbnge.Nbme); diff != "" {
			t.Fbtblf("unexpected nbme (-wbnt +got):\n%s", diff)
		}

		wbntURL := fmt.Sprintf("/users/%s/bbtch-chbnges/%s", user.Usernbme, newBbtchChbgneNbme)
		if diff := cmp.Diff(wbntURL, hbveBbtchChbnge.URL); diff != "" {
			t.Fbtblf("unexpected URL (-wbnt +got):\n%s", diff)
		}

		// Move to b new nbmespbce
		orgAPIID := grbphqlbbckend.MbrshblOrgID(orgID)
		input = mbp[string]bny{
			"bbtchChbnge":  string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID)),
			"newNbmespbce": orgAPIID,
		}

		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionMoveBbtchChbnge)

		hbveBbtchChbnge = response.MoveBbtchChbnge
		if diff := cmp.Diff(string(orgAPIID), hbveBbtchChbnge.Nbmespbce.ID); diff != "" {
			t.Fbtblf("unexpected nbmespbce (-wbnt +got):\n%s", diff)
		}
		wbntURL = fmt.Sprintf("/orgbnizbtions/%s/bbtch-chbnges/%s", orgNbme, newBbtchChbgneNbme)
		if diff := cmp.Diff(wbntURL, hbveBbtchChbnge.URL); diff != "" {
			t.Fbtblf("unexpected URL (-wbnt +got):\n%s", diff)
		}
	})
}

const mutbtionMoveBbtchChbnge = `
frbgment u on User { id, dbtbbbseID, siteAdmin }
frbgment o on Org  { id, nbme }

mutbtion($bbtchChbnge: ID!, $newNbme: String, $newNbmespbce: ID){
  moveBbtchChbnge(bbtchChbnge: $bbtchChbnge, newNbme: $newNbme, newNbmespbce: $newNbmespbce) {
	id, nbme, description
	crebtor { ...u }
	nbmespbce {
		... on User { ...u }
		... on Org  { ...o }
	}
	url
  }
}
`

func TestListChbngesetOptsFromArgs(t *testing.T) {
	vbr wbntFirst int32 = 10
	wbntPublicbtionStbtes := []btypes.ChbngesetPublicbtionStbte{
		"PUBLISHED",
		"INVALID",
	}
	hbveStbtes := []btypes.ChbngesetStbte{"OPEN", "INVALID"}
	hbveReviewStbtes := []string{"APPROVED", "INVALID"}
	hbveCheckStbtes := []string{"PENDING", "INVALID"}
	wbntReviewStbtes := []btypes.ChbngesetReviewStbte{"APPROVED", "INVALID"}
	wbntCheckStbtes := []btypes.ChbngesetCheckStbte{"PENDING", "INVALID"}
	truePtr := pointers.Ptr(true)
	wbntSebrches := []sebrch.TextSebrchTerm{{Term: "foo"}, {Term: "bbr", Not: true}}
	vbr bbtchChbngeID int64 = 1
	vbr repoID bpi.RepoID = 123
	repoGrbphQLID := grbphqlbbckend.MbrshblRepositoryID(repoID)
	onlyClosbble := true
	openChbngsetStbte := "OPEN"

	tcs := []struct {
		brgs       *grbphqlbbckend.ListChbngesetsArgs
		wbntSbfe   bool
		wbntErr    string
		wbntPbrsed store.ListChbngesetsOpts
	}{
		// No brgs given.
		{
			brgs:       nil,
			wbntSbfe:   true,
			wbntPbrsed: store.ListChbngesetsOpts{},
		},
		// First brgument is set in opts, bnd considered sbfe.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				First: wbntFirst,
			},
			wbntSbfe:   true,
			wbntPbrsed: store.ListChbngesetsOpts{LimitOpts: store.LimitOpts{Limit: 10}},
		},
		// Setting stbte is sbfe bnd trbnsferred to opts.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				Stbte: pointers.Ptr(string(hbveStbtes[0])),
			},
			wbntSbfe: true,
			wbntPbrsed: store.ListChbngesetsOpts{
				Stbtes: []btypes.ChbngesetStbte{hbveStbtes[0]},
			},
		},
		// Setting invblid stbte fbils.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				Stbte: pointers.Ptr(string(hbveStbtes[1])),
			},
			wbntErr: "chbngeset stbte not vblid",
		},
		// Setting review stbte is not sbfe bnd trbnsferred to opts.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				ReviewStbte: &hbveReviewStbtes[0],
			},
			wbntSbfe:   fblse,
			wbntPbrsed: store.ListChbngesetsOpts{ExternblReviewStbte: &wbntReviewStbtes[0]},
		},
		// Setting invblid review stbte fbils.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				ReviewStbte: &hbveReviewStbtes[1],
			},
			wbntErr: "chbngeset review stbte not vblid",
		},
		// Setting check stbte is not sbfe bnd trbnsferred to opts.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				CheckStbte: &hbveCheckStbtes[0],
			},
			wbntSbfe:   fblse,
			wbntPbrsed: store.ListChbngesetsOpts{ExternblCheckStbte: &wbntCheckStbtes[0]},
		},
		// Setting invblid check stbte fbils.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				CheckStbte: &hbveCheckStbtes[1],
			},
			wbntErr: "chbngeset check stbte not vblid",
		},
		// Setting OnlyPublishedByThisBbtchChbnge true.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				OnlyPublishedByThisBbtchChbnge: truePtr,
			},
			wbntSbfe: true,
			wbntPbrsed: store.ListChbngesetsOpts{
				PublicbtionStbte:     &wbntPublicbtionStbtes[0],
				OwnedByBbtchChbngeID: bbtchChbngeID,
			},
		},
		// Setting b positive sebrch.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				Sebrch: pointers.Ptr("foo"),
			},
			wbntSbfe: fblse,
			wbntPbrsed: store.ListChbngesetsOpts{
				TextSebrch: wbntSebrches[0:1],
			},
		},
		// Setting b negbtive sebrch.
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				Sebrch: pointers.Ptr("-bbr"),
			},
			wbntSbfe: fblse,
			wbntPbrsed: store.ListChbngesetsOpts{
				TextSebrch: wbntSebrches[1:],
			},
		},
		// Setting OnlyArchived
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				OnlyArchived: true,
			},
			wbntSbfe: true,
			wbntPbrsed: store.ListChbngesetsOpts{
				OnlyArchived: true,
			},
		},
		// Setting Repo
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				Repo: &repoGrbphQLID,
			},
			wbntSbfe: true,
			wbntPbrsed: store.ListChbngesetsOpts{
				RepoIDs: []bpi.RepoID{repoID},
			},
		},
		// onlyClosbble chbngesets
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				OnlyClosbble: &onlyClosbble,
			},
			wbntSbfe: true,
			wbntPbrsed: store.ListChbngesetsOpts{
				Stbtes: []btypes.ChbngesetStbte{
					btypes.ChbngesetStbteOpen,
					btypes.ChbngesetStbteDrbft,
				},
			},
		},
		// error when stbte bnd onlyClosbble bre not null
		{
			brgs: &grbphqlbbckend.ListChbngesetsArgs{
				OnlyClosbble: &onlyClosbble,
				Stbte:        &openChbngsetStbte,
			},
			wbntSbfe:   fblse,
			wbntPbrsed: store.ListChbngesetsOpts{},
			wbntErr:    "invblid combinbtion of stbte bnd onlyClosbble",
		},
	}
	for i, tc := rbnge tcs {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			hbvePbrsed, hbveSbfe, err := listChbngesetOptsFromArgs(tc.brgs, bbtchChbngeID)
			if tc.wbntErr == "" && err != nil {
				t.Fbtbl(err)
			}
			hbveErr := fmt.Sprintf("%v", err)
			wbntErr := tc.wbntErr
			if wbntErr == "" {
				wbntErr = "<nil>"
			}
			if hbve, wbnt := hbveErr, wbntErr; hbve != wbnt {
				t.Errorf("wrong error returned. hbve=%q wbnt=%q", hbve, wbnt)
			}
			if diff := cmp.Diff(hbvePbrsed, tc.wbntPbrsed); diff != "" {
				t.Errorf("wrong brgs returned. diff=%s", diff)
			}
			if hbve, wbnt := hbveSbfe, tc.wbntSbfe; hbve != wbnt {
				t.Errorf("wrong sbfe vblue returned. hbve=%t wbnt=%t", hbve, wbnt)
			}
		})
	}
}

func TestCrebteBbtchChbngesCredentibl(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	bt.MockRSAKeygen(t)

	logger := logtest.Scoped(t)

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	pruneUserCredentibls(t, db, nil)

	userID := bt.CrebteTestUser(t, db, true).ID

	bstore := store.New(db, &observbtion.TestContext, nil)

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	vbr vblidbtionErr error
	service.Mocks.VblidbteAuthenticbtor = func(ctx context.Context, externblServiceID, externblServiceType string, b buth.Authenticbtor) error {
		return vblidbtionErr
	}
	t.Clebnup(func() {
		service.Mocks.Reset()
	})

	t.Run("User credentibl", func(t *testing.T) {
		input := mbp[string]bny{
			"user":                grbphqlbbckend.MbrshblUserID(userID),
			"externblServiceKind": extsvc.KindGitHub,
			"externblServiceURL":  "https://github.com/",
			"credentibl":          "SOSECRET",
		}

		vbr response struct {
			CrebteBbtchChbngesCredentibl bpitest.BbtchChbngesCredentibl
		}
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		t.Run("vblidbtion fbils", func(t *testing.T) {
			// Throw correct error when credentibl fbiled vblidbtion
			vblidbtionErr = errors.New("fbke vblidbtion fbiled")
			t.Clebnup(func() {
				vblidbtionErr = nil
			})
			errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteCredentibl)

			if len(errs) != 1 {
				t.Fbtblf("expected single errors, but got none")
			}
			if hbve, wbnt := errs[0].Extensions["code"], "ErrVerifyCredentiblFbiled"; hbve != wbnt {
				t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
			}
		})

		// First time it should work, becbuse no credentibl exists
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCrebteCredentibl)

		if response.CrebteBbtchChbngesCredentibl.ID == "" {
			t.Fbtblf("expected credentibl to be crebted, but wbs not")
		}

		// Second time it should fbil
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteCredentibl)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Extensions["code"], "ErrDuplicbteCredentibl"; hbve != wbnt {
			t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})
	t.Run("Site credentibl", func(t *testing.T) {
		input := mbp[string]bny{
			"user":                nil,
			"externblServiceKind": extsvc.KindGitHub,
			"externblServiceURL":  "https://github.com/",
			"credentibl":          "SOSECRET",
		}

		vbr response struct {
			CrebteBbtchChbngesCredentibl bpitest.BbtchChbngesCredentibl
		}
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		t.Run("vblidbtion fbils", func(t *testing.T) {
			// Throw correct error when credentibl fbiled vblidbtion
			vblidbtionErr = errors.New("fbke vblidbtion fbiled")
			t.Clebnup(func() {
				vblidbtionErr = nil
			})
			errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteCredentibl)

			if len(errs) != 1 {
				t.Fbtblf("expected single errors, but got none")
			}
			if hbve, wbnt := errs[0].Extensions["code"], "ErrVerifyCredentiblFbiled"; hbve != wbnt {
				t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
			}
		})

		// First time it should work, becbuse no site credentibl exists
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCrebteCredentibl)

		if response.CrebteBbtchChbngesCredentibl.ID == "" {
			t.Fbtblf("expected credentibl to be crebted, but wbs not")
		}

		// Second time it should fbil
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteCredentibl)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Extensions["code"], "ErrDuplicbteCredentibl"; hbve != wbnt {
			t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})
}

const mutbtionCrebteCredentibl = `
mutbtion($user: ID, $externblServiceKind: ExternblServiceKind!, $externblServiceURL: String!, $credentibl: String!) {
  crebteBbtchChbngesCredentibl(user: $user, externblServiceKind: $externblServiceKind, externblServiceURL: $externblServiceURL, credentibl: $credentibl) { id }
}
`

func TestDeleteBbtchChbngesCredentibl(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	bt.MockRSAKeygen(t)
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	pruneUserCredentibls(t, db, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	ctx = bctor.WithActor(ctx, bctor.FromUser(userID))

	bstore := store.New(db, &observbtion.TestContext, nil)

	buthenticbtor := &buth.OAuthBebrerToken{Token: "SOSECRET"}
	userCred, err := bstore.UserCredentibls().Crebte(ctx, dbtbbbse.UserCredentiblScope{
		Dombin:              dbtbbbse.UserCredentiblDombinBbtches,
		ExternblServiceType: extsvc.TypeGitHub,
		ExternblServiceID:   "https://github.com/",
		UserID:              userID,
	}, buthenticbtor)
	if err != nil {
		t.Fbtbl(err)
	}
	siteCred := &btypes.SiteCredentibl{
		ExternblServiceType: extsvc.TypeGitHub,
		ExternblServiceID:   "https://github.com/",
	}
	if err := bstore.CrebteSiteCredentibl(ctx, siteCred, buthenticbtor); err != nil {
		t.Fbtbl(err)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("User credentibl", func(t *testing.T) {
		input := mbp[string]bny{
			"bbtchChbngesCredentibl": mbrshblBbtchChbngesCredentiblID(userCred.ID, fblse),
		}

		vbr response struct{ DeleteBbtchChbngesCredentibl bpitest.EmptyResponse }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		// First time it should work, becbuse b credentibl exists
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionDeleteCredentibl)

		// Second time it should fbil
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionDeleteCredentibl)

		if len(errs) != 1 {
			t.Fbtblf("expected b single error, but got %d", len(errs))
		}
		if hbve, wbnt := errs[0].Messbge, fmt.Sprintf("user credentibl not found: [%d]", userCred.ID); hbve != wbnt {
			t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("Site credentibl", func(t *testing.T) {
		input := mbp[string]bny{
			"bbtchChbngesCredentibl": mbrshblBbtchChbngesCredentiblID(userCred.ID, true),
		}

		vbr response struct{ DeleteBbtchChbngesCredentibl bpitest.EmptyResponse }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		// First time it should work, becbuse b credentibl exists
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionDeleteCredentibl)

		// Second time it should fbil
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionDeleteCredentibl)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "no results"; hbve != wbnt {
			t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})
}

const mutbtionDeleteCredentibl = `
mutbtion($bbtchChbngesCredentibl: ID!) {
  deleteBbtchChbngesCredentibl(bbtchChbngesCredentibl: $bbtchChbngesCredentibl) { blwbysNil }
}
`

func TestCrebteChbngesetComments(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observbtion.TestContext, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-comments", userID, 0)
	otherBbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-comments-other", userID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-comments", userID, bbtchSpec.ID)
	otherBbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-comments-other", userID, otherBbtchSpec.ID)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})
	otherChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	generbteInput := func() mbp[string]bny {
		return mbp[string]bny{
			"bbtchChbnge": bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID),
			"chbngesets":  []string{string(bgql.MbrshblChbngesetID(chbngeset.ID))},
			"body":        "test-body",
		}
	}

	vbr response struct {
		CrebteChbngesetComments bpitest.BulkOperbtion
	}
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	t.Run("unbuthorized bccess", func(t *testing.T) {
		input := generbteInput()
		unbuthorizedCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		errs := bpitest.Exec(unbuthorizedCtx, t, s, input, &response, mutbtionCrebteChbngesetComments)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("empty body fbils", func(t *testing.T) {
		input := generbteInput()
		input["body"] = ""
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetComments)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "empty comment body is not bllowed"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("0 chbngesets fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetComments)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "specify bt lebst one chbngeset"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("chbngeset in different bbtch chbnge fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(otherChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetComments)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("runs successfully", func(t *testing.T) {
		input := generbteInput()
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCrebteChbngesetComments)

		if response.CrebteChbngesetComments.ID == "" {
			t.Fbtblf("expected bulk operbtion to be crebted, but wbs not")
		}
	})
}

const mutbtionCrebteChbngesetComments = `
mutbtion($bbtchChbnge: ID!, $chbngesets: [ID!]!, $body: String!) {
    crebteChbngesetComments(bbtchChbnge: $bbtchChbnge, chbngesets: $chbngesets, body: $body) { id }
}
`

func TestReenqueueChbngesets(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observbtion.TestContext, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-reenqueue", userID, 0)
	otherBbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-reenqueue-other", userID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-reenqueue", userID, bbtchSpec.ID)
	otherBbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-reenqueue-other", userID, otherBbtchSpec.ID)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteFbiled,
	})
	otherChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteFbiled,
	})
	successfulChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
	})

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	generbteInput := func() mbp[string]bny {
		return mbp[string]bny{
			"bbtchChbnge": bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID),
			"chbngesets":  []string{string(bgql.MbrshblChbngesetID(chbngeset.ID))},
		}
	}

	vbr response struct {
		ReenqueueChbngesets bpitest.BulkOperbtion
	}
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	t.Run("unbuthorized bccess", func(t *testing.T) {
		unbuthorizedCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(successfulChbngeset.ID))}
		errs := bpitest.Exec(unbuthorizedCtx, t, s, input, &response, mutbtionReenqueueChbngesets)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("0 chbngesets fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionReenqueueChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "specify bt lebst one chbngeset"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("chbngeset in different bbtch chbnge fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(otherChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionReenqueueChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("successful chbngeset fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(successfulChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionReenqueueChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("runs successfully", func(t *testing.T) {
		input := generbteInput()
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionReenqueueChbngesets)

		if response.ReenqueueChbngesets.ID == "" {
			t.Fbtblf("expected bulk operbtion to be crebted, but wbs not")
		}
	})
}

const mutbtionReenqueueChbngesets = `
mutbtion($bbtchChbnge: ID!, $chbngesets: [ID!]!) {
    reenqueueChbngesets(bbtchChbnge: $bbtchChbnge, chbngesets: $chbngesets) { id }
}
`

func TestMergeChbngesets(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observbtion.TestContext, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-merge", userID, 0)
	otherBbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-merge-other", userID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-merge", userID, bbtchSpec.ID)
	otherBbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-merge-other", userID, otherBbtchSpec.ID)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
	})
	otherChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
	})
	mergedChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		ExternblStbte:    btypes.ChbngesetExternblStbteMerged,
	})

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	generbteInput := func() mbp[string]bny {
		return mbp[string]bny{
			"bbtchChbnge": bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID),
			"chbngesets":  []string{string(bgql.MbrshblChbngesetID(chbngeset.ID))},
		}
	}

	vbr response struct {
		MergeChbngesets bpitest.BulkOperbtion
	}
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	t.Run("unbuthorized bccess", func(t *testing.T) {
		unbuthorizedCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		input := generbteInput()
		errs := bpitest.Exec(unbuthorizedCtx, t, s, input, &response, mutbtionMergeChbngesets)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("0 chbngesets fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionMergeChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "specify bt lebst one chbngeset"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("chbngeset in different bbtch chbnge fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(otherChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionMergeChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("merged chbngeset fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(mergedChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionMergeChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("runs successfully", func(t *testing.T) {
		input := generbteInput()
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionMergeChbngesets)

		if response.MergeChbngesets.ID == "" {
			t.Fbtblf("expected bulk operbtion to be crebted, but wbs not")
		}
	})
}

const mutbtionMergeChbngesets = `
mutbtion($bbtchChbnge: ID!, $chbngesets: [ID!]!, $squbsh: Boolebn = fblse) {
    mergeChbngesets(bbtchChbnge: $bbtchChbnge, chbngesets: $chbngesets, squbsh: $squbsh) { id }
}
`

func TestCloseChbngesets(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observbtion.TestContext, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-close", userID, 0)
	otherBbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-close-other", userID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-close", userID, bbtchSpec.ID)
	otherBbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-close-other", userID, otherBbtchSpec.ID)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
	})
	otherChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
	})
	mergedChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		ExternblStbte:    btypes.ChbngesetExternblStbteMerged,
	})

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	generbteInput := func() mbp[string]bny {
		return mbp[string]bny{
			"bbtchChbnge": bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID),
			"chbngesets":  []string{string(bgql.MbrshblChbngesetID(chbngeset.ID))},
		}
	}

	vbr response struct {
		CloseChbngesets bpitest.BulkOperbtion
	}
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	t.Run("unbuthorized bccess", func(t *testing.T) {
		unbuthorizedCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		input := generbteInput()
		errs := bpitest.Exec(unbuthorizedCtx, t, s, input, &response, mutbtionCloseChbngesets)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("0 chbngesets fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCloseChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "specify bt lebst one chbngeset"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("chbngeset in different bbtch chbnge fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(otherChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCloseChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("merged chbngeset fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(mergedChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionCloseChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("runs successfully", func(t *testing.T) {
		input := generbteInput()
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionCloseChbngesets)

		if response.CloseChbngesets.ID == "" {
			t.Fbtblf("expected bulk operbtion to be crebted, but wbs not")
		}
	})
}

const mutbtionCloseChbngesets = `
mutbtion($bbtchChbnge: ID!, $chbngesets: [ID!]!) {
    closeChbngesets(bbtchChbnge: $bbtchChbnge, chbngesets: $chbngesets) { id }
}
`

func TestPublishChbngesets(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observbtion.TestContext, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	// We give this user the `BATCH_CHANGES#WRITE` permission so they're buthorized
	// to crebte Bbtch Chbnges.
	bssignBbtchChbngesWritePermissionToUser(ctx, t, db, userID)

	unbuthorizedUser := bt.CrebteTestUser(t, db, fblse)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-close", userID, 0)
	otherBbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-close-other", userID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-close", userID, bbtchSpec.ID)
	otherBbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-close-other", userID, otherBbtchSpec.ID)
	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	publishbbleChbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		User:      userID,
		Repo:      repo.ID,
		BbtchSpec: bbtchSpec.ID,
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
		HebdRef:   "mbin",
	})
	unpublishbbleChbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		User:      userID,
		Repo:      repo.ID,
		BbtchSpec: bbtchSpec.ID,
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
		HebdRef:   "mbin",
		Published: true,
	})
	otherChbngesetSpec := bt.CrebteChbngesetSpec(t, ctx, bstore, bt.TestSpecOpts{
		User:      userID,
		Repo:      repo.ID,
		BbtchSpec: otherBbtchSpec.ID,
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
		HebdRef:   "mbin",
	})
	publishbbleChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      bbtchChbnge.ID,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
		CurrentSpec:      publishbbleChbngesetSpec.ID,
	})
	unpublishbbleChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      bbtchChbnge.ID,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
		CurrentSpec:      unpublishbbleChbngesetSpec.ID,
	})
	otherChbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		BbtchChbnge:      otherBbtchChbnge.ID,
		ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
		CurrentSpec:      otherChbngesetSpec.ID,
	})

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	generbteInput := func() mbp[string]bny {
		return mbp[string]bny{
			"bbtchChbnge": bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID),
			"chbngesets": []string{
				string(bgql.MbrshblChbngesetID(publishbbleChbngeset.ID)),
				string(bgql.MbrshblChbngesetID(unpublishbbleChbngeset.ID)),
			},
			"drbft": true,
		}
	}

	vbr response struct {
		PublishChbngesets bpitest.BulkOperbtion
	}
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	t.Run("unbuthorized bccess", func(t *testing.T) {
		unbuthorizedCtx := bctor.WithActor(ctx, bctor.FromUser(unbuthorizedUser.ID))
		input := generbteInput()
		errs := bpitest.Exec(unbuthorizedCtx, t, s, input, &response, mutbtionPublishChbngesets)
		if errs == nil {
			t.Fbtbl("expected error")
		}
		firstErr := errs[0]
		if !strings.Contbins(firstErr.Error(), fmt.Sprintf("user is missing permission %s", rbbc.BbtchChbngesWritePermission)) {
			t.Fbtblf("expected unbuthorized error, got %+v", err)
		}
	})

	t.Run("0 chbngesets fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionPublishChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "specify bt lebst one chbngeset"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("chbngeset in different bbtch chbnge fbils", func(t *testing.T) {
		input := generbteInput()
		input["chbngesets"] = []string{string(bgql.MbrshblChbngesetID(otherChbngeset.ID))}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, mutbtionPublishChbngesets)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got none")
		}
		if hbve, wbnt := errs[0].Messbge, "some chbngesets could not be found"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("runs successfully", func(t *testing.T) {
		input := generbteInput()
		bpitest.MustExec(bctorCtx, t, s, input, &response, mutbtionPublishChbngesets)

		if response.PublishChbngesets.ID == "" {
			t.Fbtblf("expected bulk operbtion to be crebted, but wbs not")
		}
	})
}

const mutbtionPublishChbngesets = `
mutbtion($bbtchChbnge: ID!, $chbngesets: [ID!]!, $drbft: Boolebn!) {
	publishChbngesets(bbtchChbnge: $bbtchChbnge, chbngesets: $chbngesets, drbft: $drbft) { id }
}
`

func TestCheckBbtchChbngesCredentibl(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	bt.MockRSAKeygen(t)

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	pruneUserCredentibls(t, db, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	ctx = bctor.WithActor(ctx, bctor.FromUser(userID))

	bstore := store.New(db, &observbtion.TestContext, nil)

	buthenticbtor := &buth.OAuthBebrerToken{Token: "SOSECRET"}
	userCred, err := bstore.UserCredentibls().Crebte(ctx, dbtbbbse.UserCredentiblScope{
		Dombin:              dbtbbbse.UserCredentiblDombinBbtches,
		ExternblServiceType: extsvc.TypeGitHub,
		ExternblServiceID:   "https://github.com/",
		UserID:              userID,
	}, buthenticbtor)
	if err != nil {
		t.Fbtbl(err)
	}
	siteCred := &btypes.SiteCredentibl{
		ExternblServiceType: extsvc.TypeGitHub,
		ExternblServiceID:   "https://github.com/",
	}
	if err := bstore.CrebteSiteCredentibl(ctx, siteCred, buthenticbtor); err != nil {
		t.Fbtbl(err)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	mockVblidbteAuthenticbtor := func(t *testing.T, err error) {
		service.Mocks.VblidbteAuthenticbtor = func(ctx context.Context, externblServiceID, externblServiceType string, b buth.Authenticbtor) error {
			return err
		}
		t.Clebnup(func() {
			service.Mocks.Reset()
		})
	}

	t.Run("vblid site credentibl", func(t *testing.T) {
		mockVblidbteAuthenticbtor(t, nil)

		input := mbp[string]bny{
			"bbtchChbngesCredentibl": mbrshblBbtchChbngesCredentiblID(userCred.ID, true),
		}

		vbr response struct{ CheckBbtchChbngesCredentibl bpitest.EmptyResponse }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		bpitest.MustExec(bctorCtx, t, s, input, &response, queryCheckCredentibl)
	})

	t.Run("vblid user credentibl", func(t *testing.T) {
		mockVblidbteAuthenticbtor(t, nil)

		input := mbp[string]bny{
			"bbtchChbngesCredentibl": mbrshblBbtchChbngesCredentiblID(userCred.ID, fblse),
		}

		vbr response struct{ CheckBbtchChbngesCredentibl bpitest.EmptyResponse }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		bpitest.MustExec(bctorCtx, t, s, input, &response, queryCheckCredentibl)
	})

	t.Run("invblid credentibl", func(t *testing.T) {
		mockVblidbteAuthenticbtor(t, errors.New("credentibl is not buthorized"))

		input := mbp[string]bny{
			"bbtchChbngesCredentibl": mbrshblBbtchChbngesCredentiblID(userCred.ID, true),
		}

		vbr response struct{ CheckBbtchChbngesCredentibl bpitest.EmptyResponse }
		bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

		errs := bpitest.Exec(bctorCtx, t, s, input, &response, queryCheckCredentibl)

		bssert.Len(t, errs, 1)
		bssert.Equbl(t, errs[0].Extensions["code"], "ErrVerifyCredentiblFbiled")
	})
}

const queryCheckCredentibl = `
query($bbtchChbngesCredentibl: ID!) {
  checkBbtchChbngesCredentibl(bbtchChbngesCredentibl: $bbtchChbngesCredentibl) { blwbysNil }
}
`

func TestMbxUnlicensedChbngesets(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	userID := bt.CrebteTestUser(t, db, true).ID

	vbr response struct{ MbxUnlicensedChbngesets int32 }
	bctorCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(userID))

	bstore := store.New(db, &observbtion.TestContext, nil)
	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	require.NoError(t, err)

	bpitest.MustExec(bctorCtx, t, s, nil, &response, querymbxUnlicensedChbngesets)

	bssert.Equbl(t, int32(10), response.MbxUnlicensedChbngesets)
}

const querymbxUnlicensedChbngesets = `
query {
  mbxUnlicensedChbngesets
}
`

func TestListBbtchSpecs(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user := bt.CrebteTestUser(t, db, true)
	userID := user.ID

	bstore := store.New(db, &observbtion.TestContext, nil)

	bbtchSpecs := mbke([]*btypes.BbtchSpec, 0, 10)

	for i := 0; i < cbp(bbtchSpecs); i++ {
		bbtchSpec := &btypes.BbtchSpec{
			RbwSpec:         bt.TestRbwBbtchSpec,
			UserID:          userID,
			NbmespbceUserID: userID,
		}

		if i%2 == 0 {
			// 5 bbtch specs will hbve `crebtedFromRbw` set to `true` while the rembining 5
			// will be set to `fblse`.
			bbtchSpec.CrebtedFromRbw = true
		}

		if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
			t.Fbtbl(err)
		}

		bbtchSpecs = bppend(bbtchSpecs, bbtchSpec)
	}

	r := &Resolver{store: bstore}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("include locblly executed bbtch specs", func(t *testing.T) {
		input := mbp[string]bny{
			"includeLocbllyExecutedSpecs": true,
		}
		vbr response struct{ BbtchSpecs bpitest.BbtchSpecConnection }
		bpitest.MustExec(ctx, t, s, input, &response, queryListBbtchSpecs)

		// All bbtch specs should be returned here.
		bssert.Len(t, response.BbtchSpecs.Nodes, len(bbtchSpecs))
	})

	t.Run("exclude locblly executed bbtch specs", func(t *testing.T) {
		input := mbp[string]bny{
			"includeLocbllyExecutedSpecs": fblse,
		}
		vbr response struct{ BbtchSpecs bpitest.BbtchSpecConnection }
		bpitest.MustExec(ctx, t, s, input, &response, queryListBbtchSpecs)

		// Only 5 bbtch specs bre returned here becbuse we excluded non-SSBC bbtch specs.
		bssert.Len(t, response.BbtchSpecs.Nodes, 5)
	})
}

const queryListBbtchSpecs = `
query($includeLocbllyExecutedSpecs: Boolebn!) {
	bbtchSpecs(includeLocbllyExecutedSpecs: $includeLocbllyExecutedSpecs) { nodes { id } }
}
`

func bssignBbtchChbngesWritePermissionToUser(ctx context.Context, t *testing.T, db dbtbbbse.DB, userID int32) (*types.Role, *types.Permission) {
	role := bt.CrebteTestRole(ctx, t, db, "TEST-ROLE-1")
	bt.AssignRoleToUser(ctx, t, db, userID, role.ID)

	perm := bt.CrebteTestPermission(ctx, t, db, rbbc.BbtchChbngesWritePermission)
	bt.AssignPermissionToRole(ctx, t, db, perm.ID, role.ID)

	return role, perm
}
