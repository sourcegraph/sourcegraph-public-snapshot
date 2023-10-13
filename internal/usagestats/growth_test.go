package usagestats

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGrowthStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	defer func() {
		timeNow = time.Now
	}()

	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	t.Run("getUsersGrowthStatistics", func(t *testing.T) {
		createUsersQuery := `
			INSERT INTO users (id, username, created_at, deleted_at)
			VALUES
				-- created user
				(1, 'u1', $1::timestamp - interval '1 day', NULL),
				-- deleted user
				(2, 'u2', $1::timestamp - interval '1 months', $1::timestamp - interval '1 days'),
				-- retained user
				(3, 'u3', $1::timestamp - interval '1 months', NULL),
				-- resurrected user
				(4, 'u4', $1::timestamp - interval '1 months', NULL),
				-- churned user
				(5, 'u5', $1::timestamp - interval '1 months', NULL),
				-- not used in stats
				(6, 'u6', $1::timestamp - interval '2 months', NULL)`
		if _, err := db.ExecContext(context.Background(), createUsersQuery, now); err != nil {
			t.Fatal(err)
		}

		createEventLogsQuery := `
			INSERT INTO event_logs (user_id, name, argument, url, anonymous_user_id, source, version, timestamp)
			VALUES
				-- retained user
				(3, 'SomeEvent', '{}', 'https://sourcegraph.test:3443/search', '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 months'),
				(3, 'SomeEvent', '{}', 'https://sourcegraph.test:3443/search', '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
				-- resurrected user
				(4, 'SomeEvent', '{}', 'https://sourcegraph.test:3443/search', '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
				-- churned user
				(5, 'SomeEvent', '{}', 'https://sourcegraph.test:3443/search', '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 months'),
				-- not used in stats
				(6, 'SomeEvent', '{}', 'https://sourcegraph.test:3443/search', '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '2 months')`
		if _, err := db.ExecContext(context.Background(), createEventLogsQuery, now); err != nil {
			t.Fatal(err)
		}

		actual, err := getUsersGrowthStatistics(ctx, db)
		if err != nil {
			t.Fatal(err)
		}

		expected := &usersGrowthStatistics{
			createdUsers:     1,
			deletedUsers:     1,
			retainedUsers:    1,
			resurrectedUsers: 1,
			churnedUsers:     1,
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("getAccessRequestsGrowthStatistics", func(t *testing.T) {
		createAccessRequestsQuery := `
			INSERT INTO access_requests
				(id, created_at, updated_at, name, email, status)
			VALUES
				(1, $1::timestamp - interval '1 day', $1::timestamp - interval '1 day', 'a1', 'a1@example.com', 'PENDING'),
				(2, $1::timestamp - interval '1 days', $1::timestamp - interval '1 days', 'a2', 'a2@example.com', 'APPROVED'),
				(3, $1::timestamp - interval '1 days', $1::timestamp - interval '1 days', 'a3', 'a3@example.com', 'REJECTED'),
				(4, $1::timestamp - interval '1 months', $1::timestamp - interval '1 months', 'a4', 'a4@example.cmo', 'PENDING')`
		if _, err := db.ExecContext(context.Background(), createAccessRequestsQuery, now); err != nil {
			t.Fatal(err)
		}

		actual, err := getAccessRequestsGrowthStatistics(ctx, db)
		if err != nil {
			t.Fatal(err)
		}

		expected := &accessRequestsGrowthStatistics{
			pendingAccessRequests:  1,
			approvedAccessRequests: 1,
			rejectedAccessRequests: 1,
		}
		assert.Equal(t, expected, actual)
	})
}
