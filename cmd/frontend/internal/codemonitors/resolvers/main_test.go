package resolvers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func insertTestUser(t *testing.T, db database.DB, name string, isAdmin bool) *types.User {
	t.Helper()

	u, err := db.Users().Create(context.Background(), database.NewUser{Username: name})
	require.NoError(t, err)

	err = db.Users().SetIsSiteAdmin(context.Background(), u.ID, isAdmin)
	require.NoError(t, err)

	return u
}

func addUserToOrg(t *testing.T, db database.DB, userID int32, orgID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO org_members (org_id, user_id) VALUES (%s, %s)", orgID, userID)

	_, err := db.ExecContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...)
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
	defaultActions := []*graphqlbackend.CreateActionArgs{
		{Email: &graphqlbackend.CreateActionEmailArgs{
			Enabled:    true,
			Priority:   "NORMAL",
			Recipients: []graphql.ID{defaultOwner},
			Header:     "test header"}},
	}

	options := options{
		actions:   defaultActions,
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
		Trigger: &graphqlbackend.CreateTriggerArgs{Query: "repo:foo type:commit"},
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
func newTestResolver(t *testing.T, db database.DB) *Resolver {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time { return now }
	return newResolverWithClock(logtest.Scoped(t), db, clock)
}

// newResolverWithClock is used in tests to set the clock manually.
func newResolverWithClock(logger log.Logger, db database.DB, clock func() time.Time) *Resolver {
	mockDB := dbmocks.NewMockDBFrom(db)
	mockDB.CodeMonitorsFunc.SetDefaultReturn(database.CodeMonitorsWithClock(db, clock))
	return &Resolver{logger: logger, db: mockDB}
}

func marshalDateTime(t testing.TB, ts time.Time) string {
	t.Helper()

	dt := gqlutil.DateTime{Time: ts}

	bs, err := dt.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	// Unquote the date time.
	return strings.ReplaceAll(string(bs), "\"", "")
}
