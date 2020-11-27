package resolvers

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
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

type Option interface {
	apply(*options)
}

type hook func() error

type options struct {
	actions   []*graphqlbackend.CreateActionArgs
	owner     graphql.ID
	postHooks []hook
}

type actionOption struct {
	actions []*graphqlbackend.CreateActionArgs
}

func (a actionOption) apply(opts *options) {
	opts.actions = a.actions
}

func WithActions(actions []*graphqlbackend.CreateActionArgs) Option {
	return actionOption{actions: actions}
}

type ownerOption struct {
	owner graphql.ID
}

func (o ownerOption) apply(opts *options) {
	opts.owner = o.owner
}

func WithOwner(owner graphql.ID) Option {
	return ownerOption{owner: owner}
}

type postHookOption struct {
	hooks []hook
}

func (h postHookOption) apply(opts *options) {
	opts.postHooks = h.hooks
}

func WithPostHooks(hooks []hook) Option {
	return postHookOption{hooks: hooks}
}

// insertTestMonitorWithOpts is a test helper that creates monitors for test
// purposes with sensible defaults. You can override the defaults by providing
// (optional) opts.
func (r *Resolver) insertTestMonitorWithOpts(ctx context.Context, t *testing.T, opts ...Option) (graphqlbackend.MonitorResolver, error) {
	t.Helper()

	defaultOwner := relay.MarshalID("User", actor.FromContext(ctx).UID)
	defaultAction := []*graphqlbackend.CreateActionArgs{
		{Email: &graphqlbackend.CreateActionEmailArgs{
			Enabled:    true,
			Priority:   "NORMAL",
			Recipients: []graphql.ID{defaultOwner},
			Header:     "test header"}}}

	options := options{
		actions:   defaultAction,
		owner:     defaultOwner,
		postHooks: nil,
	}
	for _, opt := range opts {
		opt.apply(&options)
	}
	m, err := r.CreateCodeMonitor(ctx, &graphqlbackend.CreateCodeMonitorArgs{
		Monitor: &graphqlbackend.CreateMonitorArgs{
			Namespace:   options.owner,
			Description: "test monitor",
			Enabled:     true,
		},
		Trigger: &graphqlbackend.CreateTriggerArgs{Query: "repo:foo"},
		Actions: options.actions,
	})
	if err != nil {
		return nil, err
	}
	for _, h := range options.postHooks {
		err = h()
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// newTestResolver returns a Resolver with stopped clock, which is useful to
// compare input and outputs in tests.
func newTestResolver(t *testing.T) *Resolver {
	t.Helper()

	now := timeutil.Now()
	clock := func() time.Time { return now }
	return newResolverWithClock(dbconn.Global, clock).(*Resolver)
}

func (r *Resolver) monitorForIDInt32(ctx context.Context, t *testing.T, monitorID int64) (graphqlbackend.MonitorResolver, error) {
	t.Helper()

	q := sqlf.Sprintf("SELECT id, created_by, created_at, changed_by, changed_at, description, enabled, namespace_user_id, namespace_org_id FROM cm_monitors WHERE id = %s", monitorID)
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
