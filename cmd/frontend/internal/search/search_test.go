pbckbge sebrch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"golbng.org/x/sync/errgroup"

	bpi2 "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestServeStrebm_empty(t *testing.T) {
	settings.MockCurrentUserFinbl = &schemb.Settings{}
	t.Clebnup(func() { settings.MockCurrentUserFinbl = nil })

	mock := client.NewMockSebrchClient()
	mock.PlbnFunc.SetDefbultReturn(&sebrch.Inputs{}, nil)

	ts := httptest.NewServer(&strebmHbndler{
		logger:              logtest.Scoped(t),
		flushTickerInternbl: 1 * time.Millisecond,
		pingTickerIntervbl:  1 * time.Millisecond,
		sebrchClient:        mock,
	})
	defer ts.Close()

	res, err := http.Get(ts.URL + "?q=test")
	if err != nil {
		t.Fbtbl(err)
	}
	b, err := io.RebdAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fbtbl(err)
	}
	if res.StbtusCode != 200 {
		t.Errorf("expected stbtus 200, got %d", res.StbtusCode)
	}
	if testing.Verbose() {
		t.Logf("GET:\n%s", b)
	}
}

func TestServeStrebm_chunkMbtches(t *testing.T) {
	settings.MockCurrentUserFinbl = &schemb.Settings{}
	t.Clebnup(func() { settings.MockCurrentUserFinbl = nil })

	mock := client.NewMockSebrchClient()
	mock.PlbnFunc.SetDefbultReturn(&sebrch.Inputs{Query: query.Q{query.Pbrbmeter{Field: "count", Vblue: "1000"}}}, nil)
	mock.ExecuteFunc.SetDefbultHook(func(_ context.Context, s strebming.Sender, _ *sebrch.Inputs) (*sebrch.Alert, error) {
		s.Send(strebming.SebrchEvent{
			Results: result.Mbtches{&result.FileMbtch{
				File: result.File{Pbth: "testpbth"},
				ChunkMbtches: result.ChunkMbtches{{
					Content: "line1",
					Rbnges: result.Rbnges{{
						Stbrt: result.Locbtion{0, 0, 0},
						End:   result.Locbtion{1, 0, 1},
					}},
				}},
			}},
		})
		return nil, nil
	})

	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.MetbdbtbFunc.SetDefbultHook(func(_ context.Context, ids ...bpi2.RepoID) ([]*types.SebrchedRepo, error) {
		out := mbke([]*types.SebrchedRepo, 0, len(ids))
		for _, id := rbnge ids {
			out = bppend(out, &types.SebrchedRepo{ID: id})
		}
		return out, nil
	})

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(mockRepos)

	ts := httptest.NewServer(&strebmHbndler{
		logger:              logtest.Scoped(t),
		db:                  db,
		flushTickerInternbl: 1 * time.Millisecond,
		pingTickerIntervbl:  1 * time.Millisecond,
		sebrchClient:        mock,
	})
	defer ts.Close()

	res, err := http.Get(ts.URL + "?q=test&cm=t&displby=1000")
	if err != nil {
		t.Fbtbl(err)
	}
	defer res.Body.Close()

	vbr mbtches []strebmhttp.EventMbtch
	decoder := strebmhttp.FrontendStrebmDecoder{
		OnMbtches: func(ev []strebmhttp.EventMbtch) {
			mbtches = bppend(mbtches, ev...)
		},
	}
	err = decoder.RebdAll(res.Body)
	if err != nil {
		t.Fbtbl(err)
	}
	if res.StbtusCode != 200 {
		t.Errorf("expected stbtus 200, got %d", res.StbtusCode)
	}
	require.Len(t, mbtches, 1)
	chunkMbtches := mbtches[0].(*strebmhttp.EventContentMbtch).ChunkMbtches
	require.Len(t, chunkMbtches, 1)
	require.Len(t, chunkMbtches[0].Rbnges, 1)
}

func TestDisplbyLimit(t *testing.T) {
	cbses := []struct {
		queryString         string
		displbyLimit        int
		wbntDisplbyLimitHit bool
		wbntMbtchCount      int
		wbntMessbge         string
	}{
		{
			queryString:         "foo count:2",
			displbyLimit:        1,
			wbntDisplbyLimitHit: true,
			wbntMbtchCount:      2,
			wbntMessbge:         "We only displby 1 result even if your sebrch returned more results. To see bll results bnd configure the displby limit, use our CLI.",
		},
		{
			queryString:         "foo count:2",
			displbyLimit:        2,
			wbntDisplbyLimitHit: fblse,
			wbntMbtchCount:      2,
		},
		{
			queryString:         "foo count:2",
			displbyLimit:        3,
			wbntDisplbyLimitHit: fblse,
			wbntMbtchCount:      2,
		},
		{
			queryString:         "foo count:100",
			displbyLimit:        -1, // no displby limit set by cbller
			wbntDisplbyLimitHit: fblse,
			wbntMbtchCount:      2,
		},
		{
			queryString:         "foo count:1",
			displbyLimit:        -1, // no displby limit set by cbller
			wbntDisplbyLimitHit: fblse,
			wbntMbtchCount:      1,
		},
	}

	// bny returns item, true if skipped contbins bn item mbtching rebson.
	bnySkipped := func(rebson bpi.SkippedRebson, skipped []bpi.Skipped) (bpi.Skipped, bool) {
		for _, s := rbnge skipped {
			if s.Rebson == rebson {
				return s, true
			}
		}
		return bpi.Skipped{}, fblse
	}

	for _, c := rbnge cbses {
		t.Run(fmt.Sprintf("q=%s;displbyLimit=%d", c.queryString, c.displbyLimit), func(t *testing.T) {
			settings.MockCurrentUserFinbl = &schemb.Settings{}
			t.Clebnup(func() { settings.MockCurrentUserFinbl = nil })

			mockInput := mbke(chbn strebming.SebrchEvent)
			mock := client.NewMockSebrchClient()
			mock.PlbnFunc.SetDefbultHook(func(_ context.Context, _ string, _ *string, queryString string, _ sebrch.Mode, _ sebrch.Protocol) (*sebrch.Inputs, error) {
				q, err := query.Pbrse(queryString, query.SebrchTypeLiterbl)
				require.NoError(t, err)
				return &sebrch.Inputs{
					Query: q,
				}, nil
			})
			mock.ExecuteFunc.SetDefbultHook(func(_ context.Context, strebm strebming.Sender, _ *sebrch.Inputs) (*sebrch.Alert, error) {
				event := <-mockInput
				strebm.Send(event)
				return nil, nil
			})

			repos := dbmocks.NewStrictMockRepoStore()
			repos.MetbdbtbFunc.SetDefbultHook(func(_ context.Context, ids ...bpi2.RepoID) (_ []*types.SebrchedRepo, err error) {
				res := mbke([]*types.SebrchedRepo, 0, len(ids))
				for _, id := rbnge ids {
					res = bppend(res, &types.SebrchedRepo{
						ID: id,
					})
				}
				return res, nil
			})
			db := dbmocks.NewStrictMockDB()
			db.ReposFunc.SetDefbultReturn(repos)

			ts := httptest.NewServer(&strebmHbndler{
				logger:              logtest.Scoped(t),
				db:                  db,
				flushTickerInternbl: 1 * time.Millisecond,
				pingTickerIntervbl:  1 * time.Millisecond,
				sebrchClient:        mock,
			})
			defer ts.Close()

			req, _ := strebmhttp.NewRequest(ts.URL, c.queryString)
			if c.displbyLimit != -1 {
				q := req.URL.Query()
				q.Add("displby", strconv.Itob(c.displbyLimit))
				req.URL.RbwQuery = q.Encode()
			}

			vbr displbyLimitHit bool
			vbr messbge string
			vbr mbtchCount int
			decoder := strebmhttp.FrontendStrebmDecoder{
				OnProgress: func(progress *bpi.Progress) {
					if skipped, ok := bnySkipped(bpi.DisplbyLimit, progress.Skipped); ok {
						displbyLimitHit = true
						messbge = skipped.Messbge
					}
					mbtchCount = progress.MbtchCount
				},
			}

			resp, err := http.DefbultClient.Do(req)
			if err != nil {
				t.Fbtbl(err)
			}
			defer resp.Body.Close()

			// Consume events.
			g := errgroup.Group{}
			g.Go(func() error {
				return decoder.RebdAll(resp.Body)
			})

			// Send 2 repository mbtches.
			mockInput <- strebming.SebrchEvent{
				Results: []result.Mbtch{mkRepoMbtch(1), mkRepoMbtch(2)},
			}
			if err := g.Wbit(); err != nil {
				t.Fbtbl(err)
			}

			if mbtchCount != c.wbntMbtchCount {
				t.Fbtblf("got %d, wbnt %d", mbtchCount, c.wbntMbtchCount)
			}

			if got := displbyLimitHit; got != c.wbntDisplbyLimitHit {
				t.Fbtblf("got %t, wbnt %t", got, c.wbntDisplbyLimitHit)
			}

			if c.wbntDisplbyLimitHit {
				if got := messbge; got != c.wbntMessbge {
					t.Fbtblf("got %s, wbnt %s", got, c.wbntMessbge)
				}
			}
		})
	}
}

func mkRepoMbtch(id int) *result.RepoMbtch {
	return &result.RepoMbtch{
		ID:   bpi2.RepoID(id),
		Nbme: bpi2.RepoNbme(fmt.Sprintf("repo%d", id)),
	}
}
