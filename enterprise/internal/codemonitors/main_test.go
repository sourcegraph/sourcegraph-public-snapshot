package codemonitors

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

const (
	testQuery       = "repo:github\\.com/sourcegraph/sourcegraph func type:diff patternType:literal"
	testDescription = "test description"
)

func (s *codeMonitorStore) insertTestMonitor(ctx context.Context, t *testing.T) (*Monitor, error) {
	t.Helper()

	owner := relay.MarshalID("User", actor.FromContext(ctx).UID)
	args := &graphqlbackend.CreateCodeMonitorArgs{
		Monitor: &graphqlbackend.CreateMonitorArgs{
			Namespace:   owner,
			Description: testDescription,
			Enabled:     true,
		},
		Trigger: &graphqlbackend.CreateTriggerArgs{
			Query: testQuery,
		},
		Actions: []*graphqlbackend.CreateActionArgs{
			{
				Email: &graphqlbackend.CreateActionEmailArgs{
					Enabled:    true,
					Priority:   "NORMAL",
					Recipients: []graphql.ID{owner},
					Header:     "test header 1"},
			},
			{
				Email: &graphqlbackend.CreateActionEmailArgs{
					Enabled:    true,
					Priority:   "CRITICAL",
					Recipients: []graphql.ID{owner},
					Header:     "test header 2"},
			},
		},
	}
	// Create monitor.
	m, err := s.CreateMonitor(ctx, args.Monitor)
	if err != nil {
		return nil, err
	}

	// Create trigger.
	err = s.CreateQueryTrigger(ctx, m.ID, args.Trigger.Query)
	if err != nil {
		return nil, err
	}

	for _, a := range args.Actions {
		if a.Email != nil {
			e, err := s.CreateEmailAction(ctx, m.ID, &EmailActionArgs{
				Enabled:  a.Email.Enabled,
				Priority: a.Email.Priority,
				Header:   a.Email.Header,
			})
			if err != nil {
				return nil, err
			}

			for _, recipient := range a.Email.Recipients {
				var userID, orgID int32
				if err := graphqlbackend.UnmarshalNamespaceID(recipient, &userID, &orgID); err != nil {
					return nil, err
				}

				err := s.CreateRecipient(ctx, e.ID, nilOrInt32(userID), nilOrInt32(orgID))
				if err != nil {
					return nil, err
				}
			}
		}
		// TODO(camdencheek): add other action types (webhooks) here
	}
	return m, nil
}

func newTestStore(t *testing.T) (context.Context, *codeMonitorStore) {
	ctx := actor.WithInternalActor(context.Background())
	db := dbtesting.GetDB(t)
	now := time.Now().Truncate(time.Microsecond)
	return ctx, NewStoreWithClock(db, func() time.Time { return now })
}

func newTestUser(ctx context.Context, t *testing.T) (name string, id int32, namespace graphql.ID, userContext context.Context) {
	t.Helper()

	name = "cm-user1"
	id = insertTestUser(t, dbconn.Global, name, true)
	namespace = relay.MarshalID("User", id)
	ctx = actor.WithActor(ctx, actor.FromUser(id))
	return name, id, namespace, ctx
}

func insertTestUser(t *testing.T, db *sql.DB, name string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", name, isAdmin)
	err := db.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}
	return userID
}
