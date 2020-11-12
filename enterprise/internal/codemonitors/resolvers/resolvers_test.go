package resolvers

import (
	"context"
	"database/sql"
	"fmt"
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
	r := newTestResolver(t)

	userID := insertTestUser(t, dbconn.Global, "cm-user1", true)

	want := &monitor{
		id:              1,
		createdBy:       userID,
		createdAt:       r.clock(),
		changedBy:       userID,
		changedAt:       r.clock(),
		description:     "test monitor",
		enabled:         true,
		namespaceUserID: &userID,
		namespaceOrgID:  nil,
	}

	// Create a monitor
	ctx = actor.WithActor(ctx, actor.FromUser(userID))
	ns := relay.MarshalID("User", userID)
	got, err := r.insertTestMonitor(ctx, t, ns)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got.(*monitor)) {
		t.Fatalf("\ngot:\t %+v,\nwant:\t %+v", got, want)
	}

	// Toggle field enabled from true to false
	got, err = r.ToggleCodeMonitor(ctx, &graphqlbackend.ToggleCodeMonitorArgs{
		Id:      relay.MarshalID(monitorKind, got.(*monitor).id),
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.(*monitor).enabled {
		t.Fatalf("got enabled=%T, want enabled=%T", got.(*monitor).enabled, false)
	}

	// Delete code monitor
	_, err = r.DeleteCodeMonitor(ctx, &graphqlbackend.DeleteCodeMonitorArgs{Id: got.ID()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.monitorForIDInt32(ctx, t, got.(*monitor).id)
	if err == nil {
		t.Fatalf("monitor should have been deleted")
	}
}

func TestIsAllowedToEdit(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	dbtesting.SetupGlobalTestDB(t)

	// Setup users and org
	member := insertTestUser(t, dbconn.Global, "cm-user1", false)
	notMember := insertTestUser(t, dbconn.Global, "cm-user2", false)
	siteAdmin := insertTestUser(t, dbconn.Global, "cm-user3", true)

	admContext := actor.WithActor(context.Background(), actor.FromUser(siteAdmin))
	org, err := db.Orgs.Create(admContext, "cm-test-org", nil)
	if err != nil {
		t.Fatal(err)
	}
	addUserToOrg(t, dbconn.Global, member, org.ID)

	r := newTestResolver(t)

	// Create a monitor and set org as owner.
	ns := relay.MarshalID("Org", org.ID)
	m, err := r.insertTestMonitor(admContext, t, ns)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		user    int32
		allowed bool
	}{
		{
			user:    member,
			allowed: true,
		},
		{
			user:    notMember,
			allowed: false,
		},
		{
			user:    siteAdmin,
			allowed: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("user %d", tt.user), func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), actor.FromUser(tt.user))
			if err := r.isAllowedToEdit(ctx, m.ID()); (err != nil) == tt.allowed {
				t.Fatalf("unexpected permissions for user %d", tt.user)
			}
		})
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
			}}},
	})
}

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
