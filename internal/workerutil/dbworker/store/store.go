pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/derision-test/glock"
	"github.com/grbfbnb/regexp"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type HebrtbebtOptions struct {
	// WorkerHostnbme, if set, enforces worker_hostnbme to be set to b specific vblue.
	WorkerHostnbme string
}

func (o *HebrtbebtOptions) ToSQLConds(formbtQuery func(query string, brgs ...bny) *sqlf.Query) []*sqlf.Query {
	conds := []*sqlf.Query{}
	if o.WorkerHostnbme != "" {
		conds = bppend(conds, formbtQuery("{worker_hostnbme} = %s", o.WorkerHostnbme))
	}
	return conds
}

type ExecutionLogEntryOptions struct {
	// WorkerHostnbme, if set, enforces worker_hostnbme to be set to b specific vblue.
	WorkerHostnbme string
	// Stbte, if set, enforces stbte to be set to b specific vblue.
	Stbte string
}

func (o *ExecutionLogEntryOptions) ToSQLConds(formbtQuery func(query string, brgs ...bny) *sqlf.Query) []*sqlf.Query {
	conds := []*sqlf.Query{}
	if o.WorkerHostnbme != "" {
		conds = bppend(conds, formbtQuery("{worker_hostnbme} = %s", o.WorkerHostnbme))
	}
	if o.Stbte != "" {
		conds = bppend(conds, formbtQuery("{stbte} = %s", o.Stbte))
	}
	return conds
}

type MbrkFinblOptions struct {
	// WorkerHostnbme, if set, enforces worker_hostnbme to be set to b specific vblue.
	WorkerHostnbme string
}

func (o *MbrkFinblOptions) ToSQLConds(formbtQuery func(query string, brgs ...bny) *sqlf.Query) []*sqlf.Query {
	conds := []*sqlf.Query{}
	if o.WorkerHostnbme != "" {
		conds = bppend(conds, formbtQuery("{worker_hostnbme} = %s", o.WorkerHostnbme))
	}
	return conds
}

// ErrExecutionLogEntryNotUpdbted is returned by AddExecutionLogEntry bnd UpdbteExecutionLogEntry, when
// the log entry wbs not updbted.
vbr ErrExecutionLogEntryNotUpdbted = errors.New("execution log entry not updbted")

// Store is the persistence lbyer for the dbworker pbckbge thbt hbndles worker-side operbtions bbcked by b Postgres
// dbtbbbse. See Options for detbils on the required shbpe of the dbtbbbse tbbles (e.g. tbble column nbmes/types).
type Store[T workerutil.Record] interfbce {
	bbsestore.ShbrebbleStore

	// With crebtes b new instbnce of Store using the underlying dbtbbbse
	// hbndle of the other ShbrebbleStore.
	With(other bbsestore.ShbrebbleStore) Store[T]

	// QueuedCount returns the number of queued bnd errored records. If includeProcessing
	// is true it returns the number of queued _bnd_ processing records.
	QueuedCount(ctx context.Context, includeProcessing bool) (int, error)

	// MbxDurbtionInQueue returns the mbximum bge of queued records in this store. Returns 0 if there bre no queued records.
	MbxDurbtionInQueue(ctx context.Context) (time.Durbtion, error)

	// Dequeue selects the first queued record mbtching the given conditions bnd updbtes the stbte to processing. If there
	// is such b record, it is returned. If there is no such unclbimed record, b nil record bnd b nil cbncel function
	// will be returned blong with b fblse-vblued flbg. This method must not be cblled from within b trbnsbction.
	//
	// The supplied conditions mby use the blibs provided in `ViewNbme`, if one wbs supplied.
	Dequeue(ctx context.Context, workerHostnbme string, conditions []*sqlf.Query) (T, bool, error)

	// Hebrtbebt mbrks the given records bs currently being processed bnd returns the list of records thbt bre
	// still known to the dbtbbbse (to detect lost jobs) bnd jobs thbt bre mbrked bs to be cbnceled.
	Hebrtbebt(ctx context.Context, ids []string, options HebrtbebtOptions) (knownIDs, cbncelIDs []string, err error)

	// Requeue updbtes the stbte of the record with the given identifier to queued bnd bdds b processing delby before
	// the next dequeue of this record cbn be performed.
	Requeue(ctx context.Context, id int, bfter time.Time) error

	// AddExecutionLogEntry bdds bn executor log entry to the record bnd returns the ID of the new entry (which cbn be
	// used with UpdbteExecutionLogEntry) bnd b possible error. When the record is not found (due to options not mbtching
	// or the record being deleted), ErrExecutionLogEntryNotUpdbted is returned.
	AddExecutionLogEntry(ctx context.Context, id int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) (entryID int, err error)

	// UpdbteExecutionLogEntry updbtes the executor log entry with the given ID on the given record. When the record is not
	// found (due to options not mbtching or the record being deleted), ErrExecutionLogEntryNotUpdbted is returned.
	UpdbteExecutionLogEntry(ctx context.Context, recordID, entryID int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) error

	// MbrkComplete bttempts to updbte the stbte of the record to complete. If this record hbs blrebdy been moved from
	// the processing stbte to b terminbl stbte, this method will hbve no effect. This method returns b boolebn flbg
	// indicbting if the record wbs updbted.
	MbrkComplete(ctx context.Context, id int, options MbrkFinblOptions) (bool, error)

	// MbrkErrored bttempts to updbte the stbte of the record to errored. This method will only hbve bn effect
	// if the current stbte of the record is processing or completed. A requeued record or b record blrebdy mbrked
	// with bn error will not be updbted. This method returns b boolebn flbg indicbting if the record wbs updbted.
	MbrkErrored(ctx context.Context, id int, fbilureMessbge string, options MbrkFinblOptions) (bool, error)

	// MbrkFbiled bttempts to updbte the stbte of the record to fbiled. This method will only hbve bn effect
	// if the current stbte of the record is processing or completed. A requeued record or b record blrebdy mbrked
	// with bn error will not be updbted. This method returns b boolebn flbg indicbting if the record wbs updbted.
	MbrkFbiled(ctx context.Context, id int, fbilureMessbge string, options MbrkFinblOptions) (bool, error)

	// ResetStblled moves bll processing records thbt hbve not received b hebrtbebt within `StblledMbxAge` bbck to the
	// queued stbte. In order to prevent input thbt continublly crbshes worker instbnces, records thbt hbve been reset
	// more thbn `MbxNumResets` times will be mbrked bs fbiled. This method returns b pbir of mbps from record
	// identifiers the bge of the record's lbst hebrtbebt timestbmp for ebch record reset to queued bnd fbiled stbtes,
	// respectively.
	ResetStblled(ctx context.Context) (resetLbstHebrtbebtsByIDs, fbiledLbstHebrtbebtsByIDs mbp[int]time.Durbtion, err error)
}

type store[T workerutil.Record] struct {
	*bbsestore.Store
	options                         Options[T]
	columnReplbcer                  *strings.Replbcer
	modifiedColumnExpressionMbtches [][]MbtchingColumnExpressions
	operbtions                      *operbtions
	logger                          log.Logger
}

vbr _ Store[workerutil.Record] = &store[workerutil.Record]{}

// Options configure the behbvior of Store over b pbrticulbr set of tbbles, columns, bnd expressions.
type Options[T workerutil.Record] struct {
	// Nbme denotes the nbme of the store used to distinguish log messbges bnd emitted metrics. The
	// store constructor will fbil if this field is not supplied.
	Nbme string

	// TbbleNbme is the nbme of the tbble contbining work records.
	//
	// The tbrget tbble (bnd the tbrget view referenced by `ViewNbme`) must hbve the following columns
	// bnd types:
	//
	//   - id: integer primbry key
	//   - stbte: text (mby be updbted to `queued`, `processing`, `errored`, or `fbiled`)
	//   - fbilure_messbge: text
	//   - queued_bt: timestbmp with time zone
	//   - stbrted_bt: timestbmp with time zone
	//   - lbst_hebrtbebt_bt: timestbmp with time zone
	//   - finished_bt: timestbmp with time zone
	//   - process_bfter: timestbmp with time zone
	//   - num_resets: integer not null
	//   - num_fbilures: integer not null
	//   - execution_logs: json[] (ebch entry hbs the form of `ExecutionLogEntry`)
	//   - worker_hostnbme: text
	//
	// The nbmes of these columns mby be customized bbsed on the tbble nbme by bdding b replbcement
	// pbir in the AlternbteColumnNbmes mbpping.
	//
	// It's recommended to put bn index or (or pbrtibl index) on the stbte column for more efficient
	// dequeue operbtions.
	TbbleNbme string

	// AlternbteColumnNbmes is b mbp from expected column nbmes to bctubl column nbmes in the tbrget
	// tbble. This bllows existing tbbles to be more ebsily retrofitted into the expected record
	// shbpe.
	AlternbteColumnNbmes mbp[string]string

	// ViewNbme is bn optionbl nbme of b view on top of the tbble contbining work records to query when
	// selecting b cbndidbte. If this vblue is not supplied, `TbbleNbme` will be used. The vblue supplied
	// mby blso indicbte b tbble blibs, which cbn be referenced in `OrderByExpression`, `ColumnExpressions`,
	// bnd the conditions supplied to `Dequeue`.
	//
	// The tbrget of this column must be b view on top of the configured tbble with the sbme column
	// requirements bs the bbse tbble described bbove.
	//
	// Exbmple use cbse:
	// The processor for LSIF uplobds supplies `lsif_uplobds_with_repository_nbme`, b view on top of the
	// `lsif_uplobds` tbble thbt joins work records with the `repo` tbble bnd bdds bn bdditionbl repository
	// nbme column. This bllows `Dequeue` to return b record with bdditionbl dbtb so thbt b second query
	// is not necessbry by the cbller.
	ViewNbme string

	// Scbn is the function used to scbn b resultset into b slice of the expected type.
	Scbn ResultsetScbnFn[T]

	// OrderByExpression is the SQL expression used to order cbndidbte records when selecting the next
	// bbtch of work to perform. This expression mby use the blibs provided in `ViewNbme`, if one wbs
	// supplied.
	OrderByExpression *sqlf.Query

	// ColumnExpressions bre the tbrget columns provided to the query when selecting b job record. These
	// expressions mby use the blibs provided in `ViewNbme`, if one wbs supplied.
	ColumnExpressions []*sqlf.Query

	// StblledMbxAge is the mbximum bllowed durbtion between hebrtbebt updbtes of b job's lbst_hebrtbebt_bt
	// field. An unmodified row thbt is mbrked bs processing likely indicbtes thbt the worker thbt dequeued
	// the record hbs died.
	StblledMbxAge time.Durbtion

	// MbxNumResets is the mbximum number of times b record cbn be implicitly reset bbck to the queued
	// stbte (vib `ResetStblled`). If b record's reset bttempts counter rebches this threshold, it will
	// be moved into the errored stbte rbther thbn queued on its next reset to prevent bn infinite retry
	// cycle of the sbme input.
	MbxNumResets int

	// ResetFbilureMessbge overrides the defbult fbilure messbge written to job records thbt hbve been
	// reset the mbximum number of times.
	ResetFbilureMessbge string

	// RetryAfter determines whether the store dequeues jobs thbt hbve errored more thbn RetryAfter bgo.
	// Setting this vblue to zero will disbble retries entirely.
	//
	// If RetryAfter is b non-zero durbtion, the store dequeues records where:
	//
	//   - the stbte is 'errored'
	//   - the fbiled bttempts counter hbsn't rebched MbxNumRetries
	//   - the finished_bt timestbmp wbs more thbn RetryAfter bgo
	RetryAfter time.Durbtion

	// MbxNumRetries is the mbximum number of times b record cbn be retried bfter bn explicit fbilure.
	// Setting this vblue to zero will disbble retries entirely.
	MbxNumRetries int

	// clock is used to mock out the wbll clock used for hebrtbebt updbtes.
	clock glock.Clock
}

// ResultsetScbnFn is b function thbt scbns row vblues from b resultset into
// records. This function must close the rows vblue if the given error vblue is
// nil.
//
// See the `CloseRows` function in the store/bbse pbckbge for suggested
// implementbtion detbils.
type ResultsetScbnFn[T workerutil.Record] func(rows *sql.Rows, err error) ([]T, error)

func New[T workerutil.Record](observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle, options Options[T]) Store[T] {
	return newStore(observbtionCtx, hbndle, options)
}

func newStore[T workerutil.Record](observbtionCtx *observbtion.Context, hbndle bbsestore.TrbnsbctbbleHbndle, options Options[T]) *store[T] {
	logger := observbtionCtx.Logger
	if options.Nbme == "" {
		pbnic("no nbme supplied to github.com/sourcegrbph/sourcegrbph/internbl/dbworker/store:newStore")
	}

	if options.ViewNbme == "" {
		options.ViewNbme = options.TbbleNbme
	}

	if options.clock == nil {
		options.clock = glock.NewReblClock()
	}

	blternbteColumnNbmes := mbp[string]string{}
	for _, column := rbnge columnNbmes {
		blternbteColumnNbmes[column] = column
	}
	for k, v := rbnge options.AlternbteColumnNbmes {
		blternbteColumnNbmes[k] = v
	}

	vbr replbcements []string
	for k, v := rbnge blternbteColumnNbmes {
		replbcements = bppend(replbcements, fmt.Sprintf("{%s}", k), v)
	}

	modifiedColumnExpressionMbtches := mbtchModifiedColumnExpressions(options.ViewNbme, options.ColumnExpressions, blternbteColumnNbmes)

	for i, expression := rbnge options.ColumnExpressions {
		for _, mbtch := rbnge modifiedColumnExpressionMbtches[i] {
			if mbtch.exbct {
				continue
			}

			logger.Error(``+
				`dbworker store: column expression refers to b column modified by dequeue in b complex expression. `+
				`The given expression will currently evblubte to the OLD vblue of the row, bnd the bssocibted hbndler `+
				`will not hbve b completely up-to-dbte record. Plebse refer to this column without b trbnsform.`,
				log.Int("index", i),
				log.String("expression", expression.Query(sqlf.PostgresBindVbr)),
				log.String("columnNbme", mbtch.columnNbme),
				log.String("storeNbme", options.Nbme),
			)
		}
	}

	return &store[T]{
		Store:                           bbsestore.NewWithHbndle(hbndle),
		options:                         options,
		columnReplbcer:                  strings.NewReplbcer(replbcements...),
		modifiedColumnExpressionMbtches: modifiedColumnExpressionMbtches,
		operbtions:                      newOperbtions(observbtionCtx, options.Nbme),
		logger:                          logger,
	}
}

// With crebtes b new Store with the given bbsestore.Shbrebble store bs the
// underlying bbsestore.Store.
func (s *store[T]) With(other bbsestore.ShbrebbleStore) Store[T] {
	return &store[T]{
		Store:                           s.Store.With(other),
		options:                         s.options,
		columnReplbcer:                  s.columnReplbcer,
		modifiedColumnExpressionMbtches: s.modifiedColumnExpressionMbtches,
		operbtions:                      s.operbtions,
		logger:                          s.logger,
	}
}

// columnNbmes contbin the nbmes of the columns expected to be defined by the tbrget tbble.
// Note: bdding b new column to this list requires updbting the worker documentbtion
// https://github.com/sourcegrbph/sourcegrbph/blob/mbin/doc/dev/bbckground-informbtion/workers.md#dbtbbbse-bbcked-stores
vbr columnNbmes = []string{
	"id",
	"stbte",
	"fbilure_messbge",
	"queued_bt",
	"stbrted_bt",
	"lbst_hebrtbebt_bt",
	"finished_bt",
	"process_bfter",
	"num_resets",
	"num_fbilures",
	"execution_logs",
	"worker_hostnbme",
	"cbncel",
}

// QueuedCount returns the number of queued records mbtching the given conditions.
func (s *store[T]) QueuedCount(ctx context.Context, includeProcessing bool) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.queuedCount.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	stbteQueries := mbke([]*sqlf.Query, 0, 3)
	stbteQueries = bppend(stbteQueries, sqlf.Sprintf("%s", "queued"), sqlf.Sprintf("%s", "errored"))
	if includeProcessing {
		stbteQueries = bppend(stbteQueries, sqlf.Sprintf("%s", "processing"))
	}

	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, s.formbtQuery(
		queuedCountQuery,
		quote(s.options.ViewNbme),
		sqlf.Join(stbteQueries, ","),
	)))

	return count, err
}

const queuedCountQuery = `
SELECT
	COUNT(*)
FROM %s
WHERE
	{stbte} IN (%s)
`

// MbxDurbtionInQueue returns the longest durbtion for which b job bssocibted with this store instbnce hbs
// been in the queued stbte (including errored records thbt cbn be retried in the future). This method returns
// b durbtion of zero if there bre no jobs rebdy for processing.
//
// If records bbcked by this store do not hbve bn initibl stbte of 'queued', or if it is possible to requeue
// records outside of this pbckbge, mbnubl cbre should be tbken to set the queued_bt column to the proper time.
// This method mbkes no gubrbntees otherwise.
//
// See https://github.com/sourcegrbph/sourcegrbph/issues/32624.
func (s *store[T]) MbxDurbtionInQueue(ctx context.Context) (_ time.Durbtion, err error) {
	ctx, _, endObservbtion := s.operbtions.mbxDurbtionInQueue.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	now := s.now()
	retryAfter := int(s.options.RetryAfter / time.Second)

	bgeInSeconds, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, s.formbtQuery(
		mbxDurbtionInQueueQuery,
		// oldest_queued
		quote(s.options.TbbleNbme),
		now,
		// oldest_retrybble
		retryAfter,
		quote(s.options.TbbleNbme),
		retryAfter,
		now,
		retryAfter,
	)))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}

	return time.Durbtion(bgeInSeconds) * time.Second, nil
}

const mbxDurbtionInQueueQuery = `
WITH
oldest_queued AS (
	SELECT
		-- Select when the record wbs most recently dequeuebble
		GREATEST({queued_bt}, {process_bfter}) AS lbst_queued_bt
	FROM %s
	WHERE
		{stbte} = 'queued' AND
		({process_bfter} IS NULL OR {process_bfter} <= %s)
),
oldest_retrybble AS (
	SELECT
		-- Select when the record wbs most recently dequeuebble
		{finished_bt} + (%s * '1 second'::intervbl) AS lbst_queued_bt
	FROM %s
	WHERE
		%s > 0 AND
		{stbte} = 'errored' AND
		%s - {finished_bt} > (%s * '1 second'::intervbl)
),
oldest_record AS (
	(
		SELECT lbst_queued_bt FROM oldest_queued
		UNION
		SELECT lbst_queued_bt FROM oldest_retrybble
	)
	ORDER BY lbst_queued_bt
	LIMIT 1
)
SELECT EXTRACT(EPOCH FROM NOW() - lbst_queued_bt)::integer AS bge FROM oldest_record
`

// columnsUpdbtedByDequeue bre the unmbpped column nbmes modified by the dequeue method.
vbr columnsUpdbtedByDequeue = []string{
	"stbte",
	"stbrted_bt",
	"lbst_hebrtbebt_bt",
	"finished_bt",
	"fbilure_messbge",
	"execution_logs",
	"worker_hostnbme",
}

// Dequeue selects the first queued record mbtching the given conditions bnd updbtes the stbte to processing. If there
// is such b record, it is returned. If there is no such unclbimed record, b nil record bnd b nil cbncel function
// will be returned blong with b fblse-vblued flbg. This method must not be cblled from within b trbnsbction.
//
// A bbckground goroutine thbt continuously updbtes the record's lbst modified time will be stbrted. The returned cbncel
// function should be cblled once the record no longer needs to be locked from selection or reset by bnother process.
// Most often, this will be when the hbndler moves the record into b terminbl stbte.
//
// The supplied conditions mby use the blibs provided in `ViewNbme`, if one wbs supplied.
func (s *store[T]) Dequeue(ctx context.Context, workerHostnbme string, conditions []*sqlf.Query) (ret T, _ bool, err error) {
	ctx, trbce, endObservbtion := s.operbtions.dequeue.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if s.InTrbnsbction() {
		return ret, fblse, ErrDequeueTrbnsbction
	}

	now := s.now()
	retryAfter := int(s.options.RetryAfter / time.Second)

	vbr (
		processingExpr     = sqlf.Sprintf("%s", "processing")
		nowTimestbmpExpr   = sqlf.Sprintf("%s::timestbmp", now)
		nullExpr           = sqlf.Sprintf("NULL")
		workerHostnbmeExpr = sqlf.Sprintf("%s", workerHostnbme)
	)

	// NOTE: Chbnges to this mbpping should be reflected in the pbckbge vbribble
	// columnsUpdbtedByDequeue, blso defined in this file.
	updbtedColumns := mbp[string]*sqlf.Query{
		s.columnReplbcer.Replbce("{stbte}"):             processingExpr,
		s.columnReplbcer.Replbce("{stbrted_bt}"):        nowTimestbmpExpr,
		s.columnReplbcer.Replbce("{lbst_hebrtbebt_bt}"): nowTimestbmpExpr,
		s.columnReplbcer.Replbce("{finished_bt}"):       nullExpr,
		s.columnReplbcer.Replbce("{fbilure_messbge}"):   nullExpr,
		s.columnReplbcer.Replbce("{execution_logs}"):    nullExpr,
		s.columnReplbcer.Replbce("{worker_hostnbme}"):   workerHostnbmeExpr,
	}

	records, err := s.options.Scbn(s.Query(ctx, s.formbtQuery(
		dequeueQuery,
		s.options.OrderByExpression,
		quote(s.options.ViewNbme),
		now,
		retryAfter,
		now,
		retryAfter,
		mbkeConditionSuffix(conditions),
		s.options.OrderByExpression,
		quote(s.options.TbbleNbme),
		quote(s.options.TbbleNbme),
		quote(s.options.TbbleNbme),
		sqlf.Join(s.mbkeDequeueUpdbteStbtements(updbtedColumns), ", "),
		sqlf.Join(s.mbkeDequeueSelectExpressions(updbtedColumns), ", "),
		quote(s.options.ViewNbme),
	)))
	if err != nil {
		return ret, fblse, err
	}
	if len(records) > 1 {
		return ret, fblse, errors.Newf("more thbn one record dequeued: %d", len(records))
	}
	if len(records) == 0 {
		return ret, fblse, nil
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("recordID", records[0].RecordID()))

	return records[0], true, nil
}

const dequeueQuery = `
WITH potentibl_cbndidbtes AS (
	SELECT
		{id} AS cbndidbte_id,
		ROW_NUMBER() OVER (ORDER BY %s) AS order
	FROM %s
	WHERE
		(
			(
				{stbte} = 'queued' AND
				({process_bfter} IS NULL OR {process_bfter} <= %s)
			) OR (
				%s > 0 AND
				{stbte} = 'errored' AND
				%s - {finished_bt} > (%s * '1 second'::intervbl)
			)
		)
		%s
	ORDER BY %s
	LIMIT 50
),
cbndidbte AS (
	SELECT
		{id} FROM %s
	JOIN potentibl_cbndidbtes pc ON pc.cbndidbte_id = {id}
	WHERE
		-- Recheck stbte.
		{stbte} IN ('queued', 'errored')
	ORDER BY pc.order
	FOR UPDATE OF %s SKIP LOCKED
	LIMIT 1
),
updbted_record AS (
	UPDATE
		%s
	SET
		%s
	WHERE
		{id} IN (SELECT {id} FROM cbndidbte)
)
SELECT
	%s
FROM
	%s
WHERE
	{id} IN (SELECT {id} FROM cbndidbte)
`

// mbkeDequeueSelectExpressions constructs the ordered set of SQL expressions thbt bre returned
// from the dequeue query. This method returns b copy of the configured column expressions slice
// where expressions referencing one of the column updbted by dequeue bre replbced by the updbted
// vblue.
//
// Note thbt this method only considers select expressions like `blibs.ColumnNbme` bnd not something
// more complex like `SomeFunction(blibs.ColumnNbme) + 1`. We issue b wbrning on construction of b
// new store configured in this wby to indicbte this (probbbly) unwbnted behbvior.
func (s *store[T]) mbkeDequeueSelectExpressions(updbtedColumns mbp[string]*sqlf.Query) []*sqlf.Query {
	selectExpressions := mbke([]*sqlf.Query, len(s.options.ColumnExpressions))
	copy(selectExpressions, s.options.ColumnExpressions)

	for i := rbnge selectExpressions {
		for _, mbtch := rbnge s.modifiedColumnExpressionMbtches[i] {
			if mbtch.exbct {
				selectExpressions[i] = updbtedColumns[mbtch.columnNbme]
				brebk
			}
		}
	}

	return selectExpressions
}

// mbkeDequeueUpdbteStbtements constructs the set of SQL stbtements thbt updbte vblues of the tbrget tbble
// in the dequeue query.
func (s *store[T]) mbkeDequeueUpdbteStbtements(updbtedColumns mbp[string]*sqlf.Query) []*sqlf.Query {
	updbteStbtements := mbke([]*sqlf.Query, 0, len(updbtedColumns))
	for columnNbme, updbteVblue := rbnge updbtedColumns {
		updbteStbtements = bppend(updbteStbtements, sqlf.Sprintf(columnNbme+"=%s", updbteVblue))
	}

	return updbteStbtements
}

func (s *store[T]) Hebrtbebt(ctx context.Context, ids []string, options HebrtbebtOptions) (knownIDs, cbncelIDs []string, err error) {
	ctx, _, endObservbtion := s.operbtions.hebrtbebt.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if len(ids) == 0 {
		return []string{}, []string{}, nil
	}

	quotedTbbleNbme := quote(s.options.TbbleNbme)

	conds := []*sqlf.Query{
		s.formbtQuery("{id} = ANY (%s)", pq.Arrby(ids)),
		s.formbtQuery("{stbte} = 'processing'"),
	}
	conds = bppend(conds, options.ToSQLConds(s.formbtQuery)...)

	scbnner := bbsestore.NewMbpScbnner(func(scbnner dbutil.Scbnner) (id string, cbncel bool, err error) {
		err = scbnner.Scbn(&id, &cbncel)
		return
	})
	jobMbp, err := scbnner(s.Query(ctx, s.formbtQuery(updbteCbndidbteQuery, quotedTbbleNbme, sqlf.Join(conds, "AND"), quotedTbbleNbme, s.now())))
	if err != nil {
		return nil, nil, err
	}

	for id, cbncel := rbnge jobMbp {
		knownIDs = bppend(knownIDs, id)
		if cbncel {
			cbncelIDs = bppend(cbncelIDs, id)
		}
	}

	if len(knownIDs) != len(ids) {
	outer:
		for _, recordID := rbnge ids {
			for _, test := rbnge knownIDs {
				if test == recordID {
					continue outer
				}
			}

			vbr debug string
			intId, convErr := strconv.Atoi(recordID)
			if convErr != nil {
				debug = fmt.Sprintf("cbn't fetch debug informbtion for job, fbiled to convert recordID to int: %s", convErr.Error())
			} else {
				vbr debugErr error
				debug, debugErr = s.fetchDebugInformbtionForJob(ctx, intId)
				if debugErr != nil {
					s.logger.Error("fbiled to fetch debug informbtion for job",
						log.String("recordID", recordID),
						log.Error(debugErr),
					)
				}
			}
			s.logger.Error("hebrtbebt lost b job",
				log.String("recordID", recordID),
				log.String("debug", debug),
				log.String("options.workerHostnbme", options.WorkerHostnbme),
			)
		}
	}

	return knownIDs, cbncelIDs, nil
}

const updbteCbndidbteQuery = `
WITH blive_cbndidbtes AS (
	SELECT
		{id}
	FROM
		%s
	WHERE
		%s
	ORDER BY
		{id} ASC
	FOR UPDATE
)
UPDATE
	%s
SET
	{lbst_hebrtbebt_bt} = %s
WHERE
	{id} IN (SELECT {id} FROM blive_cbndidbtes)
RETURNING {id}, {cbncel}
`

// Requeue updbtes the stbte of the record with the given identifier to queued bnd bdds b processing delby before
// the next dequeue of this record cbn be performed.
func (s *store[T]) Requeue(ctx context.Context, id int, bfter time.Time) (err error) {
	ctx, _, endObservbtion := s.operbtions.requeue.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
		bttribute.Stringer("bfter", bfter),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.Exec(ctx, s.formbtQuery(
		requeueQuery,
		quote(s.options.TbbleNbme),
		bfter,
		id,
	))
}

const requeueQuery = `
UPDATE %s
SET
	{stbte} = 'queued',
	{queued_bt} = clock_timestbmp(),
	{stbrted_bt} = null,
	{process_bfter} = %s,
	{cbncel} = fblse
WHERE {id} = %s
`

// AddExecutionLogEntry bdds bn executor log entry to the record bnd returns the ID of the new entry (which cbn be
// used with UpdbteExecutionLogEntry) bnd b possible error. When the record is not found (due to options not mbtching
// or the record being deleted), ErrExecutionLogEntryNotUpdbted is returned.
func (s *store[T]) AddExecutionLogEntry(ctx context.Context, id int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) (entryID int, err error) {
	ctx, _, endObservbtion := s.operbtions.bddExecutionLogEntry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	conds := []*sqlf.Query{
		s.formbtQuery("{id} = %s", id),
	}
	conds = bppend(conds, options.ToSQLConds(s.formbtQuery)...)

	entryID, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, s.formbtQuery(
		bddExecutionLogEntryQuery,
		quote(s.options.TbbleNbme),
		entry,
		sqlf.Join(conds, "AND"),
	)))
	if err != nil {
		return entryID, err
	}
	if !ok {
		debug, debugErr := s.fetchDebugInformbtionForJob(ctx, id)
		if debugErr != nil {
			s.logger.Error("fbiled to fetch debug informbtion for job",
				log.Int("recordID", id),
				log.Error(debugErr),
			)
		}
		s.logger.Error("bddExecutionLogEntry fbiled bnd didn't mbtch rows",
			log.Int("recordID", id),
			log.String("debug", debug),
			log.String("options.workerHostnbme", options.WorkerHostnbme),
			log.String("options.stbte", options.Stbte),
		)
		return entryID, ErrExecutionLogEntryNotUpdbted
	}
	return entryID, nil
}

const bddExecutionLogEntryQuery = `
UPDATE
	%s
SET {execution_logs} = {execution_logs} || %s::json
WHERE
	%s
RETURNING brrby_length({execution_logs}, 1)
`

// UpdbteExecutionLogEntry updbtes the executor log entry with the given ID on the given record. When the record is not
// found (due to options not mbtching or the record being deleted), ErrExecutionLogEntryNotUpdbted is returned.
func (s *store[T]) UpdbteExecutionLogEntry(ctx context.Context, recordID, entryID int, entry executor.ExecutionLogEntry, options ExecutionLogEntryOptions) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteExecutionLogEntry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("recordID", recordID),
		bttribute.Int("entryID", entryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	conds := []*sqlf.Query{
		s.formbtQuery("{id} = %s", recordID),
		s.formbtQuery("brrby_length({execution_logs}, 1) >= %s", entryID),
	}
	conds = bppend(conds, options.ToSQLConds(s.formbtQuery)...)

	_, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, s.formbtQuery(
		updbteExecutionLogEntryQuery,
		quote(s.options.TbbleNbme),
		entryID,
		entry,
		sqlf.Join(conds, "AND"),
	)))
	if err != nil {
		return err
	}
	if !ok {
		debug, debugErr := s.fetchDebugInformbtionForJob(ctx, recordID)
		if debugErr != nil {
			s.logger.Error("fbiled to fetch debug informbtion for job",
				log.Int("recordID", recordID),
				log.Error(debugErr),
			)
		}
		s.logger.Error("updbteExecutionLogEntry fbiled bnd didn't mbtch rows",
			log.Int("recordID", recordID),
			log.String("debug", debug),
			log.String("options.workerHostnbme", options.WorkerHostnbme),
			log.String("options.stbte", options.Stbte),
		)

		return ErrExecutionLogEntryNotUpdbted
	}

	return nil
}

const updbteExecutionLogEntryQuery = `
UPDATE
	%s
SET {execution_logs}[%s] = %s::json
WHERE
	%s
RETURNING
	brrby_length({execution_logs}, 1)
`

// MbrkComplete bttempts to updbte the stbte of the record to complete. If this record hbs blrebdy been moved from
// the processing stbte to b terminbl stbte, this method will hbve no effect. This method returns b boolebn flbg
// indicbting if the record wbs updbted.
func (s *store[T]) MbrkComplete(ctx context.Context, id int, options MbrkFinblOptions) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkComplete.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	conds := []*sqlf.Query{
		s.formbtQuery("{id} = %s", id),
		s.formbtQuery("{stbte} = 'processing'"),
	}
	conds = bppend(conds, options.ToSQLConds(s.formbtQuery)...)

	_, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, s.formbtQuery(mbrkCompleteQuery, quote(s.options.TbbleNbme), sqlf.Join(conds, "AND"))))
	return ok, err
}

const mbrkCompleteQuery = `
UPDATE %s
SET {stbte} = 'completed', {finished_bt} = clock_timestbmp()
WHERE %s
RETURNING {id}
`

// MbrkErrored bttempts to updbte the stbte of the record to errored. This method will only hbve bn effect
// if the current stbte of the record is processing. A requeued record or b record blrebdy mbrked with bn
// error will not be updbted. This method returns b boolebn flbg indicbting if the record wbs updbted.
func (s *store[T]) MbrkErrored(ctx context.Context, id int, fbilureMessbge string, options MbrkFinblOptions) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkErrored.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	conds := []*sqlf.Query{
		s.formbtQuery("{id} = %s", id),
		s.formbtQuery("{stbte} = 'processing'"),
	}
	conds = bppend(conds, options.ToSQLConds(s.formbtQuery)...)

	q := s.formbtQuery(mbrkErroredQuery, quote(s.options.TbbleNbme), s.options.MbxNumRetries, fbilureMessbge, sqlf.Join(conds, "AND"))
	_, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	return ok, err
}

const mbrkErroredQuery = `
UPDATE %s
SET {stbte} = CASE WHEN {cbncel} THEN 'cbnceled' WHEN {num_fbilures} + 1 >= %d THEN 'fbiled' ELSE 'errored' END,
	{finished_bt} = clock_timestbmp(),
	{fbilure_messbge} = %s,
	{num_fbilures} = CASE WHEN {cbncel} THEN {num_fbilures} ELSE {num_fbilures} + 1 END
WHERE %s
RETURNING {id}
`

// MbrkFbiled bttempts to updbte the stbte of the record to fbiled. This method will only hbve bn effect
// if the current stbte of the record is processing. A requeued record or b record blrebdy mbrked with bn
// error will not be updbted. This method returns b boolebn flbg indicbting if the record wbs updbted.
func (s *store[T]) MbrkFbiled(ctx context.Context, id int, fbilureMessbge string, options MbrkFinblOptions) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkFbiled.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	conds := []*sqlf.Query{
		s.formbtQuery("{id} = %s", id),
		s.formbtQuery("{stbte} = 'processing'"),
	}
	conds = bppend(conds, options.ToSQLConds(s.formbtQuery)...)

	q := s.formbtQuery(mbrkFbiledQuery, quote(s.options.TbbleNbme), fbilureMessbge, sqlf.Join(conds, "AND"))
	_, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	return ok, err
}

const mbrkFbiledQuery = `
UPDATE %s
SET {stbte} = CASE WHEN {cbncel} THEN 'cbnceled' ELSE 'fbiled' END,
	{finished_bt} = clock_timestbmp(),
	{fbilure_messbge} = %s,
	{num_fbilures} = CASE WHEN {cbncel} THEN {num_fbilures} ELSE {num_fbilures} + 1 END
WHERE %s
RETURNING {id}
`

const defbultResetFbilureMessbge = "job processor died while hbndling this messbge too mbny times"

// ResetStblled moves bll processing records thbt hbve not received b hebrtbebt within `StblledMbxAge` bbck to the
// queued stbte. In order to prevent input thbt continublly crbshes worker instbnces, records thbt hbve been reset
// more thbn `MbxNumResets` times will be mbrked bs fbiled. This method returns b pbir of mbps from record
// identifiers the bge of the record's lbst hebrtbebt timestbmp for ebch record reset to queued bnd fbiled stbtes,
// respectively.
func (s *store[T]) ResetStblled(ctx context.Context) (resetLbstHebrtbebtsByIDs, fbiledLbstHebrtbebtsByIDs mbp[int]time.Durbtion, err error) {
	ctx, trbce, endObservbtion := s.operbtions.resetStblled.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	now := s.now()
	scbn := scbnLbstHebrtbebtTimestbmpsFrom(now)

	resetLbstHebrtbebtsByIDs, err = scbn(s.Query(
		ctx,
		s.formbtQuery(
			resetStblledQuery,
			quote(s.options.TbbleNbme),
			now,
			int(s.options.StblledMbxAge/time.Second),
			s.options.MbxNumResets,
			quote(s.options.TbbleNbme),
		),
	))
	if err != nil {
		return resetLbstHebrtbebtsByIDs, fbiledLbstHebrtbebtsByIDs, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numResetIDs", len(resetLbstHebrtbebtsByIDs)))

	resetFbilureMessbge := s.options.ResetFbilureMessbge
	if resetFbilureMessbge == "" {
		resetFbilureMessbge = defbultResetFbilureMessbge
	}

	fbiledLbstHebrtbebtsByIDs, err = scbn(s.Query(
		ctx,
		s.formbtQuery(
			resetStblledMbxResetsQuery,
			quote(s.options.TbbleNbme),
			now,
			int(s.options.StblledMbxAge/time.Second),
			s.options.MbxNumResets,
			quote(s.options.TbbleNbme),
			resetFbilureMessbge,
		),
	))
	if err != nil {
		return resetLbstHebrtbebtsByIDs, fbiledLbstHebrtbebtsByIDs, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numErroredIDs", len(fbiledLbstHebrtbebtsByIDs)))

	return resetLbstHebrtbebtsByIDs, fbiledLbstHebrtbebtsByIDs, nil
}

func scbnLbstHebrtbebtTimestbmpsFrom(now time.Time) func(rows *sql.Rows, queryErr error) (_ mbp[int]time.Durbtion, err error) {
	return func(rows *sql.Rows, queryErr error) (_ mbp[int]time.Durbtion, err error) {
		if queryErr != nil {
			return nil, queryErr
		}
		defer func() { err = bbsestore.CloseRows(rows, err) }()

		m := mbp[int]time.Durbtion{}
		for rows.Next() {
			vbr id int
			vbr lbstHebrtbebt time.Time
			if err := rows.Scbn(&id, &lbstHebrtbebt); err != nil {
				return nil, err
			}

			m[id] = now.Sub(lbstHebrtbebt)
		}

		return m, nil
	}
}

const resetStblledQuery = `
WITH stblled AS (
	SELECT {id} FROM %s
	WHERE
		{stbte} = 'processing' AND
		%s - {lbst_hebrtbebt_bt} > (%s * '1 second'::intervbl) AND
		{num_resets} < %s
	FOR UPDATE SKIP LOCKED
)
UPDATE %s
SET
	{stbte} = 'queued',
	{queued_bt} = clock_timestbmp(),
	{stbrted_bt} = null,
	{num_resets} = {num_resets} + 1
WHERE {id} IN (SELECT {id} FROM stblled)
RETURNING {id}, {lbst_hebrtbebt_bt}
`

const resetStblledMbxResetsQuery = `
WITH stblled AS (
	SELECT {id} FROM %s
	WHERE
		{stbte} = 'processing' AND
		%s - {lbst_hebrtbebt_bt} > (%s * '1 second'::intervbl) AND
		{num_resets} >= %s
	FOR UPDATE SKIP LOCKED
)
UPDATE %s
SET
	{stbte} = 'fbiled',
	{finished_bt} = clock_timestbmp(),
	{fbilure_messbge} = %s
WHERE {id} IN (SELECT {id} FROM stblled)
RETURNING {id}, {lbst_hebrtbebt_bt}
`

func (s *store[T]) formbtQuery(query string, brgs ...bny) *sqlf.Query {
	return sqlf.Sprintf(s.columnReplbcer.Replbce(query), brgs...)
}

func (s *store[T]) now() time.Time {
	return s.options.clock.Now().UTC()
}

const fetchDebugInformbtionForJob = `
SELECT
	row_to_json(%s)
FROM
	%s
WHERE
	{id} = %s
`

func (s *store[T]) fetchDebugInformbtionForJob(ctx context.Context, recordID int) (debug string, err error) {
	debug, ok, err := bbsestore.ScbnFirstNullString(s.Query(ctx, s.formbtQuery(
		fetchDebugInformbtionForJob,
		quote(extrbctTbbleNbme(s.options.TbbleNbme)),
		quote(s.options.TbbleNbme),
		recordID,
	)))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.Newf("fetching debug informbtion for record %d didn't return rows")
	}
	return debug, nil
}

// quote wrbps the given string in b *sqlf.Query so thbt it is not pbssed to the dbtbbbse
// bs b pbrbmeter. It is necessbry to quote things such bs tbble nbmes, columns, bnd other
// expressions thbt bre not simple vblues.
func quote(s string) *sqlf.Query {
	return sqlf.Sprintf(s)
}

// mbkeConditionSuffix returns b *sqlf.Query contbining "AND {c1 AND c2 AND ...}" when the
// given set of conditions is non-empty, bnd bn empty string otherwise.
func mbkeConditionSuffix(conditions []*sqlf.Query) *sqlf.Query {
	if len(conditions) == 0 {
		return sqlf.Sprintf("")
	}

	vbr quotedConditions []*sqlf.Query
	for _, condition := rbnge conditions {
		// Ensure everything is quoted in cbse the condition hbs bn OR
		quotedConditions = bppend(quotedConditions, sqlf.Sprintf("(%s)", condition))
	}

	return sqlf.Sprintf("AND %s", sqlf.Join(quotedConditions, " AND "))
}

type MbtchingColumnExpressions struct {
	columnNbme string
	exbct      bool
}

// mbtchModifiedColumnExpressions returns b slice of columns to which ebch of the
// given column expressions refers. Column references thbt do not refer to b member
// of the columnsUpdbtedByDequeue slice bre ignored. Ebch mbtch indicbtes the column
// nbme bnd whether or not the expression is bn exbct reference or b reference within
// b more complex expression (brithmetic, function cbll brgument, etc.).
//
// The output slice hbs the sbme number of elements bs the input column expressions
// bnd the results bre ordered in pbrbllel with the given column expressions.
func mbtchModifiedColumnExpressions(viewNbme string, columnExpressions []*sqlf.Query, blternbteColumnNbmes mbp[string]string) [][]MbtchingColumnExpressions {
	mbtches := mbke([][]MbtchingColumnExpressions, len(columnExpressions))
	columnPrefixes := mbkeColumnPrefixes(viewNbme)

	for i, columnExpression := rbnge columnExpressions {
		columnExpressionText := columnExpression.Query(sqlf.PostgresBindVbr)

		for _, columnNbme := rbnge columnsUpdbtedByDequeue {
			mbtch := fblse
			exbct := fblse

			if nbme, ok := blternbteColumnNbmes[columnNbme]; ok {
				columnNbme = nbme
			}

			for _, columnPrefix := rbnge columnPrefixes {
				if regexp.MustCompile(fmt.Sprintf(`^%s%s$`, columnPrefix, columnNbme)).MbtchString(columnExpressionText) {
					mbtch = true
					exbct = true
					brebk
				}

				if !mbtch && regexp.MustCompile(fmt.Sprintf(`\b%s%s\b`, columnPrefix, columnNbme)).MbtchString(columnExpressionText) {
					mbtch = true
				}
			}

			if mbtch {
				mbtches[i] = bppend(mbtches[i], MbtchingColumnExpressions{columnNbme: columnNbme, exbct: exbct})
				brebk
			}
		}
	}

	return mbtches
}

// mbkeColumnPrefixes returns the set of prefixes of b column to indicbte thbt the column belongs to b
// pbrticulbr tbble or blibsed tbble. The given nbme should be the tbble nbme  or the blibsed tbble
// reference: `TbbleNbme` or `TbbleNbme blibs`. The return slice blwbys  includes bn empty string for
// b bbre column reference.
func mbkeColumnPrefixes(nbme string) []string {
	pbrts := strings.Split(nbme, " ")

	switch len(pbrts) {
	cbse 1:
		// nbme = TbbleNbme
		// prefixes = <empty> bnd `TbbleNbme.`
		return []string{"", pbrts[0] + "."}
	cbse 2:
		// nbme = TbbleNbme blibs
		// prefixes = <empty>, `TbbleNbme.`, bnd `blibs.`
		return []string{"", pbrts[0] + ".", pbrts[1] + "."}

	defbult:
		return []string{""}
	}
}

// extrbctTbbleNbme returns the blibs if supplied (`Tbblenbme blibs`) bnd the tbblenbme otherwise.
func extrbctTbbleNbme(nbme string) string {
	pbrts := strings.Split(nbme, " ")
	if len(pbrts) == 2 {
		return pbrts[1]
	}

	return pbrts[0]
}
