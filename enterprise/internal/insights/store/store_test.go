package store

import (
	bytes2 "bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/storage"

	"github.com/sourcegraph/sourcegraph/internal/observation"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"

	"github.com/keisku/gorilla"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/jwilder/encoding/simple8b"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestSeriesPoints(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))

	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	// Confirm we get no results initially.
	points, err := store.SeriesPoints(ctx, SeriesPointsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("SeriesPoints", []SeriesPoint{}).Equal(t, points)

	// Insert some fake data.
	_, err = insightsDB.ExecContext(context.Background(), `
INSERT INTO repo_names(name) VALUES ('github.com/gorilla/mux-original');
INSERT INTO repo_names(name) VALUES ('github.com/gorilla/mux-renamed');
SELECT setseed(0.5);
INSERT INTO series_points(
    time,
	series_id,
    value,
    repo_id,
    repo_name_id,
    original_repo_name_id)
SELECT time,
    'somehash',
    random()*80 - 40,
    2,
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-renamed'),
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-original')
	FROM GENERATE_SERIES(CURRENT_TIMESTAMP::date - INTERVAL '30 weeks', CURRENT_TIMESTAMP::date, '2 weeks') AS time;
`)
	if err != nil {
		t.Fatal(err)
	}

	time := func(s string) *time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return &v
	}

	t.Run("all data points", func(t *testing.T) {
		// Confirm we get all data points.
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(points)
		autogold.Want("SeriesPoints(2).len", int(16)).Equal(t, len(points))
	})

	t.Run("subset of data", func(t *testing.T) {
		// Confirm we can get a subset of data points.
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{
			From: time("2020-03-01T00:00:00Z"),
			To:   time("2020-06-01T00:00:00Z"),
		})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("SeriesPoints(3).len", int(0)).Equal(t, len(points))
	})

	t.Run("latest 3 points", func(t *testing.T) {
		// Confirm we can get a subset of data points.
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{
			Limit: 3,
		})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("SeriesPoints(4).len", int(3)).Equal(t, len(points))
	})

	t.Run("include list", func(t *testing.T) {
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{Included: []api.RepoID{2}})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(16, len(points)); diff != "" {
			t.Errorf("unexpected results from include list: %v", diff)
		}
	})
	t.Run("exclude list", func(t *testing.T) {
		points, err = store.SeriesPoints(ctx, SeriesPointsOpts{Excluded: []api.RepoID{2}})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(0, len(points)); diff != "" {
			t.Errorf("unexpected results from include list: %v", diff)
		}
	})
}

func TestCountData(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	timeValue := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	timePtr := func(s string) *time.Time {
		t := timeValue(s)
		return &t
	}
	optionalString := func(v string) *string { return &v }
	optionalRepoID := func(v api.RepoID) *api.RepoID { return &v }

	// Record some duplicate data points.
	for _, record := range []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: timeValue("2020-03-01T00:00:00Z"), Value: 1.1},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "two",
			Point:       SeriesPoint{Time: timeValue("2020-03-02T00:00:00Z"), Value: 2.2},
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "two",
			Point:       SeriesPoint{Time: timeValue("2020-03-02T00:01:00Z"), Value: 2.2},
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "three",
			Point:       SeriesPoint{Time: timeValue("2020-03-03T00:00:00Z"), Value: 3.3},
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "three",
			Point:       SeriesPoint{Time: timeValue("2020-03-03T00:01:00Z"), Value: 3.3},
			PersistMode: RecordMode,
		},
	} {
		if err := store.RecordSeriesPoint(ctx, record); err != nil {
			t.Fatal(err)
		}
	}

	// How many data points on 02-29?
	numDataPoints, err := store.CountData(ctx, CountDataOpts{
		From: timePtr("2020-02-29T00:00:00Z"),
		To:   timePtr("2020-02-29T23:59:59Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("first", int(0)).Equal(t, numDataPoints)

	// How many data points on 03-01?
	numDataPoints, err = store.CountData(ctx, CountDataOpts{
		From: timePtr("2020-03-01T00:00:00Z"),
		To:   timePtr("2020-03-01T23:59:59Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("second", int(1)).Equal(t, numDataPoints)

	// How many data points from 03-01 to 03-04?
	numDataPoints, err = store.CountData(ctx, CountDataOpts{
		From: timePtr("2020-03-01T00:00:00Z"),
		To:   timePtr("2020-03-04T23:59:59Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("third", int(5)).Equal(t, numDataPoints)
}

func TestRecordSeriesPoints(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	optionalString := func(v string) *string { return &v }
	optionalRepoID := func(v api.RepoID) *api.RepoID { return &v }

	current := time.Date(2021, time.September, 10, 10, 0, 0, 0, time.UTC)

	for _, record := range []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current, Value: 1.1},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current.Add(-time.Hour * 24 * 14), Value: 2.2},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current.Add(-time.Hour * 24 * 28), Value: 3.3},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: RecordMode,
		},
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current.Add(-time.Hour * 24 * 42), Value: 3.3},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: RecordMode,
		},
	} {
		if err := store.RecordSeriesPoint(ctx, record); err != nil {
			t.Fatal(err)
		}
	}

	want := []SeriesPoint{
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 42),
			Value:    3.3,
		},
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 28),
			Value:    3.3,
		},
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 14),
			Value:    2.2,
		},
		{
			SeriesID: "one",
			Time:     current,
			Value:    1.1,
		},
	}

	// Confirm we get the expected data back.
	points, err := store.SeriesPoints(ctx, SeriesPointsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("len(points)", int(4)).Equal(t, len(points))
	if diff := cmp.Diff(4, len(points)); diff != "" {
		t.Errorf("len(points): %v", diff)
	}
	if diff := cmp.Diff(want[0], points[0]); diff != "" {
		t.Errorf("points[0].String(): %v", diff)
	}
	if diff := cmp.Diff(want[1], points[1]); diff != "" {
		t.Errorf("points[1].String(): %v", diff)
	}
	if diff := cmp.Diff(want[2], points[2]); diff != "" {
		t.Errorf("points[2].String(): %v", diff)
	}
	if diff := cmp.Diff(want[3], points[3]); diff != "" {
		t.Errorf("points[3].String(): %v", diff)
	}
}

func TestRecordSeriesPointsSnapshotOnly(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	optionalString := func(v string) *string { return &v }
	optionalRepoID := func(v api.RepoID) *api.RepoID { return &v }

	current := time.Date(2021, time.September, 10, 10, 0, 0, 0, time.UTC)

	for _, record := range []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current, Value: 1.1},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: SnapshotMode,
		},
	} {
		if err := store.RecordSeriesPoint(ctx, record); err != nil {
			t.Fatal(err)
		}
	}

	// check snapshots table has a row
	row := store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(snapshotsTable)))
	if row.Err() != nil {
		t.Fatal(row.Err())
	}

	want := 1
	var got int
	err := row.Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected count from snapshots table (want/got): %v", diff)
	}

	// check recordings table has no rows
	row = store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(recordingTable)))
	if row.Err() != nil {
		t.Fatal(row.Err())
	}

	want = 0
	err = row.Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected count from recordings table (want/got): %v", diff)
	}
}

func TestRecordSeriesPointsRecordingOnly(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	optionalString := func(v string) *string { return &v }
	optionalRepoID := func(v api.RepoID) *api.RepoID { return &v }

	current := time.Date(2021, time.September, 10, 10, 0, 0, 0, time.UTC)

	for _, record := range []RecordSeriesPointArgs{
		{
			SeriesID:    "one",
			Point:       SeriesPoint{Time: current, Value: 1.1},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: RecordMode,
		},
	} {
		if err := store.RecordSeriesPoint(ctx, record); err != nil {
			t.Fatal(err)
		}
	}

	// check snapshots table has a row
	row := store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(snapshotsTable)))
	if row.Err() != nil {
		t.Fatal(row.Err())
	}

	want := 0
	var got int
	err := row.Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected count from snapshots table (want/got): %v", diff)
	}

	// check recordings table has no rows
	row = store.QueryRow(ctx, sqlf.Sprintf("select count(*) from %s", sqlf.Sprintf(recordingTable)))
	if row.Err() != nil {
		t.Fatal(row.Err())
	}

	want = 1
	err = row.Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected count from recordings table (want/got): %v", diff)
	}
}

func TestDeleteSnapshots(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	optionalString := func(v string) *string { return &v }
	optionalRepoID := func(v api.RepoID) *api.RepoID { return &v }

	current := time.Date(2021, time.September, 10, 10, 0, 0, 0, time.UTC)

	seriesID := "one"
	for _, record := range []RecordSeriesPointArgs{
		{
			SeriesID:    seriesID,
			Point:       SeriesPoint{Time: current, Value: 1.1},
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: SnapshotMode,
		},
		{
			SeriesID:    seriesID,
			Point:       SeriesPoint{Time: current.Add(time.Hour), Value: 1.1}, // offsetting the time by an hour so that the point is not deduplicated
			RepoName:    optionalString("repo1"),
			RepoID:      optionalRepoID(3),
			PersistMode: RecordMode,
		},
	} {
		if err := store.RecordSeriesPoint(ctx, record); err != nil {
			t.Fatal(err)
		}
	}

	// first check that we have one recording and one snapshot
	points, err := store.SeriesPoints(ctx, SeriesPointsOpts{SeriesID: &seriesID})
	if err != nil {
		t.Fatal(err)
	}
	got := len(points)
	want := 2
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected count of series points prior to deleting snapshots (want/got): %v", diff)
	}
	err = store.DeleteSnapshots(ctx, &types.InsightSeries{SeriesID: seriesID})
	if err != nil {
		t.Fatal(err)
	}
	// now verify that the remaining point is the recording
	points, err = store.SeriesPoints(ctx, SeriesPointsOpts{SeriesID: &seriesID})
	if err != nil {
		t.Fatal(err)
	}
	got = len(points)
	want = 1
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected count of series points after deleting snapshots (want/got): %v", diff)
	}
	autogold.Equal(t, points, autogold.ExportedOnly())
}

func TestValues(t *testing.T) {
	ids := []api.RepoID{1, 2, 3, 4, 5, 6}
	got := values(ids)
	want := "VALUES (1),(2),(3),(4),(5),(6)"

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected values string: %v", diff)
	}
}

func TestDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	now := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsdb := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))

	repoName := "reallygreatrepo"
	repoId := api.RepoID(5)

	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	timeseriesStore := NewWithClock(insightsdb, permStore, clock)

	err := timeseriesStore.RecordSeriesPoints(ctx, []RecordSeriesPointArgs{
		{
			SeriesID: "series1",
			Point: SeriesPoint{
				SeriesID: "series1",
				Time:     now,
				Value:    50,
			},
			RepoName:    &repoName,
			RepoID:      &repoId,
			PersistMode: RecordMode,
		},
		{
			SeriesID: "series1",
			Point: SeriesPoint{
				SeriesID: "series1",
				Time:     now,
				Value:    50,
			},
			RepoName:    &repoName,
			RepoID:      &repoId,
			PersistMode: SnapshotMode,
		},
		{
			SeriesID: "series2",
			Point: SeriesPoint{
				SeriesID: "series2",
				Time:     now,
				Value:    25,
			},
			RepoName:    &repoName,
			RepoID:      &repoId,
			PersistMode: RecordMode,
		},
		{
			SeriesID: "series2",
			Point: SeriesPoint{
				SeriesID: "series2",
				Time:     now,
				Value:    25,
			},
			RepoName:    &repoName,
			RepoID:      &repoId,
			PersistMode: SnapshotMode,
		},
	})
	if err != nil {
		t.Error(err)
	}

	err = timeseriesStore.Delete(ctx, "series1")
	if err != nil {
		t.Fatal(err)
	}

	getCountForSeries := func(ctx context.Context, timeseriesStore *Store, mode PersistMode, seriesId string) int {
		table, err := getTableForPersistMode(mode)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(table)
		q := sqlf.Sprintf("select count(*) from %s where series_id = %s;", sqlf.Sprintf(table), seriesId)
		row := timeseriesStore.QueryRow(ctx, q)
		val, err := basestore.ScanInt(row)
		if err != nil {
			t.Fatal(err)
		}
		return val
	}

	if getCountForSeries(ctx, timeseriesStore, RecordMode, "series1") != 0 {
		t.Errorf("expected 0 count for series1 in record table")
	}
	if getCountForSeries(ctx, timeseriesStore, SnapshotMode, "series1") != 0 {
		t.Errorf("expected 0 count for series1 in snapshot table")
	}

	if getCountForSeries(ctx, timeseriesStore, RecordMode, "series2") != 1 {
		t.Errorf("expected 1 count for series2 in record table")
	}
	if getCountForSeries(ctx, timeseriesStore, SnapshotMode, "series2") != 1 {
		t.Errorf("expected 1 count for series2 in snapshot table")
	}
}

func TestIntegerEncoding(t *testing.T) {
	vals := make([]int32, 0)
	current := int32(500)
	vals = append(vals, current)
	for i := 0; i <= 24; i++ {
		r := int32(rand.IntnRange(-10, 10))
		n := current + r
		current = n
		vals = append(vals, n)
	}
	s := samples(vals).ToDelta()
	t.Log(s)

	v := convert(s)

	// t.Log(vals)
	//
	// compressed, err := simple8b.EncodeAll(vals)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log(compressed)
	//
	// count, err := simple8b.Count(compressed[0])
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// dst := make([]uint64, len(vals))
	// decoded, err := simple8b.DecodeAll(dst, compressed)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log(dst)
	// t.Log(decoded)

	encoder := simple8b.NewEncoder()
	encoder.SetValues(v)
	bytes, err := encoder.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(bytes)
	t.Log(len(bytes))
	t.Log(hex.EncodeToString(bytes))

	decoder := simple8b.NewDecoder(bytes)
	for decoder.Next() {
		t.Log(decoder.Read())
	}
}

func TestTimeEncoding(t *testing.T) {
	input := make([]time.Time, 0)
	start := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 100; i++ {
		input = append(input, start.AddDate(0, 0, i))
	}

	tt := timeSamples(input)
	t.Log(tt)

	delts := tt.ToDelta()
	t.Log("delts")
	t.Log(delts)

	encoder := simple8b.NewEncoder()
	encoder.SetValues(delts)
	bytes, err := encoder.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(bytes))
	oldSize := 8 * len(input)
	newSize := len(bytes)

	diff := (float64(newSize-oldSize) / float64(oldSize)) * 100

	t.Log(fmt.Sprintf("bytes: %d, regular: %d, compression: %f", len(bytes), oldSize, math.Abs(diff)))
	// t.Log(fmt.Sprintf("bytes: %d, regular: %d, compression: %f", len(bytes), oldSize, float64(len(bytes)*100)/float64(oldSize)))
	// t.Log(hex.EncodeToString(bytes))

	got := make([]uint64, 0, len(input))
	decoder := simple8b.NewDecoder(bytes)
	for decoder.Next() {
		// t.Log(decoder.Read())
		got = append(got, decoder.Read())
	}
	uncompd := uncompressTimes(got)
	t.Log(uncompd)
	require.ElementsMatchf(t, uncompd, tt, "not equal")
}

type samples []int32

func (s samples) ToDelta() []int32 {
	result := make([]int32, len(s))
	result[0] = s[0]
	for i := 1; i < len(s); i++ {
		result[i] = s[i] - s[i-1]
	}
	return result
}

func convert(in []int32) []uint64 {
	out := make([]uint64, len(in))
	for i := range in {
		out[i] = uint64(in[i])
	}
	return out
}

type timeSamples []time.Time

func (t timeSamples) ToDelta() []uint64 {
	result := make([]uint64, len(t))
	result[0] = uint64(t[0].Unix())
	for i := 1; i < len(t); i++ {
		result[i] = uint64(t[i].Sub(t[i-1]).Hours())
	}
	return result
}

func uncompressTimes(in []uint64) timeSamples {
	result := make(timeSamples, len(in))
	current := time.Unix(int64(in[0]), 0)
	result[0] = current

	for i := 1; i < len(in); i++ {
		temp := current.Add(time.Hour * time.Duration(in[i]))
		result[i] = temp
		current = temp
	}
	return result
}

func TestGorilla(t *testing.T) {
	type sample struct {
		at    time.Time
		value float64
	}

	ss := make([]*sample, 0)

	input := make([]time.Time, 0)
	start := time.Date(2021, 1, 1, 5, 30, 0, 0, time.UTC)
	for i := 0; i < 12; i++ {
		input = append(input, start.AddDate(0, i, 0))
		ss = append(ss, &sample{
			at:    start.AddDate(0, 0, i),
			value: float64(rand.IntnRange(0, 500000)),
		})
	}
	t.Log(ss)

	buf := new(bytes2.Buffer)
	// header := uint32(start.AddDate(0, 0, -1).Unix())
	header := uint32(start.Unix())

	c, finish, err := gorilla.NewCompressor(buf, header)
	if err != nil {
		t.Fatal(err)
	}

	for _, t2 := range input {
		// t2 := ss[i]
		// err := c.Compress(uint32(t2.at.Unix()), t2.value)
		err := c.Compress(uint32(t2.Unix()), float64(rand.IntnRange(0, 500000)))
		if err != nil {
			t.Fatal(err)
		}
	}

	err = finish()
	if err != nil {
		t.Fatal(err)
	}

	size := len(buf.Bytes())
	t.Log(size)
	t.Log(hex.EncodeToString(buf.Bytes()))

	// 100 * 8 = time
	// 100 * 8 = values

	// 4.5 bytes per sample

	// currently
	// 8 time + 8 bytes value = 16 per row

	decompressor, header, err := gorilla.NewDecompressor(buf)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]sample, 0)
	iter := decompressor.Iterator()
	for iter.Next() {
		ts, v := iter.At()
		t.Log(ts, v)
		got = append(got, sample{
			at:    time.Unix(int64(ts), 0),
			value: v,
		})
	}
	t.Log("got")
	for _, s := range got {
		t.Log(s.at)
		t.Log(s.value)
	}
	// t.Log(got)
}

type sample struct {
	at    time.Time
	value float64
}

type dataForRepo struct {
	repoId   int
	repoName string
	ss       []RawSample
}

func TestLoadStuff(t *testing.T) {
	ctx := context.Background()
	dsn := `postgres://sourcegraph:sourcegraph@localhost:5432/sourcegraph`
	handle, err := connections.EnsureNewCodeInsightsDB(dsn, "app", &observation.TestContext)
	if err != nil {
		t.Fatal(err)
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, handle)
	//
	// seriesId := "29M8bMLrYk2I54tMRUwRMhCnLti"
	//
	permStore := NewInsightPermissionStore(db)
	store := New(edb.NewInsightsDB(handle), permStore)

	// seriesId := rand.String(10)
	seriesId := "findme"
	t.Log(seriesId)
	rawId := 8

	numRepos := 10000
	repoIdMin := 20000
	repoIdMax := repoIdMin + numRepos

	valMax := 100000

	numSamples := 120

	times := make([]time.Time, 0)
	start := time.Date(2021, 1, 1, 5, 30, 0, 0, time.UTC)
	for i := 0; i < numSamples; i++ {
		times = append(times, start.AddDate(0, i, 0))
	}

	var allData []dataForRepo

	for i := repoIdMin; i < repoIdMax; i++ {
		repoId := i
		name := fmt.Sprintf("repo-%d", repoId)

		var smps []RawSample
		for _, current := range times {
			smps = append(smps, RawSample{
				Time:  uint32(current.Unix()),
				Value: float64(rand.IntnRange(0, valMax+1)),
			})
		}

		allData = append(allData, dataForRepo{
			repoId:   i,
			repoName: name,
			ss:       smps,
		})
	}

	t.Log(len(allData))

	for _, datum := range allData {
		err := store.StoreAlternateFormat(ctx, UncompressedRow{
			altFormatRowMetadata: altFormatRowMetadata{
				RepoId: uint32(datum.repoId),
			},
			Samples: datum.ss,
		}, uint32(rawId))
		if err != nil {
			t.Fatalf("failed on repo_id: %d %s", datum.repoId, err.Error())
		}
	}

	// err = writeNew(ctx, store, rawId, allData)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	err = writeOld(ctx, store, seriesId, allData)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func writeOld(ctx context.Context, store *Store, seriesId string, allData []dataForRepo) error {
	for i, datum := range allData {
		if i%100 == 0 {
			println(fmt.Sprintf("old %d", i))
		}
		for _, sample := range datum.ss {
			rn := datum.repoName
			id := api.RepoID(datum.repoId)
			if err := store.RecordSeriesPoint(ctx, RecordSeriesPointArgs{
				SeriesID: seriesId,
				Point: SeriesPoint{
					SeriesID: seriesId,
					Time:     time.Unix(int64(sample.Time), 0),
					Value:    sample.Value,
				},
				RepoName:    &rn,
				RepoID:      &id,
				PersistMode: "record",
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func BenchmarkReadGeneratedSeries(b *testing.B) {
	ctx := context.Background()
	dsn := `postgres://sourcegraph:sourcegraph@localhost:5432/sourcegraph`
	handle, err := connections.EnsureNewCodeInsightsDB(dsn, "app", &observation.TestContext)
	if err != nil {
		b.Fatal(err)
	}
	logger := logtest.Scoped(b)
	db := database.NewDB(logger, handle)

	// seriesId := "cnf79vsfwc"
	// plainId := 5

	// mzvgb7cltt

	// seriesId := "mzvgb7cltt"
	seriesId := "findme"
	// plainId := 8

	permStore := NewInsightPermissionStore(db)
	store := New(edb.NewInsightsDB(handle), permStore)

	var newFormat []SeriesPoint

	// b.Run("new format", func(b *testing.B) {
	// 	got, err := loadNew(ctx, store, plainId)
	// 	if err != nil {
	// 		b.Fatal(err)
	// 	}
	// 	newFormat = toTs(got, seriesId)
	// })

	b.Run("new format", func(b *testing.B) {
		rows, err := store.LoadAlternateFormat(ctx, SeriesPointsOpts{SeriesID: &seriesId})
		if err != nil {
			b.Fatal(err)
		}
		newFormat = ToTimeseries(rows, seriesId)
	})

	b.Log(len(newFormat))

	// var oldFormat []SeriesPoint
	// opts := SeriesPointsOpts{SeriesID: &seriesId}
	// b.Run("old format", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		oldFormat, err = store.LoadSeriesInMem(ctx, opts)
	// 		if err != nil {
	// 			b.Fatal(err)
	// 		}
	// 		sort.Slice(oldFormat, func(i, j int) bool {
	// 			return oldFormat[i].Time.Before(oldFormat[j].Time)
	// 		})
	// 	}
	// })

	// b.Log(oldFormat)
	b.Log(newFormat)
}

type newRow struct {
	RepoId  int
	Smpls   []sample
	Capture *string
}

func TestConvert(t *testing.T) {
	ctx := context.Background()
	dsn := `postgres://sourcegraph:sourcegraph@localhost:5432/sourcegraph`
	handle, err := connections.EnsureNewCodeInsightsDB(dsn, "app", &observation.TestContext)
	if err != nil {
		t.Fatal(err)
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, handle)
	permStore := NewInsightPermissionStore(db)
	store := New(edb.NewInsightsDB(handle), permStore)

	insightStore := NewInsightStore(edb.NewInsightsDB(handle))
	dataSeries, err := insightStore.GetDataSeries(ctx, GetDataSeriesArgs{})
	if err != nil {
		t.Fatal(err)
	}

	cvtr := NewConverter(store)

	for _, series := range dataSeries {
		t.Log(series.SeriesID)
		if series.DataFormat == storage.Uncompressed {
			err = cvtr.Convert(ctx, series, storage.Gorilla)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("completed series:%s %d", series.SeriesID, series.ID)
		}
	}
}

func TestAppend(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	clock := timeutil.Now
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t))
	postgres := database.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	store := NewWithClock(insightsDB, permStore, clock)

	start := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	tsk := TimeSeriesKey{
		SeriesId: 1,
		RepoId:   1,
	}
	err := store.Append(ctx, tsk, generateSamples(start, 5))
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.LoadAlternateFormat(ctx, SeriesPointsOpts{Key: &tsk})
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
