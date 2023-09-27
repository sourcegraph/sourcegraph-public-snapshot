pbckbge store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
)

func testStoreChbngesetEvents(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	issueComment := &github.IssueComment{
		DbtbbbseID: 443827703,
		Author: github.Actor{
			AvbtbrURL: "https://bvbtbrs0.githubusercontent.com/u/1976?v=4",
			Login:     "sqs",
			URL:       "https://github.com/sqs",
		},
		Editor:              nil,
		AuthorAssocibtion:   "MEMBER",
		Body:                "> Just to be sure: you mebn the \"sebrchFilters\" \"Filters\" should be lowercbse, not the \"Sebrch Filters\" from the description, right?\r\n\r\nNo, the prose “Sebrch Filters” should hbve the F lowercbsed to fit with our style guide preference for sentence cbse over title cbse. (Cbn’t find this comment on the GitHub mobile interfbce bnymore so quoting the embil.)",
		URL:                 "https://github.com/sourcegrbph/sourcegrbph/pull/999#issuecomment-443827703",
		CrebtedAt:           clock.Now(),
		UpdbtedAt:           clock.Now(),
		IncludesCrebtedEdit: fblse,
	}

	events := mbke([]*btypes.ChbngesetEvent, 0, 3)
	kinds := []btypes.ChbngesetEventKind{
		btypes.ChbngesetEventKindGitHubCommented,
		btypes.ChbngesetEventKindGitHubClosed,
		btypes.ChbngesetEventKindGitHubAssigned,
	}

	t.Run("Upsert", func(t *testing.T) {
		for i := 0; i < cbp(events); i++ {
			e := &btypes.ChbngesetEvent{
				ChbngesetID: int64(i + 1),
				Kind:        kinds[i],
				Key:         issueComment.Key(),
				CrebtedAt:   clock.Now(),
				Metbdbtb:    issueComment,
			}

			events = bppend(events, e)
		}

		// Verify thbt no duplicbtes bre introduced bnd no error is returned.
		for i := 0; i < 2; i++ {
			err := s.UpsertChbngesetEvents(ctx, events...)
			if err != nil {
				t.Fbtbl(err)
			}
		}

		for _, hbve := rbnge events {
			if hbve.ID == 0 {
				t.Fbtbl("id should not be zero")
			}

			wbnt := hbve.Clone()

			wbnt.ID = hbve.ID
			wbnt.CrebtedAt = clock.Now()
			wbnt.UpdbtedAt = clock.Now()

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := s.CountChbngesetEvents(ctx, CountChbngesetEventsOpts{})
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := count, len(events); hbve != wbnt {
			t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
		}

		count, err = s.CountChbngesetEvents(ctx, CountChbngesetEventsOpts{ChbngesetID: 1})
		if err != nil {
			t.Fbtbl(err)
		}

		if hbve, wbnt := count, 1; hbve != wbnt {
			t.Fbtblf("hbve count: %d, wbnt: %d", hbve, wbnt)
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			wbnt := events[0]
			opts := GetChbngesetEventOpts{ID: wbnt.ID}

			hbve, err := s.GetChbngesetEvent(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ByKey", func(t *testing.T) {
			wbnt := events[0]
			opts := GetChbngesetEventOpts{
				ChbngesetID: wbnt.ChbngesetID,
				Kind:        wbnt.Kind,
				Key:         wbnt.Key,
			}

			hbve, err := s.GetChbngesetEvent(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetChbngesetEventOpts{ID: 0xdebdbeef}

			_, hbve := s.GetChbngesetEvent(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("ByChbngesetIDs", func(t *testing.T) {
			for i := 1; i <= len(events); i++ {
				opts := ListChbngesetEventsOpts{ChbngesetIDs: []int64{int64(i)}}

				ts, next, err := s.ListChbngesetEvents(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := next, int64(0); hbve != wbnt {
					t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
				}

				hbve, wbnt := ts, events[i-1:i]
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d events, wbnt: %d", len(hbve), len(wbnt))
				}

				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}
			}

			{
				opts := ListChbngesetEventsOpts{ChbngesetIDs: []int64{}}

				for i := 1; i <= len(events); i++ {
					opts.ChbngesetIDs = bppend(opts.ChbngesetIDs, int64(i))
				}

				ts, next, err := s.ListChbngesetEvents(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := next, int64(0); hbve != wbnt {
					t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
				}

				hbve, wbnt := ts, events
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d events, wbnt: %d", len(hbve), len(wbnt))
				}
			}
		})

		t.Run("ByKinds", func(t *testing.T) {
			for _, k := rbnge kinds {
				opts := ListChbngesetEventsOpts{Kinds: []btypes.ChbngesetEventKind{k}}

				ts, next, err := s.ListChbngesetEvents(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := next, int64(0); hbve != wbnt {
					t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
				}

				if hbve, wbnt := len(ts), 1; hbve != wbnt {
					t.Fbtblf("listed %d events for %q, wbnt: %d", hbve, k, wbnt)
				}

				if hbve, wbnt := ts[0].Kind, k; hbve != wbnt {
					t.Fbtblf("listed %q events, wbnt of kind: %q", hbve, wbnt)
				}
			}

			{
				opts := ListChbngesetEventsOpts{Kinds: []btypes.ChbngesetEventKind{}}

				for _, e := rbnge events {
					opts.Kinds = bppend(opts.Kinds, e.Kind)
				}

				ts, next, err := s.ListChbngesetEvents(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				if hbve, wbnt := next, int64(0); hbve != wbnt {
					t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
				}

				hbve, wbnt := ts, events
				if len(hbve) != len(wbnt) {
					t.Fbtblf("listed %d events, wbnt: %d", len(hbve), len(wbnt))
				}
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(events); i++ {
				cs, next, err := s.ListChbngesetEvents(ctx, ListChbngesetEventsOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fbtbl(err)
				}

				{
					hbve, wbnt := next, int64(0)
					if i < len(events) {
						wbnt = events[i].ID
					}

					if hbve != wbnt {
						t.Fbtblf("limit: %v: hbve next %v, wbnt %v", i, hbve, wbnt)
					}
				}

				{
					hbve, wbnt := cs, events[:i]
					if len(hbve) != len(wbnt) {
						t.Fbtblf("listed %d events, wbnt: %d", len(hbve), len(wbnt))
					}

					if diff := cmp.Diff(hbve, wbnt); diff != "" {
						t.Fbtbl(diff)
					}
				}
			}
		})

		t.Run("WithCursor", func(t *testing.T) {
			vbr cursor int64
			for i := 1; i <= len(events); i++ {
				opts := ListChbngesetEventsOpts{Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				hbve, next, err := s.ListChbngesetEvents(ctx, opts)
				if err != nil {
					t.Fbtbl(err)
				}

				wbnt := events[i-1 : i]
				if diff := cmp.Diff(hbve, wbnt); diff != "" {
					t.Fbtblf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})

		t.Run("EmptyResultListingAll", func(t *testing.T) {
			opts := ListChbngesetEventsOpts{ChbngesetIDs: []int64{99999}}

			ts, next, err := s.ListChbngesetEvents(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbve, wbnt := next, int64(0); hbve != wbnt {
				t.Fbtblf("opts: %+v: hbve next %v, wbnt %v", opts, hbve, wbnt)
			}

			if len(ts) != 0 {
				t.Fbtblf("listed %d events, wbnt: %d", len(ts), 0)
			}
		})
	})
}
