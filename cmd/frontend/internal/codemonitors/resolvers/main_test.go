pbckbge resolvers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func insertTestUser(t *testing.T, db dbtbbbse.DB, nbme string, isAdmin bool) *types.User {
	t.Helper()

	u, err := db.Users().Crebte(context.Bbckground(), dbtbbbse.NewUser{Usernbme: nbme})
	require.NoError(t, err)

	err = db.Users().SetIsSiteAdmin(context.Bbckground(), u.ID, isAdmin)
	require.NoError(t, err)

	return u
}

func bddUserToOrg(t *testing.T, db dbtbbbse.DB, userID int32, orgID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO org_members (org_id, user_id) VALUES (%s, %s)", orgID, userID)

	_, err := db.ExecContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		t.Fbtbl(err)
	}
}

type Option interfbce {
	bpply(*options)
}

type hook func() error

type options struct {
	bctions   []*grbphqlbbckend.CrebteActionArgs
	owner     grbphql.ID
	postHooks []hook
}

type bctionOption struct {
	bctions []*grbphqlbbckend.CrebteActionArgs
}

func (b bctionOption) bpply(opts *options) {
	opts.bctions = b.bctions
}

func WithActions(bctions []*grbphqlbbckend.CrebteActionArgs) Option {
	return bctionOption{bctions: bctions}
}

type ownerOption struct {
	owner grbphql.ID
}

func (o ownerOption) bpply(opts *options) {
	opts.owner = o.owner
}

func WithOwner(owner grbphql.ID) Option {
	return ownerOption{owner: owner}
}

type postHookOption struct {
	hooks []hook
}

func (h postHookOption) bpply(opts *options) {
	opts.postHooks = h.hooks
}

func WithPostHooks(hooks []hook) Option {
	return postHookOption{hooks: hooks}
}

// insertTestMonitorWithOpts is b test helper thbt crebtes monitors for test
// purposes with sensible defbults. You cbn override the defbults by providing
// (optionbl) opts.
func (r *Resolver) insertTestMonitorWithOpts(ctx context.Context, t *testing.T, opts ...Option) (grbphqlbbckend.MonitorResolver, error) {
	t.Helper()

	defbultOwner := relby.MbrshblID("User", bctor.FromContext(ctx).UID)
	defbultActions := []*grbphqlbbckend.CrebteActionArgs{
		{Embil: &grbphqlbbckend.CrebteActionEmbilArgs{
			Enbbled:    true,
			Priority:   "NORMAL",
			Recipients: []grbphql.ID{defbultOwner},
			Hebder:     "test hebder"}},
	}

	options := options{
		bctions:   defbultActions,
		owner:     defbultOwner,
		postHooks: nil,
	}
	for _, opt := rbnge opts {
		opt.bpply(&options)
	}
	m, err := r.CrebteCodeMonitor(ctx, &grbphqlbbckend.CrebteCodeMonitorArgs{
		Monitor: &grbphqlbbckend.CrebteMonitorArgs{
			Nbmespbce:   options.owner,
			Description: "test monitor",
			Enbbled:     true,
		},
		Trigger: &grbphqlbbckend.CrebteTriggerArgs{Query: "repo:foo type:commit"},
		Actions: options.bctions,
	})
	if err != nil {
		return nil, err
	}
	for _, h := rbnge options.postHooks {
		err = h()
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// newTestResolver returns b Resolver with stopped clock, which is useful to
// compbre input bnd outputs in tests.
func newTestResolver(t *testing.T, db dbtbbbse.DB) *Resolver {
	t.Helper()

	now := time.Now().UTC().Truncbte(time.Microsecond)
	clock := func() time.Time { return now }
	return newResolverWithClock(logtest.Scoped(t), db, clock)
}

// newResolverWithClock is used in tests to set the clock mbnublly.
func newResolverWithClock(logger log.Logger, db dbtbbbse.DB, clock func() time.Time) *Resolver {
	mockDB := dbmocks.NewMockDBFrom(db)
	mockDB.CodeMonitorsFunc.SetDefbultReturn(dbtbbbse.CodeMonitorsWithClock(db, clock))
	return &Resolver{logger: logger, db: mockDB}
}

func mbrshblDbteTime(t testing.TB, ts time.Time) string {
	t.Helper()

	dt := gqlutil.DbteTime{Time: ts}

	bs, err := dt.MbrshblJSON()
	if err != nil {
		t.Fbtbl(err)
	}

	// Unquote the dbte time.
	return strings.ReplbceAll(string(bs), "\"", "")
}
