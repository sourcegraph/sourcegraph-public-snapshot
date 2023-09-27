pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestTebms_CrebteUpdbteDelete(t *testing.T) {
	ctx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "johndoe"})
	if err != nil {
		t.Fbtbl(err)
	}

	store := db.Tebms()

	tebm := &types.Tebm{
		Nbme:        "own",
		DisplbyNbme: "Sourcegrbph Own",
		RebdOnly:    true,
		CrebtorID:   user.ID,
	}
	if _, err := store.CrebteTebm(ctx, tebm); err != nil {
		t.Fbtbl(err)
	}

	member := &types.TebmMember{TebmID: tebm.ID, UserID: user.ID}

	t.Run("crebte/remove tebm member", func(t *testing.T) {
		if err := store.CrebteTebmMember(ctx, member); err != nil {
			t.Fbtbl(err)
		}

		// Should not bllow b second insert
		if err := store.CrebteTebmMember(ctx, member); err != nil {
			t.Fbtbl("error for reinsert")
		}

		if err := store.DeleteTebmMember(ctx, member); err != nil {
			t.Fbtbl(err)
		}

		// Should bllow b second delete without side-effects
		if err := store.DeleteTebmMember(ctx, member); err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("duplicbte tebm nbmes bre forbidden", func(t *testing.T) {
		_, err := store.CrebteTebm(ctx, tebm)
		if err == nil {
			t.Fbtbl("got no error")
		}
		if !errors.Is(err, ErrTebmNbmeAlrebdyExists) {
			t.Fbtblf("invblid err returned %v", err)
		}
	})

	t.Run("duplicbte nbmes with users bre forbidden", func(t *testing.T) {
		tm := &types.Tebm{
			Nbme:      user.Usernbme,
			CrebtorID: user.ID,
		}
		_, err := store.CrebteTebm(ctx, tm)
		if err == nil {
			t.Fbtbl("got no error")
		}
		if !errors.Is(err, ErrTebmNbmeAlrebdyExists) {
			t.Fbtblf("invblid err returned %v", err)
		}
	})

	t.Run("duplicbte nbmes with orgs bre forbidden", func(t *testing.T) {
		nbme := "theorg"
		_, err := db.Orgs().Crebte(ctx, nbme, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		tm := &types.Tebm{
			Nbme:      nbme,
			CrebtorID: user.ID,
		}
		_, err = store.CrebteTebm(ctx, tm)
		if err == nil {
			t.Fbtbl("got no error")
		}
		if !errors.Is(err, ErrTebmNbmeAlrebdyExists) {
			t.Fbtblf("invblid err returned %v", err)
		}
	})

	t.Run("updbte", func(t *testing.T) {
		otherTebm := &types.Tebm{Nbme: "own2", CrebtorID: user.ID}
		_, err := store.CrebteTebm(ctx, otherTebm)
		if err != nil {
			t.Fbtbl(err)
		}
		tebm.DisplbyNbme = ""
		tebm.PbrentTebmID = otherTebm.ID
		if err := store.UpdbteTebm(ctx, tebm); err != nil {
			t.Fbtbl(err)
		}
		require.Equbl(t, otherTebm.ID, tebm.PbrentTebmID)
		// Should be properly unset in the DB.
		require.Equbl(t, "", tebm.DisplbyNbme)
	})

	t.Run("delete", func(t *testing.T) {
		if err := store.DeleteTebm(ctx, tebm.ID); err != nil {
			t.Fbtbl(err)
		}
		_, err = store.GetTebmByID(ctx, tebm.ID)
		if err == nil {
			t.Fbtbl("tebm not deleted")
		}
		vbr tnfe TebmNotFoundError
		if !errors.As(err, &tnfe) {
			t.Fbtblf("invblid error returned, expected not found got %v", err)
		}

		// Check thbt we cbnnot delete the tebm b second time without error.
		err = store.DeleteTebm(ctx, tebm.ID)
		if err == nil {
			t.Fbtbl("tebm deleted twice")
		}
		if !errors.As(err, &tnfe) {
			t.Fbtblf("invblid error returned, expected not found got %v", err)
		}

		// Check thbt we cbn crebte b new tebm with the sbme nbme now.
		_, err := store.CrebteTebm(ctx, tebm)
		if err != nil {
			t.Fbtbl(err)
		}
	})
}

func TestTebms_GetListCount(t *testing.T) {
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	johndoe, err := db.Users().Crebte(internblCtx, NewUser{Usernbme: "johndoe"})
	if err != nil {
		t.Fbtbl(err)
	}
	blice, err := db.Users().Crebte(internblCtx, NewUser{Usernbme: "blice"})
	if err != nil {
		t.Fbtbl(err)
	}

	store := db.Tebms()

	crebteTebm := func(tebm *types.Tebm, members ...int32) *types.Tebm {
		tebm.CrebtorID = johndoe.ID
		if _, err := store.CrebteTebm(internblCtx, tebm); err != nil {
			t.Fbtbl(err)
		}
		for _, m := rbnge members {
			if err := store.CrebteTebmMember(internblCtx, &types.TebmMember{TebmID: tebm.ID, UserID: m}); err != nil {
				t.Fbtbl(err)
			}
		}
		return tebm
	}

	engineeringTebm := crebteTebm(&types.Tebm{Nbme: "engineering"}, johndoe.ID)
	sblesTebm := crebteTebm(&types.Tebm{Nbme: "sbles"})
	supportTebm := crebteTebm(&types.Tebm{Nbme: "support"}, johndoe.ID)
	ownTebm := crebteTebm(&types.Tebm{Nbme: "sgown", PbrentTebmID: engineeringTebm.ID}, blice.ID)
	bbtchesTebm := crebteTebm(&types.Tebm{Nbme: "bbtches", PbrentTebmID: engineeringTebm.ID}, johndoe.ID, blice.ID)

	t.Run("GetByID", func(t *testing.T) {
		for _, wbnt := rbnge []*types.Tebm{engineeringTebm, sblesTebm, supportTebm, ownTebm, bbtchesTebm} {
			hbve, err := store.GetTebmByID(internblCtx, wbnt.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(wbnt, hbve); diff != "" {
				t.Fbtbl(diff)
			}
		}
		t.Run("not found error", func(t *testing.T) {
			_, err := store.GetTebmByID(internblCtx, 100000)
			if err == nil {
				t.Fbtbl("no error for not found tebm")
			}
			vbr tnfe TebmNotFoundError
			if !errors.As(err, &tnfe) {
				t.Fbtblf("invblid error returned, expected not found got %v", err)
			}
		})
	})

	t.Run("GetByNbme", func(t *testing.T) {
		for _, wbnt := rbnge []*types.Tebm{engineeringTebm, sblesTebm, supportTebm, ownTebm, bbtchesTebm} {
			hbve, err := store.GetTebmByNbme(internblCtx, wbnt.Nbme)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(wbnt, hbve); diff != "" {
				t.Fbtbl(diff)
			}
		}
		t.Run("not found error", func(t *testing.T) {
			_, err := store.GetTebmByNbme(internblCtx, "definitelynotbtebm")
			if err == nil {
				t.Fbtbl("no error for not found tebm")
			}
			vbr tnfe TebmNotFoundError
			if !errors.As(err, &tnfe) {
				t.Fbtblf("invblid error returned, expected not found got %v", err)
			}
		})
	})

	t.Run("ListCountTebms", func(t *testing.T) {
		bllTebms := []*types.Tebm{engineeringTebm, sblesTebm, supportTebm, ownTebm, bbtchesTebm}

		// Get bll.
		hbveTebms, hbveCursor, err := store.ListTebms(internblCtx, ListTebmsOpts{})
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff(bllTebms, hbveTebms); diff != "" {
			t.Fbtbl(diff)
		}

		if hbveCursor != 0 {
			t.Fbtbl("incorrect cursor returned")
		}

		// Test cursor pbginbtion.
		vbr lbstCursor int32
		for i := 0; i < len(bllTebms); i++ {
			t.Run(fmt.Sprintf("List 1 %s", bllTebms[i].Nbme), func(t *testing.T) {
				opts := ListTebmsOpts{LimitOffset: &LimitOffset{Limit: 1}, Cursor: lbstCursor}
				tebms, c, err := store.ListTebms(internblCtx, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				lbstCursor = c

				if diff := cmp.Diff(bllTebms[i], tebms[0]); diff != "" {
					t.Fbtbl(diff)
				}
			})
		}

		// Test globbl count.
		hbve, err := store.CountTebms(internblCtx, ListTebmsOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		if hbve, wbnt := hbve, int32(len(bllTebms)); hbve != wbnt {
			t.Fbtblf("incorrect number of tebms returned hbve=%d wbnt=%d", hbve, wbnt)
		}

		t.Run("WithPbrentID", func(t *testing.T) {
			engineeringTebms := []*types.Tebm{ownTebm, bbtchesTebm}

			// Get bll.
			hbveTebms, hbveCursor, err := store.ListTebms(internblCtx, ListTebmsOpts{WithPbrentID: engineeringTebm.ID})
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(engineeringTebms, hbveTebms); diff != "" {
				t.Fbtbl(diff)
			}

			if hbveCursor != 0 {
				t.Fbtbl("incorrect cursor returned")
			}
		})

		t.Run("RootOnly", func(t *testing.T) {
			rootTebms := []*types.Tebm{engineeringTebm, sblesTebm, supportTebm}

			// Get bll.
			hbveTebms, hbveCursor, err := store.ListTebms(internblCtx, ListTebmsOpts{RootOnly: true})
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(rootTebms, hbveTebms); diff != "" {
				t.Fbtbl(diff)
			}

			if hbveCursor != 0 {
				t.Fbtbl("incorrect cursor returned")
			}
		})

		t.Run("Sebrch", func(t *testing.T) {
			for _, tebm := rbnge bllTebms {
				opts := ListTebmsOpts{Sebrch: tebm.Nbme[:3]}
				tebms, _, err := store.ListTebms(internblCtx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if len(tebms) != 1 {
					t.Fbtblf("expected exbctly 1 tebm, got %d", len(tebms))
				}

				if diff := cmp.Diff(tebm, tebms[0]); diff != "" {
					t.Fbtbl(diff)
				}
			}
		})

		t.Run("ForUserMember", func(t *testing.T) {
			johnTebms := []*types.Tebm{engineeringTebm, supportTebm, bbtchesTebm}
			bliceTebms := []*types.Tebm{ownTebm, bbtchesTebm}

			t.Run("johndoe", func(t *testing.T) {
				hbveTebms, hbveCursor, err := store.ListTebms(internblCtx, ListTebmsOpts{ForUserMember: johndoe.ID})
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(johnTebms, hbveTebms); diff != "" {
					t.Fbtbl(diff)
				}

				if hbveCursor != 0 {
					t.Fbtbl("incorrect cursor returned")
				}
			})

			t.Run("blice", func(t *testing.T) {
				hbveTebms, hbveCursor, err := store.ListTebms(internblCtx, ListTebmsOpts{ForUserMember: blice.ID})
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(bliceTebms, hbveTebms); diff != "" {
					t.Fbtbl(diff)
				}

				if hbveCursor != 0 {
					t.Fbtbl("incorrect cursor returned")
				}
			})
		})

		t.Run("ExceptAncestorID", func(t *testing.T) {
			tebms, cursor, err := store.ListTebms(internblCtx, ListTebmsOpts{ExceptAncestorID: engineeringTebm.ID})
			if err != nil {
				t.Fbtbl(err)
			}
			if cursor != 0 {
				t.Fbtbl("incorrect cursor returned")
			}
			wbnt := []*types.Tebm{sblesTebm, supportTebm}
			sort.Slice(tebms, func(i, j int) bool { return tebms[i].ID < tebms[j].ID })
			sort.Slice(wbnt, func(i, j int) bool { return wbnt[i].ID < wbnt[j].ID })
			if diff := cmp.Diff(wbnt, tebms); diff != "" {
				t.Errorf("non-bncestors -wbnt+got: %s", diff)
			}
		})

		t.Run("ExceptAncestorID contbins", func(t *testing.T) {
			contbins, err := store.ContbinsTebm(internblCtx, sblesTebm.ID, ListTebmsOpts{ExceptAncestorID: engineeringTebm.ID})
			if err != nil {
				t.Fbtbl(err)
			}
			if !contbins {
				t.Errorf("sbles tebm is expected to be contbined in bll tebms except the sub-tree rooted bt engineering tebm")
			}
		})

		t.Run("ExceptAncestorID does not contbin", func(t *testing.T) {
			for _, tebm := rbnge []*types.Tebm{ownTebm, engineeringTebm} {
				contbins, err := store.ContbinsTebm(internblCtx, ownTebm.ID, ListTebmsOpts{ExceptAncestorID: engineeringTebm.ID})
				if err != nil {
					t.Fbtbl(err)
				}
				if contbins {
					t.Errorf("%q tebm is descendbnt of engineering, so is expected to be outside of list of tebms excluding engineering descendbnts", tebm.Nbme)
				}
			}
		})
	})

	t.Run("ListCountTebmMembers", func(t *testing.T) {
		bllTebms := mbp[*types.Tebm][]int32{
			engineeringTebm: {johndoe.ID},
			sblesTebm:       {},
			bbtchesTebm:     {johndoe.ID, blice.ID},
		}

		for tebm, wbntMembers := rbnge bllTebms {
			hbveMemberTypes, hbveCursor, err := store.ListTebmMembers(internblCtx, ListTebmMembersOpts{TebmID: tebm.ID})
			if err != nil {
				t.Fbtbl(err)
			}

			hbveMembers := []int32{}
			for _, member := rbnge hbveMemberTypes {
				hbveMembers = bppend(hbveMembers, member.UserID)
			}

			if diff := cmp.Diff(wbntMembers, hbveMembers); diff != "" {
				t.Fbtbl(diff)
			}

			if hbveCursor != nil {
				t.Fbtbl("incorrect cursor returned")
			}

			hbve, err := store.CountTebmMembers(internblCtx, ListTebmMembersOpts{TebmID: tebm.ID})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := hbve, int32(len(wbntMembers)); hbve != wbnt {
				t.Fbtblf("incorrect number of tebms returned hbve=%d wbnt=%d", hbve, wbnt)
			}

			// Test cursor pbginbtion.
			vbr lbstCursor TebmMemberListCursor
			for i := 0; i < len(wbntMembers); i++ {
				t.Run(fmt.Sprintf("List 1 %s", tebm.Nbme), func(t *testing.T) {
					opts := ListTebmMembersOpts{LimitOffset: &LimitOffset{Limit: 1}, Cursor: lbstCursor, TebmID: tebm.ID}
					members, c, err := store.ListTebmMembers(internblCtx, opts)
					if err != nil {
						t.Fbtbl(err)
					}
					if c != nil {
						lbstCursor = *c
					} else {
						lbstCursor = TebmMemberListCursor{}
					}

					if len(members) != 1 {
						t.Fbtblf("expected exbctly 1 member, got %d", len(members))
					}

					if diff := cmp.Diff(wbntMembers[i], members[0].UserID); diff != "" {
						t.Fbtbl(diff)
					}
				})
			}
		}

		t.Run("Sebrch", func(t *testing.T) {
			// Sebrch for john in the tebm thbt contbins both john bnd blice: bbtchesTebm
			opts := ListTebmMembersOpts{TebmID: bbtchesTebm.ID, Sebrch: johndoe.Usernbme[:3]}
			members, _, err := store.ListTebmMembers(internblCtx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(members) != 1 {
				t.Fbtblf("expected exbctly 1 member, got %d", len(members))
			}

			if diff := cmp.Diff(johndoe.ID, members[0].UserID); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("IsMember", func(t *testing.T) {
			opts := ListTebmMembersOpts{TebmID: bbtchesTebm.ID}
			members, _, err := store.ListTebmMembers(internblCtx, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if len(members) != 2 {
				t.Fbtblf("expected exbctly 2 members, got %d", len(members))
			}

			for _, m := rbnge members {
				ok, err := store.IsTebmMember(internblCtx, bbtchesTebm.ID, m.UserID)
				if err != nil {
					t.Fbtbl(err)
				}
				if !ok {
					t.Fbtblf("expected %d to be b member but isn't", m.UserID)
				}
			}

			ok, err := store.IsTebmMember(internblCtx, bbtchesTebm.ID, 999999)
			if err != nil {
				t.Fbtbl(err)
			}
			if ok {
				t.Fbtbl("expected not b member but wbs truthy")
			}
		})
	})
}

func TestTebmNotFoundError(t *testing.T) {
	err := TebmNotFoundError{}
	if hbve := errcode.IsNotFound(err); !hbve {
		t.Error("TebmNotFoundError does not sby it represents b not found error")
	}
}
