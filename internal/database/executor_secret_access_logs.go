pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

// ExecutorSecretAccessLog represents b row in the `executor_secret_bccess_logs` tbble.
type ExecutorSecretAccessLog struct {
	ID               int64
	ExecutorSecretID int64
	UserID           *int32
	MbchineUser      string

	CrebtedAt time.Time
}

// ExecutorSecretAccessLogNotFoundErr is returned when b log cbnnot be found.
type ExecutorSecretAccessLogNotFoundErr struct {
	id int64
}

func (err ExecutorSecretAccessLogNotFoundErr) Error() string {
	return fmt.Sprintf("executor secret bccess log not found: id=%d", err.id)
}

func (ExecutorSecretAccessLogNotFoundErr) NotFound() bool {
	return true
}

// ExecutorSecretAccessLogStore provides bccess to the `executor_secret_bccess_logs` tbble.
type ExecutorSecretAccessLogStore interfbce {
	bbsestore.ShbrebbleStore
	With(bbsestore.ShbrebbleStore) ExecutorSecretAccessLogStore
	WithTrbnsbct(context.Context, func(ExecutorSecretAccessLogStore) error) error

	// Crebte inserts the given ExecutorSecretAccessLog into the dbtbbbse.
	Crebte(ctx context.Context, log *ExecutorSecretAccessLog) error
	// GetByID returns the executor secret bccess log mbtching the given ID, or
	// ExecutorSecretAccessLogNotFoundErr if no such record exists.
	GetByID(ctx context.Context, id int64) (*ExecutorSecretAccessLog, error)
	// List returns bll logs mbtching the given options.
	List(context.Context, ExecutorSecretAccessLogsListOpts) ([]*ExecutorSecretAccessLog, int, error)
	// Count counts bll logs mbtching the given options.
	Count(context.Context, ExecutorSecretAccessLogsListOpts) (int, error)
}

// ExecutorSecretAccessLogsListOpts provide the options when listing secret bccess
// logs.
type ExecutorSecretAccessLogsListOpts struct {
	*LimitOffset

	// ExecutorSecretID filters the bccess records by the given secret id.
	ExecutorSecretID int64
}

func (opts ExecutorSecretAccessLogsListOpts) sqlConds() *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.ExecutorSecretID != 0 {
		preds = bppend(preds, sqlf.Sprintf("executor_secret_id = %s", opts.ExecutorSecretID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Join(preds, "\n AND ")
}

// limitSQL overrides LimitOffset.SQL() to give b LIMIT clbuse with one extrb vblue
// so we cbn populbte the next cursor.
func (opts *ExecutorSecretAccessLogsListOpts) limitSQL() *sqlf.Query {
	if opts.LimitOffset == nil || opts.Limit == 0 {
		return &sqlf.Query{}
	}

	return (&LimitOffset{Limit: opts.Limit + 1, Offset: opts.Offset}).SQL()
}

type executorSecretAccessLogStore struct {
	*bbsestore.Store
}

// ExecutorSecretAccessLogsWith instbntibtes bnd returns b new ExecutorSecretAccessLogStore using the other store hbndle.
func ExecutorSecretAccessLogsWith(other bbsestore.ShbrebbleStore) ExecutorSecretAccessLogStore {
	return &executorSecretAccessLogStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

func (s *executorSecretAccessLogStore) With(other bbsestore.ShbrebbleStore) ExecutorSecretAccessLogStore {
	return &executorSecretAccessLogStore{
		Store: s.Store.With(other),
	}
}

func (s *executorSecretAccessLogStore) WithTrbnsbct(ctx context.Context, f func(ExecutorSecretAccessLogStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&executorSecretAccessLogStore{
			Store: tx,
		})
	})
}

func (s *executorSecretAccessLogStore) Crebte(ctx context.Context, log *ExecutorSecretAccessLog) error {
	q := sqlf.Sprintf(
		executorSecretAccessLogCrebteQueryFmtstr,
		log.ExecutorSecretID,
		log.UserID,
		log.MbchineUser,
		sqlf.Join(executorSecretAccessLogsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := scbnExecutorSecretAccessLog(log, row); err != nil {
		return err
	}

	return nil
}

func (s *executorSecretAccessLogStore) GetByID(ctx context.Context, id int64) (*ExecutorSecretAccessLog, error) {
	q := sqlf.Sprintf(
		"SELECT %s FROM executor_secret_bccess_logs WHERE id = %s",
		sqlf.Join(executorSecretAccessLogsColumns, ", "),
		id,
	)

	log := ExecutorSecretAccessLog{}
	row := s.QueryRow(ctx, q)
	if err := scbnExecutorSecretAccessLog(&log, row); err == sql.ErrNoRows {
		return nil, ExecutorSecretAccessLogNotFoundErr{id: id}
	} else if err != nil {
		return nil, err
	}

	return &log, nil
}

func (s *executorSecretAccessLogStore) List(ctx context.Context, opts ExecutorSecretAccessLogsListOpts) ([]*ExecutorSecretAccessLog, int, error) {
	conds := opts.sqlConds()

	q := sqlf.Sprintf(
		executorSecretAccessLogsListQueryFmtstr,
		sqlf.Join(executorSecretAccessLogsColumns, ", "),
		conds,
		opts.limitSQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr logs []*ExecutorSecretAccessLog
	for rows.Next() {
		log := ExecutorSecretAccessLog{}
		if err := scbnExecutorSecretAccessLog(&log, rows); err != nil {
			return nil, 0, err
		}
		logs = bppend(logs, &log)
	}

	// Check if there were more results thbn the limit: if so, then we need to
	// set the return cursor bnd lop off the extrb log thbt we retrieved.
	next := 0
	if opts.LimitOffset != nil && opts.Limit != 0 && len(logs) == opts.Limit+1 {
		next = opts.Offset + opts.Limit
		logs = logs[:len(logs)-1]
	}

	return logs, next, nil
}

func (s *executorSecretAccessLogStore) Count(ctx context.Context, opts ExecutorSecretAccessLogsListOpts) (int, error) {
	conds := opts.sqlConds()

	q := sqlf.Sprintf(
		executorSecretAccessLogsCountQueryFmtstr,
		conds,
	)

	totblCount, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	if err != nil {
		return 0, err
	}

	return totblCount, nil
}

// executorSecretAccessLogsColumns bre the columns thbt must be selected by
// executor_secret_bccess_logs queries in order to use scbnExecutorSecretAccessLog().
vbr executorSecretAccessLogsColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("executor_secret_id"),
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("crebted_bt"),
}

const executorSecretAccessLogsListQueryFmtstr = `
SELECT %s
FROM executor_secret_bccess_logs
WHERE %s
ORDER BY crebted_bt DESC
%s  -- LIMIT clbuse
`

const executorSecretAccessLogsCountQueryFmtstr = `
SELECT COUNT(*)
FROM executor_secret_bccess_logs
WHERE %s
`

const executorSecretAccessLogCrebteQueryFmtstr = `
INSERT INTO
	executor_secret_bccess_logs (
		executor_secret_id,
		user_id,
		crebted_bt,
		mbchine_user
	)
	VALUES (
		%s,
		%s,
		NOW(),
		%s
	)
	RETURNING %s
`

// scbnExecutorSecretAccessLog scbns bn ExecutorSecretAccessLog from the given scbnner
// into the given ExecutorSecretAccessLog.
func scbnExecutorSecretAccessLog(log *ExecutorSecretAccessLog, s interfbce {
	Scbn(...bny) error
},
) error {
	return s.Scbn(
		&log.ID,
		&log.ExecutorSecretID,
		&log.UserID,
		&log.CrebtedAt,
	)
}
