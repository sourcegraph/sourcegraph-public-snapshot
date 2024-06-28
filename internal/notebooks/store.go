package notebooks

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrNotebookNotFound = errors.New("notebook not found")
var ErrNotebookStarNotFound = errors.New("notebook star not found")

type NotebooksOrderByOption uint8

const (
	NotebooksOrderByID NotebooksOrderByOption = iota
	NotebooksOrderByUpdatedAt
	NotebooksOrderByCreatedAt
	NotebooksOrderByStarCount
)

type ListNotebooksPageOptions struct {
	First int32
	After int64
}

type ListNotebookStarsPageOptions struct {
	First int32
	After int64
}

type ListNotebooksOptions struct {
	Query             string
	CreatorUserID     int32
	StarredByUserID   int32
	NamespaceUserID   int32
	NamespaceOrgID    int32
	OrderBy           NotebooksOrderByOption
	OrderByDescending bool
}

func (blocks NotebookBlocks) Value() (driver.Value, error) {
	return json.Marshal(blocks)
}

func (blocks *NotebookBlocks) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &blocks)
}

func Notebooks(db database.DB) NotebooksStore {
	store := basestore.NewWithHandle(db.Handle())
	return &notebooksStore{store}
}

type NotebooksStore interface {
	basestore.ShareableStore
	GetNotebook(ctx context.Context, notebookID int64) (*Notebook, error)
	CreateNotebook(ctx context.Context, notebook *Notebook) (*Notebook, error)
	UpdateNotebook(ctx context.Context, notebook *Notebook) (*Notebook, error)
	DeleteNotebook(ctx context.Context, notebookID int64) error
	ListNotebooks(ctx context.Context, pageOpts ListNotebooksPageOptions, opts ListNotebooksOptions) ([]*Notebook, error)
	CountNotebooks(ctx context.Context, opts ListNotebooksOptions) (int64, error)

	GetNotebookStar(ctx context.Context, notebookID int64, userID int32) (*NotebookStar, error)
	CreateNotebookStar(ctx context.Context, notebookID int64, userID int32) (*NotebookStar, error)
	DeleteNotebookStar(ctx context.Context, notebookID int64, userID int32) error
	ListNotebookStars(ctx context.Context, pageOpts ListNotebookStarsPageOptions, notebookID int64) ([]*NotebookStar, error)
	CountNotebookStars(ctx context.Context, notebookID int64) (int64, error)
}

type notebooksStore struct {
	*basestore.Store
}

func (s *notebooksStore) Transact(ctx context.Context) (*notebooksStore, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &notebooksStore{Store: txBase}, nil
}

const notebooksPermissionsConditionFmtStr = `(
	-- Bypass permission check
	%s
	-- Happy path of public notebooks
	OR notebooks.public
	-- Private notebooks are available only to its creator
	OR (notebooks.namespace_user_id IS NOT NULL AND notebooks.namespace_user_id = %d)
	-- Private org notebooks are available only to its members
	OR (notebooks.namespace_org_id IS NOT NULL AND EXISTS (SELECT FROM org_members om WHERE om.org_id = notebooks.namespace_org_id AND om.user_id = %d))
)`

var notebookColumns = []*sqlf.Query{
	sqlf.Sprintf("notebooks.id"),
	sqlf.Sprintf("notebooks.title"),
	sqlf.Sprintf("notebooks.blocks"),
	sqlf.Sprintf("notebooks.public"),
	sqlf.Sprintf("notebooks.creator_user_id"),
	sqlf.Sprintf("notebooks.updater_user_id"),
	sqlf.Sprintf("notebooks.namespace_user_id"),
	sqlf.Sprintf("notebooks.namespace_org_id"),
	sqlf.Sprintf("notebooks.created_at"),
	sqlf.Sprintf("notebooks.updated_at"),
	sqlf.Sprintf("pattern_type"),
}

func notebooksPermissionsCondition(ctx context.Context) *sqlf.Query {
	a := actor.FromContext(ctx)
	authenticatedUserID := int32(0)
	bypassPermissionsCheck := a.Internal
	if !bypassPermissionsCheck && a.IsAuthenticated() {
		authenticatedUserID = a.UID
	}
	return sqlf.Sprintf(notebooksPermissionsConditionFmtStr, bypassPermissionsCheck, authenticatedUserID, authenticatedUserID)
}

const listNotebooksFmtStr = `
SELECT %s
FROM notebooks
%s -- optional JOIN clauses
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
%s -- optional GROUP BY clause
ORDER BY %s
LIMIT %d
OFFSET %d
`

const countNotebooksFmtStr = `
SELECT COUNT(DISTINCT notebooks.id)
FROM notebooks
%s -- optional JOIN clauses
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
`

func getNotebooksOrderByClause(orderBy NotebooksOrderByOption, descending bool) *sqlf.Query {
	orderDirection := "ASC"
	if descending {
		orderDirection = "DESC"
	}
	switch orderBy {
	case NotebooksOrderByCreatedAt:
		return sqlf.Sprintf("notebooks.created_at " + orderDirection)
	case NotebooksOrderByUpdatedAt:
		return sqlf.Sprintf("notebooks.updated_at " + orderDirection)
	case NotebooksOrderByID:
		return sqlf.Sprintf("notebooks.id " + orderDirection)
	case NotebooksOrderByStarCount:
		return sqlf.Sprintf("(SELECT COUNT(*) FROM notebook_stars WHERE notebook_id = notebooks.id) " + orderDirection)
	}
	panic("invalid NotebooksOrderByOption option")
}

func getNotebooksJoins(opts ListNotebooksOptions) *sqlf.Query {
	if opts.StarredByUserID != 0 {
		return sqlf.Sprintf("LEFT JOIN notebook_stars ON notebooks.id = notebook_stars.notebook_id")
	}
	return sqlf.Sprintf("")
}

func getNotebooksGroupByClause(opts ListNotebooksOptions) *sqlf.Query {
	if opts.StarredByUserID != 0 {
		return sqlf.Sprintf("GROUP BY notebooks.id")
	}
	return sqlf.Sprintf("")
}

func scanNotebook(scanner dbutil.Scanner) (*Notebook, error) {
	n := &Notebook{}
	err := scanner.Scan(
		&n.ID,
		&n.Title,
		&n.Blocks,
		&n.Public,
		&dbutil.NullInt32{N: &n.CreatorUserID},
		&dbutil.NullInt32{N: &n.UpdaterUserID},
		&dbutil.NullInt32{N: &n.NamespaceUserID},
		&dbutil.NullInt32{N: &n.NamespaceOrgID},
		&n.CreatedAt,
		&n.UpdatedAt,
		&n.PatternType,
	)
	if err != nil {
		return nil, err
	}
	return n, err
}

func scanNotebooks(rows *sql.Rows) ([]*Notebook, error) {
	var notebooks []*Notebook
	for rows.Next() {
		n, err := scanNotebook(rows)
		if err != nil {
			return nil, err
		}
		notebooks = append(notebooks, n)
	}
	return notebooks, nil
}

// Special characters used by TSQUERY we need to omit to prevent syntax errors.
// See: https://www.postgresql.org/docs/12/datatype-textsearch.html#DATATYPE-TSQUERY
var postgresTextSearchSpecialCharsRegex = lazyregexp.New(`&|!|\||\(|\)|:`)

func toPostgresTextSearchQuery(query string) string {
	tokens := strings.Fields(postgresTextSearchSpecialCharsRegex.ReplaceAllString(query, " "))
	prefixTokens := make([]string, len(tokens))
	for idx, token := range tokens {
		// :* is used for prefix matching
		prefixTokens[idx] = fmt.Sprintf("%s:*", token)
	}
	return strings.Join(prefixTokens, " & ")
}

func getNotebooksQueryCondition(opts ListNotebooksOptions) (*sqlf.Query, error) {
	if opts.NamespaceUserID != 0 && opts.NamespaceOrgID != 0 {
		return nil, errors.New("notebook list options NamespaceUserID and NamespaceOrgID are mutually exclusive")
	}

	conds := []*sqlf.Query{}
	if opts.CreatorUserID != 0 {
		conds = append(conds, sqlf.Sprintf("notebooks.creator_user_id = %d", opts.CreatorUserID))
	}
	if opts.NamespaceUserID != 0 {
		conds = append(conds, sqlf.Sprintf("notebooks.namespace_user_id = %d", opts.NamespaceUserID))
	}
	if opts.NamespaceOrgID != 0 {
		conds = append(conds, sqlf.Sprintf("notebooks.namespace_org_id = %d", opts.NamespaceOrgID))
	}
	if opts.StarredByUserID != 0 {
		conds = append(conds, sqlf.Sprintf("notebook_stars.user_id = %d", opts.StarredByUserID))
	}
	if opts.Query != "" {
		conds = append(
			conds,
			sqlf.Sprintf("(notebooks.title ILIKE %s OR notebooks.blocks_tsvector @@ to_tsquery('english', %s))", "%"+opts.Query+"%", toPostgresTextSearchQuery(opts.Query)),
		)
	}
	if len(conds) == 0 {
		// If no conditions are present, append a catch-all condition to avoid a SQL syntax error
		conds = append(conds, sqlf.Sprintf("1 = 1"))
	}
	return sqlf.Join(conds, "\n AND"), nil
}

func (s *notebooksStore) ListNotebooks(ctx context.Context, pageOpts ListNotebooksPageOptions, opts ListNotebooksOptions) ([]*Notebook, error) {
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
			getNotebooksGroupByClause(opts),
			getNotebooksOrderByClause(opts.OrderBy, opts.OrderByDescending),
			pageOpts.First,
			pageOpts.After,
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotebooks(rows)
}

func (s *notebooksStore) CountNotebooks(ctx context.Context, opts ListNotebooksOptions) (int64, error) {
	queryCondition, err := getNotebooksQueryCondition(opts)
	if err != nil {
		return -1, err
	}
	var count int64
	err = s.QueryRow(ctx,
		sqlf.Sprintf(
			countNotebooksFmtStr,
			getNotebooksJoins(opts),
			notebooksPermissionsCondition(ctx),
			queryCondition,
		),
	).Scan(&count)
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
	notebook, err := scanNotebook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotebookNotFound
	} else if err != nil {
		return nil, err
	}
	return notebook, nil
}

const insertNotebookFmtStr = `
INSERT INTO notebooks (title, blocks, public, creator_user_id, updater_user_id, namespace_user_id, namespace_org_id) VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

func (s *notebooksStore) CreateNotebook(ctx context.Context, n *Notebook) (*Notebook, error) {
	err := validateNotebookBlocks(n.Blocks)
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
			dbutil.NullInt32Column(n.CreatorUserID),
			dbutil.NullInt32Column(n.UpdaterUserID),
			dbutil.NullInt32Column(n.NamespaceUserID),
			dbutil.NullInt32Column(n.NamespaceOrgID),
			sqlf.Join(notebookColumns, ","),
		),
	)
	return scanNotebook(row)
}

const deleteNotebookFmtStr = `DELETE FROM notebooks WHERE id = %d`

// 🚨 SECURITY: The caller must ensure that the actor has permission to delete the notebook.
func (s *notebooksStore) DeleteNotebook(ctx context.Context, id int64) error {
	return s.Exec(ctx, sqlf.Sprintf(deleteNotebookFmtStr, id))
}

const updateNotebookFmtStr = `
UPDATE notebooks
SET
	title = %s,
	blocks = %s,
	public = %s,
	updater_user_id = %d,
	namespace_user_id = %d,
	namespace_org_id = %d,
	updated_at = now()
WHERE id = %d
RETURNING %s
`

// 🚨 SECURITY: The caller must ensure that the actor has permission to update the notebook.
func (s *notebooksStore) UpdateNotebook(ctx context.Context, n *Notebook) (*Notebook, error) {
	err := validateNotebookBlocks(n.Blocks)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			updateNotebookFmtStr,
			n.Title,
			n.Blocks,
			n.Public,
			dbutil.NullInt32Column(n.UpdaterUserID),
			dbutil.NullInt32Column(n.NamespaceUserID),
			dbutil.NullInt32Column(n.NamespaceOrgID),
			n.ID,
			sqlf.Join(notebookColumns, ","),
		),
	)
	return scanNotebook(row)
}

func scanNotebookStar(scanner dbutil.Scanner) (*NotebookStar, error) {
	star := &NotebookStar{}
	err := scanner.Scan(&star.NotebookID, &star.UserID, &star.CreatedAt)
	if err != nil {
		return nil, err
	}
	return star, nil
}

const getNotebookStarFmtStr = `SELECT notebook_id, user_id, created_at FROM notebook_stars WHERE notebook_id = %d AND user_id = %d`

// 🚨 SECURITY: The caller must ensure that the actor has permission to create the star for the notebook.
func (s *notebooksStore) GetNotebookStar(ctx context.Context, notebookID int64, userID int32) (*NotebookStar, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf(getNotebookStarFmtStr, notebookID, userID))
	star, err := scanNotebookStar(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotebookStarNotFound
	} else if err != nil {
		return nil, err
	}
	return star, nil
}

const insertNotebookStarFmtStr = `INSERT INTO notebook_stars (notebook_id, user_id) VALUES (%d, %d) RETURNING notebook_id, user_id, created_at`

// 🚨 SECURITY: The caller must ensure that the actor has permission to create the star for the notebook.
func (s *notebooksStore) CreateNotebookStar(ctx context.Context, notebookID int64, userID int32) (*NotebookStar, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf(insertNotebookStarFmtStr, notebookID, userID))
	return scanNotebookStar(row)
}

const deleteNotebookStarFmtStr = `DELETE FROM notebook_stars WHERE notebook_id = %d AND user_id = %d`

// 🚨 SECURITY: The caller must ensure that the actor has permission to delete the star for the notebook.
func (s *notebooksStore) DeleteNotebookStar(ctx context.Context, notebookID int64, userID int32) error {
	return s.Exec(ctx, sqlf.Sprintf(deleteNotebookStarFmtStr, notebookID, userID))
}

const listNotebookStarsFmtStr = `
SELECT notebook_id, user_id, created_at
FROM notebook_stars
WHERE notebook_id = %d
ORDER BY created_at DESC
LIMIT %d
OFFSET %d
`

// 🚨 SECURITY: The caller must ensure that the actor has permission to access the notebook.
func (s *notebooksStore) ListNotebookStars(ctx context.Context, pageOpts ListNotebookStarsPageOptions, notebookID int64) ([]*NotebookStar, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(listNotebookStarsFmtStr, notebookID, pageOpts.First, pageOpts.After))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notebookStars []*NotebookStar
	for rows.Next() {
		star, err := scanNotebookStar(rows)
		if err != nil {
			return nil, err
		}
		notebookStars = append(notebookStars, star)
	}
	return notebookStars, nil
}

const countNotebookStarsFmtStr = `SELECT COUNT(*) FROM notebook_stars WHERE notebook_id = %d`

// 🚨 SECURITY: The caller must ensure that the actor has permission to access the notebook.
func (s *notebooksStore) CountNotebookStars(ctx context.Context, notebookID int64) (int64, error) {
	var count int64
	err := s.QueryRow(ctx, sqlf.Sprintf(countNotebookStarsFmtStr, notebookID)).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}
