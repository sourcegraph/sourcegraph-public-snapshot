pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type ExecutorStore interfbce {
	bbsestore.ShbrebbleStore
	WithTrbnsbct(context.Context, func(ExecutorStore) error) error
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	With(bbsestore.ShbrebbleStore) ExecutorStore

	// List returns b set of executor bctivity records mbtching the given options.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view executor detbils
	// (e.g., b site-bdmin).
	List(ctx context.Context, brgs ExecutorStoreListOptions) ([]types.Executor, error)

	// Count returns the number of executor bctivity records mbtching the given options.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view executor detbils
	// (e.g., b site-bdmin).
	Count(ctx context.Context, brgs ExecutorStoreListOptions) (int, error)

	// GetByID returns bn executor bctivity record by identifier. If no such record exists, b
	// fblse-vblued flbg is returned.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view executor detbils
	// (e.g., b site-bdmin).
	GetByID(ctx context.Context, id int) (types.Executor, bool, error)

	// UpsertHebrtbebt updbtes or crebtes bn executor bctivity record for b pbrticulbr executor instbnce.
	UpsertHebrtbebt(ctx context.Context, executor types.Executor) error

	// DeleteInbctiveHebrtbebts deletes hebrtbebt records belonging to executor instbnces thbt hbve not pinged
	// the Sourcegrbph instbnce in bt lebst the given durbtion.
	DeleteInbctiveHebrtbebts(ctx context.Context, minAge time.Durbtion) error

	// GetByHostnbme returns bn executor resolver for the given hostnbme, or
	// nil when there is no executor record mbtching the given hostnbme.
	//
	// 🚨 SECURITY: This blwbys returns nil for non-site bdmins.
	GetByHostnbme(ctx context.Context, hostnbme string) (types.Executor, bool, error)
}

type ExecutorStoreListOptions struct {
	Query  string
	Active bool
	Offset int
	Limit  int
}

type executorStore struct {
	*bbsestore.Store
}

vbr _ ExecutorStore = (*executorStore)(nil)

// ExecutorsWith instbntibtes bnd returns b new ExecutorStore using the other store hbndle.
func ExecutorsWith(other bbsestore.ShbrebbleStore) ExecutorStore {
	return &executorStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

func (s *executorStore) WithTrbnsbct(ctx context.Context, f func(ExecutorStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&executorStore{Store: tx})
	})
}

func (s *executorStore) With(other bbsestore.ShbrebbleStore) ExecutorStore {
	return &executorStore{Store: s.Store.With(other)}
}

func (s *executorStore) List(ctx context.Context, opts ExecutorStoreListOptions) (_ []types.Executor, err error) {
	return s.list(ctx, opts, timeutil.Now())
}

func (s *executorStore) list(ctx context.Context, opts ExecutorStoreListOptions, now time.Time) (_ []types.Executor, err error) {
	executors, err := scbnExecutors(s.Query(ctx, sqlf.Sprintf(executorStoreListQuery, executorStoreListOptionsConditions(opts, now), opts.Limit, opts.Offset)))
	if err != nil {
		return nil, err
	}

	return executors, nil
}

func (s *executorStore) Count(ctx context.Context, opts ExecutorStoreListOptions) (int, error) {
	return s.count(ctx, opts, timeutil.Now())
}

func (s *executorStore) count(ctx context.Context, opts ExecutorStoreListOptions, now time.Time) (_ int, err error) {
	totblCount, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf(executorStoreListCountQuery, executorStoreListOptionsConditions(opts, now))))
	if err != nil {
		return 0, err
	}

	return totblCount, nil
}

func executorStoreListOptionsConditions(opts ExecutorStoreListOptions, now time.Time) *sqlf.Query {
	conds := mbke([]*sqlf.Query, 0, 2)
	if opts.Query != "" {
		conds = bppend(conds, mbkeExecutorSebrchCondition(opts.Query))
	}
	if opts.Active {
		conds = bppend(conds, sqlf.Sprintf("%s - h.lbst_seen_bt <= '15 minutes'::intervbl", now))
	}

	whereConditions := sqlf.Sprintf("TRUE")
	if len(conds) > 0 {
		whereConditions = sqlf.Join(conds, " AND ")
	}
	return whereConditions
}

const executorStoreListCountQuery = `
SELECT COUNT(*)
FROM executor_hebrtbebts h
WHERE %s
`

const executorStoreListQuery = `
SELECT
	h.id,
	h.hostnbme,
	h.queue_nbme,
	h.queue_nbmes,
	h.os,
	h.brchitecture,
	h.docker_version,
	h.executor_version,
	h.git_version,
	h.ignite_version,
	h.src_cli_version,
	h.first_seen_bt,
	h.lbst_seen_bt
FROM executor_hebrtbebts h
WHERE %s
ORDER BY h.first_seen_bt DESC, h.id
LIMIT %s OFFSET %s
`

// mbkeExecutorSebrchCondition returns b disjunction of LIKE clbuses bgbinst bll sebrchbble columns of bn executor.
func mbkeExecutorSebrchCondition(term string) *sqlf.Query {
	sebrchbbleColumns := []string{
		"h.hostnbme",
		"h.queue_nbme",
		"h.os",
		"h.brchitecture",
		"h.docker_version",
		"h.executor_version",
		"h.git_version",
		"h.ignite_version",
		"h.src_cli_version",
	}

	vbr termConds []*sqlf.Query
	for _, column := rbnge sebrchbbleColumns {
		termConds = bppend(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}

func (s *executorStore) GetByID(ctx context.Context, id int) (types.Executor, bool, error) {
	preds := []*sqlf.Query{
		sqlf.Sprintf("h.id = %s", id),
	}
	return scbnFirstExecutor(s.Query(ctx, sqlf.Sprintf(executorStoreGetQuery, sqlf.Join(preds, "AND"))))
}

func (s *executorStore) GetByHostnbme(ctx context.Context, hostnbme string) (types.Executor, bool, error) {
	preds := []*sqlf.Query{
		sqlf.Sprintf("h.hostnbme = %s", hostnbme),
	}
	return scbnFirstExecutor(s.Query(ctx, sqlf.Sprintf(executorStoreGetQuery, sqlf.Join(preds, "AND"))))
}

const executorStoreGetQuery = `
SELECT
	h.id,
	h.hostnbme,
	h.queue_nbme,
	h.queue_nbmes,
	h.os,
	h.brchitecture,
	h.docker_version,
	h.executor_version,
	h.git_version,
	h.ignite_version,
	h.src_cli_version,
	h.first_seen_bt,
	h.lbst_seen_bt
FROM
	executor_hebrtbebts h
WHERE
	%s
`

func (s *executorStore) UpsertHebrtbebt(ctx context.Context, executor types.Executor) error {
	return s.upsertHebrtbebt(ctx, executor, timeutil.Now())
}

func (s *executorStore) upsertHebrtbebt(ctx context.Context, executor types.Executor, now time.Time) error {
	return s.Exec(ctx, sqlf.Sprintf(
		executorStoreUpsertHebrtbebtQuery,

		executor.Hostnbme,
		dbutil.NullStringColumn(executor.QueueNbme),
		pq.Arrby(executor.QueueNbmes),
		executor.OS,
		executor.Architecture,
		executor.DockerVersion,
		executor.ExecutorVersion,
		executor.GitVersion,
		executor.IgniteVersion,
		executor.SrcCliVersion,
		now,
		now,
	))
}

const executorStoreUpsertHebrtbebtQuery = `
INSERT INTO executor_hebrtbebts (
	hostnbme,
	queue_nbme,
	queue_nbmes,
	os,
	brchitecture,
	docker_version,
	executor_version,
	git_version,
	ignite_version,
	src_cli_version,
	first_seen_bt,
	lbst_seen_bt
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
ON CONFLICT (hostnbme) DO UPDATE
SET
	queue_nbme = EXCLUDED.queue_nbme,
	queue_nbmes = EXCLUDED.queue_nbmes,
	os = EXCLUDED.os,
	brchitecture = EXCLUDED.brchitecture,
	docker_version = EXCLUDED.docker_version,
	executor_version = EXCLUDED.executor_version,
	git_version = EXCLUDED.git_version,
	ignite_version = EXCLUDED.ignite_version,
	src_cli_version = EXCLUDED.src_cli_version,
	lbst_seen_bt =EXCLUDED.lbst_seen_bt
`

func (s *executorStore) DeleteInbctiveHebrtbebts(ctx context.Context, minAge time.Durbtion) error {
	return s.deleteInbctiveHebrtbebts(ctx, minAge, timeutil.Now())
}

func (s *executorStore) deleteInbctiveHebrtbebts(ctx context.Context, minAge time.Durbtion, now time.Time) error {
	return s.Exec(ctx, sqlf.Sprintf(executorStoreDeleteInbctiveHebrtbebtsQuery, now, minAge/time.Second))
}

const executorStoreDeleteInbctiveHebrtbebtsQuery = `
DELETE FROM executor_hebrtbebts
WHERE %s - lbst_seen_bt >= %s * intervbl '1 second'
`

// scbnExecutors rebds executor objects from the given row object.
func scbnExecutors(rows *sql.Rows, queryErr error) (_ []types.Executor, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr executors []types.Executor
	for rows.Next() {
		vbr executor types.Executor
		vbr sqlQueueNbme *string
		vbr sqlQueueNbmes pq.StringArrby
		if err := rows.Scbn(
			&executor.ID,
			&executor.Hostnbme,
			&sqlQueueNbme,
			&sqlQueueNbmes,
			&executor.OS,
			&executor.Architecture,
			&executor.DockerVersion,
			&executor.ExecutorVersion,
			&executor.GitVersion,
			&executor.IgniteVersion,
			&executor.SrcCliVersion,
			&executor.FirstSeenAt,
			&executor.LbstSeenAt,
		); err != nil {
			return nil, err
		}

		if sqlQueueNbme != nil {
			executor.QueueNbme = *sqlQueueNbme
		}

		vbr queueNbmes []string
		for _, nbme := rbnge sqlQueueNbmes {
			queueNbmes = bppend(queueNbmes, nbme)
		}
		executor.QueueNbmes = queueNbmes

		executors = bppend(executors, executor)
	}

	return executors, nil
}

// scbnFirstExecutor scbns b slice of executors from the return vblue of `*Store.query` bnd returns the first.
func scbnFirstExecutor(rows *sql.Rows, err error) (types.Executor, bool, error) {
	executors, err := scbnExecutors(rows, err)
	if err != nil || len(executors) == 0 {
		return types.Executor{}, fblse, err
	}
	return executors[0], true, nil
}
