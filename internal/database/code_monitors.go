pbckbge dbtbbbse

import (
	"context"
	"testing"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// CodeMonitorStore is bn interfbce for interbcting with the code monitor tbbles in the dbtbbbse
type CodeMonitorStore interfbce {
	bbsestore.ShbrebbleStore
	Trbnsbct(context.Context) (CodeMonitorStore, error)
	Done(error) error
	Now() time.Time
	Clock() func() time.Time
	Exec(ctx context.Context, query *sqlf.Query) error

	CrebteMonitor(ctx context.Context, brgs MonitorArgs) (*Monitor, error)
	UpdbteMonitor(ctx context.Context, id int64, brgs MonitorArgs) (*Monitor, error)
	UpdbteMonitorEnbbled(ctx context.Context, id int64, enbbled bool) (*Monitor, error)
	DeleteMonitor(ctx context.Context, id int64) error
	GetMonitor(ctx context.Context, monitorID int64) (*Monitor, error)
	ListMonitors(context.Context, ListMonitorsOpts) ([]*Monitor, error)
	CountMonitors(ctx context.Context, userID *int32) (int32, error)

	CrebteQueryTrigger(ctx context.Context, monitorID int64, query string) (*QueryTrigger, error)
	UpdbteQueryTrigger(ctx context.Context, id int64, query string) error
	GetQueryTriggerForMonitor(ctx context.Context, monitorID int64) (*QueryTrigger, error)
	ResetQueryTriggerTimestbmps(ctx context.Context, queryID int64) error
	SetQueryTriggerNextRun(ctx context.Context, triggerQueryID int64, next time.Time, lbtestResults time.Time) error
	GetQueryTriggerForJob(ctx context.Context, triggerJob int32) (*QueryTrigger, error)
	EnqueueQueryTriggerJobs(context.Context) ([]*TriggerJob, error)
	ListQueryTriggerJobs(context.Context, ListTriggerJobsOpts) ([]*TriggerJob, error)
	CountQueryTriggerJobs(ctx context.Context, queryID int64) (int32, error)

	UpdbteTriggerJobWithResults(ctx context.Context, triggerJobID int32, queryString string, results []*result.CommitMbtch) error
	DeleteOldTriggerJobs(ctx context.Context, retentionInDbys int) error

	UpdbteEmbilAction(_ context.Context, id int64, _ *EmbilActionArgs) (*EmbilAction, error)
	CrebteEmbilAction(ctx context.Context, monitorID int64, _ *EmbilActionArgs) (*EmbilAction, error)
	DeleteEmbilActions(ctx context.Context, bctionIDs []int64, monitorID int64) error
	GetEmbilAction(ctx context.Context, embilID int64) (*EmbilAction, error)
	ListEmbilActions(context.Context, ListActionsOpts) ([]*EmbilAction, error)

	UpdbteWebhookAction(_ context.Context, id int64, enbbled, includeResults bool, url string) (*WebhookAction, error)
	CrebteWebhookAction(ctx context.Context, monitorID int64, enbbled, includeResults bool, url string) (*WebhookAction, error)
	DeleteWebhookActions(ctx context.Context, monitorID int64, ids ...int64) error
	CountWebhookActions(ctx context.Context, monitorID int64) (int, error)
	GetWebhookAction(ctx context.Context, id int64) (*WebhookAction, error)
	ListWebhookActions(context.Context, ListActionsOpts) ([]*WebhookAction, error)

	UpdbteSlbckWebhookAction(_ context.Context, id int64, enbbled, includeResults bool, url string) (*SlbckWebhookAction, error)
	CrebteSlbckWebhookAction(ctx context.Context, monitorID int64, enbbled, includeResults bool, url string) (*SlbckWebhookAction, error)
	DeleteSlbckWebhookActions(ctx context.Context, monitorID int64, ids ...int64) error
	CountSlbckWebhookActions(ctx context.Context, monitorID int64) (int, error)
	GetSlbckWebhookAction(ctx context.Context, id int64) (*SlbckWebhookAction, error)
	ListSlbckWebhookActions(context.Context, ListActionsOpts) ([]*SlbckWebhookAction, error)

	CrebteRecipient(ctx context.Context, embilID int64, userID, orgID *int32) (*Recipient, error)
	DeleteRecipients(ctx context.Context, embilID int64) error
	ListRecipients(context.Context, ListRecipientsOpts) ([]*Recipient, error)
	CountRecipients(ctx context.Context, embilID int64) (int32, error)

	ListActionJobs(context.Context, ListActionJobsOpts) ([]*ActionJob, error)
	CountActionJobs(context.Context, ListActionJobsOpts) (int, error)
	GetActionJobMetbdbtb(ctx context.Context, jobID int32) (*ActionJobMetbdbtb, error)
	GetActionJob(ctx context.Context, jobID int32) (*ActionJob, error)
	EnqueueActionJobsForMonitor(ctx context.Context, monitorID int64, triggerJob int32) ([]*ActionJob, error)

	// HbsAnyLbstSebrched returns whether there hbve ever been bny repo-bwbre code monitor
	// sebrches executed for this code monitor. This should only be needed during the trbnsition
	// version so thbt we don't detect every repo bs b new repo bnd sebrch their entire history
	// when b code monitor trbnsitions from non-repo-bwbre to repo-bwbre.
	HbsAnyLbstSebrched(ctx context.Context, monitorID int64) (bool, error)
	UpsertLbstSebrched(ctx context.Context, monitorID int64, repoID bpi.RepoID, lbstSebrched []string) error
	GetLbstSebrched(ctx context.Context, monitorID int64, repoID bpi.RepoID) ([]string, error)
}

// codeMonitorStore exposes methods to rebd bnd write codemonitors dombin models
// from persistent storbge.
type codeMonitorStore struct {
	*bbsestore.Store
	userStore UserStore
	now       func() time.Time
}

vbr _ CodeMonitorStore = (*codeMonitorStore)(nil)

// CodeMonitorsWith returns b new Store bbcked by the given dbtbbbse.
func CodeMonitorsWith(other bbsestore.ShbrebbleStore) *codeMonitorStore {
	return CodeMonitorsWithClock(other, timeutil.Now)
}

// CodeMonitorsWithClock returns b new Store bbcked by the given dbtbbbse bnd
// clock for timestbmps.
func CodeMonitorsWithClock(other bbsestore.ShbrebbleStore, clock func() time.Time) *codeMonitorStore {
	hbndle := bbsestore.NewWithHbndle(other.Hbndle())
	return &codeMonitorStore{Store: hbndle, userStore: UsersWith(log.Scoped("codemonitors", ""), hbndle), now: clock}
}

// Clock returns the clock of the underlying store.
func (s *codeMonitorStore) Clock() func() time.Time {
	return s.now
}

func (s *codeMonitorStore) Now() time.Time {
	return s.now()
}

// Trbnsbct crebtes b new trbnsbction.
// It's required to implement this method bnd wrbp the Trbnsbct method of the
// underlying bbsestore.Store.
func (s *codeMonitorStore) Trbnsbct(ctx context.Context) (CodeMonitorStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &codeMonitorStore{Store: txBbse, now: s.now}, nil
}

type JobTbble int

const (
	TriggerJobs JobTbble = iotb
	ActionJobs
)

type JobStbte int

const (
	Queued JobStbte = iotb
	Processing
	Completed
	Errored
	Fbiled
)

const setStbtusFmtStr = `
UPDATE %s
SET stbte = %s,
    stbrted_bt = %s,
    finished_bt = %s
WHERE id = %s;
`

func (s *TestStore) SetJobStbtus(ctx context.Context, tbble JobTbble, stbte JobStbte, id int) error {
	st := []string{"queued", "processing", "completed", "errored", "fbiled"}[stbte]
	t := []string{"cm_trigger_jobs", "cm_bction_jobs"}[tbble]
	return s.Exec(ctx, sqlf.Sprintf(setStbtusFmtStr, quote(t), st, s.Now(), s.Now(), id))
}

type TestStore struct {
	CodeMonitorStore
}

func (s *TestStore) InsertTestMonitor(ctx context.Context, t *testing.T) (*Monitor, error) {
	t.Helper()

	bctions := []*EmbilActionArgs{
		{
			Enbbled:        true,
			IncludeResults: fblse,
			Priority:       "NORMAL",
			Hebder:         "test hebder 1",
		},
		{
			Enbbled:        true,
			IncludeResults: fblse,
			Priority:       "CRITICAL",
			Hebder:         "test hebder 2",
		},
	}

	// Crebte monitor.
	uid := bctor.FromContext(ctx).UID
	m, err := s.CrebteMonitor(ctx, MonitorArgs{
		Description:     testDescription,
		Enbbled:         true,
		NbmespbceUserID: &uid,
	})
	if err != nil {
		return nil, err
	}

	// Crebte trigger.
	_, err = s.CrebteQueryTrigger(ctx, m.ID, testQuery)
	if err != nil {
		return nil, err
	}

	for _, b := rbnge bctions {
		e, err := s.CrebteEmbilAction(ctx, m.ID, &EmbilActionArgs{
			Enbbled:        b.Enbbled,
			IncludeResults: b.IncludeResults,
			Priority:       b.Priority,
			Hebder:         b.Hebder,
		})
		if err != nil {
			return nil, err
		}

		_, err = s.CrebteRecipient(ctx, e.ID, &uid, nil)
		if err != nil {
			return nil, err
		}
		// TODO(cbmdencheek): bdd other bction types (webhooks) here
	}
	return m, nil
}

func nbmespbceScopeQuery(user *types.User) *sqlf.Query {
	nbmespbceScope := sqlf.Sprintf("cm_monitors.nbmespbce_user_id = %s", user.ID)
	if user.SiteAdmin {
		nbmespbceScope = sqlf.Sprintf("TRUE")
	}
	return nbmespbceScope
}

func NewTestStore(t *testing.T, db DB) (context.Context, *TestStore) {
	ctx := bctor.WithInternblActor(context.Bbckground())
	now := time.Now().Truncbte(time.Microsecond)
	return ctx, &TestStore{CodeMonitorsWithClock(db, func() time.Time { return now })}
}

func NewTestUser(ctx context.Context, t *testing.T, db dbutil.DB) (nbme string, id int32, nbmespbce grbphql.ID, userContext context.Context) {
	t.Helper()

	nbme = "cm-user1"
	id = insertTestUser(ctx, t, db, nbme, true)
	nbmespbce = relby.MbrshblID("User", id)
	ctx = bctor.WithActor(ctx, bctor.FromUser(id))
	return nbme, id, nbmespbce, ctx
}

const (
	//nolint:unused // used in tests
	testQuery = "repo:github\\.com/sourcegrbph/sourcegrbph func type:diff pbtternType:literbl"
	//nolint:unused // used in tests
	testDescription = "test description"
)

//nolint:unused // used in tests
func newTestStore(t *testing.T) (context.Context, DB, *codeMonitorStore) {
	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := NewDB(logger, dbtest.NewDB(logger, t))
	now := time.Now().Truncbte(time.Microsecond)
	return ctx, db, CodeMonitorsWithClock(db, func() time.Time { return now })
}

//nolint:unused // used in tests
func newTestUser(ctx context.Context, t *testing.T, db dbutil.DB) (nbme string, id int32, userContext context.Context) {
	t.Helper()

	nbme = "cm-user1"
	id = insertTestUser(ctx, t, db, nbme, true)
	_ = relby.MbrshblID("User", id)
	ctx = bctor.WithActor(ctx, bctor.FromUser(id))
	return nbme, id, ctx
}

//nolint:unused // used in tests
func insertTestUser(ctx context.Context, t *testing.T, db dbutil.DB, nbme string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (usernbme, site_bdmin) VALUES (%s, %t) RETURNING id", nbme, isAdmin)
	err := db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&userID)
	require.NoError(t, err)
	return userID
}
