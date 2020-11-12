package resolvers

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func init() {
	dbtesting.DBNameSuffix = "codemonitorsdb"
}

func TestCreateCodeMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	username := "code-monitors-resolver-user"
	userID := insertTestUser(t, dbconn.Global, username, true)
	_, err := db.Orgs.Create(ctx, "test-org", nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	r := newResolverWithClock(dbconn.Global, clock)

	want := &monitor{
		id:              1,
		createdBy:       userID,
		createdAt:       clock(),
		changedBy:       userID,
		changedAt:       clock(),
		description:     "banana",
		enabled:         true,
		namespaceUserID: &userID,
		namespaceOrgID:  nil,
	}

	// Create a monitor
	ctx = actor.WithActor(ctx, actor.FromUser(userID))
	ns := relay.MarshalID("User", userID)
	got, err := r.CreateCodeMonitor(ctx, &graphqlbackend.CreateCodeMonitorArgs{
		Namespace:   ns,
		Description: "banana",
		Enabled:     true,
		Trigger:     &graphqlbackend.CreateTriggerArgs{Query: "repo:foo"},
		Actions: []*graphqlbackend.CreateActionArgs{
			{Email: &graphqlbackend.CreateActionEmailArgs{
				Enabled:    false,
				Priority:   "NORMAL",
				Recipients: []graphql.ID{ns},
				Header:     "test header",
			}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got.(*monitor)) {
		t.Fatalf("\ngot:\t %+v,\nwant:\t %+v", got, want)
	}
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
