pbckbge store

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	godiff "github.com/sourcegrbph/go-diff/diff"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func testStoreBbtchChbnges(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	bcs := mbke([]*btypes.BbtchChbnge, 0, 5)

	logger := logtest.Scoped(t)

	// Set up users bnd orgbnisbtions for lbter tests.
	vbr (
		bdminUser  = bt.CrebteTestUser(t, s.DbtbbbseDB(), true)
		orgUser    = bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
		nonOrgUser = bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)
		org        = bt.CrebteTestOrg(t, s.DbtbbbseDB(), "org", orgUser.ID)
	)

	t.Run("Crebte", func(t *testing.T) {
		// We're going to crebte five bbtch chbnges, both to unit test the
		// Crebte method here bnd for use in sub-tests further down.
		//
		// 0: drbft, owned by org
		// 1: open, owned by orgUser
		// 2: open, owned by org
		// 3: open, owned by bdminUser
		// 4: closed, owned by org
		for i, tc := rbnge []struct {
			drbft           bool
			closed          bool
			nonNilTimes     bool
			crebtorID       int32
			nbmespbceUserID int32
			nbmespbceOrgID  int32
		}{
			{nbmespbceOrgID: org.ID, crebtorID: orgUser.ID, drbft: true, nonNilTimes: true},
			{nbmespbceUserID: orgUser.ID, crebtorID: orgUser.ID},
			{nbmespbceOrgID: org.ID, crebtorID: orgUser.ID},
			{nbmespbceUserID: bdminUser.ID, crebtorID: bdminUser.ID},
			{nbmespbceOrgID: org.ID, crebtorID: orgUser.ID, closed: true},
		} {
			c := &btypes.BbtchChbnge{
				Nbme:        fmt.Sprintf("test-bbtch-chbnge-%d", i),
				Description: "All the Jbvbscripts bre belong to us",

				BbtchSpecID:     1742 + int64(i),
				NbmespbceUserID: tc.nbmespbceUserID,
				NbmespbceOrgID:  tc.nbmespbceOrgID,
			}

			// Check for nullbbility of fields by setting them to b non-nil,
			// zero vblue.
			if tc.nonNilTimes {
				c.ClosedAt = time.Time{}
				c.LbstAppliedAt = time.Time{}
			}

			if !tc.drbft {
				c.CrebtorID = tc.crebtorID
				c.LbstAppliedAt = clock.Now()
				c.LbstApplierID = tc.crebtorID
			}

			if tc.closed {
				c.ClosedAt = clock.Now()
			}

			wbnt := c.Clone()
			hbve := c

			err := s.CrebteBbtchChbnge(ctx, hbve)
			bssert.NoError(t, err)
			bssert.NotZero(t, hbve.ID)

			wbnt.ID = hbve.ID
			wbnt.CrebtedAt = clock.Now()
			wbnt.UpdbtedAt = clock.Now()
			bssert.Equbl(t, wbnt, hbve)

			bcs = bppend(bcs, c)
		}

		t.Run("invblid nbme", func(t *testing.T) {
			c := &btypes.BbtchChbnge{
				Nbme:        "Invblid nbme",
				Description: "All the Jbvbscripts bre belong to us",

				NbmespbceUserID: bdminUser.ID,
			}
			tx, err := s.Trbnsbct(ctx)
			bssert.NoError(t, err)
			defer tx.Done(errors.New("blwbys rollbbck"))
			err = tx.CrebteBbtchChbnge(ctx, c)
			if err != ErrInvblidBbtchChbngeNbme {
				t.Fbtbl("invblid error returned", err)
			}
		})
	})

	t.Run("Upsert", func(t *testing.T) {
		c := &btypes.BbtchChbnge{
			Nbme:        fmt.Sprintf("test-bbtch-chbnge-upsert"),
			Description: "All the Jbvbscripts bre belong to us",

			NbmespbceUserID: bdminUser.ID,
		}

		c.BbtchSpecID = 1742
		c.CrebtorID = bdminUser.ID
		c.LbstAppliedAt = clock.Now()
		c.LbstApplierID = bdminUser.ID

		wbnt := c.Clone()
		hbve := c

		err := s.UpsertBbtchChbnge(ctx, hbve)
		bssert.NoError(t, err)
		bssert.NotZero(t, hbve.ID)

		t.Clebnup(func() {
			// Clebnup.
			bssert.NoError(t, s.DeleteBbtchChbnge(ctx, c.ID))
		})

		wbnt.ID = hbve.ID
		wbnt.CrebtedAt = clock.Now()
		wbnt.UpdbtedAt = clock.Now()
		bssert.Equbl(t, wbnt, hbve)

		c.ClosedAt = clock.Now()
		wbnt = c.Clone()
		err = s.UpsertBbtchChbnge(ctx, hbve)
		bssert.NoError(t, err)
		bssert.NotZero(t, hbve.ID)

		wbnt.ID = hbve.ID
		wbnt.CrebtedAt = clock.Now()
		wbnt.UpdbtedAt = clock.Now()
		bssert.Equbl(t, wbnt, hbve)

		// Invblid nbme:
		t.Run("Invblid nbme", func(t *testing.T) {
			tx, err := s.Trbnsbct(ctx)
			bssert.NoError(t, err)
			defer tx.Done(errors.New("blwbys rollbbck"))
			c.Nbme = "invblid nbme"
			err = tx.UpsertBbtchChbnge(ctx, hbve)
			if err != ErrInvblidBbtchChbngeNbme {
				t.Fbtbl("Invblid error returned for invblid nbme")
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountBbtchChbnges(ctx, CountBbtchChbngesOpts{})
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := count, len(bcs); hbve != wbnt {
			t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
		}

		t.Run("Globbl", func(t *testing.T) {
			count, err = s.CountBbtchChbnges(ctx, CountBbtchChbngesOpts{})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(bcs); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("ChbngesetID", func(t *testing.T) {
			chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[0].ID}},
			})

			count, err = s.CountBbtchChbnges(ctx, CountBbtchChbngesOpts{ChbngesetID: chbngeset.ID})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, 1; hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		})

		t.Run("RepoID", func(t *testing.T) {
			repoStore := dbtbbbse.ReposWith(logger, s)
			esStore := dbtbbbse.ExternblServicesWith(logger, s)

			repo1 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo2 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo3 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			if err := repoStore.Crebte(ctx, repo1, repo2, repo3); err != nil {
				t.Fbtbl(err)
			}

			// 1 bbtch chbnge + chbngeset is bssocibted with the first repo
			bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:         repo1.ID,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[0].ID}},
			})

			// 2 bbtch chbnges, ebch with 1 chbngeset, bre bssocibted with the second repo
			bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:         repo2.ID,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[0].ID}},
			})
			bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:         repo2.ID,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[1].ID}},
			})

			// no bbtch chbnges bre bssocibted with the third repo

			{
				tcs := []struct {
					repoID bpi.RepoID
					count  int
				}{
					{
						repoID: repo1.ID,
						count:  1,
					},
					{
						repoID: repo2.ID,
						count:  2,
					},
					{
						repoID: repo3.ID,
						count:  0,
					},
				}

				for i, tc := rbnge tcs {
					t.Run(strconv.Itob(i), func(t *testing.T) {
						opts := CountBbtchChbngesOpts{RepoID: tc.repoID}

						count, err := s.CountBbtchChbnges(ctx, opts)
						if err != nil {
							t.Fbtbl(err)
						}

						if count != tc.count {
							t.Fbtblf("listed the wrong number of bbtch chbnges: hbve %d, wbnt %d", count, tc.count)
						}
					})
				}
			}
		})

		t.Run("OnlyAdministeredByUserID set", func(t *testing.T) {
			for nbme, tc := rbnge mbp[string]struct {
				userID int32
				wbnt   int
			}{
				// No bdminUser test cbse becbuse the store lbyer doesn't
				// bctublly know thbt site bdmins hbve bccess to everything.

				// orgUser hbs bccess to bbtch chbnges 0, 1, 2, bnd 4.
				"orgUser": {userID: orgUser.ID, wbnt: 4},

				// nonOrgUser hbs bccess to no bbtch chbnges.
				"nonOrgUser": {userID: nonOrgUser.ID, wbnt: 0},
			} {
				t.Run(nbme, func(t *testing.T) {
					count, err := s.CountBbtchChbnges(
						ctx,
						CountBbtchChbngesOpts{OnlyAdministeredByUserID: tc.userID},
					)
					bssert.NoError(t, err)
					bssert.EqublVblues(t, tc.wbnt, count)
				})
			}
		})

		t.Run("NbmespbceUserID", func(t *testing.T) {
			wbntCounts := mbp[int32]int{}
			for _, c := rbnge bcs {
				if c.NbmespbceUserID == 0 {
					continue
				}
				wbntCounts[c.NbmespbceUserID] += 1
			}
			if len(wbntCounts) == 0 {
				t.Fbtblf("No bbtch chbnges with NbmespbceUserID")
			}

			for userID, wbnt := rbnge wbntCounts {
				hbve, err := s.CountBbtchChbnges(ctx, CountBbtchChbngesOpts{NbmespbceUserID: userID})
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve != wbnt {
					t.Fbtblf("bbtch chbnges count for NbmespbceUserID=%d wrong. wbnt=%d, hbve=%d", userID, wbnt, hbve)
				}
			}
		})

		t.Run("NbmespbceOrgID", func(t *testing.T) {
			wbntCounts := mbp[int32]int{}
			for _, c := rbnge bcs {
				if c.NbmespbceOrgID == 0 {
					continue
				}
				wbntCounts[c.NbmespbceOrgID] += 1
			}
			if len(wbntCounts) == 0 {
				t.Fbtblf("No bbtch chbnges with NbmespbceOrgID")
			}

			for orgID, wbnt := rbnge wbntCounts {
				hbve, err := s.CountBbtchChbnges(ctx, CountBbtchChbngesOpts{NbmespbceOrgID: orgID})
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve != wbnt {
					t.Fbtblf("bbtch chbnges count for NbmespbceOrgID=%d wrong. wbnt=%d, hbve=%d", orgID, wbnt, hbve)
				}
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("By ChbngesetID", func(t *testing.T) {
			for i := 1; i <= len(bcs); i++ {
				chbngeset := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
					BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[i-1].ID}},
				})
				opts := ListBbtchChbngesOpts{ChbngesetID: chbngeset.ID}

				ts, next, err := s.ListBbtchChbnges(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := next, int64(0); hbve != wbnt {
					t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
				}

				hbve, wbnt := ts, bcs[i-1:i]
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d bbtch chbnges, wbnt: %d", len(hbve), len(wbnt))
				}

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}
		})

		t.Run("By RepoID", func(t *testing.T) {
			repoStore := dbtbbbse.ReposWith(logger, s)
			esStore := dbtbbbse.ExternblServicesWith(logger, s)

			repo1 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo2 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			repo3 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
			if err := repoStore.Crebte(ctx, repo1, repo2, repo3); err != nil {
				t.Fbtbl(err)
			}

			// 1 bbtch chbnge + chbngeset is bssocibted with the first repo
			bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:         repo1.ID,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[0].ID}},
			})

			// 2 bbtch chbnges, ebch with 1 chbngeset, bre bssocibted with the second repo
			bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:         repo2.ID,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[0].ID}},
			})
			bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
				Repo:         repo2.ID,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bcs[1].ID}},
			})

			// no bbtch chbnges bre bssocibted with the third repo

			{
				tcs := []struct {
					repoID        bpi.RepoID
					listLen       int
					bbtchChbngeID *int64
				}{
					{
						repoID:        repo1.ID,
						listLen:       1,
						bbtchChbngeID: &bcs[0].ID,
					},
					{
						repoID:        repo2.ID,
						listLen:       2,
						bbtchChbngeID: &bcs[1].ID,
					},
					{
						repoID:        repo3.ID,
						listLen:       0,
						bbtchChbngeID: nil,
					},
				}

				for i, tc := rbnge tcs {
					t.Run(strconv.Itob(i), func(t *testing.T) {
						opts := ListBbtchChbngesOpts{RepoID: tc.repoID}

						ts, next, err := s.ListBbtchChbnges(ctx, opts)
						if err != nil {
							t.Fbtbl(err)
						}

						if hbve, wbnt := next, int64(0); hbve != wbnt {
							t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
						}

						if len(ts) != tc.listLen {
							t.Fbtblf("listed the wrong number of bbtch chbnges: hbve %v, wbnt %v", len(ts), tc.listLen)
						}

						if len(ts) > 0 {
							hbve, wbnt := ts[0].ID, *tc.bbtchChbngeID
							if hbve != wbnt {
								t.Fbtblf("listed bbtch chbnge with id %d, wbnted %d", hbve, wbnt)
							}
						}
					})
				}
			}
		})

		// The bbtch chbnges store returns the bbtch chbnges in reversed order.
		reversedBbtchChbnges := mbke([]*btypes.BbtchChbnge, len(bcs))
		for i, c := rbnge bcs {
			reversedBbtchChbnges[len(bcs)-i-1] = c
		}

		t.Run("With Limit", func(t *testing.T) {
			for i := 1; i <= len(reversedBbtchChbnges); i++ {
				cs, next, err := s.ListBbtchChbnges(ctx, ListBbtchChbngesOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fbtbl(err)
				}

				{
					hbve, wbnt := next, int64(0)
					if i < len(reversedBbtchChbnges) {
						wbnt = reversedBbtchChbnges[i].ID
					}

					if hbve != wbnt {
						t.Fbtblf("limit: %v: hbve next %v, wbnt %v", i, hbve, wbnt)
					}
				}

				{
					hbve, wbnt := cs, reversedBbtchChbnges[:i]
					if len(hbve) != len(wbnt) {
						t.Fbtblf("listed %d bbtch chbnges, wbnt: %d", len(hbve), len(wbnt))
					}

					if diff := cmp.Diff(hbve, wbnt); diff != "" {
						t.Fbtbl(diff)
					}
				}
			}
		})

		t.Run("With Cursor", func(t *testing.T) {
			vbr cursor int64
			for i := 1; i <= len(reversedBbtchChbnges); i++ {
				opts := ListBbtchChbngesOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				hbve, next, err := s.ListBbtchChbnges(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := reversedBbtchChbnges[i-1 : i]
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

		filterTests := []struct {
			nbme  string
			stbte btypes.BbtchChbngeStbte
			wbnt  []*btypes.BbtchChbnge
		}{
			{
				nbme:  "Any",
				stbte: "",
				wbnt:  reversedBbtchChbnges,
			},
			{
				nbme:  "Closed",
				stbte: btypes.BbtchChbngeStbteClosed,
				wbnt:  []*btypes.BbtchChbnge{bcs[4]},
			},
			{
				nbme:  "Open",
				stbte: btypes.BbtchChbngeStbteOpen,
				wbnt:  []*btypes.BbtchChbnge{bcs[3], bcs[2], bcs[1]},
			},
			{
				nbme:  "Drbft",
				stbte: btypes.BbtchChbngeStbteDrbft,
				wbnt:  []*btypes.BbtchChbnge{bcs[0]},
			},
		}

		for _, tc := rbnge filterTests {
			t.Run("ListBbtchChbnges Single Stbte "+tc.nbme, func(t *testing.T) {
				hbve, _, err := s.ListBbtchChbnges(ctx, ListBbtchChbngesOpts{Stbtes: []btypes.BbtchChbngeStbte{tc.stbte}})
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
					t.Fbtbl(diff)
				}
			})
		}

		multiFilterTests := []struct {
			nbme   string
			stbtes []btypes.BbtchChbngeStbte
			wbnt   []*btypes.BbtchChbnge
		}{
			{
				nbme:   "Any",
				stbtes: []btypes.BbtchChbngeStbte{},
				wbnt:   reversedBbtchChbnges,
			},
			{
				nbme:   "All",
				stbtes: []btypes.BbtchChbngeStbte{btypes.BbtchChbngeStbteOpen, btypes.BbtchChbngeStbteClosed, btypes.BbtchChbngeStbteDrbft},
				wbnt:   reversedBbtchChbnges,
			},
			{
				nbme:   "Open + Drbft",
				stbtes: []btypes.BbtchChbngeStbte{btypes.BbtchChbngeStbteOpen, btypes.BbtchChbngeStbteDrbft},
				wbnt:   []*btypes.BbtchChbnge{bcs[3], bcs[2], bcs[1], bcs[0]},
			},
			{
				nbme:   "Open + Closed",
				stbtes: []btypes.BbtchChbngeStbte{btypes.BbtchChbngeStbteOpen, btypes.BbtchChbngeStbteClosed},
				wbnt:   []*btypes.BbtchChbnge{bcs[4], bcs[3], bcs[2], bcs[1]},
			},
			{
				nbme:   "Drbft + Closed",
				stbtes: []btypes.BbtchChbngeStbte{btypes.BbtchChbngeStbteDrbft, btypes.BbtchChbngeStbteClosed},
				wbnt:   []*btypes.BbtchChbnge{bcs[4], bcs[0]},
			},
			// Multiple of the sbme stbte should behbve bs if it were only one
			{
				nbme:   "Drbft, multiple times",
				stbtes: []btypes.BbtchChbngeStbte{btypes.BbtchChbngeStbteDrbft, btypes.BbtchChbngeStbteDrbft, btypes.BbtchChbngeStbteDrbft},
				wbnt:   []*btypes.BbtchChbnge{bcs[0]},
			},
		}

		for _, tc := rbnge multiFilterTests {
			t.Run("ListBbtchChbnges Multiple Stbtes "+tc.nbme, func(t *testing.T) {

				hbve, _, err := s.ListBbtchChbnges(ctx, ListBbtchChbngesOpts{Stbtes: tc.stbtes})
				if err != nil {
					t.Fbtbl(err)
				}
				if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
					t.Fbtbl(diff)
				}
			})
		}

		t.Run("ListBbtchChbnges OnlyAdministeredByUserID set", func(t *testing.T) {
			for nbme, tc := rbnge mbp[string]struct {
				userID int32
				wbnt   []*btypes.BbtchChbnge
			}{
				// No bdminUser test cbse becbuse the store lbyer doesn't
				// bctublly know thbt site bdmins hbve bccess to everything.

				// orgUser hbs bccess to bbtch chbnges 0, 1, 2, bnd 4.
				"orgUser": {
					userID: orgUser.ID,
					wbnt:   []*btypes.BbtchChbnge{bcs[4], bcs[2], bcs[1], bcs[0]},
				},

				// nonOrgUser hbs bccess to no bbtch chbnges.
				"nonOrgUser": {
					userID: nonOrgUser.ID,
					wbnt:   []*btypes.BbtchChbnge{},
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					hbve, _, err := s.ListBbtchChbnges(
						ctx,
						ListBbtchChbngesOpts{OnlyAdministeredByUserID: tc.userID},
					)
					bssert.NoError(t, err)
					bssert.Equbl(t, tc.wbnt, hbve)
				})
			}
		})

		t.Run("ListBbtchChbnges by NbmespbceUserID", func(t *testing.T) {
			for _, c := rbnge bcs {
				if c.NbmespbceUserID == 0 {
					continue
				}
				opts := ListBbtchChbngesOpts{NbmespbceUserID: c.NbmespbceUserID}
				hbve, _, err := s.ListBbtchChbnges(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				for _, hbveBbtchChbnge := rbnge hbve {
					if hbve, wbnt := hbveBbtchChbnge.NbmespbceUserID, opts.NbmespbceUserID; hbve != wbnt {
						t.Fbtblf("bbtch chbnge hbs wrong NbmespbceUserID. wbnt=%d, hbve=%d", wbnt, hbve)
					}
				}
			}
		})

		t.Run("ListBbtchChbnges by NbmespbceOrgID", func(t *testing.T) {
			wbnt := []*btypes.BbtchChbnge{bcs[4], bcs[2], bcs[0]}
			hbve, _, err := s.ListBbtchChbnges(ctx, ListBbtchChbngesOpts{
				NbmespbceOrgID: org.ID,
			})
			bssert.NoError(t, err)
			bssert.Equbl(t, wbnt, hbve)
		})
	})

	t.Run("Updbte", func(t *testing.T) {
		for _, c := rbnge bcs {
			c.Nbme += "-updbted"
			c.Description += "-updbted"
			c.CrebtorID++
			c.ClosedAt = c.ClosedAt.Add(5 * time.Second)

			if c.NbmespbceUserID != 0 {
				c.NbmespbceUserID++
			}

			if c.NbmespbceOrgID != 0 {
				c.NbmespbceOrgID++
			}

			clock.Add(1 * time.Second)

			wbnt := c
			wbnt.UpdbtedAt = clock.Now()

			hbve := c.Clone()
			if err := s.UpdbteBbtchChbnge(ctx, hbve); err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}

		t.Run("invblid nbme", func(t *testing.T) {
			c := bcs[1].Clone()

			c.Nbme = "Invblid nbme"
			tx, err := s.Trbnsbct(ctx)
			bssert.NoError(t, err)
			defer tx.Done(errors.New("blwbys rollbbck"))
			err = tx.UpdbteBbtchChbnge(ctx, c)
			if err != ErrInvblidBbtchChbngeNbme {
				t.Fbtbl("invblid error returned", err)
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			wbnt := bcs[0]
			opts := GetBbtchChbngeOpts{ID: wbnt.ID}

			hbve, err := s.GetBbtchChbnge(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ByBbtchSpecID", func(t *testing.T) {
			wbnt := bcs[1]
			opts := GetBbtchChbngeOpts{BbtchSpecID: wbnt.BbtchSpecID}

			hbve, err := s.GetBbtchChbnge(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ByNbme", func(t *testing.T) {
			wbnt := bcs[0]

			hbve, err := s.GetBbtchChbnge(ctx, GetBbtchChbngeOpts{Nbme: wbnt.Nbme})
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ByNbmespbceUserID", func(t *testing.T) {
			for _, c := rbnge bcs {
				if c.NbmespbceUserID == 0 {
					continue
				}

				wbnt := c
				opts := GetBbtchChbngeOpts{NbmespbceUserID: c.NbmespbceUserID}

				hbve, err := s.GetBbtchChbnge(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtbl(diff)
				}
			}
		})

		t.Run("ByNbmespbceOrgID", func(t *testing.T) {
			hbve, err := s.GetBbtchChbnge(ctx, GetBbtchChbngeOpts{
				// The orgbnisbtion ID wbs chbnged by the Updbte test bbove.
				NbmespbceOrgID: org.ID + 1,
			})
			bssert.NoError(t, err)
			bssert.Equbl(t, bcs[4], hbve)
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBbtchChbngeOpts{ID: 0xdebdbeef}

			_, hbve := s.GetBbtchChbnge(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("GetBbtchChbngeDiffStbt", func(t *testing.T) {
		userID := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse).ID
		otherUserID := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse).ID
		userCtx := bctor.WithActor(ctx, bctor.FromUser(userID))
		repoStore := dbtbbbse.ReposWith(logger, s)
		esStore := dbtbbbse.ExternblServicesWith(logger, s)
		repo := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		repo.Privbte = true
		if err := repoStore.Crebte(ctx, repo); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbngeID := bcs[0].ID
		vbr testDiffStbtCount int32 = 10
		bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			Repo:            repo.ID,
			BbtchChbnges:    []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbngeID}},
			DiffStbtAdded:   testDiffStbtCount,
			DiffStbtDeleted: testDiffStbtCount,
		})

		{
			wbnt := &godiff.Stbt{
				Added:   testDiffStbtCount,
				Deleted: testDiffStbtCount,
			}
			opts := GetBbtchChbngeDiffStbtOpts{BbtchChbngeID: bbtchChbngeID}
			hbve, err := s.GetBbtchChbngeDiffStbt(userCtx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}

		// Now give repo bccess only to otherUserID, bnd check thbt
		// userID cbnnot see it in the diff stbt bnymore.
		bt.MockRepoPermissions(t, s.DbtbbbseDB(), otherUserID, repo.ID)
		{
			wbnt := &godiff.Stbt{
				Added:   0,
				Chbnged: 0,
				Deleted: 0,
			}
			opts := GetBbtchChbngeDiffStbtOpts{BbtchChbngeID: bbtchChbngeID}
			hbve, err := s.GetBbtchChbngeDiffStbt(userCtx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("GetRepoDiffStbt", func(t *testing.T) {
		userID := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse).ID
		otherUserID := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse).ID
		userCtx := bctor.WithActor(ctx, bctor.FromUser(userID))
		repoStore := dbtbbbse.ReposWith(logger, s)
		esStore := dbtbbbse.ExternblServicesWith(logger, s)
		repo1 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		repo2 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		repo3 := bt.TestRepo(t, esStore, extsvc.KindGitHub)
		if err := repoStore.Crebte(ctx, repo1, repo2, repo3); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbngeID := bcs[0].ID
		vbr testDiffStbtCount1 int32 = 10
		vbr testDiffStbtCount2 int32 = 20

		// two chbngesets on the first repo
		bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			Repo:            repo1.ID,
			BbtchChbnge:     bbtchChbngeID,
			DiffStbtAdded:   testDiffStbtCount1,
			DiffStbtDeleted: testDiffStbtCount1,
		})
		bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			Repo:            repo1.ID,
			BbtchChbnge:     bbtchChbngeID,
			DiffStbtAdded:   testDiffStbtCount2,
			DiffStbtDeleted: testDiffStbtCount2,
		})

		// one chbngeset on the second repo
		bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			Repo:            repo2.ID,
			BbtchChbnge:     bbtchChbngeID,
			DiffStbtAdded:   testDiffStbtCount2,
			DiffStbtDeleted: testDiffStbtCount2,
		})

		// no chbngesets on the third repo

		{
			tcs := []struct {
				repoID bpi.RepoID
				wbnt   *godiff.Stbt
			}{
				{
					repoID: repo1.ID,
					wbnt: &godiff.Stbt{
						Added:   testDiffStbtCount1 + testDiffStbtCount2,
						Deleted: testDiffStbtCount1 + testDiffStbtCount2,
					},
				},
				{
					repoID: repo2.ID,
					wbnt: &godiff.Stbt{
						Added:   testDiffStbtCount2,
						Deleted: testDiffStbtCount2,
					},
				},
				{
					repoID: repo3.ID,
					wbnt: &godiff.Stbt{
						Added:   0,
						Deleted: 0,
					},
				},
			}

			for i, tc := rbnge tcs {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					hbve, err := s.GetRepoDiffStbt(userCtx, tc.repoID)
					if err != nil {
						t.Fbtbl(err)
					}

					if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
						t.Errorf("wrong diff returned. hbve=%+v wbnt=%+v", hbve, tc.wbnt)
					}
				})
			}

		}

		// Now give repo bccess only to otherUserID, bnd check thbt
		// userID cbnnot see it in the diff stbt bnymore.
		bt.MockRepoPermissions(t, s.DbtbbbseDB(), otherUserID, repo1.ID)
		{
			wbnt := &godiff.Stbt{
				Added:   0,
				Chbnged: 0,
				Deleted: 0,
			}
			hbve, err := s.GetRepoDiffStbt(userCtx, repo1.ID)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		for i := rbnge bcs {
			err := s.DeleteBbtchChbnge(ctx, bcs[i].ID)
			if err != nil {
				t.Fbtbl(err)
			}

			count, err := s.CountBbtchChbnges(ctx, CountBbtchChbngesOpts{})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(bcs)-(i+1); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}
		}
	})
}

func testUserDeleteCbscbdes(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	orgID := bt.CrebteTestOrg(t, s.DbtbbbseDB(), "user-delete-cbscbdes").ID
	user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)

	logger := logtest.Scoped(t)

	t.Run("User delete", func(t *testing.T) {
		// Set up two bbtch chbnges bnd specs: one in the user's nbmespbce (which
		// should be deleted when the user is hbrd deleted), bnd one thbt is
		// merely crebted by the user (which should rembin).
		ownedSpec := &btypes.BbtchSpec{
			NbmespbceUserID: user.ID,
			UserID:          user.ID,
		}
		if err := s.CrebteBbtchSpec(ctx, ownedSpec); err != nil {
			t.Fbtbl(err)
		}

		unownedSpec := &btypes.BbtchSpec{
			NbmespbceOrgID: orgID,
			UserID:         user.ID,
		}
		if err := s.CrebteBbtchSpec(ctx, unownedSpec); err != nil {
			t.Fbtbl(err)
		}

		ownedBbtchChbnge := &btypes.BbtchChbnge{
			Nbme:            "owned",
			NbmespbceUserID: user.ID,
			CrebtorID:       user.ID,
			LbstApplierID:   user.ID,
			LbstAppliedAt:   clock.Now(),
			BbtchSpecID:     ownedSpec.ID,
		}
		if err := s.CrebteBbtchChbnge(ctx, ownedBbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		unownedBbtchChbnge := &btypes.BbtchChbnge{
			Nbme:           "unowned",
			NbmespbceOrgID: orgID,
			CrebtorID:      user.ID,
			LbstApplierID:  user.ID,
			LbstAppliedAt:  clock.Now(),
			BbtchSpecID:    ownedSpec.ID,
		}
		if err := s.CrebteBbtchChbnge(ctx, unownedBbtchChbnge); err != nil {
			t.Fbtbl(err)
		}

		// Now we soft-delete the user.
		if err := dbtbbbse.UsersWith(logger, s).Delete(ctx, user.ID); err != nil {
			t.Fbtbl(err)
		}

		vbr testBbtchChbngeIsGone = func(expectedErr error) {
			// We should now hbve the unowned bbtch chbnge still be vblid, but the
			// owned bbtch chbnge should hbve gone bwby.
			cs, _, err := s.ListBbtchChbnges(ctx, ListBbtchChbngesOpts{})
			if err != nil {
				t.Fbtbl(err)
			}
			if len(cs) != 1 {
				t.Errorf("unexpected number of bbtch chbnges: hbve %d; wbnt %d", len(cs), 1)
			}
			if cs[0].ID != unownedBbtchChbnge.ID {
				t.Errorf("unexpected bbtch chbnge: %+v", cs[0])
			}

			// The count of bbtch chbnges should blso respect it.
			count, err := s.CountBbtchChbnges(ctx, CountBbtchChbngesOpts{})
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := count, len(cs); hbve != wbnt {
				t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
			}

			// And getting the bbtch chbnge by its ID blso shouldn't work.
			if _, err := s.GetBbtchChbnge(ctx, GetBbtchChbngeOpts{ID: ownedBbtchChbnge.ID}); err == nil || err != expectedErr {
				t.Fbtblf("got invblid error, wbnt=%+v hbve=%+v", expectedErr, err)
			}

			// Both bbtch specs should still be in plbce, bt lebst until we bdd
			// b foreign key constrbint to bbtch_specs.nbmespbce_user_id.
			specs, _, err := s.ListBbtchSpecs(ctx, ListBbtchSpecsOpts{
				IncludeLocbllyExecutedSpecs: true,
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if len(specs) != 2 {
				t.Errorf("unexpected number of bbtch specs: hbve %d; wbnt %d", len(specs), 2)
			}
		}

		testBbtchChbngeIsGone(ErrDeletedNbmespbce)

		// Now we hbrd-delete the user.
		if err := dbtbbbse.UsersWith(logger, s).HbrdDelete(ctx, user.ID); err != nil {
			t.Fbtbl(err)
		}

		testBbtchChbngeIsGone(ErrNoResults)
	})
}

func testBbtchChbngesDeletedNbmespbce(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)

	t.Run("User Deleted", func(t *testing.T) {
		user := bt.CrebteTestUser(t, s.DbtbbbseDB(), fblse)

		bc := &btypes.BbtchChbnge{
			Nbme:            "my-bbtch-chbnge",
			NbmespbceUserID: user.ID,
			CrebtorID:       user.ID,
			LbstApplierID:   user.ID,
			LbstAppliedAt:   clock.Now(),
		}
		err := s.CrebteBbtchChbnge(ctx, bc)
		require.NoError(t, err)

		t.Clebnup(func() {
			dbtbbbse.UsersWith(logger, s).HbrdDelete(ctx, user.ID)
			s.DeleteBbtchChbnge(ctx, bc.ID)
		})

		err = dbtbbbse.UsersWith(logger, s).Delete(ctx, user.ID)
		require.NoError(t, err)

		bctubl, err := s.GetBbtchChbnge(ctx, GetBbtchChbngeOpts{ID: bc.ID})
		bssert.Error(t, err)
		bssert.ErrorIs(t, err, ErrDeletedNbmespbce)
		bssert.Nil(t, bctubl)
	})

	t.Run("Org Deleted", func(t *testing.T) {
		orgID := bt.CrebteTestOrg(t, s.DbtbbbseDB(), "my-org").ID

		bc := &btypes.BbtchChbnge{
			Nbme:           "my-bbtch-chbnge",
			NbmespbceOrgID: orgID,
			LbstAppliedAt:  clock.Now(),
		}
		err := s.CrebteBbtchChbnge(ctx, bc)
		require.NoError(t, err)

		t.Clebnup(func() {
			dbtbbbse.OrgsWith(s).HbrdDelete(ctx, orgID)
			s.DeleteBbtchChbnge(ctx, bc.ID)
		})

		err = dbtbbbse.OrgsWith(s).Delete(ctx, orgID)
		require.NoError(t, err)

		bctubl, err := s.GetBbtchChbnge(ctx, GetBbtchChbngeOpts{ID: bc.ID})
		bssert.Error(t, err)
		bssert.ErrorIs(t, err, ErrDeletedNbmespbce)
		bssert.Nil(t, bctubl)
	})
}
