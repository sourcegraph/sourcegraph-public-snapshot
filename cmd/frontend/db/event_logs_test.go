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

	t.Run("EmptyName", func(t *testing.T) {
		err := EventLogs.Insert(ctx, &UserEvent{})
		if err == nil {
			t.Errorf("got %+v, want %+v", err, errors.New("empty event name"))
		} else if !strings.Contains(err.Error(), "empty event name") {
			t.Fatal(err)
		}
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		err := EventLogs.Insert(ctx, &UserEvent{
			Name: "test_event",
		})
		if err == nil {
			t.Errorf("got %+v, want %+v", err, errors.New("one of UserID or AnonymousUserID must have valid value"))
		} else if !strings.Contains(err.Error(), "must have valid value") {
			t.Fatal(err)
		}
	})

	t.Run("ValidInsert", func(t *testing.T) {
		err := EventLogs.Insert(ctx, &UserEvent{
			Name:   "test_event",
			UserID: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
