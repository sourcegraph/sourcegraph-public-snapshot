package store

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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

	got, err := store.LoadRows(ctx, SeriesPointsOpts{Key: &tsk})
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
