package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestSeriesPoints(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	clock := timeutil.Now
	timescale, cleanup := dbtesting.TimescaleDB(t)
	defer cleanup()
	store := NewWithClock(timescale, clock)

	// Confirm we get no results initially.
	points, err := store.SeriesPoints(ctx, SeriesPointsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("SeriesPoints", []SeriesPoint{}).Equal(t, points)

	// Insert some fake data.
	_, err = timescale.Exec(`
INSERT INTO repo_names(name) VALUES ('github.com/gorilla/mux-original');
INSERT INTO repo_names(name) VALUES ('github.com/gorilla/mux-renamed');
INSERT INTO metadata(metadata) VALUES ('{"hello": "world", "languages": ["Go", "Python", "Java"]}');
SELECT setseed(0.5);
INSERT INTO series_points(
    time,
	series_id,
    value,
    metadata_id,
    repo_id,
    repo_name_id,
    original_repo_name_id)
SELECT time,
    'somehash',
    random()*80 - 40,
    (SELECT id FROM metadata WHERE metadata = '{"hello": "world", "languages": ["Go", "Python", "Java"]}'),
    2,
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-renamed'),
    (SELECT id FROM repo_names WHERE name = 'github.com/gorilla/mux-original')
	FROM GENERATE_SERIES(CURRENT_TIMESTAMP::date - INTERVAL '6 months', CURRENT_TIMESTAMP::date, '2 weeks') AS time;
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
		autogold.Want("SeriesPoints(2).len", int(14)).Equal(t, len(points))
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

}

func TestCountData(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	clock := timeutil.Now
	timescale, cleanup := dbtesting.TimescaleDB(t)
	defer cleanup()
	store := NewWithClock(timescale, clock)

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
			SeriesID: "one",
			Point:    SeriesPoint{Time: timeValue("2020-03-01T00:00:00Z"), Value: 1.1},
			RepoName: optionalString("repo1"),
			RepoID:   optionalRepoID(3),
			Metadata: map[string]interface{}{"some": "data"},
		},
		{
			SeriesID: "two",
			Point:    SeriesPoint{Time: timeValue("2020-03-02T00:00:00Z"), Value: 2.2},
			Metadata: []interface{}{"some", "data", "two"},
		},
		{
			SeriesID: "two",
			Point:    SeriesPoint{Time: timeValue("2020-03-02T00:01:00Z"), Value: 2.2},
			Metadata: []interface{}{"some", "data", "two"},
		},
		{
			SeriesID: "three",
			Point:    SeriesPoint{Time: timeValue("2020-03-03T00:00:00Z"), Value: 3.3},
			Metadata: nil,
		},
		{
			SeriesID: "three",
			Point:    SeriesPoint{Time: timeValue("2020-03-03T00:01:00Z"), Value: 3.3},
			Metadata: nil,
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
	t.Parallel()

	ctx := context.Background()
	clock := timeutil.Now
	timescale, cleanup := dbtesting.TimescaleDB(t)
	defer cleanup()
	store := NewWithClock(timescale, clock)

	optionalString := func(v string) *string { return &v }
	optionalRepoID := func(v api.RepoID) *api.RepoID { return &v }

	current := time.Now().Truncate(24 * time.Hour)

	// Record points that will verify last-observation carried forward. The last point should roll over two time frames.
	// Metadata is currently not queried and will not resolve to reduce cardinality.
	for _, record := range []RecordSeriesPointArgs{
		{
			SeriesID: "one",
			Point:    SeriesPoint{Time: current, Value: 1.1},
			RepoName: optionalString("repo1"),
			RepoID:   optionalRepoID(3),
			Metadata: map[string]interface{}{"some": "data"},
		},
		{
			SeriesID: "one",
			Point:    SeriesPoint{Time: current.Add(-time.Hour * 24 * 15), Value: 2.2},
			RepoName: optionalString("repo1"),
			RepoID:   optionalRepoID(3),
			Metadata: []interface{}{"some", "data", "two"},
		},
		{
			SeriesID: "one",
			Point:    SeriesPoint{Time: current.Add(-time.Hour * 24 * 43), Value: 3.3},
			RepoName: optionalString("repo1"),
			RepoID:   optionalRepoID(3),
			Metadata: nil,
		},
	} {
		if err := store.RecordSeriesPoint(ctx, record); err != nil {
			t.Fatal(err)
		}
	}

	want := []SeriesPoint{
		{
			SeriesID: "one",
			Time:     current,
			Value:    1.1,
		},
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 14),
			Value:    2.2,
		},
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 28),
			Value:    3.3,
		},
		{
			SeriesID: "one",
			Time:     current.Add(-time.Hour * 24 * 42),
			Value:    3.3,
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

	// TODO: future: once querying by RepoName and/or OriginalRepoName is possible, test that here:
	// // Confirm querying by repo name works as expected.
	// forRepoNamePoints, err := store.SeriesPoints(ctx, SeriesPointsOpts{RepoName: optionalString("repo1")})
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// autogold.Want("len(forRepoNamePoints)", nil).Equal(t, len(forRepoNamePoints))
	// autogold.Want("forRepoNamePoints[0].String()", nil).Equal(t, forRepoNamePoints[0].String())
	//
	// // Confirm querying by original repo name works as expected.
	// forOriginalRepoNamePoints, err := store.SeriesPoints(ctx, SeriesPointsOpts{OriginalRepoName: optionalString("repo1")})
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// autogold.Want("len(forOriginalRepoNamePoints)", nil).Equal(t, len(forOriginalRepoNamePoints))
	// autogold.Want("forOriginalRepoNamePoints[0].String()", nil).Equal(t, forOriginalRepoNamePoints[0].String())
}
