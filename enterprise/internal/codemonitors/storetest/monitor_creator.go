package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

const (
	testQuery       = "repo:github\\.com/sourcegraph/sourcegraph func type:diff patternType:literal"
	testDescription = "test description"
)

type TestStore struct {
	codemonitors.CodeMonitorStore
}

func (s *TestStore) InsertTestMonitor(ctx context.Context, t *testing.T) (*codemonitors.Monitor, error) {
	t.Helper()

	actions := []*codemonitors.EmailActionArgs{
		{
			Enabled:  true,
			Priority: "NORMAL",
			Header:   "test header 1",
		},
		{
			Enabled:  true,
			Priority: "CRITICAL",
			Header:   "test header 2",
		},
	}

	// Create monitor.
	uid := actor.FromContext(ctx).UID
	m, err := s.CreateMonitor(ctx, codemonitors.MonitorArgs{
		Description:     testDescription,
		Enabled:         true,
		NamespaceUserID: &uid,
	})
	if err != nil {
		return nil, err
	}

	// Create trigger.
	err = s.CreateQueryTrigger(ctx, m.ID, testQuery)
	if err != nil {
		return nil, err
	}

	for _, a := range actions {
		e, err := s.CreateEmailAction(ctx, m.ID, &codemonitors.EmailActionArgs{
			Enabled:  a.Enabled,
			Priority: a.Priority,
			Header:   a.Header,
		})
		if err != nil {
			return nil, err
		}

		err = s.CreateRecipient(ctx, e.ID, &uid, nil)
		if err != nil {
			return nil, err
		}
		// TODO(camdencheek): add other action types (webhooks) here
	}
	return m, nil
}

func NewTestStore(t *testing.T) (context.Context, *TestStore) {
	ctx := actor.WithInternalActor(context.Background())
	db := dbtesting.GetDB(t)
	now := time.Now().Truncate(time.Microsecond)
	return ctx, &TestStore{codemonitors.NewStoreWithClock(db, func() time.Time { return now })}
}

func NewTestUser(ctx context.Context, t *testing.T, db dbutil.DB) (name string, id int32, namespace graphql.ID, userContext context.Context) {
	t.Helper()

	name = "cm-user1"
	id = insertTestUser(t, db, name, true)
	namespace = relay.MarshalID("User", id)
	ctx = actor.WithActor(ctx, actor.FromUser(id))
	return name, id, namespace, ctx
}

func insertTestUser(t *testing.T, db dbutil.DB, name string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", name, isAdmin)
	err := db.QueryRowContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}
	return userID
}
