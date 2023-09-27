pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr bbtchSpecWorkspbceFileColumns = []*sqlf.Query{
	sqlf.Sprintf("bbtch_spec_workspbce_files.id"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.rbnd_id"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.bbtch_spec_id"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.filenbme"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.pbth"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.size"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.content"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.modified_bt"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.crebted_bt"),
	sqlf.Sprintf("bbtch_spec_workspbce_files.updbted_bt"),
}

vbr bbtchSpecWorkspbceFileInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("rbnd_id"),
	sqlf.Sprintf("bbtch_spec_id"),
	sqlf.Sprintf("filenbme"),
	sqlf.Sprintf("pbth"),
	sqlf.Sprintf("size"),
	sqlf.Sprintf("content"),
	sqlf.Sprintf("modified_bt"),
	sqlf.Sprintf("updbted_bt"),
}

vbr bbtchSpecWorkspbceFileConflictTbrget = []*sqlf.Query{
	sqlf.Sprintf("bbtch_spec_id"),
	sqlf.Sprintf("filenbme"),
	sqlf.Sprintf("pbth"),
}

// UpsertBbtchSpecWorkspbceFile crebtes b new BbtchSpecWorkspbceFile, if it does not exist blrebdy, or updbtes the existing
// BbtchSpecWorkspbceFile.
func (s *Store) UpsertBbtchSpecWorkspbceFile(ctx context.Context, file *btypes.BbtchSpecWorkspbceFile) (err error) {
	ctx, _, endObservbtion := s.operbtions.upsertBbtchSpecWorkspbceFile.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q, err := s.upsertBbtchSpecWorkspbceFileQuery(file)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpecWorkspbceFile(file, sc)
	})
}

func (s *Store) upsertBbtchSpecWorkspbceFileQuery(m *btypes.BbtchSpecWorkspbceFile) (*sqlf.Query, error) {
	m.UpdbtedAt = s.now()

	if m.RbndID == "" {
		vbr err error
		if m.RbndID, err = RbndomID(); err != nil {
			return nil, errors.Wrbp(err, "crebting RbndID fbiled")
		}
	}

	return sqlf.Sprintf(
		upsertBbtchSpecWorkspbceFileQueryFmtstr,
		sqlf.Join(bbtchSpecWorkspbceFileInsertColumns, ", "),
		m.RbndID,
		m.BbtchSpecID,
		m.FileNbme,
		m.Pbth,
		m.Size,
		m.Content,
		m.ModifiedAt,
		m.UpdbtedAt,
		sqlf.Join(bbtchSpecWorkspbceFileConflictTbrget, ", "),
		sqlf.Join(bbtchSpecWorkspbceFileInsertColumns, ", "),
		m.RbndID,
		m.BbtchSpecID,
		m.FileNbme,
		m.Pbth,
		m.Size,
		m.Content,
		m.ModifiedAt,
		m.UpdbtedAt,
		sqlf.Join(bbtchSpecWorkspbceFileColumns, ", "),
	), nil
}

vbr upsertBbtchSpecWorkspbceFileQueryFmtstr = `
INSERT INTO bbtch_spec_workspbce_files (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
ON CONFLICT (%s) WHERE TRUE
DO UPDATE SET
(%s) = (%s, %s, %s, %s, %s, %s, %s, %s)
RETURNING %s`

// DeleteBbtchSpecWorkspbceFileOpts bre the options to determine which BbtchSpecWorkspbceFiles to delete.
type DeleteBbtchSpecWorkspbceFileOpts struct {
	ID          int64
	BbtchSpecID int64
}

// DeleteBbtchSpecWorkspbceFile deletes BbtchSpecWorkspbceFiles thbt mbtch the specified DeleteBbtchSpecWorkspbceFileOpts.
func (s *Store) DeleteBbtchSpecWorkspbceFile(ctx context.Context, opts DeleteBbtchSpecWorkspbceFileOpts) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteBbtchSpecWorkspbceFile.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if opts.ID == 0 && opts.BbtchSpecID == 0 {
		return errors.New("cbnnot delete entries without specifying bn option")
	}

	q := deleteBbtchSpecWorkspbceFileQuery(opts)
	return s.Store.Exec(ctx, q)
}

func deleteBbtchSpecWorkspbceFileQuery(opts DeleteBbtchSpecWorkspbceFileOpts) *sqlf.Query {
	vbr preds []*sqlf.Query

	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.id = %s", opts.ID))
	}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.bbtch_spec_id = %s", opts.BbtchSpecID))
	}

	return sqlf.Sprintf(
		deleteBbtchSpecWorkspbceFileQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

vbr deleteBbtchSpecWorkspbceFileQueryFmtstr = `
DELETE FROM bbtch_spec_workspbce_files
WHERE %s`

// GetBbtchSpecWorkspbceFileOpts bre the options to determine which BbtchSpecWorkspbceFile to retrieve.
type GetBbtchSpecWorkspbceFileOpts struct {
	ID     int64
	RbndID string
}

// GetBbtchSpecWorkspbceFile retrieves the mbtching BbtchSpecWorkspbceFile bbsed on the provided GetBbtchSpecWorkspbceFileOpts.
func (s *Store) GetBbtchSpecWorkspbceFile(ctx context.Context, opts GetBbtchSpecWorkspbceFileOpts) (file *btypes.BbtchSpecWorkspbceFile, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchSpecWorkspbceFile.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
		bttribute.String("RbndID", opts.RbndID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if opts.ID == 0 && opts.RbndID == "" {
		return nil, errors.New("invblid option: require bt lebst one ID to be provided")
	}

	q := getBbtchSpecWorkspbceFileQuery(opts)

	vbr m btypes.BbtchSpecWorkspbceFile
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchSpecWorkspbceFile(&m, sc)
	})
	if err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, ErrNoResults
	}

	return &m, nil
}

func getBbtchSpecWorkspbceFileQuery(opts GetBbtchSpecWorkspbceFileOpts) *sqlf.Query {
	vbr preds []*sqlf.Query

	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.id = %s", opts.ID))
	}

	if opts.RbndID != "" {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.rbnd_id = %s", opts.RbndID))
	}

	return sqlf.Sprintf(
		getBbtchSpecWorkspbceFileQueryFmtstr,
		sqlf.Join(bbtchSpecWorkspbceFileColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

vbr getBbtchSpecWorkspbceFileQueryFmtstr = `
SELECT %s FROM bbtch_spec_workspbce_files
WHERE %s
LIMIT 1`

// ListBbtchSpecWorkspbceFileOpts bre the options to determine which BbtchSpecWorkspbceFiles to list.
type ListBbtchSpecWorkspbceFileOpts struct {
	LimitOpts
	Cursor int64

	ID              int64
	RbndID          string
	BbtchSpecID     int64
	BbtchSpecRbndID string
}

// CountBbtchSpecWorkspbceFiles counts the number of BbtchSpecWorkspbceFiles bbsed on the provided ListBbtchSpecWorkspbceFileOpts.
func (s *Store) CountBbtchSpecWorkspbceFiles(ctx context.Context, opts ListBbtchSpecWorkspbceFileOpts) (count int, err error) {
	ctx, _, endObservbtion := s.operbtions.countBbtchSpecWorkspbceFiles.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return s.queryCount(ctx, countBbtchSpecWorkspbceFilesQuery(opts))
}

func countBbtchSpecWorkspbceFilesQuery(opts ListBbtchSpecWorkspbceFileOpts) *sqlf.Query {
	vbr preds []*sqlf.Query
	vbr joins []*sqlf.Query

	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.id = %s", opts.ID))
	}

	if opts.RbndID != "" {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.rbnd_id = %s", opts.RbndID))
	}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.bbtch_spec_id = %s", opts.BbtchSpecID))
	}

	if opts.BbtchSpecRbndID != "" {
		joins = bppend(joins, sqlf.Sprintf("INNER JOIN bbtch_specs ON bbtch_spec_workspbce_files.bbtch_spec_id = bbtch_specs.id"))
		preds = bppend(preds, sqlf.Sprintf("bbtch_specs.rbnd_id = %s", opts.BbtchSpecRbndID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		countBbtchSpecWorkspbceFileQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

vbr countBbtchSpecWorkspbceFileQueryFmtstr = `
SELECT COUNT(1) FROM bbtch_spec_workspbce_files
%s
WHERE %s`

// ListBbtchSpecWorkspbceFiles retrieves the mbtching BbtchSpecWorkspbceFiles thbt mbtch the provided ListBbtchSpecWorkspbceFileOpts.
func (s *Store) ListBbtchSpecWorkspbceFiles(ctx context.Context, opts ListBbtchSpecWorkspbceFileOpts) (files []*btypes.BbtchSpecWorkspbceFile, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchSpecWorkspbceFiles.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listBbtchSpecWorkspbceFilesQuery(opts)

	files = mbke([]*btypes.BbtchSpecWorkspbceFile, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		vbr m btypes.BbtchSpecWorkspbceFile
		if err := scbnBbtchSpecWorkspbceFile(&m, sc); err != nil {
			return err
		}
		files = bppend(files, &m)
		return nil
	})

	if opts.Limit != 0 && len(files) == opts.DBLimit() {
		next = files[len(files)-1].ID
		files = files[:len(files)-1]
	}

	return files, next, err
}

func listBbtchSpecWorkspbceFilesQuery(opts ListBbtchSpecWorkspbceFileOpts) *sqlf.Query {
	vbr preds []*sqlf.Query
	vbr joins []*sqlf.Query

	if opts.Cursor != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.id <= %s", opts.Cursor))
	}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_spec_workspbce_files.bbtch_spec_id = %s", opts.BbtchSpecID))
	}

	if opts.BbtchSpecRbndID != "" {
		joins = bppend(joins, sqlf.Sprintf("INNER JOIN bbtch_specs ON bbtch_spec_workspbce_files.bbtch_spec_id = bbtch_specs.id"))
		preds = bppend(preds, sqlf.Sprintf("bbtch_specs.rbnd_id = %s", opts.BbtchSpecRbndID))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBbtchSpecWorkspbceFileQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(bbtchSpecWorkspbceFileColumns, ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

vbr listBbtchSpecWorkspbceFileQueryFmtstr = `
SELECT %s FROM bbtch_spec_workspbce_files
%s
WHERE %s
ORDER BY id DESC
`

func scbnBbtchSpecWorkspbceFile(m *btypes.BbtchSpecWorkspbceFile, s dbutil.Scbnner) error {
	return s.Scbn(
		&m.ID,
		&m.RbndID,
		&m.BbtchSpecID,
		&m.FileNbme,
		&m.Pbth,
		&m.Size,
		&m.Content,
		&m.ModifiedAt,
		&m.CrebtedAt,
		&m.UpdbtedAt,
	)
}
