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
		name      string
		userEvent *UserEvent
		err       string // Stringified error
	}{
		{"EmptyName", &UserEvent{UserID: 1}, `INSERT: pq: new row for relation "event_logs" violates check constraint "event_logs_check_name_not_empty"`},
		{"InvalidUser", &UserEvent{Name: "test_event"}, `INSERT: pq: new row for relation "event_logs" violates check constraint "event_logs_check_has_user"`},

		{"ValidInsert", &UserEvent{Name: "test_event", UserID: 1}, "<nil>"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := EventLogs.Insert(ctx, tc.userEvent)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have %+v, want %+v", have, want)
			}
		})
	}
}
