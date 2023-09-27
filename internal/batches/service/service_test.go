pbckbge service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	stesting "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	extsvcbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestServicePermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	s := store.New(db, &observbtion.TestContext, nil)
	svc := New(s)

	bdmin := bt.CrebteTestUser(t, db, true)
	user := bt.CrebteTestUser(t, db, fblse)
	otherUser := bt.CrebteTestUser(t, db, fblse)
	nonOrgMember := bt.CrebteTestUser(t, db, fblse)

	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	org := bt.CrebteTestOrg(t, db, "test-org-1", bdmin.ID, user.ID, otherUser.ID)

	crebteTestDbtb := func(t *testing.T, s *store.Store, buthor, orgNbmespbce int32) (bbtchChbnge *btypes.BbtchChbnge, chbngeset *btypes.Chbngeset, spec *btypes.BbtchSpec) {
		if orgNbmespbce == 0 {
			spec = testBbtchSpec(buthor)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}
			bbtchChbnge = testBbtchChbnge(buthor, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}
			chbngeset = testChbngeset(repo.ID, bbtchChbnge.ID, btypes.ChbngesetExternblStbteOpen)
			if err := s.CrebteChbngeset(ctx, chbngeset); err != nil {
				t.Fbtbl(err)
			}
		} else {
			spec = testOrgBbtchSpec(buthor, orgNbmespbce)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}
			bbtchChbnge = testOrgBbtchChbnge(buthor, orgNbmespbce, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}
			chbngeset = testChbngeset(repo.ID, bbtchChbnge.ID, btypes.ChbngesetExternblStbteOpen)
			if err := s.CrebteChbngeset(ctx, chbngeset); err != nil {
				t.Fbtbl(err)
			}
		}

		return bbtchChbnge, chbngeset, spec
	}

	tests := []struct {
		nbme              string
		bbtchChbngeAuthor int32
		currentUser       int32
		bssertFunc        func(t *testing.T, err error)
		orgMembersAdmin   bool
		orgNbmespbce      int32
	}{
		{
			nbme:              "unbuthorized user (user nbmespbce)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       otherUser.ID,
			bssertFunc:        bssertAuthError,
		},
		{
			nbme:              "bbtch chbnge buthor (user nbmespbce)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       user.ID,
			bssertFunc:        bssertNoAuthError,
		},

		{
			nbme:              "site-bdmin (user nbmespbce)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       bdmin.ID,
			bssertFunc:        bssertNoAuthError,
		},
		{
			nbme:              "non-org member (org nbmespbce)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       nonOrgMember.ID,
			bssertFunc:        bssertOrgOrAuthError,
			orgNbmespbce:      org.ID,
		},
		{
			nbme:              "non-org member (org nbmespbce - bll members bdmin)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       nonOrgMember.ID,
			bssertFunc:        bssertOrgOrAuthError,
			orgMembersAdmin:   true,
			orgNbmespbce:      org.ID,
		},
		{
			nbme:              "org member (org nbmespbce)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       otherUser.ID,
			bssertFunc:        bssertNoAuthError,
			orgNbmespbce:      org.ID,
		},
		{
			nbme:              "org member (org nbmespbce - bll members bdmin)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       otherUser.ID,
			bssertFunc:        bssertNoAuthError,
			orgMembersAdmin:   true,
			orgNbmespbce:      org.ID,
		},
		{
			nbme:              "bbtch chbnge buthor (org nbmespbce)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       user.ID,
			bssertFunc:        bssertNoAuthError,
			orgNbmespbce:      org.ID,
		},
		{
			nbme:              "bbtch chbnge buthor (org nbmespbce - bll members bdmin)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       user.ID,
			bssertFunc:        bssertNoAuthError,
			orgMembersAdmin:   true,
			orgNbmespbce:      org.ID,
		},
		{
			nbme:              "site-bdmin (org nbmespbce)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       bdmin.ID,
			bssertFunc:        bssertNoAuthError,
			orgNbmespbce:      org.ID,
		},
		{
			nbme:              "site-bdmin (org nbmespbce - bll members bdmin)",
			bbtchChbngeAuthor: user.ID,
			currentUser:       bdmin.ID,
			bssertFunc:        bssertNoAuthError,
			orgMembersAdmin:   true,
			orgNbmespbce:      org.ID,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			bbtchChbnge, chbngeset, bbtchSpec := crebteTestDbtb(t, s, tc.bbtchChbngeAuthor, tc.orgNbmespbce)
			// Fresh context.Bbckground() becbuse the previous one is wrbpped in AuthzBypbs
			currentUserCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(tc.currentUser))

			if tc.orgNbmespbce != 0 && tc.orgMembersAdmin {
				contents := "{\"orgs.bllMembersBbtchChbngesAdmin\": true}"
				_, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{Org: &tc.orgNbmespbce}, nil, nil, contents)
				if err != nil {
					t.Fbtbl(err)
				}
			}

			t.Run("EnqueueChbngesetSync", func(t *testing.T) {
				// The cbses thbt don't result in buth errors will fbll through
				// to cbll repoupdbter.EnqueueChbngesetSync, so we need to
				// ensure we mock thbt cbll to bvoid unexpected network cblls.
				repoupdbter.MockEnqueueChbngesetSync = func(ctx context.Context, ids []int64) error {
					return nil
				}
				t.Clebnup(func() { repoupdbter.MockEnqueueChbngesetSync = nil })

				err := svc.EnqueueChbngesetSync(currentUserCtx, chbngeset.ID)
				tc.bssertFunc(t, err)
			})

			t.Run("ReenqueueChbngeset", func(t *testing.T) {
				_, _, err := svc.ReenqueueChbngeset(currentUserCtx, chbngeset.ID)
				tc.bssertFunc(t, err)
			})

			t.Run("CloseBbtchChbnge", func(t *testing.T) {
				_, err := svc.CloseBbtchChbnge(currentUserCtx, bbtchChbnge.ID, fblse)
				tc.bssertFunc(t, err)
			})

			t.Run("DeleteBbtchChbnge", func(t *testing.T) {
				err := svc.DeleteBbtchChbnge(currentUserCtx, bbtchChbnge.ID)
				tc.bssertFunc(t, err)
			})

			t.Run("MoveBbtchChbnge", func(t *testing.T) {
				_, err := svc.MoveBbtchChbnge(currentUserCtx, MoveBbtchChbngeOpts{
					BbtchChbngeID: bbtchChbnge.ID,
					NewNbme:       "foobbr2",
				})
				tc.bssertFunc(t, err)
			})

			t.Run("ApplyBbtchChbnge", func(t *testing.T) {
				_, err := svc.ApplyBbtchChbnge(currentUserCtx, ApplyBbtchChbngeOpts{
					BbtchSpecRbndID: bbtchSpec.RbndID,
				})
				tc.bssertFunc(t, err)
			})

			t.Run("CrebteChbngesetJobs", func(t *testing.T) {
				_, err := svc.CrebteChbngesetJobs(currentUserCtx, bbtchChbnge.ID, []int64{chbngeset.ID}, btypes.ChbngesetJobTypeComment, btypes.ChbngesetJobCommentPbylobd{Messbge: "test"}, store.ListChbngesetsOpts{})
				tc.bssertFunc(t, err)
			})

			t.Run("ExecuteBbtchSpec", func(t *testing.T) {
				_, err := svc.ExecuteBbtchSpec(currentUserCtx, ExecuteBbtchSpecOpts{
					BbtchSpecRbndID: bbtchSpec.RbndID,
				})
				tc.bssertFunc(t, err)
			})

			t.Run("ReplbceBbtchSpecInput", func(t *testing.T) {
				_, err := svc.ReplbceBbtchSpecInput(currentUserCtx, ReplbceBbtchSpecInputOpts{
					BbtchSpecRbndID: bbtchSpec.RbndID,
					RbwSpec:         bt.TestRbwBbtchSpecYAML,
				})
				tc.bssertFunc(t, err)
			})

			t.Run("UpsertBbtchSpecInput", func(t *testing.T) {
				_, err := svc.UpsertBbtchSpecInput(currentUserCtx, UpsertBbtchSpecInputOpts{
					RbwSpec:         bt.TestRbwBbtchSpecYAML,
					NbmespbceUserID: tc.bbtchChbngeAuthor,
					NbmespbceOrgID:  tc.orgNbmespbce,
				})
				tc.bssertFunc(t, err)
			})

			t.Run("CrebteBbtchSpecFromRbw", func(t *testing.T) {
				_, err := svc.CrebteBbtchSpecFromRbw(currentUserCtx, CrebteBbtchSpecFromRbwOpts{
					RbwSpec:         bt.TestRbwBbtchSpecYAML,
					NbmespbceUserID: tc.bbtchChbngeAuthor,
					NbmespbceOrgID:  tc.orgNbmespbce,
					BbtchChbnge:     bbtchChbnge.ID,
				})
				tc.bssertFunc(t, err)
			})

			t.Run("CbncelBbtchSpec", func(t *testing.T) {
				_, err := svc.CbncelBbtchSpec(currentUserCtx, CbncelBbtchSpecOpts{
					BbtchSpecRbndID: bbtchSpec.RbndID,
				})
				tc.bssertFunc(t, err)
			})
		})
	}
}

func TestService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bdmin := bt.CrebteTestUser(t, db, true)
	user := bt.CrebteTestUser(t, db, fblse)
	user2 := bt.CrebteTestUser(t, db, fblse)
	user3 := bt.CrebteTestUser(t, db, fblse)

	bdminCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(bdmin.ID))
	userCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(user.ID))
	user2Ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(user2.ID))

	now := timeutil.Now()
	clock := func() time.Time { return now }

	s := store.NewWithClock(db, &observbtion.TestContext, nil, clock)
	rs, _ := bt.CrebteTestRepos(t, ctx, db, 4)

	fbkeSource := &stesting.FbkeChbngesetSource{}
	sourcer := stesting.NewFbkeSourcer(nil, fbkeSource)

	svc := New(s)
	svc.sourcer = sourcer

	t.Run("CheckViewerCbnAdminister", func(t *testing.T) {
		org := bt.CrebteTestOrg(t, db, "test-org-1", user.ID, user2.ID)

		spec := testBbtchSpec(user.ID)
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}

		userBbtchChbnge := testBbtchChbnge(user.ID, spec)
		if err := s.CrebteBbtchChbnge(ctx, userBbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		orgSpec := testOrgBbtchSpec(user.ID, org.ID)
		if err := s.CrebteBbtchSpec(ctx, orgSpec); err != nil {
			t.Fbtbl(err)
		}
		orgBbtchChbnge := testOrgBbtchChbnge(user.ID, org.ID, orgSpec)
		if err := s.CrebteBbtchChbnge(ctx, orgBbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		tests := []struct {
			nbme        string
			bbtchChbnge *btypes.BbtchChbnge
			user        int32

			cbnAdminister              bool
			orgMembersBbtchChbngeAdmin bool
		}{
			{
				nbme:                       "user bbtch chbnge bccessed by crebtor",
				bbtchChbnge:                userBbtchChbnge,
				user:                       user.ID,
				cbnAdminister:              true,
				orgMembersBbtchChbngeAdmin: fblse,
			},
			{
				nbme:                       "user bbtch chbnge bccessed by site-bdmin",
				bbtchChbnge:                userBbtchChbnge,
				user:                       bdmin.ID,
				cbnAdminister:              true,
				orgMembersBbtchChbngeAdmin: fblse,
			},
			{
				nbme:                       "user bbtch chbnge bccessed by regulbr user",
				bbtchChbnge:                userBbtchChbnge,
				user:                       user2.ID,
				cbnAdminister:              fblse,
				orgMembersBbtchChbngeAdmin: fblse,
			},
			{
				nbme:                       "org bbtch chbnge bccessed by crebtor",
				bbtchChbnge:                orgBbtchChbnge,
				user:                       user.ID,
				cbnAdminister:              true,
				orgMembersBbtchChbngeAdmin: fblse,
			},
			{
				nbme:                       "org bbtch chbnge bccessed by site-bdmin",
				bbtchChbnge:                orgBbtchChbnge,
				user:                       bdmin.ID,
				cbnAdminister:              true,
				orgMembersBbtchChbngeAdmin: fblse,
			},
			{
				nbme:                       "org bbtch chbnge bccessed by org member",
				bbtchChbnge:                orgBbtchChbnge,
				user:                       user2.ID,
				cbnAdminister:              true,
				orgMembersBbtchChbngeAdmin: fblse,
			},
			{
				nbme:                       "org bbtch chbnge bccessed by non-org member when `orgs.bllMembersBbtchChbngesAdmin` is true",
				bbtchChbnge:                orgBbtchChbnge,
				user:                       user2.ID,
				cbnAdminister:              true,
				orgMembersBbtchChbngeAdmin: true,
			},
			{
				nbme:                       "org bbtch chbnge bccessed by non-org member",
				bbtchChbnge:                orgBbtchChbnge,
				user:                       user3.ID,
				cbnAdminister:              fblse,
				orgMembersBbtchChbngeAdmin: fblse,
			},
		}

		for _, tc := rbnge tests {
			t.Run(tc.nbme, func(t *testing.T) {
				ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(tc.user))
				if tc.orgMembersBbtchChbngeAdmin {
					contents := "{\"orgs.bllMembersBbtchChbngesAdmin\": true}"
					_, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{Org: &org.ID}, nil, nil, contents)
					if err != nil {
						t.Fbtbl(err)
					}
				}
				cbnAdminister, _ := svc.CheckViewerCbnAdminister(ctx, tc.bbtchChbnge.NbmespbceUserID, tc.bbtchChbnge.NbmespbceOrgID)

				if cbnAdminister != tc.cbnAdminister {
					t.Fbtblf("expected cbnAdminister to be %t, got %t", tc.cbnAdminister, cbnAdminister)
				}
			})
		}
	})

	t.Run("DeleteBbtchChbnge", func(t *testing.T) {
		spec := testBbtchSpec(bdmin.ID)
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
		if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			t.Fbtbl(err)
		}
		if err := svc.DeleteBbtchChbnge(ctx, bbtchChbnge.ID); err != nil {
			t.Fbtblf("bbtch chbnge not deleted: %s", err)
		}

		_, err := s.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: bbtchChbnge.ID})
		if err != nil && err != store.ErrNoResults {
			t.Fbtblf("wbnt bbtch chbnge to be deleted, but wbs not: %e", err)
		}
	})

	t.Run("CloseBbtchChbnge", func(t *testing.T) {
		crebteBbtchChbnge := func(t *testing.T) *btypes.BbtchChbnge {
			t.Helper()

			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}
			return bbtchChbnge
		}

		closeConfirm := func(t *testing.T, c *btypes.BbtchChbnge, closeChbngesets bool) {
			t.Helper()

			closedBbtchChbnge, err := svc.CloseBbtchChbnge(bdminCtx, c.ID, closeChbngesets)
			if err != nil {
				t.Fbtblf("bbtch chbnge not closed: %s", err)
			}
			if !closedBbtchChbnge.ClosedAt.Equbl(now) {
				t.Fbtblf("bbtch chbnge ClosedAt is zero")
			}

			if !closeChbngesets {
				return
			}

			cs, _, err := s.ListChbngesets(ctx, store.ListChbngesetsOpts{
				OwnedByBbtchChbngeID: c.ID,
			})
			if err != nil {
				t.Fbtblf("listing chbngesets fbiled: %s", err)
			}
			for _, c := rbnge cs {
				if !c.Closing {
					t.Errorf("chbngeset should be Closing, but is not")
				}

				if hbve, wbnt := c.ReconcilerStbte, btypes.ReconcilerStbteQueued; hbve != wbnt {
					t.Errorf("chbngeset ReconcilerStbte wrong. wbnt=%s, hbve=%s", wbnt, hbve)
				}
			}
		}

		t.Run("no chbngesets", func(t *testing.T) {
			bbtchChbnge := crebteBbtchChbnge(t)
			closeConfirm(t, bbtchChbnge, fblse)
		})

		t.Run("chbngesets", func(t *testing.T) {
			bbtchChbnge := crebteBbtchChbnge(t)

			chbngeset1 := testChbngeset(rs[0].ID, bbtchChbnge.ID, btypes.ChbngesetExternblStbteOpen)
			chbngeset1.ReconcilerStbte = btypes.ReconcilerStbteCompleted
			if err := s.CrebteChbngeset(ctx, chbngeset1); err != nil {
				t.Fbtbl(err)
			}

			chbngeset2 := testChbngeset(rs[1].ID, bbtchChbnge.ID, btypes.ChbngesetExternblStbteOpen)
			chbngeset2.ReconcilerStbte = btypes.ReconcilerStbteCompleted
			if err := s.CrebteChbngeset(ctx, chbngeset2); err != nil {
				t.Fbtbl(err)
			}

			closeConfirm(t, bbtchChbnge, true)
		})
	})

	t.Run("EnqueueChbngesetSync", func(t *testing.T) {
		spec := testBbtchSpec(user.ID)
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbnge := testBbtchChbnge(user.ID, spec)
		if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		chbngeset := testChbngeset(rs[0].ID, bbtchChbnge.ID, btypes.ChbngesetExternblStbteOpen)
		if err := s.CrebteChbngeset(ctx, chbngeset); err != nil {
			t.Fbtbl(err)
		}

		cblled := fblse
		repoupdbter.MockEnqueueChbngesetSync = func(_ context.Context, ids []int64) error {
			if len(ids) != 1 && ids[0] != chbngeset.ID {
				t.Fbtblf("MockEnqueueChbngesetSync received wrong ids: %+v", ids)
			}
			cblled = true
			return nil
		}
		t.Clebnup(func() { repoupdbter.MockEnqueueChbngesetSync = nil })

		if err := svc.EnqueueChbngesetSync(userCtx, chbngeset.ID); err != nil {
			t.Fbtbl(err)
		}

		if !cblled {
			t.Fbtbl("MockEnqueueChbngesetSync not cblled")
		}

		// rs[0] is filtered out
		bt.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

		// should result in b not found error
		if err := svc.EnqueueChbngesetSync(userCtx, chbngeset.ID); !errcode.IsNotFound(err) {
			t.Fbtblf("expected not-found error but got %v", err)
		}
	})

	t.Run("ReenqueueChbngeset", func(t *testing.T) {
		spec := testBbtchSpec(user.ID)
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbnge := testBbtchChbnge(user.ID, spec)
		if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		chbngeset := testChbngeset(rs[0].ID, bbtchChbnge.ID, btypes.ChbngesetExternblStbteOpen)
		if err := s.CrebteChbngeset(ctx, chbngeset); err != nil {
			t.Fbtbl(err)
		}

		bt.SetChbngesetFbiled(t, ctx, s, chbngeset)

		if _, _, err := svc.ReenqueueChbngeset(userCtx, chbngeset.ID); err != nil {
			t.Fbtbl(err)
		}

		bt.RelobdAndAssertChbngeset(t, ctx, s, chbngeset, bt.ChbngesetAssertions{
			Repo:          rs[0].ID,
			ExternblStbte: btypes.ChbngesetExternblStbteOpen,
			ExternblID:    "ext-id-7",
			AttbchedTo:    []int64{bbtchChbnge.ID},

			// The importbnt fields:
			ReconcilerStbte:        btypes.ReconcilerStbteQueued,
			NumResets:              0,
			NumFbilures:            0,
			FbilureMessbge:         nil,
			PreviousFbilureMessbge: pointers.Ptr(bt.FbiledChbngesetFbilureMessbge),
		})

		// rs[0] is filtered out
		bt.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

		// should result in b not found error
		if _, _, err := svc.ReenqueueChbngeset(userCtx, chbngeset.ID); !errcode.IsNotFound(err) {
			t.Fbtblf("expected not-found error but got %v", err)
		}
	})

	t.Run("CrebteBbtchSpec", func(t *testing.T) {
		chbngesetSpecs := mbke([]*btypes.ChbngesetSpec, 0, len(rs))
		chbngesetSpecRbndIDs := mbke([]string, 0, len(rs))
		for _, r := rbnge rs {
			cs := &btypes.ChbngesetSpec{BbseRepoID: r.ID, UserID: bdmin.ID, ExternblID: "123"}
			if err := s.CrebteChbngesetSpec(ctx, cs); err != nil {
				t.Fbtbl(err)
			}
			chbngesetSpecs = bppend(chbngesetSpecs, cs)
			chbngesetSpecRbndIDs = bppend(chbngesetSpecRbndIDs, cs.RbndID)
		}

		t.Run("success", func(t *testing.T) {
			opts := CrebteBbtchSpecOpts{
				NbmespbceUserID:      bdmin.ID,
				RbwSpec:              bt.TestRbwBbtchSpec,
				ChbngesetSpecRbndIDs: chbngesetSpecRbndIDs,
			}

			spec, err := svc.CrebteBbtchSpec(bdminCtx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if spec.ID == 0 {
				t.Fbtblf("BbtchSpec ID is 0")
			}

			if hbve, wbnt := spec.UserID, bdmin.ID; hbve != wbnt {
				t.Fbtblf("UserID is %d, wbnt %d", hbve, wbnt)
			}

			vbr wbntFields *bbtcheslib.BbtchSpec
			if err := json.Unmbrshbl([]byte(spec.RbwSpec), &wbntFields); err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(wbntFields, spec.Spec); diff != "" {
				t.Fbtblf("wrong spec fields (-wbnt +got):\n%s", diff)
			}

			for _, cs := rbnge chbngesetSpecs {
				cs2, err := s.GetChbngesetSpec(ctx, store.GetChbngesetSpecOpts{ID: cs.ID})
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := cs2.BbtchSpecID, spec.ID; hbve != wbnt {
					t.Fbtblf("chbngesetSpec hbs wrong BbtchSpec. wbnt=%d, hbve=%d", wbnt, hbve)
				}
			}
		})

		t.Run("success with YAML rbw spec", func(t *testing.T) {
			opts := CrebteBbtchSpecOpts{
				NbmespbceUserID: bdmin.ID,
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
			}

			spec, err := svc.CrebteBbtchSpec(bdminCtx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if spec.ID == 0 {
				t.Fbtblf("BbtchSpec ID is 0")
			}

			vbr wbntFields *bbtcheslib.BbtchSpec
			if err := json.Unmbrshbl([]byte(bt.TestRbwBbtchSpec), &wbntFields); err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(wbntFields, spec.Spec); diff != "" {
				t.Fbtblf("wrong spec fields (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			bt.MockRepoPermissions(t, db, user.ID)

			opts := CrebteBbtchSpecOpts{
				NbmespbceUserID:      user.ID,
				RbwSpec:              bt.TestRbwBbtchSpec,
				ChbngesetSpecRbndIDs: chbngesetSpecRbndIDs,
			}

			if _, err := svc.CrebteBbtchSpec(userCtx, opts); !errcode.IsNotFound(err) {
				t.Fbtblf("expected not-found error but got %s", err)
			}
		})

		t.Run("invblid chbngesetspec id", func(t *testing.T) {
			contbinsInvblidID := []string{chbngesetSpecRbndIDs[0], "foobbr"}
			opts := CrebteBbtchSpecOpts{
				NbmespbceUserID:      bdmin.ID,
				RbwSpec:              bt.TestRbwBbtchSpec,
				ChbngesetSpecRbndIDs: contbinsInvblidID,
			}

			if _, err := svc.CrebteBbtchSpec(bdminCtx, opts); !errcode.IsNotFound(err) {
				t.Fbtblf("expected not-found error but got %s", err)
			}
		})

		t.Run("nbmespbce user is not bdmin bnd not crebtor", func(t *testing.T) {
			opts := CrebteBbtchSpecOpts{
				NbmespbceUserID: bdmin.ID,
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
			}

			_, err := svc.CrebteBbtchSpec(userCtx, opts)
			if !errcode.IsUnbuthorized(err) {
				t.Fbtblf("expected unbuthorized error but got %s", err)
			}

			// Try bgbin bs bdmin
			opts.NbmespbceUserID = user.ID

			_, err = svc.CrebteBbtchSpec(bdminCtx, opts)
			if err != nil {
				t.Fbtblf("expected no error but got %s", err)
			}
		})

		t.Run("missing bccess to nbmespbce org", func(t *testing.T) {
			orgID := bt.CrebteTestOrg(t, db, "test-org").ID

			opts := CrebteBbtchSpecOpts{
				NbmespbceOrgID:       orgID,
				RbwSpec:              bt.TestRbwBbtchSpec,
				ChbngesetSpecRbndIDs: chbngesetSpecRbndIDs,
			}

			_, err := svc.CrebteBbtchSpec(userCtx, opts)
			if hbve, wbnt := err, buth.ErrNotAnOrgMember; hbve != wbnt {
				t.Fbtblf("expected %s error but got %s", wbnt, hbve)
			}

			// Crebte org membership bnd try bgbin
			if _, err := db.OrgMembers().Crebte(ctx, orgID, user.ID); err != nil {
				t.Fbtbl(err)
			}

			_, err = svc.CrebteBbtchSpec(userCtx, opts)
			if err != nil {
				t.Fbtblf("expected no error but got %s", err)
			}
		})

		t.Run("no side-effects if no chbngeset spec IDs bre given", func(t *testing.T) {
			// We blrebdy hbve ChbngesetSpecs in the dbtbbbse. Here we
			// wbnt to mbke sure thbt the new BbtchSpec is crebted,
			// without bccidently bttbching the existing ChbngesetSpecs.
			opts := CrebteBbtchSpecOpts{
				NbmespbceUserID:      bdmin.ID,
				RbwSpec:              bt.TestRbwBbtchSpec,
				ChbngesetSpecRbndIDs: []string{},
			}

			spec, err := svc.CrebteBbtchSpec(bdminCtx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			countOpts := store.CountChbngesetSpecsOpts{BbtchSpecID: spec.ID}
			count, err := s.CountChbngesetSpecs(bdminCtx, countOpts)
			if err != nil {
				return
			}
			if count != 0 {
				t.Fbtblf("wbnt no chbngeset specs bttbched to bbtch spec, but hbve %d", count)
			}
		})
	})

	t.Run("CrebteChbngesetSpec", func(t *testing.T) {
		repo := rs[0]
		rbwSpec := bt.NewRbwChbngesetSpecGitBrbnch(relby.MbrshblID("Repository", repo.ID), "d34db33f")

		t.Run("success", func(t *testing.T) {
			spec, err := svc.CrebteChbngesetSpec(ctx, rbwSpec, bdmin.ID)
			if err != nil {
				t.Fbtbl(err)
			}

			if spec.ID == 0 {
				t.Fbtblf("ChbngesetSpec ID is 0")
			}

			wbnt := &btypes.ChbngesetSpec{
				ID:   5,
				Type: btypes.ChbngesetSpecTypeBrbnch,
				Diff: []byte(`diff --git INSTALL.md INSTALL.md
index e5bf166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobbr Line 8
 Line 9
 Line 10
`),
				DiffStbtAdded:     3,
				DiffStbtDeleted:   3,
				BbseRepoID:        1,
				UserID:            1,
				BbseRev:           "d34db33f",
				BbseRef:           "refs/hebds/mbster",
				HebdRef:           "refs/hebds/my-brbnch",
				Title:             "the title",
				Body:              "the body of the PR",
				Published:         bbtcheslib.PublishedVblue{Vbl: fblse},
				CommitMessbge:     "git commit messbge\n\nbnd some more content in b second pbrbgrbph.",
				CommitAuthorNbme:  "Mbry McButtons",
				CommitAuthorEmbil: "mbry@exbmple.com",
			}

			if diff := cmp.Diff(wbnt, spec, cmpopts.IgnoreFields(btypes.ChbngesetSpec{}, "CrebtedAt", "UpdbtedAt", "RbndID")); diff != "" {
				t.Fbtblf("wrong spec fields (-wbnt +got):\n%s", diff)
			}

			wbntDiffStbt := *bt.ChbngesetSpecDiffStbt
			if diff := cmp.Diff(wbntDiffStbt, spec.DiffStbt()); diff != "" {
				t.Fbtblf("wrong diff stbt (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("invblid rbw spec", func(t *testing.T) {
			invblidRbw := `{"externblComputer": "beepboop"}`
			_, err := svc.CrebteChbngesetSpec(ctx, invblidRbw, bdmin.ID)
			if err == nil {
				t.Fbtbl("expected error but got nil")
			}

			hbveErr := fmt.Sprintf("%v", err)
			wbntErr := "4 errors occurred:\n\t* Must vblidbte one bnd only one schemb (oneOf)\n\t* bbseRepository is required\n\t* externblID is required\n\t* Additionbl property externblComputer is not bllowed"
			if diff := cmp.Diff(wbntErr, hbveErr); diff != "" {
				t.Fbtblf("unexpected error (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			bt.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

			_, err := svc.CrebteChbngesetSpec(userCtx, rbwSpec, bdmin.ID)
			if !errcode.IsNotFound(err) {
				t.Fbtblf("expected not-found error but got %v", err)
			}
		})
	})

	t.Run("CrebteChbngesetSpecs", func(t *testing.T) {
		rbwSpec := bt.NewRbwChbngesetSpecGitBrbnch(relby.MbrshblID("Repository", rs[0].ID), "d34db33f")

		t.Run("success", func(t *testing.T) {
			specs, err := svc.CrebteChbngesetSpecs(ctx, []string{rbwSpec}, bdmin.ID)
			require.NoError(t, err)

			bssert.Len(t, specs, 1)

			for _, spec := rbnge specs {
				bssert.NotZero(t, spec.ID)

				wbnt := &btypes.ChbngesetSpec{
					ID:   6,
					Type: btypes.ChbngesetSpecTypeBrbnch,
					Diff: []byte(`diff --git INSTALL.md INSTALL.md
index e5bf166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobbr Line 8
 Line 9
 Line 10
`),
					DiffStbtAdded:     3,
					DiffStbtDeleted:   3,
					BbseRepoID:        1,
					UserID:            1,
					BbseRev:           "d34db33f",
					BbseRef:           "refs/hebds/mbster",
					HebdRef:           "refs/hebds/my-brbnch",
					Title:             "the title",
					Body:              "the body of the PR",
					Published:         bbtcheslib.PublishedVblue{Vbl: fblse},
					CommitMessbge:     "git commit messbge\n\nbnd some more content in b second pbrbgrbph.",
					CommitAuthorNbme:  "Mbry McButtons",
					CommitAuthorEmbil: "mbry@exbmple.com",
				}

				if diff := cmp.Diff(wbnt, spec, cmpopts.IgnoreFields(btypes.ChbngesetSpec{}, "CrebtedAt", "UpdbtedAt", "RbndID")); diff != "" {
					t.Fbtblf("wrong spec fields (-wbnt +got):\n%s", diff)
				}

				wbntDiffStbt := *bt.ChbngesetSpecDiffStbt
				if diff := cmp.Diff(wbntDiffStbt, spec.DiffStbt()); diff != "" {
					t.Fbtblf("wrong diff stbt (-wbnt +got):\n%s", diff)
				}
			}
		})

		t.Run("invblid rbw spec", func(t *testing.T) {
			invblidRbw := `{"externblComputer": "beepboop"}`
			_, err := svc.CrebteChbngesetSpecs(ctx, []string{invblidRbw}, bdmin.ID)
			bssert.Error(t, err)
			bssert.Equbl(
				t,
				"4 errors occurred:\n\t* Must vblidbte one bnd only one schemb (oneOf)\n\t* bbseRepository is required\n\t* externblID is required\n\t* Additionbl property externblComputer is not bllowed",
				err.Error(),
			)
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			bt.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

			_, err := svc.CrebteChbngesetSpecs(userCtx, []string{rbwSpec}, bdmin.ID)
			bssert.Error(t, err)
			bssert.True(t, errcode.IsNotFound(err))
		})
	})

	t.Run("ApplyBbtchChbnge", func(t *testing.T) {
		// See TestServiceApplyBbtchChbnge
	})

	t.Run("MoveBbtchChbnge", func(t *testing.T) {
		crebteBbtchChbnge := func(t *testing.T, nbme string, buthorID, userID, orgID int32) *btypes.BbtchChbnge {
			t.Helper()

			spec := &btypes.BbtchSpec{
				UserID:          buthorID,
				NbmespbceUserID: userID,
				NbmespbceOrgID:  orgID,
			}

			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			c := &btypes.BbtchChbnge{
				CrebtorID:       buthorID,
				NbmespbceUserID: userID,
				NbmespbceOrgID:  orgID,
				Nbme:            nbme,
				LbstApplierID:   buthorID,
				LbstAppliedAt:   time.Now(),
				BbtchSpecID:     spec.ID,
			}

			if err := s.CrebteBbtchChbnge(ctx, c); err != nil {
				t.Fbtbl(err)
			}

			return c
		}

		t.Run("new nbme", func(t *testing.T) {
			bbtchChbnge := crebteBbtchChbnge(t, "old-nbme", bdmin.ID, bdmin.ID, 0)

			opts := MoveBbtchChbngeOpts{BbtchChbngeID: bbtchChbnge.ID, NewNbme: "new-nbme"}
			moved, err := svc.MoveBbtchChbnge(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := moved.Nbme, opts.NewNbme; hbve != wbnt {
				t.Fbtblf("wrong nbme. wbnt=%q, hbve=%q", wbnt, hbve)
			}
		})

		t.Run("new user nbmespbce", func(t *testing.T) {
			bbtchChbnge := crebteBbtchChbnge(t, "old-nbme", bdmin.ID, bdmin.ID, 0)

			user2 := bt.CrebteTestUser(t, db, fblse)

			opts := MoveBbtchChbngeOpts{BbtchChbngeID: bbtchChbnge.ID, NewNbmespbceUserID: user2.ID}
			moved, err := svc.MoveBbtchChbnge(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := moved.NbmespbceUserID, opts.NewNbmespbceUserID; hbve != wbnt {
				t.Fbtblf("wrong NbmespbceUserID. wbnt=%d, hbve=%d", wbnt, hbve)
			}

			if hbve, wbnt := moved.NbmespbceOrgID, opts.NewNbmespbceOrgID; hbve != wbnt {
				t.Fbtblf("wrong NbmespbceOrgID. wbnt=%d, hbve=%d", wbnt, hbve)
			}
		})

		t.Run("new user nbmespbce but current user is not bdmin", func(t *testing.T) {
			bbtchChbnge := crebteBbtchChbnge(t, "old-nbme", user.ID, user.ID, 0)

			user2 := bt.CrebteTestUser(t, db, fblse)

			opts := MoveBbtchChbngeOpts{BbtchChbngeID: bbtchChbnge.ID, NewNbmespbceUserID: user2.ID}

			_, err := svc.MoveBbtchChbnge(userCtx, opts)
			if !errcode.IsUnbuthorized(err) {
				t.Fbtblf("expected unbuthorized error but got %s", err)
			}
		})

		t.Run("new org nbmespbce", func(t *testing.T) {
			bbtchChbnge := crebteBbtchChbnge(t, "old-nbme-1", bdmin.ID, bdmin.ID, 0)

			orgID := bt.CrebteTestOrg(t, db, "org").ID

			opts := MoveBbtchChbngeOpts{BbtchChbngeID: bbtchChbnge.ID, NewNbmespbceOrgID: orgID}
			moved, err := svc.MoveBbtchChbnge(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := moved.NbmespbceUserID, opts.NewNbmespbceUserID; hbve != wbnt {
				t.Fbtblf("wrong NbmespbceUserID. wbnt=%d, hbve=%d", wbnt, hbve)
			}

			if hbve, wbnt := moved.NbmespbceOrgID, opts.NewNbmespbceOrgID; hbve != wbnt {
				t.Fbtblf("wrong NbmespbceOrgID. wbnt=%d, hbve=%d", wbnt, hbve)
			}
		})

		t.Run("new org nbmespbce but current user is missing bccess", func(t *testing.T) {
			bbtchChbnge := crebteBbtchChbnge(t, "old-nbme-2", user.ID, user.ID, 0)

			orgID := bt.CrebteTestOrg(t, db, "org-no-bccess").ID

			opts := MoveBbtchChbngeOpts{BbtchChbngeID: bbtchChbnge.ID, NewNbmespbceOrgID: orgID}

			_, err := svc.MoveBbtchChbnge(userCtx, opts)
			if hbve, wbnt := err, buth.ErrNotAnOrgMember; !errors.Is(hbve, wbnt) {
				t.Fbtblf("expected %s error but got %s", wbnt, hbve)
			}
		})
	})

	t.Run("GetBbtchChbngeMbtchingBbtchSpec", func(t *testing.T) {
		bbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "mbtching-bbtch-spec", bdmin.ID, 0)

		hbveBbtchChbnge, err := svc.GetBbtchChbngeMbtchingBbtchSpec(ctx, bbtchSpec)
		if err != nil {
			t.Fbtblf("unexpected error: %s\n", err)
		}
		if hbveBbtchChbnge != nil {
			t.Fbtblf("expected bbtch chbnge to be nil, but is not: %+v\n", hbveBbtchChbnge)
		}

		mbtchingBbtchChbnge := &btypes.BbtchChbnge{
			Nbme:            bbtchSpec.Spec.Nbme,
			Description:     bbtchSpec.Spec.Description,
			CrebtorID:       bdmin.ID,
			NbmespbceOrgID:  bbtchSpec.NbmespbceOrgID,
			NbmespbceUserID: bbtchSpec.NbmespbceUserID,
			BbtchSpecID:     bbtchSpec.ID,
			LbstApplierID:   bdmin.ID,
			LbstAppliedAt:   time.Now(),
		}
		if err := s.CrebteBbtchChbnge(ctx, mbtchingBbtchChbnge); err != nil {
			t.Fbtblf("fbiled to crebte bbtch chbnge: %s\n", err)
		}

		t.Run("BbtchChbngeID is not provided", func(t *testing.T) {
			hbveBbtchChbnge, err = svc.GetBbtchChbngeMbtchingBbtchSpec(ctx, bbtchSpec)
			if err != nil {
				t.Fbtblf("unexpected error: %s\n", err)
			}
			if hbveBbtchChbnge == nil {
				t.Fbtblf("expected to hbve mbtching bbtch chbnge, but got nil")
			}

			if diff := cmp.Diff(mbtchingBbtchChbnge, hbveBbtchChbnge); diff != "" {
				t.Fbtblf("wrong bbtch chbnge wbs mbtched (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("BbtchChbngeID is provided", func(t *testing.T) {
			bbtchSpec2 := bt.CrebteBbtchSpec(t, ctx, s, "mbtching-bbtch-spec", bdmin.ID, mbtchingBbtchChbnge.ID)
			hbveBbtchChbnge, err = svc.GetBbtchChbngeMbtchingBbtchSpec(ctx, bbtchSpec2)
			if err != nil {
				t.Fbtblf("unexpected error: %s\n", err)
			}
			if hbveBbtchChbnge == nil {
				t.Fbtblf("expected to hbve mbtching bbtch chbnge, but got nil")
			}

			if diff := cmp.Diff(mbtchingBbtchChbnge, hbveBbtchChbnge); diff != "" {
				t.Fbtblf("wrong bbtch chbnge wbs mbtched (-wbnt +got):\n%s", diff)
			}
		})
	})

	t.Run("GetNewestBbtchSpec", func(t *testing.T) {
		older := bt.CrebteBbtchSpec(t, ctx, s, "superseding", user.ID, 0)
		newer := bt.CrebteBbtchSpec(t, ctx, s, "superseding", user.ID, 0)

		for nbme, in := rbnge mbp[string]*btypes.BbtchSpec{
			"older": older,
			"newer": newer,
		} {
			t.Run(nbme, func(t *testing.T) {
				hbve, err := svc.GetNewestBbtchSpec(ctx, s, in, user.ID)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(newer, hbve); diff != "" {
					t.Errorf("unexpected newer bbtch spec (-wbnt +hbve):\n%s", diff)
				}
			})
		}

		t.Run("different user", func(t *testing.T) {
			hbve, err := svc.GetNewestBbtchSpec(ctx, s, older, bdmin.ID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if hbve != nil {
				t.Errorf("unexpected non-nil bbtch spec: %+v", hbve)
			}
		})
	})

	t.Run("FetchUsernbmeForBitbucketServerToken", func(t *testing.T) {
		fbkeSource := &stesting.FbkeChbngesetSource{Usernbme: "my-bbs-usernbme"}
		sourcer := stesting.NewFbkeSourcer(nil, fbkeSource)

		// Crebte b fresh service for this test bs to not mess with stbte
		// possibly used by other tests.
		testSvc := New(s)
		testSvc.sourcer = sourcer

		rs, _ := bt.CrebteBbsTestRepos(t, ctx, db, 1)
		repo := rs[0]

		url := repo.ExternblRepo.ServiceID
		extType := repo.ExternblRepo.ServiceType

		usernbme, err := testSvc.FetchUsernbmeForBitbucketServerToken(ctx, url, extType, "my-token")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !fbkeSource.AuthenticbtedUsernbmeCblled {
			t.Errorf("service didn't cbll AuthenticbtedUsernbme")
		}

		if hbve, wbnt := usernbme, fbkeSource.Usernbme; hbve != wbnt {
			t.Errorf("wrong usernbme returned. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("VblidbteAuthenticbtor", func(t *testing.T) {
		t.Run("vblid", func(t *testing.T) {
			fbkeSource.AuthenticbtorIsVblid = true
			fbkeSource.VblidbteAuthenticbtorCblled = fblse
			if err := svc.VblidbteAuthenticbtor(
				ctx,
				"https://github.com/",
				extsvc.TypeGitHub,
				&extsvcbuth.OAuthBebrerToken{Token: "test123"},
			); err != nil {
				t.Fbtbl(err)
			}
			if !fbkeSource.VblidbteAuthenticbtorCblled {
				t.Fbtbl("VblidbteAuthenticbtor on Source not cblled")
			}
		})
		t.Run("invblid", func(t *testing.T) {
			fbkeSource.AuthenticbtorIsVblid = fblse
			fbkeSource.VblidbteAuthenticbtorCblled = fblse
			if err := svc.VblidbteAuthenticbtor(
				ctx,
				"https://github.com/",
				extsvc.TypeGitHub,
				&extsvcbuth.OAuthBebrerToken{Token: "test123"},
			); err == nil {
				t.Fbtbl("unexpected nil-error returned from VblidbteAuthenticbtor")
			}
			if !fbkeSource.VblidbteAuthenticbtorCblled {
				t.Fbtbl("VblidbteAuthenticbtor on Source not cblled")
			}
		})
	})

	t.Run("CrebteChbngesetJobs", func(t *testing.T) {
		spec := testBbtchSpec(bdmin.ID)
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
		if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		t.Run("crebtes jobs", func(t *testing.T) {
			chbngeset1 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:             rs[0].ID,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:      bbtchChbnge.ID,
			})
			chbngeset2 := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:             rs[1].ID,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:      bbtchChbnge.ID,
			})
			bulkOperbtionID, err := svc.CrebteChbngesetJobs(
				bdminCtx,
				bbtchChbnge.ID,
				[]int64{chbngeset1.ID, chbngeset2.ID},
				btypes.ChbngesetJobTypeComment,
				btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
				store.ListChbngesetsOpts{},
			)
			if err != nil {
				t.Fbtbl(err)
			}
			// Vblidbte the bulk operbtion exists.
			if _, err = s.GetBulkOperbtion(ctx, store.GetBulkOperbtionOpts{ID: bulkOperbtionID}); err != nil {
				t.Fbtbl(err)
			}
		})

		t.Run("chbngeset not found", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:        rs[0].ID,
				BbtchChbnge: bbtchChbnge.ID,
			})
			_, err := svc.CrebteChbngesetJobs(
				bdminCtx,
				bbtchChbnge.ID,
				[]int64{chbngeset.ID},
				btypes.ChbngesetJobTypeComment,
				btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
				store.ListChbngesetsOpts{
					ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteCompleted},
				},
			)
			if err != ErrChbngesetsForJobNotFound {
				t.Fbtblf("wrong error. wbnt=%s, got=%s", ErrChbngesetsForJobNotFound, err)
			}
		})

		t.Run("DetbchChbngesets", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}
			t.Run("bttbched chbngeset", func(t *testing.T) {
				chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
					Repo:             rs[0].ID,
					ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
					PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
					BbtchChbnge:      bbtchChbnge.ID,
					IsArchived:       fblse,
				})
				_, err := svc.CrebteChbngesetJobs(ctx, bbtchChbnge.ID, []int64{chbngeset.ID}, btypes.ChbngesetJobTypeDetbch, btypes.ChbngesetJobDetbchPbylobd{}, store.ListChbngesetsOpts{OnlyArchived: true})
				if err != ErrChbngesetsForJobNotFound {
					t.Fbtblf("wrong error. wbnt=%s, got=%s", ErrChbngesetsForJobNotFound, err)
				}
			})
			t.Run("detbched chbngeset", func(t *testing.T) {
				detbchedChbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
					Repo:             rs[2].ID,
					ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
					PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
					ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
					BbtchChbnges:     []btypes.BbtchChbngeAssoc{},
				})
				_, err := svc.CrebteChbngesetJobs(ctx, bbtchChbnge.ID, []int64{detbchedChbngeset.ID}, btypes.ChbngesetJobTypeDetbch, btypes.ChbngesetJobDetbchPbylobd{}, store.ListChbngesetsOpts{OnlyArchived: true})
				if err != ErrChbngesetsForJobNotFound {
					t.Fbtblf("wrong error. wbnt=%s, got=%s", ErrChbngesetsForJobNotFound, err)
				}
			})
		})

		t.Run("MergeChbngesets", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}
			published := btypes.ChbngesetPublicbtionStbtePublished
			openStbte := btypes.ChbngesetExternblStbteOpen
			t.Run("open chbngeset", func(t *testing.T) {
				chbngeset := bt.CrebteChbngeset(t, bdminCtx, s, bt.TestChbngesetOpts{
					Repo:             rs[0].ID,
					ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
					ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
					PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
					BbtchChbnge:      bbtchChbnge.ID,
					IsArchived:       fblse,
				})
				_, err := svc.CrebteChbngesetJobs(
					bdminCtx,
					bbtchChbnge.ID,
					[]int64{chbngeset.ID},
					btypes.ChbngesetJobTypeMerge,
					btypes.ChbngesetJobMergePbylobd{Squbsh: true},
					store.ListChbngesetsOpts{
						PublicbtionStbte: &published,
						ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteCompleted},
						ExternblStbtes:   []btypes.ChbngesetExternblStbte{openStbte},
					},
				)
				if err != nil {
					t.Fbtbl(err)
				}
			})
			t.Run("closed chbngeset", func(t *testing.T) {
				closedChbngeset := bt.CrebteChbngeset(t, bdminCtx, s, bt.TestChbngesetOpts{
					Repo:             rs[0].ID,
					ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
					ExternblStbte:    btypes.ChbngesetExternblStbteClosed,
					PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
					BbtchChbnge:      bbtchChbnge.ID,
				})
				_, err := svc.CrebteChbngesetJobs(
					bdminCtx,
					bbtchChbnge.ID,
					[]int64{closedChbngeset.ID},
					btypes.ChbngesetJobTypeMerge,
					btypes.ChbngesetJobMergePbylobd{},
					store.ListChbngesetsOpts{
						PublicbtionStbte: &published,
						ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteCompleted},
						ExternblStbtes:   []btypes.ChbngesetExternblStbte{openStbte},
					},
				)
				if err != ErrChbngesetsForJobNotFound {
					t.Fbtblf("wrong error. wbnt=%s, got=%s", ErrChbngesetsForJobNotFound, err)
				}
			})
		})

		t.Run("CloseChbngesets", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}
			published := btypes.ChbngesetPublicbtionStbtePublished
			openStbte := btypes.ChbngesetExternblStbteOpen
			t.Run("open chbngeset", func(t *testing.T) {
				chbngeset := bt.CrebteChbngeset(t, bdminCtx, s, bt.TestChbngesetOpts{
					Repo:             rs[0].ID,
					ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
					ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
					PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
					BbtchChbnge:      bbtchChbnge.ID,
					IsArchived:       fblse,
				})
				_, err := svc.CrebteChbngesetJobs(
					bdminCtx,
					bbtchChbnge.ID,
					[]int64{chbngeset.ID},
					btypes.ChbngesetJobTypeClose,
					btypes.ChbngesetJobClosePbylobd{},
					store.ListChbngesetsOpts{
						PublicbtionStbte: &published,
						ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteCompleted},
						ExternblStbtes:   []btypes.ChbngesetExternblStbte{openStbte},
					},
				)
				if err != nil {
					t.Fbtbl(err)
				}
			})
			t.Run("closed chbngeset", func(t *testing.T) {
				closedChbngeset := bt.CrebteChbngeset(t, bdminCtx, s, bt.TestChbngesetOpts{
					Repo:             rs[0].ID,
					ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
					ExternblStbte:    btypes.ChbngesetExternblStbteClosed,
					PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
					BbtchChbnge:      bbtchChbnge.ID,
				})
				_, err := svc.CrebteChbngesetJobs(
					bdminCtx,
					bbtchChbnge.ID,
					[]int64{closedChbngeset.ID},
					btypes.ChbngesetJobTypeClose,
					btypes.ChbngesetJobClosePbylobd{},
					store.ListChbngesetsOpts{
						PublicbtionStbte: &published,
						ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteCompleted},
						ExternblStbtes:   []btypes.ChbngesetExternblStbte{openStbte},
					},
				)
				if err != ErrChbngesetsForJobNotFound {
					t.Fbtblf("wrong error. wbnt=%s, got=%s", ErrChbngesetsForJobNotFound, err)
				}
			})
		})

		t.Run("PublishChbngesets", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}

			chbngeset := bt.CrebteChbngeset(t, bdminCtx, s, bt.TestChbngesetOpts{
				Repo:             rs[0].ID,
				ReconcilerStbte:  btypes.ReconcilerStbteCompleted,
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				BbtchChbnge:      bbtchChbnge.ID,
			})

			_, err := svc.CrebteChbngesetJobs(
				bdminCtx,
				bbtchChbnge.ID,
				[]int64{chbngeset.ID},
				btypes.ChbngesetJobTypePublish,
				btypes.ChbngesetJobPublishPbylobd{Drbft: true},
				store.ListChbngesetsOpts{},
			)
			if err != nil {
				t.Fbtbl(err)
			}
		})
	})

	t.Run("ExecuteBbtchSpec", func(t *testing.T) {
		bdminCtx := bctor.WithActor(ctx, bctor.FromUser(bdmin.ID))
		t.Run("success", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			// Simulbte successful resolution.
			job := &btypes.BbtchSpecResolutionJob{
				Stbte:       btypes.BbtchSpecResolutionJobStbteCompleted,
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}

			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			vbr workspbceIDs []int64
			for _, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{
					BbtchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}
				workspbceIDs = bppend(workspbceIDs, ws.ID)
			}

			// Execute BbtchSpec by crebting execution jobs
			if _, err := svc.ExecuteBbtchSpec(bdminCtx, ExecuteBbtchSpecOpts{BbtchSpecRbndID: spec.RbndID}); err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, store.ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: workspbceIDs,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(jobs) != len(rs) {
				t.Fbtblf("wrong number of execution jobs crebted. wbnt=%d, hbve=%d", len(rs), len(jobs))
			}
		})

		t.Run("cbching disbbled", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			// Simulbte successful resolution.
			job := &btypes.BbtchSpecResolutionJob{
				Stbte:       btypes.BbtchSpecResolutionJobStbteCompleted,
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}

			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			cs := &btypes.ChbngesetSpec{
				Title:      "test",
				BbseRepoID: rs[0].ID,
			}
			if err := s.CrebteChbngesetSpec(ctx, cs); err != nil {
				t.Fbtbl(err)
			}

			vbr workspbceIDs []int64
			for _, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{
					BbtchSpecID:       spec.ID,
					RepoID:            repo.ID,
					CbchedResultFound: true,
					StepCbcheResults:  mbp[int]btypes.StepCbcheResult{1: {}},
					ChbngesetSpecIDs:  []int64{cs.ID},
				}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}
				workspbceIDs = bppend(workspbceIDs, ws.ID)
			}

			tru := true
			// Execute BbtchSpec by crebting execution jobs
			if _, err := svc.ExecuteBbtchSpec(bdminCtx, ExecuteBbtchSpecOpts{
				BbtchSpecRbndID: spec.RbndID,
				// Disbble cbching.
				NoCbche: &tru,
			}); err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, store.ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: workspbceIDs,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(jobs) != len(rs) {
				t.Fbtblf("wrong number of execution jobs crebted. wbnt=%d, hbve=%d", len(rs), len(jobs))
			}

			ws, _, err := s.ListBbtchSpecWorkspbces(ctx, store.ListBbtchSpecWorkspbcesOpts{IDs: workspbceIDs})
			if err != nil {
				t.Fbtbl(err)
			}
			for _, w := rbnge ws {
				if w.CbchedResultFound {
					t.Error("cbched_result_found not reset")
				}
				if len(w.StepCbcheResults) > 0 {
					t.Error("step_cbche_results not reset")
				}
				if len(w.ChbngesetSpecIDs) > 0 {
					t.Error("chbngeset_spec_ids not reset")
				}
			}

			// Verify the chbngeset spec hbs been deleted.
			if _, err := s.GetChbngesetSpecByID(ctx, cs.ID); err == nil || err != store.ErrNoResults {
				t.Fbtbl(err)
			}

			// Verify the bbtch spec no_cbche flbg hbs been updbted.
			relobdedSpec, err := s.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: spec.ID})
			if err != nil {
				t.Fbtbl(err)
			}
			if !relobdedSpec.NoCbche {
				t.Error("no_cbche flbg on bbtch spec not updbted")
			}
		})

		t.Run("resolution not completed", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			job := &btypes.BbtchSpecResolutionJob{
				Stbte:       btypes.BbtchSpecResolutionJobStbteQueued,
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}

			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			// Execute BbtchSpec by crebting execution jobs
			_, err := svc.ExecuteBbtchSpec(bdminCtx, ExecuteBbtchSpecOpts{BbtchSpecRbndID: spec.RbndID})
			if !errors.Is(err, ErrBbtchSpecResolutionIncomplete) {
				t.Fbtblf("error hbs wrong type: %T", err)
			}
		})

		t.Run("resolution fbiled", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			fbilureMessbge := "cbt bte the homework"
			job := &btypes.BbtchSpecResolutionJob{
				Stbte:          btypes.BbtchSpecResolutionJobStbteFbiled,
				FbilureMessbge: &fbilureMessbge,
				BbtchSpecID:    spec.ID,
				InitibtorID:    bdmin.ID,
			}

			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			// Execute BbtchSpec by crebting execution jobs
			_, err := svc.ExecuteBbtchSpec(bdminCtx, ExecuteBbtchSpecOpts{BbtchSpecRbndID: spec.RbndID})
			if !errors.HbsType(err, ErrBbtchSpecResolutionErrored{}) {
				t.Fbtblf("error hbs wrong type: %T", err)
			}
		})

		t.Run("ignored/unsupported workspbce", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			// Simulbte successful resolution.
			job := &btypes.BbtchSpecResolutionJob{
				Stbte:       btypes.BbtchSpecResolutionJobStbteCompleted,
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}

			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			ignoredWorkspbce := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
				Ignored:     true,
			}

			unsupportedWorkspbce := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
				Unsupported: true,
			}
			if err := s.CrebteBbtchSpecWorkspbce(ctx, ignoredWorkspbce, unsupportedWorkspbce); err != nil {
				t.Fbtbl(err)
			}

			if _, err := svc.ExecuteBbtchSpec(bdminCtx, ExecuteBbtchSpecOpts{BbtchSpecRbndID: spec.RbndID}); err != nil {
				t.Fbtbl(err)
			}

			ids := []int64{ignoredWorkspbce.ID, unsupportedWorkspbce.ID}
			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, store.ListBbtchSpecWorkspbceExecutionJobsOpts{
				BbtchSpecWorkspbceIDs: ids,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(jobs) != 0 {
				t.Fbtblf("wrong number of execution jobs crebted. wbnt=%d, hbve=%d", len(rs), len(jobs))
			}

			for _, workspbceID := rbnge ids {
				relobded, err := s.GetBbtchSpecWorkspbce(ctx, store.GetBbtchSpecWorkspbceOpts{ID: workspbceID})
				if err != nil {
					t.Fbtbl(err)
				}
				if !relobded.Skipped {
					t.Fbtblf("workspbce not mbrked bs skipped")
				}
			}
		})
	})

	t.Run("CbncelBbtchSpec", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			spec.CrebtedFromRbw = true
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			// Simulbte successful resolution.
			job := &btypes.BbtchSpecResolutionJob{
				Stbte:       btypes.BbtchSpecResolutionJobStbteCompleted,
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}

			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			vbr jobIDs []int64
			for _, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: repo.ID}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}

				job := &btypes.BbtchSpecWorkspbceExecutionJob{
					BbtchSpecWorkspbceID: ws.ID,
					UserID:               user.ID,
				}
				if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, store.ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
					t.Fbtbl(err)
				}

				job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
				job.StbrtedAt = time.Now()
				bt.UpdbteJobStbte(t, ctx, s, job)

				jobIDs = bppend(jobIDs, job.ID)
			}

			if _, err := svc.CbncelBbtchSpec(ctx, CbncelBbtchSpecOpts{BbtchSpecRbndID: spec.RbndID}); err != nil {
				t.Fbtbl(err)
			}

			jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(ctx, store.ListBbtchSpecWorkspbceExecutionJobsOpts{
				IDs: jobIDs,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if len(jobs) != len(rs) {
				t.Fbtblf("wrong number of execution jobs crebted. wbnt=%d, hbve=%d", len(rs), len(jobs))
			}

			vbr cbnceled int
			for _, j := rbnge jobs {
				if j.Cbncel {
					cbnceled += 1
				}
			}
			if cbnceled != len(jobs) {
				t.Fbtblf("not bll jobs were cbnceled. jobs=%d, cbnceled=%d", len(jobs), cbnceled)
			}
		})

		t.Run("blrebdy completed", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			spec.CrebtedFromRbw = true
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			resolutionJob := &btypes.BbtchSpecResolutionJob{
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}
			if err := s.CrebteBbtchSpecResolutionJob(ctx, resolutionJob); err != nil {
				t.Fbtbl(err)
			}

			ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: rs[0].ID}
			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			job := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				UserID:               user.ID,
			}
			if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, store.ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
				t.Fbtbl(err)
			}

			job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted
			job.StbrtedAt = time.Now()
			job.FinishedAt = time.Now()
			bt.UpdbteJobStbte(t, ctx, s, job)

			_, err := svc.CbncelBbtchSpec(ctx, CbncelBbtchSpecOpts{BbtchSpecRbndID: spec.RbndID})
			if !errors.Is(err, ErrBbtchSpecNotCbncelbble) {
				t.Fbtblf("error hbs wrong type: %T", err)
			}
		})
	})

	t.Run("ReplbceBbtchSpecInput", func(t *testing.T) {
		crebteBbtchSpecWithWorkspbces := func(t *testing.T) *btypes.BbtchSpec {
			t.Helper()
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			job := &btypes.BbtchSpecResolutionJob{
				Stbte:       btypes.BbtchSpecResolutionJobStbteCompleted,
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}

			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			for _, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: repo.ID}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}
			}
			return spec
		}

		crebteBbtchSpecWithWorkspbcesAndChbngesetSpecs := func(t *testing.T) *btypes.BbtchSpec {
			t.Helper()

			spec := crebteBbtchSpecWithWorkspbces(t)

			for _, r := rbnge rs {
				bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
					BbtchSpec: spec.ID,
					Repo:      r.ID,
					Typ:       btypes.ChbngesetSpecTypeBrbnch,
				})
			}

			return spec
		}

		t.Run("success", func(t *testing.T) {
			spec := crebteBbtchSpecWithWorkspbces(t)

			newSpec, err := svc.ReplbceBbtchSpecInput(ctx, ReplbceBbtchSpecInputOpts{
				BbtchSpecRbndID: spec.RbndID,
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if newSpec.ID == spec.ID {
				t.Fbtblf("new bbtch spec hbs sbme ID bs old one: %d", newSpec.ID)
			}

			if newSpec.RbndID != spec.RbndID {
				t.Fbtblf("new bbtch spec hbs different RbndID. new=%s, old=%s", newSpec.RbndID, spec.RbndID)
			}
			if newSpec.UserID != spec.UserID {
				t.Fbtblf("new bbtch spec hbs different UserID. new=%d, old=%d", newSpec.UserID, spec.UserID)
			}
			if newSpec.NbmespbceUserID != spec.NbmespbceUserID {
				t.Fbtblf("new bbtch spec hbs different NbmespbceUserID. new=%d, old=%d", newSpec.NbmespbceUserID, spec.NbmespbceUserID)
			}
			if newSpec.NbmespbceOrgID != spec.NbmespbceOrgID {
				t.Fbtblf("new bbtch spec hbs different NbmespbceOrgID. new=%d, old=%d", newSpec.NbmespbceOrgID, spec.NbmespbceOrgID)
			}

			if !newSpec.CrebtedFromRbw {
				t.Fbtblf("new bbtch spec not crebtedFromRbw: %t", newSpec.CrebtedFromRbw)
			}

			resolutionJob, err := s.GetBbtchSpecResolutionJob(ctx, store.GetBbtchSpecResolutionJobOpts{
				BbtchSpecID: newSpec.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if wbnt, hbve := btypes.BbtchSpecResolutionJobStbteQueued, resolutionJob.Stbte; hbve != wbnt {
				t.Fbtblf("resolution job hbs wrong stbte. wbnt=%s, hbve=%s", wbnt, hbve)
			}

			// Assert thbt old bbtch spec is deleted
			_, err = s.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: spec.ID})
			if err != store.ErrNoResults {
				t.Fbtblf("unexpected error: %s", err)
			}
		})

		tests := []struct {
			nbme    string
			rbwSpec string
			wbntErr error
		}{
			{
				nbme:    "empty",
				rbwSpec: "",
				wbntErr: errors.New("Expected: object, given: null"),
			},
			{
				nbme:    "invblid YAML",
				rbwSpec: "invblid YAML",
				wbntErr: errors.New("Expected: object, given: string"),
			},
			{
				nbme:    "invblid nbme",
				rbwSpec: "nbme: invblid nbme",
				wbntErr: errors.New("The bbtch chbnge nbme cbn only contbin word chbrbcters, dots bnd dbshes. No whitespbce or newlines bllowed."),
			},
			{
				nbme: "requires chbngesetTemplbte when steps bre included",
				rbwSpec: `
nbme: test
on:
  - repository: github.com/sourcegrbph-testing/some-repo
steps:
  - run: echo "Hello world"
    contbiner: blpine:3`,
				wbntErr: errors.New("bbtch spec includes steps but no chbngesetTemplbte"),
			},
			{
				nbme: "unknown templbting vbribble",
				rbwSpec: `
nbme: hello
on:
  - repository: github.com/sourcegrbph-testing/some-repo
steps:
  - run: echo "Hello ${{ resopitory.nbme }}" >> messbge.txt
    contbiner: blpine:3
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Write b messbge to b text file
`,
				wbntErr: errors.New("unknown templbting vbribble: 'resopitory'"),
			},
		}

		for _, tc := rbnge tests {
			t.Run("bbtchSpec hbs invblid rbw spec: "+tc.nbme, func(t *testing.T) {
				spec := crebteBbtchSpecWithWorkspbces(t)

				_, gotErr := svc.ReplbceBbtchSpecInput(ctx, ReplbceBbtchSpecInputOpts{
					BbtchSpecRbndID: spec.RbndID,
					RbwSpec:         tc.rbwSpec,
				})

				if gotErr == nil {
					t.Fbtblf("unexpected nil error.\nwbnt=%s\n---\ngot=nil", tc.wbntErr)
				}

				if !strings.Contbins(gotErr.Error(), tc.wbntErr.Error()) {
					t.Fbtblf("unexpected error.\nwbnt=%s\n---\ngot=%s", tc.wbntErr, gotErr)
				}
			})
		}

		t.Run("bbtchSpec blrebdy hbs chbngeset specs", func(t *testing.T) {
			bssertNoChbngesetSpecs := func(t *testing.T, bbtchSpecID int64) {
				t.Helper()
				specs, _, err := s.ListChbngesetSpecs(ctx, store.ListChbngesetSpecsOpts{
					BbtchSpecID: bbtchSpecID,
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if len(specs) != 0 {
					t.Fbtblf("wrong number of chbngeset specs bttbched to bbtch spec %d: %d", bbtchSpecID, len(specs))
				}
			}

			spec := crebteBbtchSpecWithWorkspbcesAndChbngesetSpecs(t)

			newSpec, err := svc.ReplbceBbtchSpecInput(ctx, ReplbceBbtchSpecInputOpts{
				BbtchSpecRbndID: spec.RbndID,
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			bssertNoChbngesetSpecs(t, newSpec.ID)
			bssertNoChbngesetSpecs(t, spec.ID)
		})

		t.Run("hbs mount", func(t *testing.T) {
			spec := crebteBbtchSpecWithWorkspbcesAndChbngesetSpecs(t)

			_, err := svc.ReplbceBbtchSpecInput(bdminCtx, ReplbceBbtchSpecInputOpts{
				BbtchSpecRbndID: spec.RbndID,
				RbwSpec: `
nbme: test-spec
description: A test spec
steps:
  - run: /tmp/sbmple.sh
    contbiner: blpine:3
    mount:
      - pbth: /some/pbth/sbmple.sh
        mountpoint: /tmp/sbmple.sh
chbngesetTemplbte:
  title: Test Mount
  body: Test b mounted pbth
  brbnch: test
  commit:
    messbge: Test
`,
			})
			bssert.NoError(t, err)
		})
	})

	t.Run("CrebteBbtchSpecFromRbw", func(t *testing.T) {
		t.Run("bbtch chbnge isn't owned by non-bdmin user", func(t *testing.T) {
			spec := testBbtchSpec(user.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(user.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}

			_, err := svc.CrebteBbtchSpecFromRbw(user2Ctx, CrebteBbtchSpecFromRbwOpts{
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
				NbmespbceUserID: user2.ID,
				BbtchChbnge:     bbtchChbnge.ID,
			})

			bssert.Equbl(t, buth.ErrMustBeSiteAdminOrSbmeUser.Error(), err.Error())
		})

		t.Run("success - without bbtch chbnge ID", func(t *testing.T) {
			newSpec, err := svc.CrebteBbtchSpecFromRbw(bdminCtx, CrebteBbtchSpecFromRbwOpts{
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
				NbmespbceUserID: bdmin.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if !newSpec.CrebtedFromRbw {
				t.Fbtblf("bbtchSpec not crebtedFromRbw: %t", newSpec.CrebtedFromRbw)
			}

			resolutionJob, err := s.GetBbtchSpecResolutionJob(ctx, store.GetBbtchSpecResolutionJobOpts{
				BbtchSpecID: newSpec.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if wbnt, hbve := btypes.BbtchSpecResolutionJobStbteQueued, resolutionJob.Stbte; hbve != wbnt {
				t.Fbtblf("resolution job hbs wrong stbte. wbnt=%s, hbve=%s", wbnt, hbve)
			}
		})

		t.Run("success - with bbtch chbnge ID", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}

			newSpec, err := svc.CrebteBbtchSpecFromRbw(bdminCtx, CrebteBbtchSpecFromRbwOpts{
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
				NbmespbceUserID: bdmin.ID,
				BbtchChbnge:     bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if !newSpec.CrebtedFromRbw {
				t.Fbtblf("bbtchSpec not crebtedFromRbw: %t", newSpec.CrebtedFromRbw)
			}

			resolutionJob, err := s.GetBbtchSpecResolutionJob(ctx, store.GetBbtchSpecResolutionJobOpts{
				BbtchSpecID: newSpec.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if wbnt, hbve := btypes.BbtchSpecResolutionJobStbteQueued, resolutionJob.Stbte; hbve != wbnt {
				t.Fbtblf("resolution job hbs wrong stbte. wbnt=%s, hbve=%s", wbnt, hbve)
			}
		})

		t.Run("vblidbtion error", func(t *testing.T) {
			rbwSpec := bbtcheslib.BbtchSpec{
				Nbme:        "test-bbtch-chbnge",
				Description: "only importing",
				ImportChbngesets: []bbtcheslib.ImportChbngeset{
					{Repository: string(rs[0].Nbme), ExternblIDs: []bny{true, fblse}},
				},
			}

			mbrshbledRbwSpec, err := json.Mbrshbl(rbwSpec)
			if err != nil {
				t.Fbtbl(err)
			}

			_, err = svc.CrebteBbtchSpecFromRbw(bdminCtx, CrebteBbtchSpecFromRbwOpts{
				RbwSpec:         string(mbrshbledRbwSpec),
				NbmespbceUserID: bdmin.ID,
			})
			if err == nil {
				t.Fbtblf("expected error but got none")
			}
			if !strings.Contbins(err.Error(), "Invblid type. Expected: string, given: boolebn") {
				t.Fbtblf("wrong error messbge: %s", err)
			}
		})

		t.Run("hbs mount", func(t *testing.T) {
			_, err := svc.CrebteBbtchSpecFromRbw(bdminCtx, CrebteBbtchSpecFromRbwOpts{
				RbwSpec: `
nbme: test-spec
description: A test spec
steps:
  - run: /tmp/sbmple.sh
    contbiner: blpine:3
    mount:
      - pbth: /some/pbth/sbmple.sh
        mountpoint: /tmp/sbmple.sh
chbngesetTemplbte:
  title: Test Mount
  body: Test b mounted pbth
  brbnch: test
  commit:
    messbge: Test
`,
				NbmespbceUserID: bdmin.ID,
			})
			bssert.NoError(t, err)
		})
	})

	t.Run("UpsertBbtchSpecInput", func(t *testing.T) {
		bdminCtx := bctor.WithActor(ctx, bctor.FromUser(bdmin.ID))
		t.Run("new spec", func(t *testing.T) {
			newSpec, err := svc.UpsertBbtchSpecInput(bdminCtx, UpsertBbtchSpecInputOpts{
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
				NbmespbceUserID: bdmin.ID,
			})
			bssert.Nil(t, err)
			bssert.True(t, newSpec.CrebtedFromRbw)

			resolutionJob, err := s.GetBbtchSpecResolutionJob(ctx, store.GetBbtchSpecResolutionJobOpts{
				BbtchSpecID: newSpec.ID,
			})
			bssert.Nil(t, err)
			bssert.Equbl(t, btypes.BbtchSpecResolutionJobStbteQueued, resolutionJob.Stbte)
		})

		t.Run("replbced spec", func(t *testing.T) {
			oldSpec, err := svc.UpsertBbtchSpecInput(bdminCtx, UpsertBbtchSpecInputOpts{
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
				NbmespbceUserID: bdmin.ID,
			})
			bssert.Nil(t, err)
			bssert.True(t, oldSpec.CrebtedFromRbw)

			newSpec, err := svc.UpsertBbtchSpecInput(bdminCtx, UpsertBbtchSpecInputOpts{
				RbwSpec:         bt.TestRbwBbtchSpecYAML,
				NbmespbceUserID: bdmin.ID,
			})
			bssert.Nil(t, err)
			bssert.True(t, newSpec.CrebtedFromRbw)
			bssert.Equbl(t, oldSpec.RbndID, newSpec.RbndID)
			bssert.Equbl(t, oldSpec.NbmespbceUserID, newSpec.NbmespbceUserID)
			bssert.Equbl(t, oldSpec.NbmespbceOrgID, newSpec.NbmespbceOrgID)

			// Check thbt the replbced bbtch spec wbs reblly deleted.
			_, err = s.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{
				ID: oldSpec.ID,
			})
			bssert.Equbl(t, store.ErrNoResults, err)
		})

		t.Run("hbs mount", func(t *testing.T) {
			_, err := svc.UpsertBbtchSpecInput(bdminCtx, UpsertBbtchSpecInputOpts{
				RbwSpec: `
nbme: test-spec
description: A test spec
steps:
  - run: /tmp/sbmple.sh
    contbiner: blpine:3
    mount:
      - pbth: /some/pbth/sbmple.sh
        mountpoint: /tmp/sbmple.sh
chbngesetTemplbte:
  title: Test Mount
  body: Test b mounted pbth
  brbnch: test
  commit:
    messbge: Test
`,
				NbmespbceUserID: bdmin.ID,
			})
			bssert.NoError(t, err)
		})
	})

	t.Run("VblidbteChbngesetSpecs", func(t *testing.T) {
		bbtchSpec := bt.CrebteBbtchSpec(t, ctx, s, "mbtching-bbtch-spec", bdmin.ID, 0)
		conflictingRef := "refs/hebds/conflicting-hebd-ref"
		for _, opts := rbnge []bt.TestSpecOpts{
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[0].ID, BbtchSpec: bbtchSpec.ID},
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[1].ID, BbtchSpec: bbtchSpec.ID},
			{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[1].ID, BbtchSpec: bbtchSpec.ID},
			{HebdRef: conflictingRef + "-2", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[2].ID, BbtchSpec: bbtchSpec.ID},
			{HebdRef: conflictingRef + "-2", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[2].ID, BbtchSpec: bbtchSpec.ID},
			{HebdRef: conflictingRef + "-2", Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[2].ID, BbtchSpec: bbtchSpec.ID},
		} {
			bt.CrebteChbngesetSpec(t, ctx, s, opts)
		}
		err := svc.VblidbteChbngesetSpecs(ctx, bbtchSpec.ID)
		if err == nil {
			t.Fbtbl("expected error, but got none")
		}

		wbnt := `2 errors when vblidbting chbngeset specs:
* 2 chbngeset specs in repo-1-2 use the sbme brbnch: refs/hebds/conflicting-hebd-ref
* 3 chbngeset specs in repo-1-3 use the sbme brbnch: refs/hebds/conflicting-hebd-ref-2
`
		if diff := cmp.Diff(wbnt, err.Error()); diff != "" {
			t.Fbtblf("wrong error messbge: %s", diff)
		}
	})

	t.Run("ComputeBbtchSpecStbte", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			spec.CrebtedFromRbw = true
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			job := &btypes.BbtchSpecResolutionJob{
				BbtchSpecID: spec.ID,
				InitibtorID: bdmin.ID,
			}
			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			stbrtedAt := clock()
			for _, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: repo.ID}
				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}

				job := &btypes.BbtchSpecWorkspbceExecutionJob{BbtchSpecWorkspbceID: ws.ID, UserID: user.ID}
				if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, store.ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
					t.Fbtbl(err)
				}

				job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
				job.StbrtedAt = stbrtedAt
				bt.UpdbteJobStbte(t, ctx, s, job)
			}

			hbve, err := svc.LobdBbtchSpecStbts(ctx, spec)
			if err != nil {
				t.Fbtbl(err)
			}
			wbnt := btypes.BbtchSpecStbts{
				Workspbces: len(rs),
				Executions: len(rs),
				Processing: len(rs),
				StbrtedAt:  stbrtedAt,
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("wrong stbts: %s", diff)
			}
		})
	})

	t.Run("RetryBbtchSpecWorkspbces", func(t *testing.T) {
		fbilureMessbge := "this fbiled"

		t.Run("success", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			chbngesetSpec1 := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BbtchSpec: spec.ID,
				HebdRef:   "refs/hebds/my-spec",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			chbngesetSpec2 := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BbtchSpec: spec.ID,
				HebdRef:   "refs/hebds/my-spec-2",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			vbr workspbceIDs []int64
			for i, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{
					BbtchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				// This workspbce hbs the completed job bnd resulted in 2 chbngesetspecs
				if i == 2 {
					ws.ChbngesetSpecIDs = bppend(ws.ChbngesetSpecIDs, chbngesetSpec1.ID, chbngesetSpec2.ID)
				}

				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}
				workspbceIDs = bppend(workspbceIDs, ws.ID)
			}

			fbiledJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[0],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FbilureMessbge:       &fbilureMessbge,
			}
			crebteJob(t, s, fbiledJob)

			completedJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[2],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			crebteJob(t, s, completedJob)

			jobs := []*btypes.BbtchSpecWorkspbceExecutionJob{fbiledJob, completedJob}

			// RETRY
			if err := svc.RetryBbtchSpecWorkspbces(ctx, workspbceIDs); err != nil {
				t.Fbtbl(err)
			}

			bssertJobsDeleted(t, s, jobs)
			bssertChbngesetSpecsDeleted(t, s, []*btypes.ChbngesetSpec{chbngesetSpec1, chbngesetSpec2})
			bssertJobsCrebtedFor(t, s, []int64{workspbceIDs[0], workspbceIDs[1], workspbceIDs[2]})
		})

		t.Run("bbtch spec blrebdy bpplied", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bbtchChbnge := testBbtchChbnge(spec.UserID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}

			ws := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			fbiledJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FbilureMessbge:       &fbilureMessbge,
			}
			crebteJob(t, s, fbiledJob)

			// RETRY
			err := svc.RetryBbtchSpecWorkspbces(ctx, []int64{ws.ID})
			if err == nil {
				t.Fbtbl("no error")
			}
			if err.Error() != "bbtch spec blrebdy bpplied" {
				t.Fbtblf("wrong error: %s", err)
			}
		})

		t.Run("bbtch spec bssocibted with drbft bbtch chbnge", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			// Associbte with drbft bbtch chbnge
			bbtchChbnge := testDrbftBbtchChbnge(spec.UserID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}

			ws := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			fbiledJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FbilureMessbge:       &fbilureMessbge,
			}
			crebteJob(t, s, fbiledJob)

			// RETRY
			err := svc.RetryBbtchSpecWorkspbces(ctx, []int64{ws.ID})
			if err != nil {
				t.Fbtbl("unexpected error")
			}

			bssertJobsDeleted(t, s, []*btypes.BbtchSpecWorkspbceExecutionJob{fbiledJob})
			bssertJobsCrebtedFor(t, s, []int64{ws.ID})
		})

		t.Run("job not retrybble", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			ws := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			queuedJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteQueued,
			}
			crebteJob(t, s, queuedJob)

			// RETRY
			err := svc.RetryBbtchSpecWorkspbces(ctx, []int64{ws.ID})
			if err == nil {
				t.Fbtbl("no error")
			}
			if !strings.Contbins(err.Error(), "not retrybble") {
				t.Fbtblf("wrong error: %s", err)
			}
		})

		t.Run("user is not nbmespbce user bnd not bdmin", func(t *testing.T) {
			// bdmin owns bbtch spec
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			ws := testWorkspbce(spec.ID, rs[0].ID)
			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			queuedJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteQueued,
			}
			crebteJob(t, s, queuedJob)

			// userCtx uses user bs bctor
			err := svc.RetryBbtchSpecWorkspbces(userCtx, []int64{ws.ID})
			bssertAuthError(t, err)
		})
	})

	t.Run("RetryBbtchSpecExecution", func(t *testing.T) {
		fbilureMessbge := "this fbiled"

		crebteSpec := func(t *testing.T) *btypes.BbtchSpec {
			t.Helper()

			spec := testBbtchSpec(bdmin.ID)
			spec.CrebtedFromRbw = true
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}
			job := &btypes.BbtchSpecResolutionJob{
				BbtchSpecID: spec.ID,
				Stbte:       btypes.BbtchSpecResolutionJobStbteCompleted,
				InitibtorID: bdmin.ID,
			}
			if err := s.CrebteBbtchSpecResolutionJob(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			return spec
		}

		t.Run("success", func(t *testing.T) {
			spec := crebteSpec(t)

			chbngesetSpec1 := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BbtchSpec: spec.ID,
				HebdRef:   "refs/hebds/my-spec",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			chbngesetSpec2 := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BbtchSpec: spec.ID,
				HebdRef:   "refs/hebds/my-spec-2",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			vbr workspbceIDs []int64
			for i, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{
					BbtchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				// This workspbce hbs the completed job bnd resulted in 2 chbngesetspecs
				if i == 2 {
					ws.ChbngesetSpecIDs = bppend(ws.ChbngesetSpecIDs, chbngesetSpec1.ID, chbngesetSpec2.ID)
				}

				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}
				workspbceIDs = bppend(workspbceIDs, ws.ID)
			}

			fbiledJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[0],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FbilureMessbge:       &fbilureMessbge,
			}
			crebteJob(t, s, fbiledJob)

			completedJob1 := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[1],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			crebteJob(t, s, completedJob1)

			completedJob2 := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[2],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			crebteJob(t, s, completedJob2)

			// RETRY
			if err := svc.RetryBbtchSpecExecution(ctx, RetryBbtchSpecExecutionOpts{BbtchSpecRbndID: spec.RbndID}); err != nil {
				t.Fbtbl(err)
			}

			// Completed jobs should not be retried
			bssertJobsDeleted(t, s, []*btypes.BbtchSpecWorkspbceExecutionJob{fbiledJob})
			bssertChbngesetSpecsNotDeleted(t, s, []*btypes.ChbngesetSpec{chbngesetSpec1, chbngesetSpec2})
			bssertJobsCrebtedFor(t, s, []int64{workspbceIDs[0], workspbceIDs[1], workspbceIDs[2]})
		})

		t.Run("success with IncludeCompleted", func(t *testing.T) {
			spec := crebteSpec(t)

			chbngesetSpec1 := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BbtchSpec: spec.ID,
				HebdRef:   "refs/hebds/my-spec",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			chbngesetSpec2 := bt.CrebteChbngesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BbtchSpec: spec.ID,
				HebdRef:   "refs/hebds/my-spec-2",
				Typ:       btypes.ChbngesetSpecTypeBrbnch,
			})

			vbr workspbceIDs []int64
			for i, repo := rbnge rs {
				ws := &btypes.BbtchSpecWorkspbce{
					BbtchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				// This workspbce hbs the completed job bnd resulted in 2 chbngesetspecs
				if i == 2 {
					ws.ChbngesetSpecIDs = bppend(ws.ChbngesetSpecIDs, chbngesetSpec1.ID, chbngesetSpec2.ID)
				}

				if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
					t.Fbtbl(err)
				}
				workspbceIDs = bppend(workspbceIDs, ws.ID)
			}

			fbiledJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[0],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FbilureMessbge:       &fbilureMessbge,
			}
			crebteJob(t, s, fbiledJob)

			completedJob1 := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[1],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			crebteJob(t, s, completedJob1)

			completedJob2 := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: workspbceIDs[2],
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			crebteJob(t, s, completedJob2)

			// RETRY
			opts := RetryBbtchSpecExecutionOpts{BbtchSpecRbndID: spec.RbndID, IncludeCompleted: true}
			if err := svc.RetryBbtchSpecExecution(ctx, opts); err != nil {
				t.Fbtbl(err)
			}

			// Queued job should not be deleted
			bssertJobsDeleted(t, s, []*btypes.BbtchSpecWorkspbceExecutionJob{
				fbiledJob,
				completedJob1,
				completedJob2,
			})
			bssertChbngesetSpecsDeleted(t, s, []*btypes.ChbngesetSpec{chbngesetSpec1, chbngesetSpec2})
			bssertJobsCrebtedFor(t, s, []int64{workspbceIDs[0], workspbceIDs[1], workspbceIDs[2]})
		})

		t.Run("bbtch spec blrebdy bpplied", func(t *testing.T) {
			spec := crebteSpec(t)

			bbtchChbnge := testBbtchChbnge(spec.UserID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}

			ws := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			fbiledJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FbilureMessbge:       &fbilureMessbge,
			}
			crebteJob(t, s, fbiledJob)

			// RETRY
			err := svc.RetryBbtchSpecExecution(ctx, RetryBbtchSpecExecutionOpts{BbtchSpecRbndID: spec.RbndID})
			if err == nil {
				t.Fbtbl("no error")
			}
			if err.Error() != "bbtch spec blrebdy bpplied" {
				t.Fbtblf("wrong error: %s", err)
			}
		})

		t.Run("bbtch spec bssocibted with drbft bbtch chbnge", func(t *testing.T) {
			spec := testBbtchSpec(bdmin.ID)
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			// Associbte with drbft bbtch chbnge
			bbtchChbnge := testDrbftBbtchChbnge(spec.UserID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
				t.Fbtbl(err)
			}

			ws := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			fbiledJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled,
				StbrtedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FbilureMessbge:       &fbilureMessbge,
			}
			crebteJob(t, s, fbiledJob)

			// RETRY
			opts := RetryBbtchSpecExecutionOpts{BbtchSpecRbndID: spec.RbndID, IncludeCompleted: true}
			if err := svc.RetryBbtchSpecExecution(ctx, opts); err != nil {
				t.Fbtbl(err)
			}

			// Queued job should not be deleted
			bssertJobsDeleted(t, s, []*btypes.BbtchSpecWorkspbceExecutionJob{
				fbiledJob,
			})
			bssertJobsCrebtedFor(t, s, []int64{ws.ID})
		})

		t.Run("user is not nbmespbce user bnd not bdmin", func(t *testing.T) {
			// bdmin owns bbtch spec
			spec := crebteSpec(t)

			ws := testWorkspbce(spec.ID, rs[0].ID)
			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			queuedJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteQueued,
			}
			crebteJob(t, s, queuedJob)

			// userCtx uses user bs bctor
			err := svc.RetryBbtchSpecExecution(userCtx, RetryBbtchSpecExecutionOpts{BbtchSpecRbndID: spec.RbndID})
			bssertAuthError(t, err)
		})

		t.Run("bbtch spec not in finbl stbte", func(t *testing.T) {
			spec := crebteSpec(t)

			ws := &btypes.BbtchSpecWorkspbce{
				BbtchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
				t.Fbtbl(err)
			}

			queuedJob := &btypes.BbtchSpecWorkspbceExecutionJob{
				BbtchSpecWorkspbceID: ws.ID,
				Stbte:                btypes.BbtchSpecWorkspbceExecutionJobStbteQueued,
			}
			crebteJob(t, s, queuedJob)

			// RETRY
			err := svc.RetryBbtchSpecExecution(ctx, RetryBbtchSpecExecutionOpts{BbtchSpecRbndID: spec.RbndID})
			if err == nil {
				t.Fbtbl("no error")
			}
			if !errors.Is(err, ErrRetryNonFinbl) {
				t.Fbtblf("wrong error: %s", err)
			}
		})
	})

	t.Run("GetAvbilbbleBulkOperbtions", func(t *testing.T) {
		spec := testBbtchSpec(bdmin.ID)
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbnge := testBbtchChbnge(bdmin.ID, spec)
		if err := s.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		t.Run("fbiled chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteFbiled,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"REENQUEUE", "PUBLISH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("brchived chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,

				// brchived chbngeset
				IsArchived: true,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"DETACH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("unpublished chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbteUnpublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"PUBLISH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("drbft chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteDrbft,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"CLOSE", "COMMENT", "PUBLISH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("open chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"CLOSE", "COMMENT", "MERGE", "PUBLISH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("closed chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"COMMENT", "PUBLISH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("merged chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteMerged,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"COMMENT", "PUBLISH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("rebd-only chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteRebdOnly,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})

			bssert.NoError(t, err)
			bssert.Empty(t, bulkOperbtions)
		})

		t.Run("imported chbngesets", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:      btypes.ChbngesetExternblStbteOpen,
				OwnedByBbtchChbnge: 0,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					chbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})

			bssert.NoError(t, err)
			expectedBulkOperbtions := []string{"COMMENT", "CLOSE", "MERGE"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("drbft, brchived bnd fbiled chbngesets with no common bulk operbtion", func(t *testing.T) {
			fbiledChbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ReconcilerStbte:    btypes.ReconcilerStbteFbiled,
			})

			brchivedChbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,

				// brchived chbngeset
				IsArchived: true,
			})

			drbftChbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteDrbft,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					fbiledChbngeset.ID,
					brchivedChbngeset.ID,
					drbftChbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})

		t.Run("drbft, closed bnd merged chbngesets with b common bulk operbtion", func(t *testing.T) {
			drbftChbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteDrbft,
			})

			closedChbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteClosed,
			})

			mergedChbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:               rs[0].ID,
				PublicbtionStbte:   btypes.ChbngesetPublicbtionStbtePublished,
				BbtchChbnge:        bbtchChbnge.ID,
				OwnedByBbtchChbnge: bbtchChbnge.ID,
				ExternblStbte:      btypes.ChbngesetExternblStbteMerged,
			})

			bulkOperbtions, err := svc.GetAvbilbbleBulkOperbtions(ctx, GetAvbilbbleBulkOperbtionsOpts{
				Chbngesets: []int64{
					closedChbngeset.ID,
					mergedChbngeset.ID,
					drbftChbngeset.ID,
				},
				BbtchChbnge: bbtchChbnge.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			expectedBulkOperbtions := []string{"COMMENT", "PUBLISH"}
			if !bssert.ElementsMbtch(t, expectedBulkOperbtions, bulkOperbtions) {
				t.Errorf("wrong bulk operbtion type returned. wbnt=%q, hbve=%q", expectedBulkOperbtions, bulkOperbtions)
			}
		})
	})

	t.Run("UpsertEmptyBbtchChbnge", func(t *testing.T) {
		t.Run("crebtes new bbtch chbnge if it is non-existent", func(t *testing.T) {
			nbme := "rbndom-bc-nbme"

			// verify thbt the bbtch chbnge doesn't exist
			_, err := s.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{
				Nbme:            nbme,
				NbmespbceUserID: user.ID,
			})

			if err != store.ErrNoResults {
				t.Fbtblf("bbtch chbnge %s should not exist", nbme)
			}

			bbtchChbnge, err := svc.UpsertEmptyBbtchChbnge(ctx, UpsertEmptyBbtchChbngeOpts{
				Nbme:            nbme,
				NbmespbceUserID: user.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if bbtchChbnge.ID == 0 {
				t.Fbtblf("BbtchChbnge ID is 0")
			}

			if hbve, wbnt := bbtchChbnge.NbmespbceUserID, user.ID; hbve != wbnt {
				t.Fbtblf("UserID is %d, wbnt %d", hbve, wbnt)
			}
		})

		t.Run("returns existing Bbtch Chbnge", func(t *testing.T) {
			spec := &btypes.BbtchSpec{
				UserID:          user.ID,
				NbmespbceUserID: user.ID,
				NbmespbceOrgID:  0,
			}
			if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			bc := testBbtchChbnge(user.ID, spec)
			if err := s.CrebteBbtchChbnge(ctx, bc); err != nil {
				t.Fbtbl(err)
			}

			hbveBbtchChbnge, err := svc.UpsertEmptyBbtchChbnge(ctx, UpsertEmptyBbtchChbngeOpts{
				Nbme:            bc.Nbme,
				NbmespbceUserID: user.ID,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbveBbtchChbnge == nil {
				t.Fbtbl("expected to hbve mbtching bbtch chbnge, but got nil")
			}

			if hbveBbtchChbnge.ID == 0 {
				t.Fbtbl("BbtchChbnge ID is 0")
			}

			if hbveBbtchChbnge.ID != bc.ID {
				t.Fbtbl("expected sbme ID for bbtch chbnge")
			}

			if hbveBbtchChbnge.BbtchSpecID == bc.BbtchSpecID {
				t.Fbtbl("expected different spec ID for bbtch chbnge")
			}
		})
	})
}

func crebteJob(t *testing.T, s *store.Store, job *btypes.BbtchSpecWorkspbceExecutionJob) {
	t.Helper()

	if job.UserID == 0 {
		job.UserID = 1
	}

	clone := *job

	if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(context.Bbckground(), s, store.ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
		t.Fbtbl(err)
	}

	job.Stbte = clone.Stbte
	job.Cbncel = clone.Cbncel
	job.WorkerHostnbme = clone.WorkerHostnbme
	job.StbrtedAt = clone.StbrtedAt
	job.FinishedAt = clone.FinishedAt
	job.FbilureMessbge = clone.FbilureMessbge

	bt.UpdbteJobStbte(t, context.Bbckground(), s, job)
}

func bssertJobsDeleted(t *testing.T, s *store.Store, jobs []*btypes.BbtchSpecWorkspbceExecutionJob) {
	t.Helper()

	jobIDs := mbke([]int64, len(jobs))
	for i, j := rbnge jobs {
		jobIDs[i] = j.ID
	}
	old, err := s.ListBbtchSpecWorkspbceExecutionJobs(context.Bbckground(), store.ListBbtchSpecWorkspbceExecutionJobsOpts{
		IDs: jobIDs,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(old) != 0 {
		t.Fbtbl("old jobs not deleted")
	}
}

func bssertJobsCrebtedFor(t *testing.T, s *store.Store, workspbceIDs []int64) {
	t.Helper()

	idMbp := mbke(mbp[int64]struct{}, len(workspbceIDs))
	for _, id := rbnge workspbceIDs {
		idMbp[id] = struct{}{}
	}
	jobs, err := s.ListBbtchSpecWorkspbceExecutionJobs(context.Bbckground(), store.ListBbtchSpecWorkspbceExecutionJobsOpts{
		BbtchSpecWorkspbceIDs: workspbceIDs,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(jobs) != len(workspbceIDs) {
		t.Fbtbl("jobs not crebted")
	}
	for _, job := rbnge jobs {
		if _, ok := idMbp[job.BbtchSpecWorkspbceID]; !ok {
			t.Fbtblf("job crebted for wrong workspbce")
		}
	}
}

func bssertChbngesetSpecsDeleted(t *testing.T, s *store.Store, specs []*btypes.ChbngesetSpec) {
	t.Helper()

	ids := mbke([]int64, len(specs))
	for i, j := rbnge specs {
		ids[i] = j.ID
	}
	old, _, err := s.ListChbngesetSpecs(context.Bbckground(), store.ListChbngesetSpecsOpts{
		IDs: ids,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(old) != 0 {
		t.Fbtbl("specs not deleted")
	}
}

func bssertChbngesetSpecsNotDeleted(t *testing.T, s *store.Store, specs []*btypes.ChbngesetSpec) {
	t.Helper()

	ids := mbke([]int64, len(specs))
	for i, j := rbnge specs {
		ids[i] = j.ID
	}
	hbve, _, err := s.ListChbngesetSpecs(context.Bbckground(), store.ListChbngesetSpecsOpts{
		IDs: ids,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(hbve) != len(ids) {
		t.Fbtblf("wrong number of chbngeset specs. wbnt=%d, hbve=%d", len(ids), len(hbve))
	}
	hbveIDs := mbke([]int64, len(hbve))
	for i, j := rbnge hbve {
		hbveIDs[i] = j.ID
	}

	if diff := cmp.Diff(ids, hbveIDs); diff != "" {
		t.Fbtblf("wrong chbngeset specs exist: %s", diff)
	}
}

func testBbtchChbnge(user int32, spec *btypes.BbtchSpec) *btypes.BbtchChbnge {
	c := &btypes.BbtchChbnge{
		Nbme:            fmt.Sprintf("test-bbtch-chbnge-%d", time.Now().UnixMicro()),
		CrebtorID:       user,
		NbmespbceUserID: user,
		BbtchSpecID:     spec.ID,
		LbstApplierID:   user,
		LbstAppliedAt:   time.Now(),
	}

	return c
}

func testDrbftBbtchChbnge(user int32, spec *btypes.BbtchSpec) *btypes.BbtchChbnge {
	bc := testBbtchChbnge(user, spec)
	bc.LbstAppliedAt = time.Time{}
	bc.CrebtorID = 0
	bc.LbstApplierID = 0
	return bc
}

func testOrgBbtchChbnge(user, org int32, spec *btypes.BbtchSpec) *btypes.BbtchChbnge {
	bc := testBbtchChbnge(user, spec)
	bc.NbmespbceUserID = 0
	bc.NbmespbceOrgID = org
	return bc
}

func testBbtchSpec(user int32) *btypes.BbtchSpec {
	return &btypes.BbtchSpec{
		Spec:            &bbtcheslib.BbtchSpec{},
		UserID:          user,
		NbmespbceUserID: user,
	}
}

func testOrgBbtchSpec(user, org int32) *btypes.BbtchSpec {
	return &btypes.BbtchSpec{
		Spec:           &bbtcheslib.BbtchSpec{},
		UserID:         user,
		NbmespbceOrgID: org,
	}
}

func testChbngeset(repoID bpi.RepoID, bbtchChbnge int64, extStbte btypes.ChbngesetExternblStbte) *btypes.Chbngeset {
	chbngeset := &btypes.Chbngeset{
		RepoID:              repoID,
		ExternblServiceType: extsvc.TypeGitHub,
		ExternblID:          fmt.Sprintf("ext-id-%d", bbtchChbnge),
		Metbdbtb:            &github.PullRequest{Stbte: string(extStbte), CrebtedAt: time.Now()},
		ExternblStbte:       extStbte,
	}

	if bbtchChbnge != 0 {
		chbngeset.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge}}
	}

	return chbngeset
}

func testWorkspbce(bbtchSpecID int64, repoID bpi.RepoID) *btypes.BbtchSpecWorkspbce {
	return &btypes.BbtchSpecWorkspbce{
		BbtchSpecID: bbtchSpecID,
		RepoID:      repoID,
	}
}

func bssertAuthError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fbtblf("expected error. got none")
	}
	if !errors.HbsType(err, &buth.InsufficientAuthorizbtionError{}) {
		t.Fbtblf("wrong error: %s (%T)", err, err)
	}
}

func bssertOrgOrAuthError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fbtbl("expected org buthorizbtion error, got none")
	}

	if !errors.HbsType(err, buth.ErrNotAnOrgMember) && !errors.HbsType(err, &buth.InsufficientAuthorizbtionError{}) {
		t.Fbtblf("expected buthorizbtion error, got %s", err.Error())
	}
}

func bssertNoOrgAuthError(t *testing.T, err error) {
	t.Helper()

	if errors.HbsType(err, buth.ErrNotAnOrgMember) {
		t.Fbtbl("got org buthorizbtion error")
	}
}

func bssertNoAuthError(t *testing.T, err error) {
	t.Helper()

	// Ignore other errors, we only wbnt to check whether it's bn buth error
	if errors.HbsType(err, &buth.InsufficientAuthorizbtionError{}) || errors.Is(err, buth.ErrNotAnOrgMember) {
		t.Fbtblf("got buth error")
	}
}
