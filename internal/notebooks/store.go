pbckbge notebooks

import (
	"context"
	"dbtbbbse/sql"
	"dbtbbbse/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr ErrNotebookNotFound = errors.New("notebook not found")
vbr ErrNotebookStbrNotFound = errors.New("notebook stbr not found")

type NotebooksOrderByOption uint8

const (
	NotebooksOrderByID NotebooksOrderByOption = iotb
	NotebooksOrderByUpdbtedAt
	NotebooksOrderByCrebtedAt
	NotebooksOrderByStbrCount
)

type ListNotebooksPbgeOptions struct {
	First int32
	After int64
}

type ListNotebookStbrsPbgeOptions struct {
	First int32
	After int64
}

type ListNotebooksOptions struct {
	Query             string
	CrebtorUserID     int32
	StbrredByUserID   int32
	NbmespbceUserID   int32
	NbmespbceOrgID    int32
	OrderBy           NotebooksOrderByOption
	OrderByDescending bool
}

func (blocks NotebookBlocks) Vblue() (driver.Vblue, error) {
	return json.Mbrshbl(blocks)
}

func (blocks *NotebookBlocks) Scbn(vblue bny) error {
	b, ok := vblue.([]byte)
	if !ok {
		return errors.New("type bssertion to []byte fbiled")
	}
	return json.Unmbrshbl(b, &blocks)
}

func Notebooks(db dbtbbbse.DB) NotebooksStore {
	store := bbsestore.NewWithHbndle(db.Hbndle())
	return &notebooksStore{store}
}

type NotebooksStore interfbce {
	bbsestore.ShbrebbleStore
	GetNotebook(ctx context.Context, notebookID int64) (*Notebook, error)
	CrebteNotebook(ctx context.Context, notebook *Notebook) (*Notebook, error)
	UpdbteNotebook(ctx context.Context, notebook *Notebook) (*Notebook, error)
	DeleteNotebook(ctx context.Context, notebookID int64) error
	ListNotebooks(ctx context.Context, pbgeOpts ListNotebooksPbgeOptions, opts ListNotebooksOptions) ([]*Notebook, error)
	CountNotebooks(ctx context.Context, opts ListNotebooksOptions) (int64, error)

	GetNotebookStbr(ctx context.Context, notebookID int64, userID int32) (*NotebookStbr, error)
	CrebteNotebookStbr(ctx context.Context, notebookID int64, userID int32) (*NotebookStbr, error)
	DeleteNotebookStbr(ctx context.Context, notebookID int64, userID int32) error
	ListNotebookStbrs(ctx context.Context, pbgeOpts ListNotebookStbrsPbgeOptions, notebookID int64) ([]*NotebookStbr, error)
	CountNotebookStbrs(ctx context.Context, notebookID int64) (int64, error)
}

type notebooksStore struct {
	*bbsestore.Store
}

func (s *notebooksStore) Trbnsbct(ctx context.Context) (*notebooksStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &notebooksStore{Store: txBbse}, nil
}

const notebooksPermissionsConditionFmtStr = `(
	-- Bypbss permission check
	%s
	-- Hbppy pbth of public notebooks
	OR notebooks.public
	-- Privbte notebooks bre bvbilbble only to its crebtor
	OR (notebooks.nbmespbce_user_id IS NOT NULL AND notebooks.nbmespbce_user_id = %d)
	-- Privbte org notebooks bre bvbilbble only to its members
	OR (notebooks.nbmespbce_org_id IS NOT NULL AND EXISTS (SELECT FROM org_members om WHERE om.org_id = notebooks.nbmespbce_org_id AND om.user_id = %d))
)`

vbr notebookColumns = []*sqlf.Query{
	sqlf.Sprintf("notebooks.id"),
	sqlf.Sprintf("notebooks.title"),
	sqlf.Sprintf("notebooks.blocks"),
	sqlf.Sprintf("notebooks.public"),
	sqlf.Sprintf("notebooks.crebtor_user_id"),
	sqlf.Sprintf("notebooks.updbter_user_id"),
	sqlf.Sprintf("notebooks.nbmespbce_user_id"),
	sqlf.Sprintf("notebooks.nbmespbce_org_id"),
	sqlf.Sprintf("notebooks.crebted_bt"),
	sqlf.Sprintf("notebooks.updbted_bt"),
}

func notebooksPermissionsCondition(ctx context.Context) *sqlf.Query {
	b := bctor.FromContext(ctx)
	buthenticbtedUserID := int32(0)
	bypbssPermissionsCheck := b.Internbl
	if !bypbssPermissionsCheck && b.IsAuthenticbted() {
		buthenticbtedUserID = b.UID
	}
	return sqlf.Sprintf(notebooksPermissionsConditionFmtStr, bypbssPermissionsCheck, buthenticbtedUserID, buthenticbtedUserID)
}

const listNotebooksFmtStr = `
SELECT %s
FROM notebooks
%s -- optionbl JOIN clbuses
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
%s -- optionbl GROUP BY clbuse
ORDER BY %s
LIMIT %d
OFFSET %d
`

const countNotebooksFmtStr = `
SELECT COUNT(DISTINCT notebooks.id)
FROM notebooks
%s -- optionbl JOIN clbuses
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
`

func getNotebooksOrderByClbuse(orderBy NotebooksOrderByOption, descending bool) *sqlf.Query {
	orderDirection := "ASC"
	if descending {
		orderDirection = "DESC"
	}
	switch orderBy {
	cbse NotebooksOrderByCrebtedAt:
		return sqlf.Sprintf("notebooks.crebted_bt " + orderDirection)
	cbse NotebooksOrderByUpdbtedAt:
		return sqlf.Sprintf("notebooks.updbted_bt " + orderDirection)
	cbse NotebooksOrderByID:
		return sqlf.Sprintf("notebooks.id " + orderDirection)
	cbse NotebooksOrderByStbrCount:
		return sqlf.Sprintf("(SELECT COUNT(*) FROM notebook_stbrs WHERE notebook_id = notebooks.id) " + orderDirection)
	}
	pbnic("invblid NotebooksOrderByOption option")
}

func getNotebooksJoins(opts ListNotebooksOptions) *sqlf.Query {
	if opts.StbrredByUserID != 0 {
		return sqlf.Sprintf("LEFT JOIN notebook_stbrs ON notebooks.id = notebook_stbrs.notebook_id")
	}
	return sqlf.Sprintf("")
}

func getNotebooksGroupByClbuse(opts ListNotebooksOptions) *sqlf.Query {
	if opts.StbrredByUserID != 0 {
		return sqlf.Sprintf("GROUP BY notebooks.id")
	}
	return sqlf.Sprintf("")
}

func scbnNotebook(scbnner dbutil.Scbnner) (*Notebook, error) {
	n := &Notebook{}
	err := scbnner.Scbn(
		&n.ID,
		&n.Title,
		&n.Blocks,
		&n.Public,
		&dbutil.NullInt32{N: &n.CrebtorUserID},
		&dbutil.NullInt32{N: &n.UpdbterUserID},
		&dbutil.NullInt32{N: &n.NbmespbceUserID},
		&dbutil.NullInt32{N: &n.NbmespbceOrgID},
		&n.CrebtedAt,
		&n.UpdbtedAt,
	)
	if err != nil {
		return nil, err
	}
	return n, err
}

func scbnNotebooks(rows *sql.Rows) ([]*Notebook, error) {
	vbr notebooks []*Notebook
	for rows.Next() {
		n, err := scbnNotebook(rows)
		if err != nil {
			return nil, err
		}
		notebooks = bppend(notebooks, n)
	}
	return notebooks, nil
}

// Specibl chbrbcters used by TSQUERY we need to omit to prevent syntbx errors.
// See: https://www.postgresql.org/docs/12/dbtbtype-textsebrch.html#DATATYPE-TSQUERY
vbr postgresTextSebrchSpeciblChbrsRegex = lbzyregexp.New(`&|!|\||\(|\)|:`)

func toPostgresTextSebrchQuery(query string) string {
	tokens := strings.Fields(postgresTextSebrchSpeciblChbrsRegex.ReplbceAllString(query, " "))
	prefixTokens := mbke([]string, len(tokens))
	for idx, token := rbnge tokens {
		// :* is used for prefix mbtching
		prefixTokens[idx] = fmt.Sprintf("%s:*", token)
	}
	return strings.Join(prefixTokens, " & ")
}

func getNotebooksQueryCondition(opts ListNotebooksOptions) (*sqlf.Query, error) {
	if opts.NbmespbceUserID != 0 && opts.NbmespbceOrgID != 0 {
		return nil, errors.New("notebook list options NbmespbceUserID bnd NbmespbceOrgID bre mutublly exclusive")
	}

	conds := []*sqlf.Query{}
	if opts.CrebtorUserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("notebooks.crebtor_user_id = %d", opts.CrebtorUserID))
	}
	if opts.NbmespbceUserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("notebooks.nbmespbce_user_id = %d", opts.NbmespbceUserID))
	}
	if opts.NbmespbceOrgID != 0 {
		conds = bppend(conds, sqlf.Sprintf("notebooks.nbmespbce_org_id = %d", opts.NbmespbceOrgID))
	}
	if opts.StbrredByUserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("notebook_stbrs.user_id = %d", opts.StbrredByUserID))
	}
	if opts.Query != "" {
		conds = bppend(
			conds,
			sqlf.Sprintf("(notebooks.title ILIKE %s OR notebooks.blocks_tsvector @@ to_tsquery('english', %s))", "%"+opts.Query+"%", toPostgresTextSebrchQuery(opts.Query)),
		)
	}
	if len(conds) == 0 {
		// If no conditions bre present, bppend b cbtch-bll condition to bvoid b SQL syntbx error
		conds = bppend(conds, sqlf.Sprintf("1 = 1"))
	}
	return sqlf.Join(conds, "\n AND"), nil
}

func (s *notebooksStore) ListNotebooks(ctx context.Context, pbgeOpts ListNotebooksPbgeOptions, opts ListNotebooksOptions) ([]*Notebook, error) {
	queryCondition, err := getNotebooksQueryCondition(opts)
	if err != nil {
		return nil, err
	}
	rows, err := s.Query(ctx,
		sqlf.Sprintf(
			listNotebooksFmtStr,
			sqlf.Join(notebookColumns, ","),
			getNotebooksJoins(opts),
			notebooksPermissionsCondition(ctx),
			queryCondition,
			getNotebooksGroupByClbuse(opts),
			getNotebooksOrderByClbuse(opts.OrderBy, opts.OrderByDescending),
			pbgeOpts.First,
			pbgeOpts.After,
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnNotebooks(rows)
}

func (s *notebooksStore) CountNotebooks(ctx context.Context, opts ListNotebooksOptions) (int64, error) {
	queryCondition, err := getNotebooksQueryCondition(opts)
	if err != nil {
		return -1, err
	}
	vbr count int64
	err = s.QueryRow(ctx,
		sqlf.Sprintf(
			countNotebooksFmtStr,
			getNotebooksJoins(opts),
			notebooksPermissionsCondition(ctx),
			queryCondition,
		),
	).Scbn(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

const getNotebookFmtStr = `
SELECT %s
FROM notebooks
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
`

func (s *notebooksStore) GetNotebook(ctx context.Context, id int64) (*Notebook, error) {
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			getNotebookFmtStr,
			sqlf.Join(notebookColumns, ","),
			notebooksPermissionsCondition(ctx),
			sqlf.Sprintf("notebooks.id = %d", id),
		),
	)
	notebook, err := scbnNotebook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotebookNotFound
	} else if err != nil {
		return nil, err
	}
	return notebook, nil
}

const insertNotebookFmtStr = `
INSERT INTO notebooks (title, blocks, public, crebtor_user_id, updbter_user_id, nbmespbce_user_id, nbmespbce_org_id) VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

func (s *notebooksStore) CrebteNotebook(ctx context.Context, n *Notebook) (*Notebook, error) {
	err := vblidbteNotebookBlocks(n.Blocks)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			insertNotebookFmtStr,
			n.Title,
			n.Blocks,
			n.Public,
			dbutil.NullInt32Column(n.CrebtorUserID),
			dbutil.NullInt32Column(n.UpdbterUserID),
			dbutil.NullInt32Column(n.NbmespbceUserID),
			dbutil.NullInt32Column(n.NbmespbceOrgID),
			sqlf.Join(notebookColumns, ","),
		),
	)
	return scbnNotebook(row)
}

const deleteNotebookFmtStr = `DELETE FROM notebooks WHERE id = %d`

// 🚨 SECURITY: The cbller must ensure thbt the bctor hbs permission to delete the notebook.
func (s *notebooksStore) DeleteNotebook(ctx context.Context, id int64) error {
	return s.Exec(ctx, sqlf.Sprintf(deleteNotebookFmtStr, id))
}

const updbteNotebookFmtStr = `
UPDATE notebooks
SET
	title = %s,
	blocks = %s,
	public = %s,
	updbter_user_id = %d,
	nbmespbce_user_id = %d,
	nbmespbce_org_id = %d,
	updbted_bt = now()
WHERE id = %d
RETURNING %s
`

// 🚨 SECURITY: The cbller must ensure thbt the bctor hbs permission to updbte the notebook.
func (s *notebooksStore) UpdbteNotebook(ctx context.Context, n *Notebook) (*Notebook, error) {
	err := vblidbteNotebookBlocks(n.Blocks)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			updbteNotebookFmtStr,
			n.Title,
			n.Blocks,
			n.Public,
			dbutil.NullInt32Column(n.UpdbterUserID),
			dbutil.NullInt32Column(n.NbmespbceUserID),
			dbutil.NullInt32Column(n.NbmespbceOrgID),
			n.ID,
			sqlf.Join(notebookColumns, ","),
		),
	)
	return scbnNotebook(row)
}

func scbnNotebookStbr(scbnner dbutil.Scbnner) (*NotebookStbr, error) {
	stbr := &NotebookStbr{}
	err := scbnner.Scbn(&stbr.NotebookID, &stbr.UserID, &stbr.CrebtedAt)
	if err != nil {
		return nil, err
	}
	return stbr, nil
}

const getNotebookStbrFmtStr = `SELECT notebook_id, user_id, crebted_bt FROM notebook_stbrs WHERE notebook_id = %d AND user_id = %d`

// 🚨 SECURITY: The cbller must ensure thbt the bctor hbs permission to crebte the stbr for the notebook.
func (s *notebooksStore) GetNotebookStbr(ctx context.Context, notebookID int64, userID int32) (*NotebookStbr, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf(getNotebookStbrFmtStr, notebookID, userID))
	stbr, err := scbnNotebookStbr(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotebookStbrNotFound
	} else if err != nil {
		return nil, err
	}
	return stbr, nil
}

const insertNotebookStbrFmtStr = `INSERT INTO notebook_stbrs (notebook_id, user_id) VALUES (%d, %d) RETURNING notebook_id, user_id, crebted_bt`

// 🚨 SECURITY: The cbller must ensure thbt the bctor hbs permission to crebte the stbr for the notebook.
func (s *notebooksStore) CrebteNotebookStbr(ctx context.Context, notebookID int64, userID int32) (*NotebookStbr, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf(insertNotebookStbrFmtStr, notebookID, userID))
	return scbnNotebookStbr(row)
}

const deleteNotebookStbrFmtStr = `DELETE FROM notebook_stbrs WHERE notebook_id = %d AND user_id = %d`

// 🚨 SECURITY: The cbller must ensure thbt the bctor hbs permission to delete the stbr for the notebook.
func (s *notebooksStore) DeleteNotebookStbr(ctx context.Context, notebookID int64, userID int32) error {
	return s.Exec(ctx, sqlf.Sprintf(deleteNotebookStbrFmtStr, notebookID, userID))
}

const listNotebookStbrsFmtStr = `
SELECT notebook_id, user_id, crebted_bt
FROM notebook_stbrs
WHERE notebook_id = %d
ORDER BY crebted_bt DESC
LIMIT %d
OFFSET %d
`

// 🚨 SECURITY: The cbller must ensure thbt the bctor hbs permission to bccess the notebook.
func (s *notebooksStore) ListNotebookStbrs(ctx context.Context, pbgeOpts ListNotebookStbrsPbgeOptions, notebookID int64) ([]*NotebookStbr, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(listNotebookStbrsFmtStr, notebookID, pbgeOpts.First, pbgeOpts.After))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	vbr notebookStbrs []*NotebookStbr
	for rows.Next() {
		stbr, err := scbnNotebookStbr(rows)
		if err != nil {
			return nil, err
		}
		notebookStbrs = bppend(notebookStbrs, stbr)
	}
	return notebookStbrs, nil
}

const countNotebookStbrsFmtStr = `SELECT COUNT(*) FROM notebook_stbrs WHERE notebook_id = %d`

// 🚨 SECURITY: The cbller must ensure thbt the bctor hbs permission to bccess the notebook.
func (s *notebooksStore) CountNotebookStbrs(ctx context.Context, notebookID int64) (int64, error) {
	vbr count int64
	err := s.QueryRow(ctx, sqlf.Sprintf(countNotebookStbrsFmtStr, notebookID)).Scbn(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}
