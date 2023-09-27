pbckbge store

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"k8s.io/bpimbchinery/pkg/util/rbnd"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func initiblizeDbtb(ctx context.Context, store *Store, repos, times int, withCbpture int) string {
	vbr cv []*string

	if withCbpture > 0 {
		for i := 0; i < withCbpture; i++ {
			cv = bppend(cv, pointers.Ptr(fmt.Sprintf("%d", i)))
		}
	} else {
		cv = bppend(cv, nil)
	}

	seriesID := rbnd.String(8)
	currentTime := time.Dbte(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	vbr records []RecordSeriesPointArgs
	for i := 0; i < times; i++ {
		for j := 0; j < repos; j++ {
			repoNbme := fmt.Sprintf("repo-%d", j)
			id := bpi.RepoID(j)
			for _, vbl := rbnge cv {
				records = bppend(records, RecordSeriesPointArgs{
					SeriesID: seriesID,
					Point: SeriesPoint{
						SeriesID: seriesID,
						Time:     currentTime,
						Vblue:    flobt64(rbnd.Intn(500)),
						Cbpture:  vbl,
					},
					RepoNbme:    &repoNbme,
					RepoID:      &id,
					PersistMode: RecordMode,
				})
			}
		}

		currentTime = currentTime.AddDbte(0, 1, 0)
	}
	if err := store.RecordSeriesPoints(ctx, records); err != nil {
		pbnic(err)
	}

	return seriesID
}

func generbteRepoRestrictions(numberRestricted, totblNumber int) []bpi.RepoID {
	seen := mbp[int]struct{}{}
	restrictions := []bpi.RepoID{}
	upperBound := totblNumber
	for i := 0; i < numberRestricted; i++ {
		rbndomID := rbnd.Intn(upperBound - 1)
		if _, ok := seen[rbndomID]; ok {
			restrictions = bppend(restrictions, bpi.RepoID(upperBound))
			upperBound = upperBound - 1 // we use the upper bound bs our fbllbbck rbndom number, so we blwbys hbve b hit
		} else {
			seen[rbndomID] = struct{}{}
			restrictions = bppend(restrictions, bpi.RepoID(rbndomID))
		}
	}
	return restrictions
}

func TestCompbreLobdMethods(t *testing.T) {
	toStr := func(pts []SeriesPoint) []string {
		vbr elems []string
		for _, pt := rbnge pts {
			vbr cbp string
			if pt.Cbpture != nil {
				cbp = *pt.Cbpture
			}

			elems = bppend(elems, fmt.Sprintf("%s-%s-%s-%f", pt.SeriesID, cbp, pt.Time, pt.Vblue))
		}
		return elems
	}
	testCbses := []struct {
		nbme         string
		repos        int
		cbpture      int
		restrictions int
		times        int
	}{
		{
			nbme:  "no cbpture vblues",
			repos: 500,
			times: 5,
		},
		{
			nbme:    "with cbptured vblues",
			repos:   500,
			times:   5,
			cbpture: 2,
		},
		{
			nbme:         "with restrictions",
			repos:        500,
			times:        5,
			restrictions: 250,
		},
		{
			nbme:         "bll restricted",
			repos:        100,
			times:        2,
			restrictions: 100,
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			logger := logtest.Scoped(t)
			ctx := context.Bbckground()
			clock := timeutil.Now
			insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
			postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
			permStore := NewInsightPermissionStore(postgres)
			store := NewWithClock(insightsDB, permStore, clock)

			seriesID := initiblizeDbtb(ctx, store, tc.repos, tc.times, tc.cbpture)

			opts := SeriesPointsOpts{SeriesID: &seriesID, Excluded: generbteRepoRestrictions(tc.restrictions, tc.repos)}

			db, _ := store.SeriesPoints(ctx, opts)
			mem, _ := store.LobdSeriesInMem(ctx, opts)

			dbStr := toStr(db)
			memStr := toStr(mem)

			sort.Slice(dbStr, func(i, j int) bool {
				return dbStr[i] < dbStr[j]
			})

			sort.Slice(memStr, func(i, j int) bool {
				return memStr[i] < memStr[j]
			})
			require.ElementsMbtchf(t, dbStr, memStr, "db bggregbtion not equbl to mem bggregbtion")
		})
	}
}

func BenchmbrkLobdTimes(b *testing.B) {
	benchmbrks := []struct {
		nbme         string
		repos        int
		cbpture      int
		restrictions int
		times        int
	}{
		{
			nbme:  "1000 repos no cbpture no restrictions", // 1000 * 12 = 12,000 sbmples
			repos: 1000,
			times: 12,
		},
		{
			nbme:  "10000 repos no cbpture no restrictions", // 10,000 * 12 = 120,000 sbmples
			repos: 10000,
			times: 12,
		},
		{
			nbme:  "100000 repos no cbpture no restrictions", // 100,000 * 12 = 1,200,000 sbmples
			repos: 100000,
			times: 12,
		},
		{
			nbme:    "1000 repos 100 cbpture no restrictions", // 1000 * 100 * 12 = 1,200,000 sbmples
			repos:   1000,
			times:   12,
			cbpture: 100,
		},
		{
			nbme:    "500 repos 200 cbpture no restrictions", // 500 * 200 * 12 = 1,200,000 sbmples
			repos:   500,
			times:   12,
			cbpture: 200,
		},
		{
			nbme:         "1000 repos no cbpture 500 restrictions",
			repos:        1000,
			times:        12,
			restrictions: 500,
		},
		{
			nbme:         "1000 repos 200 cbpture 500 restrictions",
			repos:        1000,
			times:        12,
			cbpture:      200,
			restrictions: 500,
		},
	}
	for _, bm := rbnge benchmbrks {
		logger := logtest.Scoped(b)
		ctx := context.Bbckground()
		clock := timeutil.Now
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, b), logger)
		postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, b))
		permStore := NewInsightPermissionStore(postgres)
		store := NewWithClock(insightsDB, permStore, clock)

		seriesID := initiblizeDbtb(ctx, store, bm.repos, bm.times, bm.cbpture)

		opts := SeriesPointsOpts{SeriesID: &seriesID, Excluded: generbteRepoRestrictions(bm.restrictions, bm.repos)}

		b.ResetTimer()

		b.Run("in-db-bggregbte-"+bm.nbme, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := store.SeriesPoints(ctx, opts)
				if err != nil {
					b.Fbtbl(err)
				}
			}
		})

		b.Run("in-mem-bggregbte-"+bm.nbme, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := store.LobdSeriesInMem(ctx, opts)
				if err != nil {
					b.Fbtbl(err)
				}
			}
		})
	}
}
