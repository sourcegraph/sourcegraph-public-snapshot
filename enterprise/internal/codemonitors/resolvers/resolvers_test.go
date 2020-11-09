package resolvers

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

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

func TestCodeMonitorResolver(t *testing.T) {
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
	r := Resolver{
		db:    dbconn.Global,
		clock: clock,
	}

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
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got.(*monitor)) {
		t.Fatalf("\ngot %+v,\nwant %+v", got, want)
	}

	// Toggle the monitor to enabled=false
	got, err = r.ToggleCodeMonitor(ctx, &graphqlbackend.ToggleCodeMonitorArgs{
		Id:      got.ID(),
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled() != false {
		t.Fatalf("got %t, want %t", got.Enabled(), false)
	}

	want = &monitor{
		id:              1,
		createdBy:       userID,
		createdAt:       clock(),
		changedBy:       userID,
		changedAt:       clock(),
		description:     "apple",
		enabled:         true,
		namespaceUserID: &userID,
		namespaceOrgID:  nil,
	}

	// Edit the description of code monitor
	got, err = r.EditCodeMonitor(ctx, &graphqlbackend.EditCodeMonitorArgs{
		Id:          got.ID(),
		Enabled:     true,
		Description: "apple",
		Namespace:   ns,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got.(*monitor)) {
		t.Fatalf("\ngot %+v,\nwant %+v", got, want)
	}

	// Delete the monitor
	_, err = r.DeleteCodeMonitor(ctx, &graphqlbackend.DeleteCodeMonitorArgs{
		Id: got.ID(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.monitorForID(ctx, got.ID())
	if err == nil {
		t.Fatalf("monitor should have been deleted")
	}

	// Delete the same monitor again
	_, err = r.DeleteCodeMonitor(ctx, &graphqlbackend.DeleteCodeMonitorArgs{
		Id: got.ID(),
	})
	if err == nil {
		t.Fatalf("DeleteCodeMonitor should have returned an error")
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
