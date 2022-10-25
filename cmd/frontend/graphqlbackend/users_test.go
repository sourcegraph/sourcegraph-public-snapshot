package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUsers(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.ListFunc.SetDefaultReturn([]*types.User{{Username: "user1"}, {Username: "user2"}}, nil)
	users.CountFunc.SetDefaultReturn(2, nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				{
					users {
						nodes { username }
						totalCount
					}
				}
			`,
			ExpectedResult: `
				{
					"users": {
						"nodes": [
							{
								"username": "user1"
							},
							{
								"username": "user2"
							}
						],
						"totalCount": 2
					}
				}
			`,
		},
	})
}

func TestUsers_InactiveSince(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	schema := mustParseGraphQLSchema(t, db)

	now := time.Now()
	daysAgo := func(days int) time.Time {
		return now.Add(-time.Duration(days) * 24 * time.Hour)
	}

	var users = []struct {
		user        database.NewUser
		lastEventAt time.Time
	}{
		{user: database.NewUser{Username: "user-1", Password: "user-1"}, lastEventAt: daysAgo(1)},
		{user: database.NewUser{Username: "user-2", Password: "user-2"}, lastEventAt: daysAgo(2)},
		{user: database.NewUser{Username: "user-3", Password: "user-3"}, lastEventAt: daysAgo(3)},
		{user: database.NewUser{Username: "user-4", Password: "user-4"}, lastEventAt: daysAgo(4)},
	}

	for _, newUser := range users {
		u, err := db.Users().Create(ctx, newUser.user)
		if err != nil {
			t.Fatal(err)
		}

		event := &database.Event{
			UserID:    uint32(u.ID),
			Timestamp: newUser.lastEventAt,
			Name:      "testevent",
			Source:    "test",
		}
		if err := db.EventLogs().Insert(ctx, event); err != nil {
			t.Fatal(err)
		}
	}

	ctx = actor.WithInternalActor(ctx)

	query := `
		query InactiveUsers($since: DateTime) {
			users(inactiveSince: $since) {
				nodes { username }
				totalCount
			}
		}
	`

	RunTests(t, []*Test{
		{
			Context:   ctx,
			Schema:    schema,
			Query:     query,
			Variables: map[string]any{"since": daysAgo(4).Format(time.RFC3339Nano)},
			ExpectedResult: `
			{"users": { "nodes": [], "totalCount": 0 }}
			`,
		},
		{
			Context:   ctx,
			Schema:    schema,
			Query:     query,
			Variables: map[string]any{"since": daysAgo(3).Format(time.RFC3339Nano)},
			ExpectedResult: `
			{"users": { "nodes": [{ "username": "user-4" }], "totalCount": 1 }}
			`,
		},
		{
			Context:   ctx,
			Schema:    schema,
			Query:     query,
			Variables: map[string]any{"since": daysAgo(2).Format(time.RFC3339Nano)},
			ExpectedResult: `
			{"users": { "nodes": [{ "username": "user-3" }, { "username": "user-4" }], "totalCount": 2 }}
			`,
		},
		{
			Context:   ctx,
			Schema:    schema,
			Query:     query,
			Variables: map[string]any{"since": daysAgo(1).Format(time.RFC3339Nano)},
			ExpectedResult: `
			{"users": { "nodes": [
				{ "username": "user-2" },
				{ "username": "user-3" },
				{ "username": "user-4" }
			], "totalCount": 3 }}
			`,
		},
		{
			Context:   ctx,
			Schema:    schema,
			Query:     query,
			Variables: map[string]any{"since": daysAgo(0).Format(time.RFC3339Nano)},
			ExpectedResult: `
			{"users": { "nodes": [
				{ "username": "user-1" },
				{ "username": "user-2" },
				{ "username": "user-3" },
				{ "username": "user-4" }
			], "totalCount": 4 }}
			`,
		},
	})
}
