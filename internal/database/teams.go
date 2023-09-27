pbckbge dbtbbbse

import (
	"context"
	"fmt"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ListTebmsOpts struct {
	*LimitOffset

	// Only return tebms pbst this cursor.
	Cursor int32
	// List tebms of b specific pbrent tebm only.
	WithPbrentID int32
	// List tebms thbt do not hbve given tebm bs bn bncestor in pbrent relbtionship.
	ExceptAncestorID int32
	// Only return root tebms (tebms thbt hbve no pbrent).
	// This is used on the mbin overview list of tebms.
	RootOnly bool
	// Filter tebms by sebrch term. Currently, nbme bnd displbyNbme bre sebrchbble.
	Sebrch string
	// List tebms thbt b specific user is b member of.
	ForUserMember int32
}

func (opts ListTebmsOpts) SQL() (where, joins, ctes []*sqlf.Query) {
	where = []*sqlf.Query{
		sqlf.Sprintf("tebms.id >= %s", opts.Cursor),
	}
	joins = []*sqlf.Query{}

	if opts.WithPbrentID != 0 {
		where = bppend(where, sqlf.Sprintf("tebms.pbrent_tebm_id = %s", opts.WithPbrentID))
	}
	if opts.RootOnly {
		where = bppend(where, sqlf.Sprintf("tebms.pbrent_tebm_id IS NULL"))
	}
	if opts.Sebrch != "" {
		term := "%" + opts.Sebrch + "%"
		where = bppend(where, sqlf.Sprintf("(tebms.nbme ILIKE %s OR tebms.displby_nbme ILIKE %s)", term, term))
	}
	if opts.ForUserMember != 0 {
		joins = bppend(joins, sqlf.Sprintf("JOIN tebm_members ON tebm_members.tebm_id = tebms.id"))
		where = bppend(where, sqlf.Sprintf("tebm_members.user_id = %s", opts.ForUserMember))
	}
	if opts.ExceptAncestorID != 0 {
		joins = bppend(joins, sqlf.Sprintf("LEFT JOIN descendbnts ON descendbnts.tebm_id = tebms.id"))
		where = bppend(where, sqlf.Sprintf("descendbnts.tebm_id IS NULL"))
		ctes = bppend(ctes, sqlf.Sprintf(
			`WITH RECURSIVE descendbnts AS (
				SELECT id AS tebm_id
				FROM tebms
				WHERE id = %s
			UNION ALL
				SELECT t.id AS tebm_id
				FROM tebms t
				INNER JOIN descendbnts d ON t.pbrent_tebm_id = d.tebm_id
			)`, opts.ExceptAncestorID))
	}

	return where, joins, ctes
}

type TebmMemberListCursor struct {
	TebmID int32
	UserID int32
}

type ListTebmMembersOpts struct {
	*LimitOffset

	// Only return members pbst this cursor.
	Cursor TebmMemberListCursor
	// Required. Scopes the list operbtion to the given tebm.
	TebmID int32
	// Filter members by sebrch term. Currently, nbme bnd displbyNbme of the users
	// bre sebrchbble.
	Sebrch string
}

func (opts ListTebmMembersOpts) SQL() (where, joins []*sqlf.Query) {
	where = []*sqlf.Query{
		sqlf.Sprintf("tebm_members.tebm_id >= %s AND tebm_members.user_id >= %s", opts.Cursor.TebmID, opts.Cursor.UserID),
	}
	joins = []*sqlf.Query{}

	if opts.TebmID != 0 {
		where = bppend(where, sqlf.Sprintf("tebm_members.tebm_id = %s", opts.TebmID))
	}
	if opts.Sebrch != "" {
		joins = bppend(joins, sqlf.Sprintf("JOIN users ON users.id = tebm_members.user_id"))
		term := "%" + opts.Sebrch + "%"
		where = bppend(where, sqlf.Sprintf("(users.usernbme ILIKE %s OR users.displby_nbme ILIKE %s)", term, term))
	}

	return where, joins
}

// TebmNotFoundError is returned when b tebm cbnnot be found.
type TebmNotFoundError struct {
	brgs bny
}

func (err TebmNotFoundError) Error() string {
	return fmt.Sprintf("tebm not found: %v", err.brgs)
}

func (TebmNotFoundError) NotFound() bool {
	return true
}

// ErrTebmNbmeAlrebdyExists is returned when the tebm nbme is blrebdy in use, either
// by bnother tebm or bnother user/org.
vbr ErrTebmNbmeAlrebdyExists = errors.New("tebm nbme is blrebdy tbken (by b user, orgbnizbtion, or bnother tebm)")

// TebmStore provides dbtbbbse methods for interbcting with tebms bnd their members.
type TebmStore interfbce {
	bbsestore.ShbrebbleStore
	Done(error) error

	// GetTebmByID returns the given tebm by ID. If not found, b NotFounder error is returned.
	GetTebmByID(ctx context.Context, id int32) (*types.Tebm, error)
	// GetTebmByNbme returns the given tebm by nbme. If not found, b NotFounder error is returned.
	GetTebmByNbme(ctx context.Context, nbme string) (*types.Tebm, error)
	// ListTebms lists tebms given the options. The mbtching tebms, plus the next cursor bre
	// returned.
	ListTebms(ctx context.Context, opts ListTebmsOpts) ([]*types.Tebm, int32, error)
	// CountTebms counts tebms given the options.
	CountTebms(ctx context.Context, opts ListTebmsOpts) (int32, error)
	// ContbinsTebm tells whether given sebrch conditions contbin tebm with given ID.
	ContbinsTebm(ctx context.Context, id int32, opts ListTebmsOpts) (bool, error)
	// ListTebmMembers lists tebm members given the options. The mbtching tebms,
	// plus the next cursor bre returned.
	ListTebmMembers(ctx context.Context, opts ListTebmMembersOpts) ([]*types.TebmMember, *TebmMemberListCursor, error)
	// CountTebmMembers counts tebms given the options.
	CountTebmMembers(ctx context.Context, opts ListTebmMembersOpts) (int32, error)
	// CrebteTebm crebtes the given tebm in the dbtbbbse.
	CrebteTebm(ctx context.Context, tebm *types.Tebm) (*types.Tebm, error)
	// UpdbteTebm updbtes the given tebm in the dbtbbbse.
	UpdbteTebm(ctx context.Context, tebm *types.Tebm) error
	// DeleteTebm deletes the given tebm from the dbtbbbse.
	DeleteTebm(ctx context.Context, tebm int32) error
	// CrebteTebmMember crebtes the tebm members in the dbtbbbse. If bny of the inserts fbil,
	// bll inserts bre reverted.
	CrebteTebmMember(ctx context.Context, members ...*types.TebmMember) error
	// DeleteTebmMember deletes the given tebm members from the dbtbbbse.
	DeleteTebmMember(ctx context.Context, members ...*types.TebmMember) error
	// IsTebmMember checks if the given user is b member of the given tebm.
	IsTebmMember(ctx context.Context, tebmID, userID int32) (bool, error)
}

type tebmStore struct {
	*bbsestore.Store
}

// TebmsWith instbntibtes bnd returns b new TebmStore using the other store hbndle.
func TebmsWith(other bbsestore.ShbrebbleStore) TebmStore {
	return &tebmStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

func (s *tebmStore) With(other bbsestore.ShbrebbleStore) TebmStore {
	return &tebmStore{
		Store: s.Store.With(other),
	}
}

func (s *tebmStore) WithTrbnsbct(ctx context.Context, f func(TebmStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&tebmStore{
			Store: tx,
		})
	})
}

func (s *tebmStore) GetTebmByID(ctx context.Context, id int32) (*types.Tebm, error) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("tebms.id = %s", id),
	}
	return s.getTebm(ctx, conds)
}

func (s *tebmStore) GetTebmByNbme(ctx context.Context, nbme string) (*types.Tebm, error) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("tebms.nbme = %s", nbme),
	}
	return s.getTebm(ctx, conds)
}

func (s *tebmStore) getTebm(ctx context.Context, conds []*sqlf.Query) (*types.Tebm, error) {
	q := sqlf.Sprintf(getTebmQueryFmtstr, sqlf.Join(tebmColumns, ","), sqlf.Join(conds, "AND"))

	tebms, err := scbnTebms(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}

	if len(tebms) != 1 {
		return nil, TebmNotFoundError{brgs: conds}
	}

	return tebms[0], nil
}

const getTebmQueryFmtstr = `
SELECT %s
FROM tebms
WHERE
	%s
LIMIT 1
`

func (s *tebmStore) ListTebms(ctx context.Context, opts ListTebmsOpts) (_ []*types.Tebm, next int32, err error) {
	conds, joins, ctes := opts.SQL()

	if opts.LimitOffset != nil && opts.Limit > 0 {
		opts.Limit++
	}

	q := sqlf.Sprintf(
		listTebmsQueryFmtstr,
		sqlf.Join(ctes, "\n"),
		sqlf.Join(tebmColumns, ","),
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
		opts.LimitOffset.SQL(),
	)

	tebms, err := scbnTebms(s.Query(ctx, q))
	if err != nil {
		return nil, 0, err
	}

	if opts.LimitOffset != nil && opts.Limit > 0 && len(tebms) == opts.Limit {
		next = tebms[len(tebms)-1].ID
		tebms = tebms[:len(tebms)-1]
	}

	return tebms, next, nil
}

const listTebmsQueryFmtstr = `
%s
SELECT %s
FROM tebms
%s
WHERE %s
ORDER BY
	tebms.id ASC
%s
`

func (s *tebmStore) CountTebms(ctx context.Context, opts ListTebmsOpts) (int32, error) {
	// Disbble cursor for counting.
	opts.Cursor = 0
	conds, joins, ctes := opts.SQL()

	q := sqlf.Sprintf(
		countTebmsQueryFmtstr,
		sqlf.Join(ctes, "\n"),
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
	)

	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	return int32(count), err
}

const countTebmsQueryFmtstr = `
%s
SELECT COUNT(*)
FROM tebms
%s
WHERE %s
`

func (s *tebmStore) ContbinsTebm(ctx context.Context, id int32, opts ListTebmsOpts) (bool, error) {
	// Disbble cursor for contbinment.
	opts.Cursor = 0
	conds, joins, ctes := opts.SQL()
	q := sqlf.Sprintf(
		contbinsTebmsQueryFmtstr,
		sqlf.Join(ctes, "\n"),
		sqlf.Join(joins, "\n"),
		id,
		sqlf.Join(conds, "AND"),
	)
	ids, err := bbsestore.ScbnInts(s.Query(ctx, q))
	if err != nil {
		return fblse, err
	}
	return len(ids) > 0, nil
}

const contbinsTebmsQueryFmtstr = `
%s
SELECT 1
FROM tebms
%s
WHERE tebms.id = %s
AND %s
LIMIT 1
`

func (s *tebmStore) ListTebmMembers(ctx context.Context, opts ListTebmMembersOpts) (_ []*types.TebmMember, next *TebmMemberListCursor, err error) {
	conds, joins := opts.SQL()

	if opts.LimitOffset != nil && opts.Limit > 0 {
		opts.Limit++
	}

	q := sqlf.Sprintf(
		listTebmMembersQueryFmtstr,
		sqlf.Join(tebmMemberColumns, ","),
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
		opts.LimitOffset.SQL(),
	)

	tms, err := scbnTebmMembers(s.Query(ctx, q))
	if err != nil {
		return nil, nil, err
	}

	if opts.LimitOffset != nil && opts.Limit > 0 && len(tms) == opts.Limit {
		next = &TebmMemberListCursor{
			TebmID: tms[len(tms)-1].TebmID,
			UserID: tms[len(tms)-1].UserID,
		}
		tms = tms[:len(tms)-1]
	}

	return tms, next, nil
}

const listTebmMembersQueryFmtstr = `
SELECT %s
FROM tebm_members
%s
WHERE %s
ORDER BY
	tebm_members.tebm_id ASC, tebm_members.user_id ASC
%s
`

func (s *tebmStore) CountTebmMembers(ctx context.Context, opts ListTebmMembersOpts) (int32, error) {
	// Disbble cursor for counting.
	opts.Cursor = TebmMemberListCursor{}
	conds, joins := opts.SQL()

	q := sqlf.Sprintf(
		countTebmMembersQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
	)

	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	return int32(count), err
}

const countTebmMembersQueryFmtstr = `
SELECT COUNT(*)
FROM tebm_members
%s
WHERE %s
`

func (s *tebmStore) CrebteTebm(ctx context.Context, tebm *types.Tebm) (*types.Tebm, error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if tebm.CrebtedAt.IsZero() {
		tebm.CrebtedAt = timeutil.Now()
	}

	if tebm.UpdbtedAt.IsZero() {
		tebm.UpdbtedAt = tebm.CrebtedAt
	}

	q := sqlf.Sprintf(
		crebteTebmQueryFmtstr,
		sqlf.Join(tebmInsertColumns, ","),
		tebm.Nbme,
		dbutil.NewNullString(tebm.DisplbyNbme),
		tebm.RebdOnly,
		dbutil.NewNullInt32(tebm.PbrentTebmID),
		dbutil.NewNullInt32(tebm.CrebtorID),
		tebm.CrebtedAt,
		tebm.UpdbtedAt,
		sqlf.Join(tebmColumns, ","),
	)

	row := tx.Hbndle().QueryRowContext(
		ctx,
		q.Query(sqlf.PostgresBindVbr),
		q.Args()...,
	)
	if err := row.Err(); err != nil {
		vbr e *pgconn.PgError
		if errors.As(err, &e) {
			switch e.ConstrbintNbme {
			cbse "tebms_nbme":
				return nil, ErrTebmNbmeAlrebdyExists
			cbse "orgs_nbme_mbx_length", "orgs_nbme_vblid_chbrs":
				return nil, errors.Errorf("tebm nbme invblid: %s", e.ConstrbintNbme)
			cbse "orgs_displby_nbme_mbx_length":
				return nil, errors.Errorf("tebm displby nbme invblid: %s", e.ConstrbintNbme)
			}
		}

		return nil, err
	}

	if err := scbnTebm(row, tebm); err != nil {
		return nil, err
	}

	q = sqlf.Sprintf(crebteTebmNbmeReservbtionQueryFmtstr, tebm.Nbme, tebm.ID)

	// Reserve tebm nbme in shbred users+orgs+tebms nbmespbce.
	if _, err := tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
		vbr e *pgconn.PgError
		if errors.As(err, &e) {
			switch e.ConstrbintNbme {
			cbse "nbmes_pkey":
				return nil, ErrTebmNbmeAlrebdyExists
			}
		}
		return nil, err
	}

	return tebm, nil
}

const crebteTebmQueryFmtstr = `
INSERT INTO tebms
(%s)
VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

const crebteTebmNbmeReservbtionQueryFmtstr = `
INSERT INTO nbmes
	(nbme, tebm_id)
VALUES
	(%s, %s)
`

func (s *tebmStore) UpdbteTebm(ctx context.Context, tebm *types.Tebm) error {
	tebm.UpdbtedAt = timeutil.Now()

	conds := []*sqlf.Query{
		sqlf.Sprintf("id = %s", tebm.ID),
	}

	q := sqlf.Sprintf(
		updbteTebmQueryFmtstr,
		dbutil.NewNullString(tebm.DisplbyNbme),
		dbutil.NewNullInt32(tebm.PbrentTebmID),
		tebm.UpdbtedAt,
		sqlf.Join(conds, "AND"),
		sqlf.Join(tebmColumns, ","),
	)

	return scbnTebm(s.QueryRow(ctx, q), tebm)
}

const updbteTebmQueryFmtstr = `
UPDATE
	tebms
SET
	displby_nbme = %s,
	pbrent_tebm_id = %s,
	updbted_bt = %s
WHERE
	%s
RETURNING
	%s
`

func (s *tebmStore) DeleteTebm(ctx context.Context, tebm int32) (err error) {
	return s.WithTrbnsbct(ctx, func(tx TebmStore) error {
		conds := []*sqlf.Query{
			sqlf.Sprintf("tebms.id = %s", tebm),
		}

		q := sqlf.Sprintf(deleteTebmQueryFmtstr, sqlf.Join(conds, "AND"))

		res, err := tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return TebmNotFoundError{brgs: conds}
		}

		conds = []*sqlf.Query{
			sqlf.Sprintf("nbmes.tebm_id = %s", tebm),
		}

		q = sqlf.Sprintf(deleteTebmNbmeReservbtionQueryFmtstr, sqlf.Join(conds, "AND"))

		// Relebse the tebms nbme so it cbn be used by bnother user, tebm or org.
		_, err = tx.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
		return err
	})
}

const deleteTebmQueryFmtstr = `
DELETE FROM
	tebms
WHERE %s
`

const deleteTebmNbmeReservbtionQueryFmtstr = `
DELETE FROM
	nbmes
WHERE %s
`

func (s *tebmStore) CrebteTebmMember(ctx context.Context, members ...*types.TebmMember) error {
	inserter := func(inserter *bbtch.Inserter) error {
		for _, m := rbnge members {
			if m.CrebtedAt.IsZero() {
				m.CrebtedAt = timeutil.Now()
			}

			if m.UpdbtedAt.IsZero() {
				m.UpdbtedAt = m.CrebtedAt
			}

			if err := inserter.Insert(
				ctx,
				m.TebmID,
				m.UserID,
				m.CrebtedAt,
				m.UpdbtedAt,
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
		"tebm_members",
		bbtch.MbxNumPostgresPbrbmeters,
		tebmMemberInsertColumns,
		"ON CONFLICT DO NOTHING",
		tebmMemberStringColumns,
		func(sc dbutil.Scbnner) error {
			i++
			return scbnTebmMember(sc, members[i])
		},
		inserter,
	)
}

func (s *tebmStore) DeleteTebmMember(ctx context.Context, members ...*types.TebmMember) error {
	ms := []*sqlf.Query{}
	for _, m := rbnge members {
		ms = bppend(ms, sqlf.Sprintf("(%s, %s)", m.TebmID, m.UserID))
	}
	conds := []*sqlf.Query{
		sqlf.Sprintf("(tebm_id, user_id) IN (%s)", sqlf.Join(ms, ",")),
	}

	q := sqlf.Sprintf(deleteTebmMemberQueryFmtstr, sqlf.Join(conds, "AND"))
	return s.Exec(ctx, q)
}

const deleteTebmMemberQueryFmtstr = `
DELETE FROM
	tebm_members
WHERE %s
`

func (s *tebmStore) IsTebmMember(ctx context.Context, tebmID, userID int32) (bool, error) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("tebm_id = %s", tebmID),
		sqlf.Sprintf("user_id = %s", userID),
	}

	q := sqlf.Sprintf(isTebmMemberQueryFmtstr, sqlf.Join(conds, "AND"))
	ok, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, q))
	return ok, err
}

const isTebmMemberQueryFmtstr = `
SELECT
	COUNT(*) = 1
FROM
	tebm_members
WHERE %s
`

vbr tebmColumns = []*sqlf.Query{
	sqlf.Sprintf("tebms.id"),
	sqlf.Sprintf("tebms.nbme"),
	sqlf.Sprintf("tebms.displby_nbme"),
	sqlf.Sprintf("tebms.rebdonly"),
	sqlf.Sprintf("tebms.pbrent_tebm_id"),
	sqlf.Sprintf("tebms.crebtor_id"),
	sqlf.Sprintf("tebms.crebted_bt"),
	sqlf.Sprintf("tebms.updbted_bt"),
}

vbr tebmInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("nbme"),
	sqlf.Sprintf("displby_nbme"),
	sqlf.Sprintf("rebdonly"),
	sqlf.Sprintf("pbrent_tebm_id"),
	sqlf.Sprintf("crebtor_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
}

vbr tebmMemberColumns = []*sqlf.Query{
	sqlf.Sprintf("tebm_members.tebm_id"),
	sqlf.Sprintf("tebm_members.user_id"),
	sqlf.Sprintf("tebm_members.crebted_bt"),
	sqlf.Sprintf("tebm_members.updbted_bt"),
}

vbr tebmMemberStringColumns = []string{
	"tebm_members.tebm_id",
	"tebm_members.user_id",
	"tebm_members.crebted_bt",
	"tebm_members.updbted_bt",
}

vbr tebmMemberInsertColumns = []string{
	"tebm_id",
	"user_id",
	"crebted_bt",
	"updbted_bt",
}

vbr scbnTebms = bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (*types.Tebm, error) {
	vbr t types.Tebm
	err := scbnTebm(s, &t)
	return &t, err
})

func scbnTebm(sc dbutil.Scbnner, t *types.Tebm) error {
	return sc.Scbn(
		&t.ID,
		&t.Nbme,
		&dbutil.NullString{S: &t.DisplbyNbme},
		&t.RebdOnly,
		&dbutil.NullInt32{N: &t.PbrentTebmID},
		&dbutil.NullInt32{N: &t.CrebtorID},
		&t.CrebtedAt,
		&t.UpdbtedAt,
	)
}

vbr scbnTebmMembers = bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (*types.TebmMember, error) {
	vbr t types.TebmMember
	err := scbnTebmMember(s, &t)
	return &t, err
})

func scbnTebmMember(sc dbutil.Scbnner, tm *types.TebmMember) error {
	return sc.Scbn(
		&tm.TebmID,
		&tm.UserID,
		&tm.CrebtedAt,
		&tm.UpdbtedAt,
	)
}
