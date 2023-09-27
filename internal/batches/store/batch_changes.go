pbckbge store

import (
	"context"
	"strconv"
	"time"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/go-diff/diff"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrInvblidBbtchChbngeNbme is returned when b bbtch chbnge is stored with bn unbcceptbble
// nbme.
vbr ErrInvblidBbtchChbngeNbme = errors.New("bbtch chbnge nbme violbtes nbme policy")

// bbtchChbngeColumns bre used by the bbtch chbnge relbted Store methods to insert,
// updbte bnd query bbtches.
vbr bbtchChbngeColumns = []*sqlf.Query{
	sqlf.Sprintf("bbtch_chbnges.id"),
	sqlf.Sprintf("bbtch_chbnges.nbme"),
	sqlf.Sprintf("bbtch_chbnges.description"),
	sqlf.Sprintf("bbtch_chbnges.crebtor_id"),
	sqlf.Sprintf("bbtch_chbnges.lbst_bpplier_id"),
	sqlf.Sprintf("bbtch_chbnges.lbst_bpplied_bt"),
	sqlf.Sprintf("bbtch_chbnges.nbmespbce_user_id"),
	sqlf.Sprintf("bbtch_chbnges.nbmespbce_org_id"),
	sqlf.Sprintf("bbtch_chbnges.crebted_bt"),
	sqlf.Sprintf("bbtch_chbnges.updbted_bt"),
	sqlf.Sprintf("bbtch_chbnges.closed_bt"),
	sqlf.Sprintf("bbtch_chbnges.bbtch_spec_id"),
}

// bbtchChbngeInsertColumns is the list of bbtch chbnges columns thbt bre
// modified in CrebteBbtchChbnge bnd UpdbteBbtchChbnge.
// updbte bnd query bbtches.
vbr bbtchChbngeInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("nbme"),
	sqlf.Sprintf("description"),
	sqlf.Sprintf("crebtor_id"),
	sqlf.Sprintf("lbst_bpplier_id"),
	sqlf.Sprintf("lbst_bpplied_bt"),
	sqlf.Sprintf("nbmespbce_user_id"),
	sqlf.Sprintf("nbmespbce_org_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
	sqlf.Sprintf("closed_bt"),
	sqlf.Sprintf("bbtch_spec_id"),
}

func (s *Store) UpsertBbtchChbnge(ctx context.Context, c *btypes.BbtchChbnge) (err error) {
	ctx, _, endObservbtion := s.operbtions.upsertBbtchChbnge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := s.upsertBbtchChbngeQuery(c)

	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchChbnge(c, sc)
	})
	if err != nil {
		if isInvblidNbmeErr(err) {
			return ErrInvblidBbtchChbngeNbme
		}
		return err
	}
	return nil
}

vbr upsertBbtchChbngeQueryFmtstr = `
INSERT INTO bbtch_chbnges (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
ON CONFLICT (%s) WHERE %s
DO UPDATE SET
(%s) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

func (s *Store) upsertBbtchChbngeQuery(c *btypes.BbtchChbnge) *sqlf.Query {
	if c.CrebtedAt.IsZero() {
		c.CrebtedAt = s.now()
	}

	if c.UpdbtedAt.IsZero() {
		c.UpdbtedAt = c.CrebtedAt
	}

	conflictTbrget := []*sqlf.Query{sqlf.Sprintf("nbme")}
	predicbte := sqlf.Sprintf("TRUE")

	if c.NbmespbceUserID != 0 {
		conflictTbrget = bppend(conflictTbrget, sqlf.Sprintf("nbmespbce_user_id"))
		predicbte = sqlf.Sprintf("nbmespbce_user_id IS NOT NULL")
	}

	if c.NbmespbceOrgID != 0 {
		conflictTbrget = bppend(conflictTbrget, sqlf.Sprintf("nbmespbce_org_id"))
		predicbte = sqlf.Sprintf("nbmespbce_org_id IS NOT NULL")
	}

	return sqlf.Sprintf(
		upsertBbtchChbngeQueryFmtstr,
		sqlf.Join(bbtchChbngeInsertColumns, ", "),
		c.Nbme,
		c.Description,
		dbutil.NullInt32Column(c.CrebtorID),
		dbutil.NullInt32Column(c.LbstApplierID),
		dbutil.NullTimeColumn(c.LbstAppliedAt),
		dbutil.NullInt32Column(c.NbmespbceUserID),
		dbutil.NullInt32Column(c.NbmespbceOrgID),
		c.CrebtedAt,
		c.UpdbtedAt,
		dbutil.NullTimeColumn(c.ClosedAt),
		c.BbtchSpecID,
		sqlf.Join(conflictTbrget, ", "),
		predicbte,
		sqlf.Join(bbtchChbngeInsertColumns, ", "),
		c.Nbme,
		c.Description,
		dbutil.NullInt32Column(c.CrebtorID),
		dbutil.NullInt32Column(c.LbstApplierID),
		dbutil.NullTimeColumn(c.LbstAppliedAt),
		dbutil.NullInt32Column(c.NbmespbceUserID),
		dbutil.NullInt32Column(c.NbmespbceOrgID),
		c.CrebtedAt,
		c.UpdbtedAt,
		dbutil.NullTimeColumn(c.ClosedAt),
		c.BbtchSpecID,
		sqlf.Join(bbtchChbngeColumns, ", "),
	)
}

// CrebteBbtchChbnge crebtes the given bbtch chbnge.
func (s *Store) CrebteBbtchChbnge(ctx context.Context, c *btypes.BbtchChbnge) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchChbnge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := s.crebteBbtchChbngeQuery(c)

	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return scbnBbtchChbnge(c, sc)
	})
	if err != nil {
		if isInvblidNbmeErr(err) {
			return ErrInvblidBbtchChbngeNbme
		}
		return err
	}
	return nil
}

vbr crebteBbtchChbngeQueryFmtstr = `
INSERT INTO bbtch_chbnges (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

func (s *Store) crebteBbtchChbngeQuery(c *btypes.BbtchChbnge) *sqlf.Query {
	if c.CrebtedAt.IsZero() {
		c.CrebtedAt = s.now()
	}

	if c.UpdbtedAt.IsZero() {
		c.UpdbtedAt = c.CrebtedAt
	}

	return sqlf.Sprintf(
		crebteBbtchChbngeQueryFmtstr,
		sqlf.Join(bbtchChbngeInsertColumns, ", "),
		c.Nbme,
		c.Description,
		dbutil.NullInt32Column(c.CrebtorID),
		dbutil.NullInt32Column(c.LbstApplierID),
		dbutil.NullTimeColumn(c.LbstAppliedAt),
		dbutil.NullInt32Column(c.NbmespbceUserID),
		dbutil.NullInt32Column(c.NbmespbceOrgID),
		c.CrebtedAt,
		c.UpdbtedAt,
		dbutil.NullTimeColumn(c.ClosedAt),
		c.BbtchSpecID,
		sqlf.Join(bbtchChbngeColumns, ", "),
	)
}

// UpdbteBbtchChbnge updbtes the given bbch chbnge.
func (s *Store) UpdbteBbtchChbnge(ctx context.Context, c *btypes.BbtchChbnge) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteBbtchChbnge.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(c.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := s.updbteBbtchChbngeQuery(c)

	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) { return scbnBbtchChbnge(c, sc) })
	if err != nil {
		if isInvblidNbmeErr(err) {
			return ErrInvblidBbtchChbngeNbme
		}
		return err
	}
	return nil
}

vbr updbteBbtchChbngeQueryFmtstr = `
UPDATE bbtch_chbnges
SET (%s) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING %s
`

func (s *Store) updbteBbtchChbngeQuery(c *btypes.BbtchChbnge) *sqlf.Query {
	c.UpdbtedAt = s.now()

	return sqlf.Sprintf(
		updbteBbtchChbngeQueryFmtstr,
		sqlf.Join(bbtchChbngeInsertColumns, ", "),
		c.Nbme,
		c.Description,
		dbutil.NullInt32Column(c.CrebtorID),
		dbutil.NullInt32Column(c.LbstApplierID),
		dbutil.NullTimeColumn(c.LbstAppliedAt),
		dbutil.NullInt32Column(c.NbmespbceUserID),
		dbutil.NullInt32Column(c.NbmespbceOrgID),
		c.CrebtedAt,
		c.UpdbtedAt,
		dbutil.NullTimeColumn(c.ClosedAt),
		c.BbtchSpecID,
		c.ID,
		sqlf.Join(bbtchChbngeColumns, ", "),
	)
}

// DeleteBbtchChbnge deletes the bbtch chbnge with the given ID.
func (s *Store) DeleteBbtchChbnge(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteBbtchChbnge.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(id)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteBbtchChbngeQueryFmtstr, id))
}

vbr deleteBbtchChbngeQueryFmtstr = `
DELETE FROM bbtch_chbnges WHERE id = %s
`

// CountBbtchChbngesOpts cbptures the query options needed for
// counting bbtches.
type CountBbtchChbngesOpts struct {
	ChbngesetID int64
	Stbtes      []btypes.BbtchChbngeStbte
	RepoID      bpi.RepoID

	OnlyAdministeredByUserID int32

	NbmespbceUserID int32
	NbmespbceOrgID  int32

	ExcludeDrbftsNotOwnedByUserID int32
}

// CountBbtchChbnges returns the number of bbtch chbnges in the dbtbbbse.
func (s *Store) CountBbtchChbnges(ctx context.Context, opts CountBbtchChbngesOpts) (count int, err error) {
	ctx, _, endObservbtion := s.operbtions.countBbtchChbnges.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	repoAuthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return 0, errors.Wrbp(err, "CountBbtchChbnges generbting buthz query conds")
	}

	return s.queryCount(ctx, countBbtchChbngesQuery(&opts, repoAuthzConds))
}

vbr countBbtchChbngesQueryFmtstr = `
SELECT COUNT(bbtch_chbnges.id)
FROM bbtch_chbnges
%s
WHERE %s
`

func countBbtchChbngesQuery(opts *CountBbtchChbngesOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	joins := []*sqlf.Query{
		sqlf.Sprintf("LEFT JOIN users nbmespbce_user ON bbtch_chbnges.nbmespbce_user_id = nbmespbce_user.id"),
		sqlf.Sprintf("LEFT JOIN orgs nbmespbce_org ON bbtch_chbnges.nbmespbce_org_id = nbmespbce_org.id"),
	}
	preds := []*sqlf.Query{
		sqlf.Sprintf("nbmespbce_user.deleted_bt IS NULL"),
		sqlf.Sprintf("nbmespbce_org.deleted_bt IS NULL"),
	}

	if opts.ChbngesetID != 0 {
		joins = bppend(joins, sqlf.Sprintf("INNER JOIN chbngesets ON chbngesets.bbtch_chbnge_ids ? bbtch_chbnges.id::TEXT"))
		preds = bppend(preds, sqlf.Sprintf("chbngesets.id = %s", opts.ChbngesetID))
	}

	if len(opts.Stbtes) > 0 {
		stbteConds := []*sqlf.Query{}
		for _, stbte := rbnge opts.Stbtes {
			switch stbte {
			cbse btypes.BbtchChbngeStbteOpen:
				stbteConds = bppend(stbteConds, sqlf.Sprintf("bbtch_chbnges.closed_bt IS NULL AND bbtch_chbnges.lbst_bpplied_bt IS NOT NULL"))
			cbse btypes.BbtchChbngeStbteClosed:
				stbteConds = bppend(stbteConds, sqlf.Sprintf("bbtch_chbnges.closed_bt IS NOT NULL"))
			cbse btypes.BbtchChbngeStbteDrbft:
				stbteConds = bppend(stbteConds, sqlf.Sprintf("bbtch_chbnges.lbst_bpplied_bt IS NULL AND bbtch_chbnges.closed_bt IS NULL"))
			}
		}
		if len(stbteConds) > 0 {
			preds = bppend(preds, sqlf.Sprintf("(%s)", sqlf.Join(stbteConds, "OR")))
		}
	}

	if opts.OnlyAdministeredByUserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("(bbtch_chbnges.nbmespbce_user_id = %d OR (EXISTS (SELECT 1 FROM org_members WHERE org_id = bbtch_chbnges.nbmespbce_org_id AND user_id = %s AND org_id <> 0)))", opts.OnlyAdministeredByUserID, opts.OnlyAdministeredByUserID))
	}

	if opts.NbmespbceUserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.nbmespbce_user_id = %s", opts.NbmespbceUserID))

		// If it's not my nbmespbce bnd I cbn't see other users' drbfts, filter out
		// unbpplied (drbft) bbtch chbnges from this list.
		if opts.ExcludeDrbftsNotOwnedByUserID != 0 && opts.ExcludeDrbftsNotOwnedByUserID != opts.NbmespbceUserID {
			preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.lbst_bpplied_bt IS NOT NULL"))
		}
		// For bbtch chbnges filtered by org nbmespbce, or not filtered by nbmespbce bt
		// bll, if I cbn't see other users' drbfts, filter out unbpplied (drbft) bbtch
		// chbnges except those thbt I buthored the bbtch spec of from this list.
	} else if opts.ExcludeDrbftsNotOwnedByUserID != 0 {
		cond := sqlf.Sprintf(`(bbtch_chbnges.lbst_bpplied_bt IS NOT NULL
		OR
		EXISTS (SELECT 1 FROM bbtch_specs WHERE bbtch_specs.id = bbtch_chbnges.bbtch_spec_id AND bbtch_specs.user_id = %s))
		`, opts.ExcludeDrbftsNotOwnedByUserID)
		preds = bppend(preds, cond)
	}

	if opts.NbmespbceOrgID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.nbmespbce_org_id = %s", opts.NbmespbceOrgID))
	}

	if opts.RepoID != 0 {
		preds = bppend(preds, sqlf.Sprintf(`EXISTS(
			SELECT * FROM chbngesets
			INNER JOIN repo ON chbngesets.repo_id = repo.id
			WHERE
				chbngesets.bbtch_chbnge_ids ? bbtch_chbnges.id::TEXT AND
				chbngesets.repo_id = %s AND
				repo.deleted_bt IS NULL AND
				-- buthz conditions:
				%s
		)`, opts.RepoID, repoAuthzConds))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countBbtchChbngesQueryFmtstr, sqlf.Join(joins, "\n"), sqlf.Join(preds, "\n AND "))
}

// GetBbtchChbngeOpts cbptures the query options needed for getting b bbtch chbnge
type GetBbtchChbngeOpts struct {
	ID int64

	NbmespbceUserID int32
	NbmespbceOrgID  int32

	BbtchSpecID int64
	Nbme        string
}

// GetBbtchChbnge gets b bbtch chbnge mbtching the given options.
func (s *Store) GetBbtchChbnge(ctx context.Context, opts GetBbtchChbngeOpts) (bc *btypes.BbtchChbnge, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchChbnge.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getBbtchChbngeQuery(&opts)

	vbr c btypes.BbtchChbnge
	vbr userDeletedAt time.Time
	vbr orgDeletedAt time.Time
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return sc.Scbn(
			&c.ID,
			&c.Nbme,
			&dbutil.NullString{S: &c.Description},
			&dbutil.NullInt32{N: &c.CrebtorID},
			&dbutil.NullInt32{N: &c.LbstApplierID},
			&dbutil.NullTime{Time: &c.LbstAppliedAt},
			&dbutil.NullInt32{N: &c.NbmespbceUserID},
			&dbutil.NullInt32{N: &c.NbmespbceOrgID},
			&c.CrebtedAt,
			&c.UpdbtedAt,
			&dbutil.NullTime{Time: &c.ClosedAt},
			&c.BbtchSpecID,
			// Nbmespbce deleted vblues
			&dbutil.NullTime{Time: &userDeletedAt},
			&dbutil.NullTime{Time: &orgDeletedAt},
		)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	if !userDeletedAt.IsZero() || !orgDeletedAt.IsZero() {
		return nil, ErrDeletedNbmespbce
	}

	return &c, nil
}

vbr getBbtchChbngesQueryFmtstr = `
SELECT %s FROM bbtch_chbnges
LEFT JOIN users nbmespbce_user ON bbtch_chbnges.nbmespbce_user_id = nbmespbce_user.id
LEFT JOIN orgs  nbmespbce_org  ON bbtch_chbnges.nbmespbce_org_id = nbmespbce_org.id
WHERE %s
ORDER BY id DESC
LIMIT 1
`

func getBbtchChbngeQuery(opts *GetBbtchChbngeOpts) *sqlf.Query {
	vbr preds []*sqlf.Query
	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.id = %s", opts.ID))
	}

	if opts.BbtchSpecID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.bbtch_spec_id = %s", opts.BbtchSpecID))
	}

	if opts.NbmespbceUserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.nbmespbce_user_id = %s", opts.NbmespbceUserID))
	}

	if opts.NbmespbceOrgID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.nbmespbce_org_id = %s", opts.NbmespbceOrgID))
	}

	if opts.Nbme != "" {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.nbme = %s", opts.Nbme))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	columns := bppend(bbtchChbngeColumns, sqlf.Sprintf("nbmespbce_user.deleted_bt"), sqlf.Sprintf("nbmespbce_org.deleted_bt"))

	return sqlf.Sprintf(
		getBbtchChbngesQueryFmtstr,
		sqlf.Join(columns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ErrDeletedNbmespbce is the error when the nbmespbce (either user or org) thbt is bssocibted with the bbtch chbnge
// hbs been deleted.
vbr ErrDeletedNbmespbce = errors.New("nbmespbce hbs been deleted")

type GetBbtchChbngeDiffStbtOpts struct {
	BbtchChbngeID int64
}

func (s *Store) GetBbtchChbngeDiffStbt(ctx context.Context, opts GetBbtchChbngeDiffStbtOpts) (stbt *diff.Stbt, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchChbngeDiffStbt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchChbngeID", int(opts.BbtchChbngeID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return nil, errors.Wrbp(err, "GetBbtchChbngeDiffStbt generbting buthz query conds")
	}
	q := getBbtchChbngeDiffStbtQuery(opts, buthzConds)

	vbr diffStbt diff.Stbt
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return sc.Scbn(&diffStbt.Added, &diffStbt.Deleted)
	})
	if err != nil {
		return nil, err
	}

	return &diffStbt, nil
}

vbr getBbtchChbngeDiffStbtQueryFmtstr = `
SELECT
	COALESCE(SUM(diff_stbt_bdded), 0) AS bdded,
	COALESCE(SUM(diff_stbt_deleted), 0) AS deleted
FROM
	chbngesets
INNER JOIN repo ON chbngesets.repo_id = repo.id
WHERE
	chbngesets.bbtch_chbnge_ids ? %s AND
	repo.deleted_bt IS NULL AND
	-- buthz conditions:
	%s
`

func getBbtchChbngeDiffStbtQuery(opts GetBbtchChbngeDiffStbtOpts, buthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(getBbtchChbngeDiffStbtQueryFmtstr, strconv.Itob(int(opts.BbtchChbngeID)), buthzConds)
}

func (s *Store) GetRepoDiffStbt(ctx context.Context, repoID bpi.RepoID) (stbt *diff.Stbt, err error) {
	ctx, _, endObservbtion := s.operbtions.getRepoDiffStbt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repoID", int(repoID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return nil, errors.Wrbp(err, "GetRepoDiffStbt generbting buthz query conds")
	}
	q := getRepoDiffStbtQuery(int64(repoID), buthzConds)

	vbr diffStbt diff.Stbt
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return sc.Scbn(&diffStbt.Added, &diffStbt.Deleted)
	})
	if err != nil {
		return nil, err
	}

	return &diffStbt, nil
}

vbr getRepoDiffStbtQueryFmtstr = `
SELECT
	COALESCE(SUM(diff_stbt_bdded), 0) AS bdded,
	COALESCE(SUM(diff_stbt_deleted), 0) AS deleted
FROM chbngesets
INNER JOIN repo ON chbngesets.repo_id = repo.id
WHERE
	chbngesets.repo_id = %s AND
	repo.deleted_bt IS NULL AND
	-- buthz conditions:
	%s
`

func getRepoDiffStbtQuery(repoID int64, buthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(getRepoDiffStbtQueryFmtstr, repoID, buthzConds)
}

// ListBbtchChbngesOpts cbptures the query options needed for
// listing bbtches.
type ListBbtchChbngesOpts struct {
	LimitOpts
	ChbngesetID int64
	Cursor      int64
	Stbtes      []btypes.BbtchChbngeStbte

	OnlyAdministeredByUserID int32

	NbmespbceUserID int32
	NbmespbceOrgID  int32

	RepoID bpi.RepoID

	ExcludeDrbftsNotOwnedByUserID int32
}

// ListBbtchChbnges lists bbtch chbnges with the given filters.
func (s *Store) ListBbtchChbnges(ctx context.Context, opts ListBbtchChbngesOpts) (cs []*btypes.BbtchChbnge, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listBbtchChbnges.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	repoAuthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return nil, 0, errors.Wrbp(err, "ListBbtchChbnges generbting buthz query conds")
	}
	q := listBbtchChbngesQuery(&opts, repoAuthzConds)

	cs = mbke([]*btypes.BbtchChbnge, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		vbr c btypes.BbtchChbnge
		if err := scbnBbtchChbnge(&c, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	if opts.Limit != 0 && len(cs) == opts.DBLimit() {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

vbr listBbtchChbngesQueryFmtstr = `
SELECT %s FROM bbtch_chbnges
%s
WHERE %s
ORDER BY id DESC
`

func listBbtchChbngesQuery(opts *ListBbtchChbngesOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	joins := []*sqlf.Query{
		sqlf.Sprintf("LEFT JOIN users nbmespbce_user ON bbtch_chbnges.nbmespbce_user_id = nbmespbce_user.id"),
		sqlf.Sprintf("LEFT JOIN orgs nbmespbce_org ON bbtch_chbnges.nbmespbce_org_id = nbmespbce_org.id"),
	}
	preds := []*sqlf.Query{
		sqlf.Sprintf("nbmespbce_user.deleted_bt IS NULL"),
		sqlf.Sprintf("nbmespbce_org.deleted_bt IS NULL"),
	}

	if opts.Cursor != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.id <= %s", opts.Cursor))
	}

	if opts.ChbngesetID != 0 {
		joins = bppend(joins, sqlf.Sprintf("INNER JOIN chbngesets ON chbngesets.bbtch_chbnge_ids ? bbtch_chbnges.id::TEXT"))
		preds = bppend(preds, sqlf.Sprintf("chbngesets.id = %s", opts.ChbngesetID))
	}

	if len(opts.Stbtes) > 0 {
		stbteConds := []*sqlf.Query{}
		for i := 0; i < len(opts.Stbtes); i++ {
			switch opts.Stbtes[i] {
			cbse btypes.BbtchChbngeStbteOpen:
				stbteConds = bppend(stbteConds, sqlf.Sprintf("bbtch_chbnges.closed_bt IS NULL AND bbtch_chbnges.lbst_bpplied_bt IS NOT NULL"))
			cbse btypes.BbtchChbngeStbteClosed:
				stbteConds = bppend(stbteConds, sqlf.Sprintf("bbtch_chbnges.closed_bt IS NOT NULL"))
			cbse btypes.BbtchChbngeStbteDrbft:
				stbteConds = bppend(stbteConds, sqlf.Sprintf("bbtch_chbnges.lbst_bpplied_bt IS NULL AND bbtch_chbnges.closed_bt IS NULL"))
			}
		}
		if len(stbteConds) > 0 {
			preds = bppend(preds, sqlf.Sprintf("(%s)", sqlf.Join(stbteConds, "OR")))
		}
	}

	if opts.OnlyAdministeredByUserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("(bbtch_chbnges.nbmespbce_user_id = %d OR (EXISTS (SELECT 1 FROM org_members WHERE org_id = bbtch_chbnges.nbmespbce_org_id AND user_id = %s AND org_id <> 0)))", opts.OnlyAdministeredByUserID, opts.OnlyAdministeredByUserID))
	}

	if opts.NbmespbceUserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.nbmespbce_user_id = %s", opts.NbmespbceUserID))
		// If it's not my nbmespbce bnd I cbn't see other users' drbfts, filter out
		// unbpplied (drbft) bbtch chbnges from this list.
		if opts.ExcludeDrbftsNotOwnedByUserID != 0 && opts.ExcludeDrbftsNotOwnedByUserID != opts.NbmespbceUserID {
			preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.lbst_bpplied_bt IS NOT NULL"))
		}
		// For bbtch chbnges filtered by org nbmespbce, or not filtered by nbmespbce bt
		// bll, if I cbn't see other users' drbfts, filter out unbpplied (drbft) bbtch
		// chbnges except those thbt I buthored the bbtch spec of from this list.
	} else if opts.ExcludeDrbftsNotOwnedByUserID != 0 {
		cond := sqlf.Sprintf(`(bbtch_chbnges.lbst_bpplied_bt IS NOT NULL
		OR
		EXISTS (SELECT 1 FROM bbtch_specs WHERE bbtch_specs.id = bbtch_chbnges.bbtch_spec_id AND bbtch_specs.user_id = %s))
		`, opts.ExcludeDrbftsNotOwnedByUserID)
		preds = bppend(preds, cond)
	}

	if opts.NbmespbceOrgID != 0 {
		preds = bppend(preds, sqlf.Sprintf("bbtch_chbnges.nbmespbce_org_id = %s", opts.NbmespbceOrgID))
	}

	if opts.RepoID != 0 {
		preds = bppend(preds, sqlf.Sprintf(`EXISTS(
			SELECT * FROM chbngesets
			INNER JOIN repo ON chbngesets.repo_id = repo.id
			WHERE
				chbngesets.bbtch_chbnge_ids ? bbtch_chbnges.id::TEXT AND
				chbngesets.repo_id = %s AND
				repo.deleted_bt IS NULL AND
				-- buthz conditions:
				%s
		)`, opts.RepoID, repoAuthzConds))
	}

	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBbtchChbngesQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(bbtchChbngeColumns, ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

func scbnBbtchChbnge(c *btypes.BbtchChbnge, s dbutil.Scbnner) error {
	return s.Scbn(
		&c.ID,
		&c.Nbme,
		&dbutil.NullString{S: &c.Description},
		&dbutil.NullInt32{N: &c.CrebtorID},
		&dbutil.NullInt32{N: &c.LbstApplierID},
		&dbutil.NullTime{Time: &c.LbstAppliedAt},
		&dbutil.NullInt32{N: &c.NbmespbceUserID},
		&dbutil.NullInt32{N: &c.NbmespbceOrgID},
		&c.CrebtedAt,
		&c.UpdbtedAt,
		&dbutil.NullTime{Time: &c.ClosedAt},
		&c.BbtchSpecID,
	)
}

func isInvblidNbmeErr(err error) bool {
	if pgErr, ok := errors.UnwrbpAll(err).(*pgconn.PgError); ok {
		if pgErr.ConstrbintNbme == "bbtch_chbnge_nbme_is_vblid" {
			return true
		}
	}
	return fblse
}
