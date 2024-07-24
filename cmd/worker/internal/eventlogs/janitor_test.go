package eventlogs

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDeleteOldEventLogsInPostgres(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	err := db.EventLogs().BulkInsert(
		context.Background(),
		[]*database.Event{
			{
				UserID:    1,
				Name:      "old",
				Source:    "test",
				Timestamp: time.Now().Add(-94 * 24 * time.Hour),
			},
			{
				UserID:    1,
				Name:      "new",
				Source:    "test",
				Timestamp: time.Now(),
			},
		},
	)
	require.NoError(t, err)

	err = deleteOldEventLogsInPostgres(ctx, db)
	require.NoError(t, err)

	got, err := db.EventLogs().ListAll(ctx, database.EventLogsListOptions{})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "new", got[0].Name)
}

func TestDeleteOldSecurityEventLogsInPostgres(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			Log: &schema.Log{SecurityEventLog: &schema.SecurityEventLog{Location: "database"}},
		},
	})
	defer conf.Mock(nil)

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	err := db.SecurityEventLogs().InsertList(ctx, []*database.SecurityEvent{
		{
			Name:      "old",
			UserID:    1,
			Source:    "test",
			Timestamp: time.Now().Add(-31 * 24 * time.Hour),
		},
		{
			Name:      "new",
			UserID:    1,
			Source:    "test",
			Timestamp: time.Now(),
		},
	})
	require.NoError(t, err)

	assertSecurityEventCount(t, db, 2)
	err = deleteOldSecurityEventLogsInPostgres(ctx, db)
	require.NoError(t, err)
	assertSecurityEventCount(t, db, 1)
}

func assertSecurityEventCount(t *testing.T, db database.DB, expectedCount int) {
	t.Helper()

	row := db.SecurityEventLogs().Handle().QueryRowContext(context.Background(), "SELECT count(*) FROM security_event_logs")
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatal("couldn't read security events count")
	}
	require.Equal(t, expectedCount, count)
}
