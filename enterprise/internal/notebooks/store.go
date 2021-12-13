package notebooks

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var ErrNotebookNotFound = errors.New("notebook not found")

type NotebooksOrderByOption uint8

const (
	NotebooksOrderByID NotebooksOrderByOption = iota
	NotebooksOrderByUpdatedAt
	NotebooksOrderByCreatedAt
)

type ListNotebooksPageOptions struct {
	First int32
	After int64
}

type ListNotebooksOptions struct {
	Query             string
	CreatorUserID     int32
	OrderBy           NotebooksOrderByOption
	OrderByDescending bool
}

func (blocks NotebookBlocks) Value() (driver.Value, error) {
	return json.Marshal(blocks)
}

func (blocks *NotebookBlocks) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &blocks)
}

func Notebooks(db dbutil.DB) NotebooksStore {
	store := basestore.NewWithDB(db, sql.TxOptions{})
	return &notebooksStore{store}
}

type NotebooksStore interface {
	basestore.ShareableStore
	GetNotebook(context.Context, int64) (*Notebook, error)
	CreateNotebook(context.Context, *Notebook) (*Notebook, error)
	UpdateNotebook(context.Context, *Notebook) (*Notebook, error)
	DeleteNotebook(context.Context, int64) error
	ListNotebooks(context.Context, ListNotebooksPageOptions, ListNotebooksOptions) ([]*Notebook, error)
	CountNotebooks(context.Context, ListNotebooksOptions) (int64, error)
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
    OR (notebooks.creator_user_id IS NOT NULL AND notebooks.creator_user_id = %d)
)`

var notebookColumns = []*sqlf.Query{
	sqlf.Sprintf("notebooks.id"),
	sqlf.Sprintf("notebooks.title"),
	sqlf.Sprintf("notebooks.blocks"),
	sqlf.Sprintf("notebooks.public"),
	sqlf.Sprintf("notebooks.creator_user_id"),
	sqlf.Sprintf("notebooks.created_at"),
	sqlf.Sprintf("notebooks.updated_at"),
}

func notebooksPermissionsCondition(ctx context.Context) *sqlf.Query {
	a := actor.FromContext(ctx)
	authenticatedUserID := int32(0)
	bypassPermissionsCheck := a.Internal
	if !bypassPermissionsCheck && a.IsAuthenticated() {
		authenticatedUserID = a.UID
	}
	return sqlf.Sprintf(notebooksPermissionsConditionFmtStr, bypassPermissionsCheck, authenticatedUserID)
}

const listNotebooksFmtStr = `
SELECT %s
FROM notebooks
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
ORDER BY %s
LIMIT %d
OFFSET %d
`

const countNotebooksFmtStr = `
SELECT COUNT(*)
FROM notebooks
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
	}
	panic("invalid NotebooksOrderByOption option")
}

func scanNotebook(scanner dbutil.Scanner) (*Notebook, error) {
	n := &Notebook{}
	err := scanner.Scan(
		&n.ID,
		&n.Title,
		&n.Blocks,
		&n.Public,
		&dbutil.NullInt32{N: &n.CreatorUserID},
		&n.CreatedAt,
		&n.UpdatedAt,
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

func getNotebooksQueryCondition(opts ListNotebooksOptions) *sqlf.Query {
	conds := []*sqlf.Query{}
	if opts.CreatorUserID != 0 {
		conds = append(conds, sqlf.Sprintf("notebooks.creator_user_id = %d", opts.CreatorUserID))
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
	return sqlf.Join(conds, "\n AND")
}

func (s *notebooksStore) ListNotebooks(ctx context.Context, pageOpts ListNotebooksPageOptions, opts ListNotebooksOptions) ([]*Notebook, error) {
	rows, err := s.Query(ctx,
		sqlf.Sprintf(
			listNotebooksFmtStr,
			sqlf.Join(notebookColumns, ","),
			notebooksPermissionsCondition(ctx),
			getNotebooksQueryCondition(opts),
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
	var count int64
	err := s.QueryRow(ctx,
		sqlf.Sprintf(
			countNotebooksFmtStr,
			notebooksPermissionsCondition(ctx),
			getNotebooksQueryCondition(opts),
		),
	).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (s *notebooksStore) GetNotebook(ctx context.Context, id int64) (*Notebook, error) {
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			listNotebooksFmtStr,
			sqlf.Join(notebookColumns, ","),
			notebooksPermissionsCondition(ctx),
			sqlf.Sprintf("notebooks.id = %d", id),
			getNotebooksOrderByClause(NotebooksOrderByID, false),
			1, // limit
			0, // offset
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
INSERT INTO notebooks (title, blocks, public, creator_user_id) VALUES (%s, %s, %s, %s)
RETURNING %s
`

func (s *notebooksStore) CreateNotebook(ctx context.Context, n *Notebook) (*Notebook, error) {
	err := validateNotebookBlocks(n.Blocks)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(insertNotebookFmtStr, n.Title, n.Blocks, n.Public, nullInt32Column(n.CreatorUserID), sqlf.Join(notebookColumns, ",")),
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
		sqlf.Sprintf(updateNotebookFmtStr, n.Title, n.Blocks, n.Public, n.ID, sqlf.Join(notebookColumns, ",")),
	)
	return scanNotebook(row)
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}
