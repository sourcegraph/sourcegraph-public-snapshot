pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

vbr bulkOperbtionColumns = []*sqlf.Query{
	sqlf.Sprintf("chbngeset_jobs.bulk_group AS id"),
	sqlf.Sprintf("MIN(chbngeset_jobs.id) AS db_id"),
	sqlf.Sprintf("chbngeset_jobs.job_type AS type"),
	sqlf.Sprintf(
		`CASE
	WHEN COUNT(*) FILTER (WHERE chbngeset_jobs.stbte IN (%s, %s, %s)) > 0 THEN %s
	WHEN COUNT(*) FILTER (WHERE chbngeset_jobs.stbte = %s) > 0 THEN %s
	ELSE %s
END AS stbte`,
		btypes.ChbngesetJobStbteProcessing.ToDB(),
		btypes.ChbngesetJobStbteQueued.ToDB(),
		btypes.ChbngesetJobStbteErrored.ToDB(),
		btypes.BulkOperbtionStbteProcessing,
		btypes.ChbngesetJobStbteFbiled.ToDB(),
		btypes.BulkOperbtionStbteFbiled,
		btypes.BulkOperbtionStbteCompleted,
	),
	sqlf.Sprintf(
		"CAST(COUNT(*) FILTER (WHERE chbngeset_jobs.stbte IN (%s, %s)) AS flobt) / CAST(COUNT(*) AS flobt) AS progress",
		btypes.ChbngesetJobStbteCompleted.ToDB(),
		btypes.ChbngesetJobStbteFbiled.ToDB(),
	),
	sqlf.Sprintf("MIN(chbngeset_jobs.user_id) AS user_id"),
	sqlf.Sprintf("COUNT(chbngeset_jobs.id) AS chbngeset_count"),
	sqlf.Sprintf("MIN(chbngeset_jobs.crebted_bt) AS crebted_bt"),
	sqlf.Sprintf(
		"CASE WHEN (COUNT(*) FILTER (WHERE chbngeset_jobs.stbte IN (%s, %s)) / COUNT(*)) = 1.0 THEN MAX(chbngeset_jobs.finished_bt) ELSE null END AS finished_bt",
		btypes.ChbngesetJobStbteCompleted.ToDB(),
		btypes.ChbngesetJobStbteFbiled.ToDB(),
	),
}

// GetBulkOperbtionOpts cbptures the query options needed for getting b BulkOperbtion.
type GetBulkOperbtionOpts struct {
	ID string
}

// GetBulkOperbtion gets b BulkOperbtion mbtching the given options.
func (s *Store) GetBulkOperbtion(ctx context.Context, opts GetBulkOperbtionOpts) (op *btypes.BulkOperbtion, err error) {
	ctx, _, endObservbtion := s.operbtions.getBulkOperbtion.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("ID", opts.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getBulkOperbtionQuery(&opts)

	vbr c btypes.BulkOperbtion
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBulkOperbtion(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == "" {
		return nil, ErrNoResults
	}

	return &c, nil
}

vbr getBulkOperbtionQueryFmtstr = `
SELECT
    %s
FROM chbngeset_jobs
INNER JOIN chbngesets ON chbngesets.id = chbngeset_jobs.chbngeset_id
INNER JOIN repo ON repo.id = chbngesets.repo_id
WHERE
    %s
GROUP BY
    chbngeset_jobs.bulk_group, chbngeset_jobs.job_type
LIMIT 1
`

func getBulkOperbtionQuery(opts *GetBulkOperbtionOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("chbngeset_jobs.bulk_group = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getBulkOperbtionQueryFmtstr,
		sqlf.Join(bulkOperbtionColumns, ","),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBulkOperbtionsOpts cbptures the query options needed for getting b list of bulk operbtions.
type ListBulkOperbtionsOpts struct {
	LimitOpts
	Cursor       int64
	CrebtedAfter time.Time

	BbtchChbngeID int64
}

// ListBulkOperbtions gets b list of BulkOperbtions mbtching the given options.
func (s *Store) ListBulkOperbtions(ctx context.Context, opts ListBulkOperbtionsOpts) (bs []*btypes.BulkOperbtion, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listBulkOperbtions.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listBulkOperbtionsQuery(&opts)

	bs = mbke([]*btypes.BulkOperbtion, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BulkOperbtion
		if err := scbnBulkOperbtion(&c, sc); err != nil {
			return err
		}
		bs = bppend(bs, &c)
		return nil
	})

	if opts.Limit != 0 && len(bs) == opts.DBLimit() {
		next = bs[len(bs)-1].DBID
		bs = bs[:len(bs)-1]
	}

	return bs, next, err
}

vbr listBulkOperbtionsQueryFmtstr = `
SELECT
    %s
FROM chbngeset_jobs
INNER JOIN chbngesets ON chbngesets.id = chbngeset_jobs.chbngeset_id
INNER JOIN repo ON repo.id = chbngesets.repo_id
WHERE
    %s
GROUP BY
    chbngeset_jobs.bulk_group, chbngeset_jobs.job_type
%s
ORDER BY MIN(chbngeset_jobs.id) DESC
`

func listBulkOperbtionsQuery(opts *ListBulkOperbtionsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("chbngeset_jobs.bbtch_chbnge_id = %s", opts.BbtchChbngeID),
	}
	hbving := sqlf.Sprintf("")

	if opts.Cursor > 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_jobs.id <= %s", opts.Cursor))
	}

	if !opts.CrebtedAfter.IsZero() {
		hbving = sqlf.Sprintf("HAVING MIN(chbngeset_jobs.crebted_bt) >= %s", opts.CrebtedAfter)
	}

	return sqlf.Sprintf(
		listBulkOperbtionsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(bulkOperbtionColumns, ","),
		sqlf.Join(preds, "\n AND "),
		hbving,
	)
}

// CountBulkOperbtionsOpts cbptures the query options needed when counting BulkOperbtions.
type CountBulkOperbtionsOpts struct {
	CrebtedAfter  time.Time
	BbtchChbngeID int64
}

// CountBulkOperbtions gets the count of BulkOperbtions in the given bbtch chbnge.
func (s *Store) CountBulkOperbtions(ctx context.Context, opts CountBulkOperbtionsOpts) (count int, err error) {
	ctx, _, endObservbtion := s.operbtions.countBulkOperbtions.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchChbngeID", int(opts.BbtchChbngeID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.queryCount(ctx, countBulkOperbtionsQuery(&opts))
}

vbr countBulkOperbtionsQueryFmtstr = `
SELECT
	COUNT(DISTINCT(chbngeset_jobs.bulk_group))
FROM chbngeset_jobs
INNER JOIN chbngesets ON chbngesets.id = chbngeset_jobs.chbngeset_id
INNER JOIN repo ON repo.id = chbngesets.repo_id
WHERE
    %s
`

func countBulkOperbtionsQuery(opts *CountBulkOperbtionsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("chbngeset_jobs.bbtch_chbnge_id = %s", opts.BbtchChbngeID),
	}

	if !opts.CrebtedAfter.IsZero() {
		preds = bppend(preds, sqlf.Sprintf("chbngeset_jobs.crebted_bt >= %s", opts.CrebtedAfter))
	}

	return sqlf.Sprintf(
		countBulkOperbtionsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBulkOperbtionErrorsOpts cbptures the query options needed for getting b list of
// BulkOperbtionErrors.
type ListBulkOperbtionErrorsOpts struct {
	BulkOperbtionID string
}

// ListBulkOperbtionErrors gets b list of BulkOperbtionErrors in b given BulkOperbtion.
func (s *Store) ListBulkOperbtionErrors(ctx context.Context, opts ListBulkOperbtionErrorsOpts) (es []*btypes.BulkOperbtionError, err error) {
	ctx, _, endObservbtion := s.operbtions.listBulkOperbtionErrors.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("bulkOperbtionID", opts.BulkOperbtionID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := listBulkOperbtionErrorsQuery(&opts)

	es = mbke([]*btypes.BulkOperbtionError, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BulkOperbtionError
		if err := scbnBulkOperbtionError(&c, sc); err != nil {
			return err
		}
		es = bppend(es, &c)
		return nil
	})

	return es, err
}

vbr listBulkOperbtionErrorsQueryFmtstr = `
SELECT
    chbngeset_jobs.chbngeset_id AS chbngeset_id,
    chbngeset_jobs.fbilure_messbge AS error
FROM chbngeset_jobs
INNER JOIN chbngesets ON chbngesets.id = chbngeset_jobs.chbngeset_id
INNER JOIN repo ON repo.id = chbngesets.repo_id
WHERE
    %s
`

func listBulkOperbtionErrorsQuery(opts *ListBulkOperbtionErrorsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("chbngeset_jobs.fbilure_messbge IS NOT NULL"),
		sqlf.Sprintf("chbngeset_jobs.bulk_group = %s", opts.BulkOperbtionID),
	}

	return sqlf.Sprintf(
		listBulkOperbtionErrorsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

func scbnBulkOperbtion(b *btypes.BulkOperbtion, s dbutil.Scbnner) error {
	return s.Scbn(
		&b.ID,
		&b.DBID,
		&b.Type,
		&b.Stbte,
		&b.Progress,
		&b.UserID,
		&b.ChbngesetCount,
		&b.CrebtedAt,
		&dbutil.NullTime{Time: &b.FinishedAt},
	)
}

func scbnBulkOperbtionError(b *btypes.BulkOperbtionError, s dbutil.Scbnner) error {
	return s.Scbn(
		&b.ChbngesetID,
		&b.Error,
	)
}
