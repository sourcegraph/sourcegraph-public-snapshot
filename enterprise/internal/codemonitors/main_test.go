package codemonitors

import (
	"context"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

const (
	testQuery       = "repo:github\\.com/sourcegraph/sourcegraph func type:diff patternType:literal"
	testDescription = "test description"
)

type testFixtures struct {
	monitor    *Monitor
	query      *QueryTrigger
	emails     [2]*EmailAction
	recipients [2]*Recipient
}

func (s *codeMonitorStore) insertTestMonitor(ctx context.Context, t *testing.T) (*testFixtures, error) {
	t.Helper()

	fixtures := testFixtures{}

	actions := []*EmailActionArgs{
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
	var err error
	fixtures.monitor, err = s.CreateMonitor(ctx, MonitorArgs{
		Description:     testDescription,
		Enabled:         true,
		NamespaceUserID: &uid,
	})
	require.NoError(t, err)

	// Create trigger.
	fixtures.query, err = s.CreateQueryTrigger(ctx, fixtures.monitor.ID, testQuery)
	require.NoError(t, err)

	for i, a := range actions {
		fixtures.emails[i], err = s.CreateEmailAction(ctx, fixtures.monitor.ID, &EmailActionArgs{
			Enabled:  a.Enabled,
			Priority: a.Priority,
			Header:   a.Header,
		})
		require.NoError(t, err)

		fixtures.recipients[i], err = s.CreateRecipient(ctx, fixtures.emails[i].ID, &uid, nil)
		require.NoError(t, err)
		// TODO(camdencheek): add other action types (webhooks) here
	}
	return &fixtures, nil
}

func newTestStore(t *testing.T) (context.Context, dbutil.DB, *codeMonitorStore) {
	ctx := actor.WithInternalActor(context.Background())
	db := dbtest.NewDB(t)
	now := time.Now().Truncate(time.Microsecond)
	return ctx, db, NewStoreWithClock(db, func() time.Time { return now })
}

func newTestUser(ctx context.Context, t *testing.T, db dbutil.DB) (name string, id int32, namespace graphql.ID, userContext context.Context) {
	t.Helper()

	name = "cm-user1"
	id = insertTestUser(ctx, t, db, name, true)
	namespace = relay.MarshalID("User", id)
	ctx = actor.WithActor(ctx, actor.FromUser(id))
	return name, id, namespace, ctx
}

func insertTestUser(ctx context.Context, t *testing.T, db dbutil.DB, name string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", name, isAdmin)
	err := db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID)
	require.NoError(t, err)
	return userID
}
