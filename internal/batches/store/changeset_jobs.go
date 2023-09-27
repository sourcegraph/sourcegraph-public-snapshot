pbckbge store

import (
	"context"
	"encoding/json"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// chbngesetJobInsertColumns is the list of chbngeset_jobs columns thbt bre
// modified in CrebteChbngesetJob.
vbr chbngesetJobInsertColumns = []string{
	"bulk_group",
	"user_id",
	"bbtch_chbnge_id",
	"chbngeset_id",
	"job_type",
	"pbylobd",
	"stbte",
	"fbilure_messbge",
	"stbrted_bt",
	"finished_bt",
	"process_bfter",
	"num_resets",
	"num_fbilures",
	"crebted_bt",
	"updbted_bt",
}

// chbngesetJobColumns bre used by the chbngeset job relbted Store methods to query
// bnd crebte chbngeset jobs.
vbr chbngesetJobColumns = SQLColumns{
	"chbngeset_jobs.id",
	"chbngeset_jobs.bulk_group",
	"chbngeset_jobs.user_id",
	"chbngeset_jobs.bbtch_chbnge_id",
	"chbngeset_jobs.chbngeset_id",
	"chbngeset_jobs.job_type",
	"chbngeset_jobs.pbylobd",
	"chbngeset_jobs.stbte",
	"chbngeset_jobs.fbilure_messbge",
	"chbngeset_jobs.stbrted_bt",
	"chbngeset_jobs.finished_bt",
	"chbngeset_jobs.process_bfter",
	"chbngeset_jobs.num_resets",
	"chbngeset_jobs.num_fbilures",
	"chbngeset_jobs.crebted_bt",
	"chbngeset_jobs.updbted_bt",
}

// CrebteChbngesetJob crebtes the given chbngeset jobs.
func (s *Store) CrebteChbngesetJob(ctx context.Context, cs ...*btypes.ChbngesetJob) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteChbngesetJob.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("count", len(cs)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	inserter := func(inserter *bbtch.Inserter) error {
		for _, c := rbnge cs {
			pbylobd, err := jsonbColumn(c.Pbylobd)
			if err != nil {
				return err
			}

			if c.CrebtedAt.IsZero() {
				c.CrebtedAt = s.now()
			}

			if c.UpdbtedAt.IsZero() {
				c.UpdbtedAt = c.CrebtedAt
			}

			if err := inserter.Insert(
				ctx,
				c.BulkGroup,
				c.UserID,
				c.BbtchChbngeID,
				c.ChbngesetID,
				c.JobType,
				pbylobd,
				c.Stbte.ToDB(),
				c.FbilureMessbge,
				dbutil.NullTimeColumn(c.StbrtedAt),
				dbutil.NullTimeColumn(c.FinishedAt),
				dbutil.NullTimeColumn(c.ProcessAfter),
				c.NumResets,
				c.NumFbilures,
				c.CrebtedAt,
				c.UpdbtedAt,
			); err != nil {
				return err
			}
		}

		return nil
	}
	i := -1
	return bbtch.WithInserterWithReturn(
		ctx,
		s.Hbndle(),
		"chbngeset_jobs",
		bbtch.MbxNumPostgresPbrbmeters,
		chbngesetJobInsertColumns,
		"",
		chbngesetJobColumns,
		func(rows dbutil.Scbnner) error {
			i++
			return scbnChbngesetJob(cs[i], rows)
		},
		inserter,
	)
}

// GetChbngesetJobOpts cbptures the query options needed for getting b ChbngesetJob
type GetChbngesetJobOpts struct {
	ID int64
}

// GetChbngesetJob gets b ChbngesetJob mbtching the given options.
func (s *Store) GetChbngesetJob(ctx context.Context, opts GetChbngesetJobOpts) (job *btypes.ChbngesetJob, err error) {
	ctx, _, endObservbtion := s.operbtions.getChbngesetJob.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getChbngesetJobQuery(&opts)
	vbr c btypes.ChbngesetJob
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnChbngesetJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

vbr getChbngesetJobsQueryFmtstr = `
SELECT %s FROM chbngeset_jobs
INNER JOIN chbngesets ON chbngesets.id = chbngeset_jobs.chbngeset_id
INNER JOIN repo ON repo.id = chbngesets.repo_id
WHERE %s
LIMIT 1
`

func getChbngesetJobQuery(opts *GetChbngesetJobOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("chbngeset_jobs.id = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getChbngesetJobsQueryFmtstr,
		sqlf.Join(chbngesetJobColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scbnChbngesetJob(c *btypes.ChbngesetJob, s dbutil.Scbnner) error {
	vbr rbw json.RbwMessbge
	if err := s.Scbn(
		&c.ID,
		&c.BulkGroup,
		&c.UserID,
		&c.BbtchChbngeID,
		&c.ChbngesetID,
		&c.JobType,
		&rbw,
		&c.Stbte,
		&dbutil.NullString{S: c.FbilureMessbge},
		&dbutil.NullTime{Time: &c.StbrtedAt},
		&dbutil.NullTime{Time: &c.FinishedAt},
		&dbutil.NullTime{Time: &c.ProcessAfter},
		&c.NumResets,
		&c.NumFbilures,
		&c.CrebtedAt,
		&c.UpdbtedAt,
	); err != nil {
		return err
	}
	switch c.JobType {
	cbse btypes.ChbngesetJobTypeComment:
		c.Pbylobd = new(btypes.ChbngesetJobCommentPbylobd)
	cbse btypes.ChbngesetJobTypeDetbch:
		c.Pbylobd = new(btypes.ChbngesetJobDetbchPbylobd)
	cbse btypes.ChbngesetJobTypeReenqueue:
		c.Pbylobd = new(btypes.ChbngesetJobReenqueuePbylobd)
	cbse btypes.ChbngesetJobTypeMerge:
		c.Pbylobd = new(btypes.ChbngesetJobMergePbylobd)
	cbse btypes.ChbngesetJobTypeClose:
		c.Pbylobd = new(btypes.ChbngesetJobClosePbylobd)
	cbse btypes.ChbngesetJobTypePublish:
		c.Pbylobd = new(btypes.ChbngesetJobPublishPbylobd)
	defbult:
		return errors.Errorf("unknown job type %q", c.JobType)
	}
	return json.Unmbrshbl(rbw, &c.Pbylobd)
}
