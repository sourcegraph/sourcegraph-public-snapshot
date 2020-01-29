package db

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestEventLogs_ValidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	var testCases = []struct {
		name  string
		event *Event
		err   string // Stringified error
	}{
		{
			name:  "EmptyName",
			event: &Event{UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   `INSERT: pq: new row for relation "event_logs" violates check constraint "event_logs_check_name_not_empty"`,
		},
		{
			name:  "EmptyURL",
			event: &Event{Name: "test_event", UserID: 1, Source: "WEB"},
			err:   `INSERT: pq: new row for relation "event_logs" violates check constraint "event_logs_check_url_not_empty"`,
		},
		{
			name:  "InvalidUser",
			event: &Event{Name: "test_event", URL: "http://sourcegraph.com", Source: "WEB"},
			err:   `INSERT: pq: new row for relation "event_logs" violates check constraint "event_logs_check_has_user"`,
		},
		{
			name:  "EmptySource",
			event: &Event{Name: "test_event", URL: "http://sourcegraph.com", UserID: 1},
			err:   `INSERT: pq: new row for relation "event_logs" violates check constraint "event_logs_check_source_not_empty"`,
		},

		{
			name:  "ValidInsert",
			event: &Event{Name: "test_event", UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   "<nil>",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := EventLogs.Insert(ctx, tc.event)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have %+v, want %+v", have, want)
			}
		})
	}
}

func TestEventLogs_CountUniqueUsersPerPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	startDate, err := startOfPeriod(time.Now(), Daily, 2)
	if err != nil {
		t.Fatal(err)
	}

	events := []*Event{
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 2, Source: "WEB", Timestamp: startDate.Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 2, Source: "WEB", Timestamp: startDate.Add(time.Hour * 3)},

		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24)},
		{Name: "foo", URL: "test", UserID: 2, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 3, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour * 3)},

		{Name: "foo", URL: "test", UserID: 5, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2)},
		{Name: "foo", URL: "test", UserID: 6, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 7, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 8, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour * 3)},
	}

	for _, e := range events {
		if err := EventLogs.Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	values, err := EventLogs.CountUniqueUsersPerPeriod(ctx, Daily, startDate, 2, nil)
	if err != nil {
		t.Fatal(err)
	}

	assertUsageValue(t, values[0], startDate.Add(time.Hour*24*2), 4)
	assertUsageValue(t, values[1], startDate.Add(time.Hour*24), 3)
	assertUsageValue(t, values[2], startDate, 2)
}

func TestEventLogs_CountEventsPerPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	startDate, err := startOfPeriod(time.Now(), Daily, 2)
	if err != nil {
		t.Fatal(err)
	}

	events := []*Event{
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 3)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 4)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 5)},

		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour * 3)},

		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 1, Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour * 2)},
	}

	for _, e := range events {
		if err := EventLogs.Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	values, err := EventLogs.CountEventsPerPeriod(ctx, Daily, startDate, 2, nil)
	if err != nil {
		t.Fatal(err)
	}

	assertUsageValue(t, values[0], startDate.Add(time.Hour*24*2), 3)
	assertUsageValue(t, values[1], startDate.Add(time.Hour*24), 4)
	assertUsageValue(t, values[2], startDate, 6)
}

func TestEventLogs_PercentilesPerPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	startDate, err := startOfPeriod(time.Now(), Daily, 2)
	if err != nil {
		t.Fatal(err)
	}

	events := []*Event{
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 10}`), Source: "WEB", Timestamp: startDate},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 20}`), Source: "WEB", Timestamp: startDate.Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 30}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 40}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 3)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 50}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 4)},

		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 20}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 30}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 40}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 50}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour * 3)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 60}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24).Add(time.Hour * 4)},

		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 30}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 40}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 50}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour * 2)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 60}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour * 3)},
		{Name: "foo", URL: "test", UserID: 1, Argument: json.RawMessage(`{"durationMs": 70}`), Source: "WEB", Timestamp: startDate.Add(time.Hour * 24 * 2).Add(time.Hour * 4)},
	}

	for _, e := range events {
		if err := EventLogs.Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	values, err := EventLogs.PercentilesPerPeriod(ctx, Daily, startDate, 2, "durationMs", []float64{0.5, 0.8}, nil)
	if err != nil {
		t.Fatal(err)
	}

	assertPercentileValue(t, values[0], startDate.Add(time.Hour*24*2), []float64{50, 62})
	assertPercentileValue(t, values[1], startDate.Add(time.Hour*24), []float64{40, 52})
	assertPercentileValue(t, values[2], startDate, []float64{30, 42})
}

func assertUsageValue(t *testing.T, v UsageValue, start time.Time, count int) {
	if v.Start != start {
		t.Errorf("got Start %q, want %q", v.Start, start)
	}
	if v.Count != count {
		t.Errorf("got Count %d, want %d", v.Count, count)
	}
}

func assertPercentileValue(t *testing.T, v PercentileValue, start time.Time, values []float64) {
	if v.Start != start {
		t.Errorf("got Start %q, want %q", v.Start, start)
	}

	for i, value := range v.Values {
		if value != values[i] {
			t.Errorf("got Values[%d] %f, want %f", i, value, values[i])
		}
	}
}

//
// Temporary

func startOfPeriod(now time.Time, periodType PeriodType, periodsAgo int) (time.Time, error) {
	switch periodType {
	case Daily:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -periodsAgo), nil
	case Weekly:
		return startOfWeek(now, periodsAgo), nil
	case Monthly:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -periodsAgo, 0), nil
	}
	return time.Time{}, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
}

func startOfWeek(now time.Time, weeksAgo int) time.Time {
	if weeksAgo > 0 {
		return startOfWeek(now, 0).AddDate(0, 0, -7*weeksAgo)
	}

	// If weeksAgo == 0, start at timeNow(), and loop back by day until we hit a Sunday
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}
