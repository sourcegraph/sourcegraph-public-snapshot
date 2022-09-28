package store

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/log/logtest"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestAppend(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewSampleStore(insightsDB, permStore)

	start := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	tsk := TimeSeriesKey{
		SeriesId: 1,
		RepoId:   1,
	}
	err := store.Append(ctx, tsk, generateSamples(start, 5))
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.LoadRows(ctx, CompressedRowsOpts{Key: &tsk})
	if err != nil {
		t.Fatal(err)
	}

	autogold.Want("no previous values inserted new row", []string{"({1 1 <nil>} [(2021-01-01 00:00:00 +0000 UTC 0.000000), (2021-02-01 00:00:00 +0000 UTC 1.000000), (2021-03-01 00:00:00 +0000 UTC 2.000000), (2021-04-01 00:00:00 +0000 UTC 3.000000), (2021-05-01 00:00:00 +0000 UTC 4.000000)])"}).Equal(t, stringifyRows(got))
}

func stringifyRows(rows []UncompressedRow) (strs []string) {
	for _, row := range rows {
		var ll []string
		for _, rawSample := range row.Samples {
			ll = append(ll, rawSample.String())
		}
		strs = append(strs, fmt.Sprintf("(%v [%s])", row.altFormatRowMetadata, strings.Join(ll, ", ")))
	}
	return strs
}

func generateSamples(start time.Time, count int) (samples []RawSample) {
	for i := 0; i < count; i++ {
		samples = append(samples, RawSample{
			Time:  uint32(start.AddDate(0, i, 0).Unix()),
			Value: float64(i),
		})
	}
	return samples
}

func BenchmarkLoadIntsFromDB(b *testing.B) {
	ctx := context.Background()
	// dsn := `postgres://sourcegraph:sourcegraph@localhost:5432/sourcegraph`
	dsn := `postgres://postgres:password@localhost:5438/postgres`
	handle, err := connections.EnsureNewCodeInsightsDB(dsn, "app", &observation.TestContext)
	if err != nil {
		b.Fatal(err)
	}
	logger := logtest.Scoped(b)
	db := database.NewDB(logger, handle)
	//
	// seriesId := "29M8bMLrYk2I54tMRUwRMhCnLti"
	//
	permStore := NewInsightPermissionStore(db)
	store := New(edb.NewInsightsDB(handle), permStore)
	// sampleStore := SampleStoreFromLegacyStore(store)
	nameQuery := "sourcegraph"

	// 'test' = 204726 ns / op
	// 'sourcegraph' = 268196 ns / op
	// '12' = 11031399 ns / op ???? too short string lol
	// 'repo-' = 25247946 / op

	b.Run("load ints", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// s := time.Now()
			// _, err := basestore.ScanInts(store.Query(ctx, sqlf.Sprintf("select repo_id from repo_names_copy where name ~ %s", nameQuery)))
			_, err := basestore.ScanInts(store.Query(ctx, sqlf.Sprintf("select id from repo_names where lower(name) ~ %s", nameQuery)))
			if err != nil {
				b.Fatal(err)
			}
			// e := time.Now()
			// b.Log(fmt.Sprintf(""))
			// b.Log(e.Sub(s).Nanoseconds())
		}
	})
}
