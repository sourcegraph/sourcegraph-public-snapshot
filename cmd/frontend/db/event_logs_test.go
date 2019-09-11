package db

import (
	"context"
	"errors"
	"strings"
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
		err       string // Error substr, empty string implies error is nil
	}{
		{"EmptyName", &UserEvent{UserID: 1}, `violates check constraint "event_logs_check_name_not_empty"`},
		{"InvalidUser", &UserEvent{Name: "test_event"}, `violates check constraint "event_logs_check_has_user"`},

		{"ValidInsert", &UserEvent{Name: "test_event", UserID: 1}, ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := EventLogs.Insert(ctx, tc.userEvent)

			// Should have no error
			if tc.err == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Errorf("got %+v, want %+v", err, errors.New(tc.err))
			}
		})
	}
}
