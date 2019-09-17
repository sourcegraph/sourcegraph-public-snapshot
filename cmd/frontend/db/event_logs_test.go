package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
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
