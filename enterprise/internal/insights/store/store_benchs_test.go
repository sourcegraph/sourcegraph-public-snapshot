package store

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func initializeData(ctx context.Context, store *Store, repos, times int, withCapture int) string {
	var cv []*string
	strPtr := func(s string) *string {
		return &s
	}

	if withCapture > 0 {
		for i := 0; i < withCapture; i++ {
			cv = append(cv, strPtr(fmt.Sprintf("%d", i)))
		}
	} else {
		cv = append(cv, nil)
	}

	seriesID := rand.String(8)
	currentTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	var records []RecordSeriesPointArgs
	for i := 0; i < times; i++ {
		for j := 0; j < repos; j++ {
			repoName := fmt.Sprintf("repo-%d", j)
			id := api.RepoID(j)
			for _, val := range cv {
				records = append(records, RecordSeriesPointArgs{
					SeriesID: seriesID,
					Point: SeriesPoint{
						SeriesID: seriesID,
						Time:     currentTime,
						Value:    float64(rand.Intn(500)),
						Capture:  val,
					},
					RepoName:    &repoName,
					RepoID:      &id,
					PersistMode: RecordMode,
				})
			}
		}

		currentTime = currentTime.AddDate(0, 1, 0)
	}
	if err := store.RecordSeriesPoints(ctx, records); err != nil {
		panic(err)
	}

	return seriesID
}

func TestCompareLoadMethods(t *testing.T) {
	toStr := func(pts []SeriesPoint) []string {
		var elems []string
		for _, pt := range pts {
			var cap string
			if pt.Capture != nil {
				cap = *pt.Capture
			}

			elems = append(elems, fmt.Sprintf("%s-%s-%s-%f", pt.SeriesID, cap, pt.Time, pt.Value))
		}
		return elems
	}

	t.Run("no capture values", func(t *testing.T) {
		logger := logtest.Scoped(t)
		ctx := context.Background()
		clock := timeutil.Now
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
		postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
		permStore := NewInsightPermissionStore(postgres)
		store := NewWithClock(insightsDB, permStore, clock)

		seriesID := initializeData(ctx, store, 500, 5, 0)

		db, _ := store.SeriesPoints(ctx, SeriesPointsOpts{SeriesID: &seriesID})
		mem, _ := store.LoadSeriesInMem(ctx, SeriesPointsOpts{SeriesID: &seriesID})

		dbStr := toStr(db)
		memStr := toStr(mem)

		sort.Slice(dbStr, func(i, j int) bool {
			return dbStr[i] < dbStr[j]
		})

		sort.Slice(memStr, func(i, j int) bool {
			return memStr[i] < memStr[j]
		})
		require.ElementsMatchf(t, dbStr, memStr, "db aggregation not equal to mem aggregation")
	})

	t.Run("with captured values", func(t *testing.T) {
		logger := logtest.Scoped(t)
		ctx := context.Background()
		clock := timeutil.Now
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
		postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
		permStore := NewInsightPermissionStore(postgres)
		store := NewWithClock(insightsDB, permStore, clock)

		seriesID := initializeData(ctx, store, 500, 5, 2)

		db, _ := store.SeriesPoints(ctx, SeriesPointsOpts{SeriesID: &seriesID})
		mem, _ := store.LoadSeriesInMem(ctx, SeriesPointsOpts{SeriesID: &seriesID})

		dbStr := toStr(db)
		memStr := toStr(mem)

		sort.Slice(dbStr, func(i, j int) bool {
			return dbStr[i] < dbStr[j]
		})

		sort.Slice(memStr, func(i, j int) bool {
			return memStr[i] < memStr[j]
		})
		require.ElementsMatchf(t, dbStr, memStr, "db aggregation not equal to mem aggregation")
	})
}

func BenchmarkLoadTimes(b *testing.B) {
	benchmarks := []struct {
		name    string
		repos   int
		capture int
		times   int
	}{
		{
			name:  "1000 repos no capture no restrictions", // 1000 * 12 = 12,000 samples
			repos: 1000,
			times: 12,
		},
		{
			name:    "1000 repos 100 capture no restrictions", // 1000 * 100 * 12 = 1,200,000 samples
			repos:   1000,
			times:   12,
			capture: 100,
		},
		{
			name:    "500 repos 200 capture no restrictions", // 500 * 200 * 12 = 1,200,000 samples
			repos:   500,
			times:   12,
			capture: 200,
		},
	}
	for _, bm := range benchmarks {
		logger := logtest.Scoped(b)
		ctx := context.Background()
		clock := timeutil.Now
		insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, b))
		postgres := database.NewDB(logger, dbtest.NewDB(logger, b))
		permStore := NewInsightPermissionStore(postgres)
		store := NewWithClock(insightsDB, permStore, clock)

		seriesID := initializeData(ctx, store, bm.repos, bm.times, bm.capture)

		opts := SeriesPointsOpts{SeriesID: &seriesID}

		b.ResetTimer()

		b.Run("in-db-aggregate-"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := store.SeriesPoints(ctx, opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("in-mem-aggregate-"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := store.LoadSeriesInMem(ctx, opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
