package resolvers

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func insertTestUser(t *testing.T, db *sql.DB, name string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", name, isAdmin)

	err := db.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func addUserToOrg(t *testing.T, db *sql.DB, userID int32, orgID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO org_members (org_id, user_id) VALUES (%s, %s)", orgID, userID)

	_, err := db.Exec(q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatal(err)
	}
}

func (r *Resolver) insertTestMonitor(ctx context.Context, t *testing.T, owner graphql.ID) (graphqlbackend.MonitorResolver, error) {
	t.Helper()

	return r.CreateCodeMonitor(ctx, &graphqlbackend.CreateCodeMonitorArgs{
		Namespace:   owner,
		Description: "test monitor",
		Enabled:     true,
		Trigger:     &graphqlbackend.CreateTriggerArgs{Query: "repo:foo"},
		Actions: []*graphqlbackend.CreateActionArgs{
			{Email: &graphqlbackend.CreateActionEmailArgs{
				Enabled:    true,
				Priority:   "NORMAL",
				Recipients: []graphql.ID{owner},
				Header:     "test header",
			}},
		},
	})
}

// newTestResolver returns a Resolver with stopped clock, which is useful to
// compare input and outputs in tests.
func newTestResolver(t *testing.T) *Resolver {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	return newResolverWithClock(dbconn.Global, clock).(*Resolver)
}

func (r *Resolver) monitorForIDInt32(ctx context.Context, t *testing.T, monitorId int64) (graphqlbackend.MonitorResolver, error) {
	t.Helper()

	q := sqlf.Sprintf("SELECT id, created_by, created_at, changed_by, changed_at, description, enabled, namespace_user_id, namespace_org_id FROM cm_monitors WHERE id = %s", monitorId)
	return r.runMonitorQuery(ctx, q)
}

func marshalDateTime(t testing.TB, ts time.Time) string {
	t.Helper()

	dt := graphqlbackend.DateTime{Time: ts}

	bs, err := dt.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	// Unquote the date time.
	return strings.ReplaceAll(string(bs), "\"", "")
}
