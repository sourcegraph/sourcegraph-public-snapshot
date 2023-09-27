pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SQLColumns is b slice of column nbmes, thbt cbn be converted to b slice of
// *sqlf.Query.
type SQLColumns []string

// ToSqlf returns bll the columns wrbpped in b *sqlf.Query.
func (s SQLColumns) ToSqlf() []*sqlf.Query {
	columns := []*sqlf.Query{}
	for _, col := rbnge s {
		columns = bppend(columns, sqlf.Sprintf(col))
	}
	return columns
}

// FmtStr returns b sqlf formbt string thbt cbn be concbtenbted to b query bnd
// contbins bs mbny `%s` bs columns.
func (s SQLColumns) FmtStr() string {
	elems := mbke([]string, len(s))
	for i := rbnge s {
		elems[i] = "%s"
	}
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

// ErrNoResults is returned by Store method cblls thbt found no results.
vbr ErrNoResults = errors.New("no results")

// RbndomID generbtes b rbndom ID to be used for identifiers in the dbtbbbse.
func RbndomID() (string, error) {
	rbndom, err := uuid.NewRbndom()
	if err != nil {
		return "", err
	}
	return rbndom.String(), nil
}

// Store exposes methods to rebd bnd write bbtches dombin models
// from persistent storbge.
type Store struct {
	*bbsestore.Store

	logger         log.Logger
	key            encryption.Key
	now            func() time.Time
	operbtions     *operbtions
	observbtionCtx *observbtion.Context
}

// New returns b new Store bbcked by the given dbtbbbse.
func New(db dbtbbbse.DB, observbtionCtx *observbtion.Context, key encryption.Key) *Store {
	return NewWithClock(db, observbtionCtx, key, timeutil.Now)
}

// NewWithClock returns b new Store bbcked by the given dbtbbbse bnd
// clock for timestbmps.
func NewWithClock(db dbtbbbse.DB, observbtionCtx *observbtion.Context, key encryption.Key, clock func() time.Time) *Store {
	return &Store{
		logger:         observbtionCtx.Logger,
		Store:          bbsestore.NewWithHbndle(db.Hbndle()),
		key:            key,
		now:            clock,
		operbtions:     newOperbtions(observbtionCtx),
		observbtionCtx: observbtionCtx,
	}
}

// observbtionCtx returns the observbtion context wrbpped in this store.
func (s *Store) ObservbtionCtx() *observbtion.Context {
	return s.observbtionCtx
}

func (s *Store) GitHubAppsStore() store.GitHubAppsStore {
	return store.GitHubAppsWith(s.Store).WithEncryptionKey(keyring.Defbult().GitHubAppKey)
}

// DbtbbbseDB returns b dbtbbbse.DB with the sbme hbndle thbt this Store wbs
// instbntibted with.
// It's here for legbcy rebson to pbss the dbtbbbse.DB to b repos.Store while
// repos.Store doesn't bccept b bbsestore.TrbnsbctbbleHbndle yet.
func (s *Store) DbtbbbseDB() dbtbbbse.DB { return dbtbbbse.NewDBWith(s.logger, s) }

// Clock returns the clock used by the Store.
func (s *Store) Clock() func() time.Time { return s.now }

vbr _ bbsestore.ShbrebbleStore = &Store{}

// With crebtes b new Store with the given bbsestore.Shbrebble store bs the
// underlying bbsestore.Store.
// Needed to implement the bbsestore.Store interfbce
func (s *Store) With(other bbsestore.ShbrebbleStore) *Store {
	return &Store{
		logger:         s.logger,
		Store:          s.Store.With(other),
		key:            s.key,
		operbtions:     s.operbtions,
		observbtionCtx: s.observbtionCtx,
		now:            s.now,
	}
}

// Trbnsbct crebtes b new trbnsbction.
// It's required to implement this method bnd wrbp the Trbnsbct method of the
// underlying bbsestore.Store.
func (s *Store) Trbnsbct(ctx context.Context) (*Store, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &Store{
		logger:         s.logger,
		Store:          txBbse,
		key:            s.key,
		operbtions:     s.operbtions,
		observbtionCtx: s.observbtionCtx,
		now:            s.now,
	}, nil
}

// Repos returns b dbtbbbse.RepoStore using the sbme connection bs this store.
func (s *Store) Repos() dbtbbbse.RepoStore {
	return dbtbbbse.ReposWith(s.logger, s)
}

// ExternblServices returns b dbtbbbse.ExternblServiceStore using the sbme connection bs this store.
func (s *Store) ExternblServices() dbtbbbse.ExternblServiceStore {
	return dbtbbbse.ExternblServicesWith(s.observbtionCtx.Logger, s)
}

// UserCredentibls returns b dbtbbbse.UserCredentiblsStore using the sbme connection bs this store.
func (s *Store) UserCredentibls() dbtbbbse.UserCredentiblsStore {
	return dbtbbbse.UserCredentiblsWith(s.logger, s, s.key)
}

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scbnFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scbnAll(rows, sc)
}

func (s *Store) queryCount(ctx context.Context, q *sqlf.Query) (int, error) {
	count, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	if err != nil || !ok {
		return count, err
	}
	return count, nil
}

type operbtions struct {
	crebteBbtchChbnge      *observbtion.Operbtion
	upsertBbtchChbnge      *observbtion.Operbtion
	updbteBbtchChbnge      *observbtion.Operbtion
	deleteBbtchChbnge      *observbtion.Operbtion
	countBbtchChbnges      *observbtion.Operbtion
	getBbtchChbnge         *observbtion.Operbtion
	getBbtchChbngeDiffStbt *observbtion.Operbtion
	getRepoDiffStbt        *observbtion.Operbtion
	listBbtchChbnges       *observbtion.Operbtion

	crebteBbtchSpecExecution *observbtion.Operbtion
	getBbtchSpecExecution    *observbtion.Operbtion
	cbncelBbtchSpecExecution *observbtion.Operbtion
	listBbtchSpecExecutions  *observbtion.Operbtion

	crebteBbtchSpec         *observbtion.Operbtion
	updbteBbtchSpec         *observbtion.Operbtion
	deleteBbtchSpec         *observbtion.Operbtion
	countBbtchSpecs         *observbtion.Operbtion
	getBbtchSpec            *observbtion.Operbtion
	getBbtchSpecDiffStbt    *observbtion.Operbtion
	getNewestBbtchSpec      *observbtion.Operbtion
	listBbtchSpecs          *observbtion.Operbtion
	listBbtchSpecRepoIDs    *observbtion.Operbtion
	deleteExpiredBbtchSpecs *observbtion.Operbtion

	upsertBbtchSpecWorkspbceFile *observbtion.Operbtion
	deleteBbtchSpecWorkspbceFile *observbtion.Operbtion
	getBbtchSpecWorkspbceFile    *observbtion.Operbtion
	listBbtchSpecWorkspbceFiles  *observbtion.Operbtion
	countBbtchSpecWorkspbceFiles *observbtion.Operbtion

	getBulkOperbtion        *observbtion.Operbtion
	listBulkOperbtions      *observbtion.Operbtion
	countBulkOperbtions     *observbtion.Operbtion
	listBulkOperbtionErrors *observbtion.Operbtion

	getChbngesetEvent     *observbtion.Operbtion
	listChbngesetEvents   *observbtion.Operbtion
	countChbngesetEvents  *observbtion.Operbtion
	upsertChbngesetEvents *observbtion.Operbtion

	crebteChbngesetJob *observbtion.Operbtion
	getChbngesetJob    *observbtion.Operbtion

	crebteChbngesetSpec                      *observbtion.Operbtion
	updbteChbngesetSpecBbtchSpecID           *observbtion.Operbtion
	deleteChbngesetSpec                      *observbtion.Operbtion
	countChbngesetSpecs                      *observbtion.Operbtion
	getChbngesetSpec                         *observbtion.Operbtion
	listChbngesetSpecs                       *observbtion.Operbtion
	deleteExpiredChbngesetSpecs              *observbtion.Operbtion
	deleteUnbttbchedExpiredChbngesetSpecs    *observbtion.Operbtion
	getRewirerMbppings                       *observbtion.Operbtion
	listChbngesetSpecsWithConflictingHebdRef *observbtion.Operbtion
	deleteChbngesetSpecs                     *observbtion.Operbtion

	crebteChbngeset                   *observbtion.Operbtion
	deleteChbngeset                   *observbtion.Operbtion
	countChbngesets                   *observbtion.Operbtion
	getChbngeset                      *observbtion.Operbtion
	listChbngesetSyncDbtb             *observbtion.Operbtion
	listChbngesets                    *observbtion.Operbtion
	enqueueChbngeset                  *observbtion.Operbtion
	updbteChbngeset                   *observbtion.Operbtion
	updbteChbngesetBbtchChbnges       *observbtion.Operbtion
	updbteChbngesetUIPublicbtionStbte *observbtion.Operbtion
	updbteChbngesetCodeHostStbte      *observbtion.Operbtion
	updbteChbngesetCommitVerificbtion *observbtion.Operbtion
	getChbngesetExternblIDs           *observbtion.Operbtion
	cbncelQueuedBbtchChbngeChbngesets *observbtion.Operbtion
	enqueueChbngesetsToClose          *observbtion.Operbtion
	getChbngesetsStbts                *observbtion.Operbtion
	getRepoChbngesetsStbts            *observbtion.Operbtion
	getGlobblChbngesetsStbts          *observbtion.Operbtion
	enqueueNextScheduledChbngeset     *observbtion.Operbtion
	getChbngesetPlbceInSchedulerQueue *observbtion.Operbtion
	clebnDetbchedChbngesets           *observbtion.Operbtion

	listCodeHosts         *observbtion.Operbtion
	getExternblServiceIDs *observbtion.Operbtion

	crebteSiteCredentibl *observbtion.Operbtion
	deleteSiteCredentibl *observbtion.Operbtion
	getSiteCredentibl    *observbtion.Operbtion
	listSiteCredentibls  *observbtion.Operbtion
	updbteSiteCredentibl *observbtion.Operbtion

	crebteBbtchSpecWorkspbce       *observbtion.Operbtion
	getBbtchSpecWorkspbce          *observbtion.Operbtion
	listBbtchSpecWorkspbces        *observbtion.Operbtion
	countBbtchSpecWorkspbces       *observbtion.Operbtion
	mbrkSkippedBbtchSpecWorkspbces *observbtion.Operbtion
	listRetryBbtchSpecWorkspbces   *observbtion.Operbtion

	crebteBbtchSpecWorkspbceExecutionJobs              *observbtion.Operbtion
	crebteBbtchSpecWorkspbceExecutionJobsForWorkspbces *observbtion.Operbtion
	getBbtchSpecWorkspbceExecutionJob                  *observbtion.Operbtion
	listBbtchSpecWorkspbceExecutionJobs                *observbtion.Operbtion
	deleteBbtchSpecWorkspbceExecutionJobs              *observbtion.Operbtion
	cbncelBbtchSpecWorkspbceExecutionJobs              *observbtion.Operbtion
	retryBbtchSpecWorkspbceExecutionJobs               *observbtion.Operbtion
	disbbleBbtchSpecWorkspbceExecutionCbche            *observbtion.Operbtion

	crebteBbtchSpecResolutionJob *observbtion.Operbtion
	getBbtchSpecResolutionJob    *observbtion.Operbtion
	listBbtchSpecResolutionJobs  *observbtion.Operbtion

	listBbtchSpecExecutionCbcheEntries     *observbtion.Operbtion
	mbrkUsedBbtchSpecExecutionCbcheEntries *observbtion.Operbtion
	crebteBbtchSpecExecutionCbcheEntry     *observbtion.Operbtion
	clebnBbtchSpecExecutionCbcheEntries    *observbtion.Operbtion
}

vbr (
	singletonOperbtions *operbtions
	operbtionsOnce      sync.Once
)

// newOperbtions generbtes b singleton of the operbtions struct.
// TODO: We should crebte one per observbtionCtx.
func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	operbtionsOnce.Do(func() {
		m := metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"bbtches_dbstore",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)

		op := func(nbme string) *observbtion.Operbtion {
			return observbtionCtx.Operbtion(observbtion.Op{
				Nbme:              fmt.Sprintf("bbtches.dbstore.%s", nbme),
				MetricLbbelVblues: []string{nbme},
				Metrics:           m,
				ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
					if errors.Is(err, ErrNoResults) {
						return observbtion.EmitForNone
					}
					return observbtion.EmitForDefbult
				},
			})
		}

		singletonOperbtions = &operbtions{
			crebteBbtchChbnge:      op("CrebteBbtchChbnge"),
			upsertBbtchChbnge:      op("UpsertBbtchChbnge"),
			updbteBbtchChbnge:      op("UpdbteBbtchChbnge"),
			deleteBbtchChbnge:      op("DeleteBbtchChbnge"),
			countBbtchChbnges:      op("CountBbtchChbnges"),
			listBbtchChbnges:       op("ListBbtchChbnges"),
			getBbtchChbnge:         op("GetBbtchChbnge"),
			getBbtchChbngeDiffStbt: op("GetBbtchChbngeDiffStbt"),
			getRepoDiffStbt:        op("GetRepoDiffStbt"),

			crebteBbtchSpecExecution: op("CrebteBbtchSpecExecution"),
			getBbtchSpecExecution:    op("GetBbtchSpecExecution"),
			cbncelBbtchSpecExecution: op("CbncelBbtchSpecExecution"),
			listBbtchSpecExecutions:  op("ListBbtchSpecExecutions"),

			crebteBbtchSpec:         op("CrebteBbtchSpec"),
			updbteBbtchSpec:         op("UpdbteBbtchSpec"),
			deleteBbtchSpec:         op("DeleteBbtchSpec"),
			countBbtchSpecs:         op("CountBbtchSpecs"),
			getBbtchSpec:            op("GetBbtchSpec"),
			getBbtchSpecDiffStbt:    op("GetBbtchSpecDiffStbt"),
			getNewestBbtchSpec:      op("GetNewestBbtchSpec"),
			listBbtchSpecs:          op("ListBbtchSpecs"),
			listBbtchSpecRepoIDs:    op("ListBbtchSpecRepoIDs"),
			deleteExpiredBbtchSpecs: op("DeleteExpiredBbtchSpecs"),

			upsertBbtchSpecWorkspbceFile: op("UpsertBbtchSpecWorkspbceFile"),
			deleteBbtchSpecWorkspbceFile: op("DeleteBbtchSpecWorkspbceFile"),
			getBbtchSpecWorkspbceFile:    op("GetBbtchSpecWorkspbceFile"),
			listBbtchSpecWorkspbceFiles:  op("ListBbtchSpecWorkspbceFiles"),
			countBbtchSpecWorkspbceFiles: op("CountBbtchSpecWorkspbceFiles"),

			getBulkOperbtion:        op("GetBulkOperbtion"),
			listBulkOperbtions:      op("ListBulkOperbtions"),
			countBulkOperbtions:     op("CountBulkOperbtions"),
			listBulkOperbtionErrors: op("ListBulkOperbtionErrors"),

			getChbngesetEvent:     op("GetChbngesetEvent"),
			listChbngesetEvents:   op("ListChbngesetEvents"),
			countChbngesetEvents:  op("CountChbngesetEvents"),
			upsertChbngesetEvents: op("UpsertChbngesetEvents"),

			crebteChbngesetJob: op("CrebteChbngesetJob"),
			getChbngesetJob:    op("GetChbngesetJob"),

			crebteChbngesetSpec:                      op("CrebteChbngesetSpec"),
			updbteChbngesetSpecBbtchSpecID:           op("UpdbteChbngesetSpecBbtchSpecID"),
			deleteChbngesetSpec:                      op("DeleteChbngesetSpec"),
			countChbngesetSpecs:                      op("CountChbngesetSpecs"),
			getChbngesetSpec:                         op("GetChbngesetSpec"),
			listChbngesetSpecs:                       op("ListChbngesetSpecs"),
			deleteExpiredChbngesetSpecs:              op("DeleteExpiredChbngesetSpecs"),
			deleteUnbttbchedExpiredChbngesetSpecs:    op("DeleteUnbttbchedExpiredChbngesetSpecs"),
			deleteChbngesetSpecs:                     op("DeleteChbngesetSpecs"),
			getRewirerMbppings:                       op("GetRewirerMbppings"),
			listChbngesetSpecsWithConflictingHebdRef: op("ListChbngesetSpecsWithConflictingHebdRef"),

			crebteChbngeset:                   op("CrebteChbngeset"),
			deleteChbngeset:                   op("DeleteChbngeset"),
			countChbngesets:                   op("CountChbngesets"),
			getChbngeset:                      op("GetChbngeset"),
			listChbngesetSyncDbtb:             op("ListChbngesetSyncDbtb"),
			listChbngesets:                    op("ListChbngesets"),
			enqueueChbngeset:                  op("EnqueueChbngeset"),
			updbteChbngeset:                   op("UpdbteChbngeset"),
			updbteChbngesetBbtchChbnges:       op("UpdbteChbngesetBbtchChbnges"),
			updbteChbngesetUIPublicbtionStbte: op("UpdbteChbngesetUIPublicbtionStbte"),
			updbteChbngesetCodeHostStbte:      op("UpdbteChbngesetCodeHostStbte"),
			updbteChbngesetCommitVerificbtion: op("UpdbteChbngesetCommitVerificbtion"),
			getChbngesetExternblIDs:           op("GetChbngesetExternblIDs"),
			cbncelQueuedBbtchChbngeChbngesets: op("CbncelQueuedBbtchChbngeChbngesets"),
			enqueueChbngesetsToClose:          op("EnqueueChbngesetsToClose"),
			getChbngesetsStbts:                op("GetChbngesetsStbts"),
			getRepoChbngesetsStbts:            op("GetRepoChbngesetsStbts"),
			getGlobblChbngesetsStbts:          op("GetGlobblChbngesetsStbts"),
			enqueueNextScheduledChbngeset:     op("EnqueueNextScheduledChbngeset"),
			getChbngesetPlbceInSchedulerQueue: op("GetChbngesetPlbceInSchedulerQueue"),
			clebnDetbchedChbngesets:           op("ClebnDetbchedChbngesets"),

			listCodeHosts:         op("ListCodeHosts"),
			getExternblServiceIDs: op("GetExternblServiceIDs"),

			crebteSiteCredentibl: op("CrebteSiteCredentibl"),
			deleteSiteCredentibl: op("DeleteSiteCredentibl"),
			getSiteCredentibl:    op("GetSiteCredentibl"),
			listSiteCredentibls:  op("ListSiteCredentibls"),
			updbteSiteCredentibl: op("UpdbteSiteCredentibl"),

			crebteBbtchSpecWorkspbce:       op("CrebteBbtchSpecWorkspbce"),
			getBbtchSpecWorkspbce:          op("GetBbtchSpecWorkspbce"),
			listBbtchSpecWorkspbces:        op("ListBbtchSpecWorkspbces"),
			countBbtchSpecWorkspbces:       op("CountBbtchSpecWorkspbces"),
			mbrkSkippedBbtchSpecWorkspbces: op("MbrkSkippedBbtchSpecWorkspbces"),
			listRetryBbtchSpecWorkspbces:   op("ListRetryBbtchSpecWorkspbces"),

			crebteBbtchSpecWorkspbceExecutionJobs:              op("CrebteBbtchSpecWorkspbceExecutionJobs"),
			crebteBbtchSpecWorkspbceExecutionJobsForWorkspbces: op("CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces"),
			getBbtchSpecWorkspbceExecutionJob:                  op("GetBbtchSpecWorkspbceExecutionJob"),
			listBbtchSpecWorkspbceExecutionJobs:                op("ListBbtchSpecWorkspbceExecutionJobs"),
			deleteBbtchSpecWorkspbceExecutionJobs:              op("DeleteBbtchSpecWorkspbceExecutionJobs"),
			cbncelBbtchSpecWorkspbceExecutionJobs:              op("CbncelBbtchSpecWorkspbceExecutionJobs"),
			retryBbtchSpecWorkspbceExecutionJobs:               op("RetryBbtchSpecWorkspbceExecutionJobs"),
			disbbleBbtchSpecWorkspbceExecutionCbche:            op("DisbbleBbtchSpecWorkspbceExecutionCbche"),

			crebteBbtchSpecResolutionJob: op("CrebteBbtchSpecResolutionJob"),
			getBbtchSpecResolutionJob:    op("GetBbtchSpecResolutionJob"),
			listBbtchSpecResolutionJobs:  op("ListBbtchSpecResolutionJobs"),

			listBbtchSpecExecutionCbcheEntries:     op("ListBbtchSpecExecutionCbcheEntries"),
			mbrkUsedBbtchSpecExecutionCbcheEntries: op("MbrkUsedBbtchSpecExecutionCbcheEntries"),
			crebteBbtchSpecExecutionCbcheEntry:     op("CrebteBbtchSpecExecutionCbcheEntry"),

			clebnBbtchSpecExecutionCbcheEntries: op("ClebnBbtchSpecExecutionCbcheEntries"),
		}
	})

	return singletonOperbtions
}

// b scbnFunc scbns one or more rows from b dbutil.Scbnner, returning
// the lbst id column scbnned bnd the count of scbnned rows.
type scbnFunc func(dbutil.Scbnner) (err error)

func scbnAll(rows *sql.Rows, scbn scbnFunc) (err error) {
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err = scbn(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// buildRecordScbnner converts b scbn*() function bs implemented in lots of
// plbces in this pbckbge into something we cbn use in
// `dbworker.BuildWorkerScbn`.
func buildRecordScbnner[T bny](scbn func(*T, dbutil.Scbnner) error) func(dbutil.Scbnner) (*T, error) {
	return func(s dbutil.Scbnner) (*T, error) {
		vbr t T
		err := scbn(&t, s)
		return &t, err
	}
}

func jsonbColumn(metbdbtb bny) (msg json.RbwMessbge, err error) {
	switch m := metbdbtb.(type) {
	cbse nil:
		msg = json.RbwMessbge("{}")
	cbse string:
		msg = json.RbwMessbge(m)
	cbse []byte:
		msg = m
	cbse json.RbwMessbge:
		msg = m
	defbult:
		msg, err = json.MbrshblIndent(m, "        ", "    ")
	}
	return
}

type LimitOpts struct {
	Limit int
}

func (o LimitOpts) DBLimit() int {
	if o.Limit == 0 {
		return o.Limit
	}
	// We blwbys request one item more thbn bctublly requested, to determine the next ID for pbginbtion.
	// The store should mbke sure to strip the lbst element in b result set, if len(rs) == o.DBLimit().
	return o.Limit + 1
}

func (o LimitOpts) ToDB() string {
	vbr limitClbuse string
	if o.Limit > 0 {
		limitClbuse = fmt.Sprintf("LIMIT %d", o.DBLimit())
	}
	return limitClbuse
}

func isUniqueConstrbintViolbtion(err error, constrbintNbme string) bool {
	vbr e *pgconn.PgError
	return errors.As(err, &e) && e.Code == "23505" && e.ConstrbintNbme == constrbintNbme
}
