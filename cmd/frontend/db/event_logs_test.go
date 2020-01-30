package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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

	now := time.Now()
	startDate, _ := calcStartDate(now, Daily, 3)
	secondDay := startDate.Add(time.Hour * 24)
	thirdDay := startDate.Add(time.Hour * 24 * 2)

	events := []*Event{
		makeTestEvent(&Event{UserID: 1, Timestamp: startDate}),
		makeTestEvent(&Event{UserID: 1, Timestamp: startDate}),
		makeTestEvent(&Event{UserID: 2, Timestamp: startDate}),
		makeTestEvent(&Event{UserID: 2, Timestamp: startDate}),

		makeTestEvent(&Event{UserID: 1, Timestamp: secondDay}),
		makeTestEvent(&Event{UserID: 2, Timestamp: secondDay}),
		makeTestEvent(&Event{UserID: 3, Timestamp: secondDay}),
		makeTestEvent(&Event{UserID: 1, Timestamp: secondDay}),

		makeTestEvent(&Event{UserID: 5, Timestamp: thirdDay}),
		makeTestEvent(&Event{UserID: 6, Timestamp: thirdDay}),
		makeTestEvent(&Event{UserID: 7, Timestamp: thirdDay}),
		makeTestEvent(&Event{UserID: 8, Timestamp: thirdDay}),
	}

	for _, e := range events {
		if err := EventLogs.Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	values, err := EventLogs.CountUniqueUsersPerPeriod(ctx, Daily, now, 3, nil)
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

	now := time.Now()
	startDate, _ := calcStartDate(now, Daily, 3)
	secondDay := startDate.Add(time.Hour * 24)
	thirdDay := startDate.Add(time.Hour * 24 * 2)

	events := []*Event{
		makeTestEvent(&Event{Timestamp: startDate}),
		makeTestEvent(&Event{Timestamp: startDate}),
		makeTestEvent(&Event{Timestamp: startDate}),
		makeTestEvent(&Event{Timestamp: startDate}),
		makeTestEvent(&Event{Timestamp: startDate}),
		makeTestEvent(&Event{Timestamp: startDate}),

		makeTestEvent(&Event{Timestamp: secondDay}),
		makeTestEvent(&Event{Timestamp: secondDay}),
		makeTestEvent(&Event{Timestamp: secondDay}),
		makeTestEvent(&Event{Timestamp: secondDay}),

		makeTestEvent(&Event{Timestamp: thirdDay}),
		makeTestEvent(&Event{Timestamp: thirdDay}),
		makeTestEvent(&Event{Timestamp: thirdDay}),
	}

	for _, e := range events {
		if err := EventLogs.Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	values, err := EventLogs.CountEventsPerPeriod(ctx, Daily, now, 3, nil)
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

	now := time.Now()
	startDate, _ := calcStartDate(now, Daily, 3)
	secondDay := startDate.Add(time.Hour * 24)
	thirdDay := startDate.Add(time.Hour * 24 * 2)

	events := []*Event{
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 10}`), Timestamp: startDate}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 20}`), Timestamp: startDate}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 30}`), Timestamp: startDate}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 40}`), Timestamp: startDate}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 50}`), Timestamp: startDate}),

		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 20}`), Timestamp: secondDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 30}`), Timestamp: secondDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 40}`), Timestamp: secondDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 50}`), Timestamp: secondDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 60}`), Timestamp: secondDay}),

		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 30}`), Timestamp: thirdDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 40}`), Timestamp: thirdDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 50}`), Timestamp: thirdDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 60}`), Timestamp: thirdDay}),
		makeTestEvent(&Event{Argument: json.RawMessage(`{"durationMs": 70}`), Timestamp: thirdDay}),
	}

	for _, e := range events {
		if err := EventLogs.Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	values, err := EventLogs.PercentilesPerPeriod(ctx, Daily, now, 3, "durationMs", []float64{0.5, 0.8}, nil)
	if err != nil {
		t.Fatal(err)
	}

	assertPercentileValue(t, values[0], startDate.Add(time.Hour*24*2), []float64{50, 62})
	assertPercentileValue(t, values[1], startDate.Add(time.Hour*24), []float64{40, 52})
	assertPercentileValue(t, values[2], startDate, []float64{30, 42})
}

// makeTestEvent sets the required (uninteresting) fields that are required on insertion
// due to db constraints. This method will also add some sub-day jitter to the timestamp.
func makeTestEvent(e *Event) *Event {
	if e.UserID == 0 {
		e.UserID = 1
	}
	e.Name = "foo"
	e.URL = "test"
	e.Source = "WEB"
	e.Timestamp = e.Timestamp.Add(time.Minute * time.Duration(rand.Intn(60*12)))
	return e
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
