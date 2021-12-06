package notebooks

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
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
	OrderBy           NotebooksOrderByOption
	OrderByDescending bool
}

func Notebooks(db dbutil.DB) NotebooksStore {
	store := basestore.NewWithDB(db, sql.TxOptions{})
	return &notebooksStore{store}
}

type NotebooksStore interface {
	basestore.ShareableStore
	GetNotebook(context.Context, int64) (*Notebook, error)
	CreateNotebook(context.Context, *Notebook) (*Notebook, error)
	// TODO
	// UpdateNotebook(context.Context, *Notebook) (*Notebook, error)
	// DeleteNotebook(context.Context, int64) (*Notebook, error)
	// ListNotebooks(context.Context, ListNotebooksPageOptions, ListNotebooksOptions) ([]*Notebook, error)
	// CountNotebooks(context.Context, ListNotebooksOptions) (int, error)
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

func notebooksPermissionsCondition(ctx context.Context, db dbutil.DB) *sqlf.Query {
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

func scanNotebook(row *sql.Row) (*Notebook, error) {
	var blocksJSON json.RawMessage
	n := &Notebook{}
	err := row.Scan(
		&n.ID,
		&n.Title,
		&blocksJSON,
		&n.Public,
		&dbutil.NullInt32{N: &n.CreatorUserID},
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotebookNotFound
	} else if err != nil {
		return nil, err
	}
	var blocks []NotebookBlock
	if err = json.Unmarshal(blocksJSON, &blocks); err != nil {
		return nil, err
	}
	n.Blocks = blocks
	return n, nil
}

func (s *notebooksStore) GetNotebook(ctx context.Context, id int64) (*Notebook, error) {
	permissionsCond := notebooksPermissionsCondition(ctx, s.Handle().DB())
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(
			listNotebooksFmtStr,
			sqlf.Join(notebookColumns, ","),
			permissionsCond,
			sqlf.Sprintf("notebooks.id = %d", id),
			getNotebooksOrderByClause(NotebooksOrderByID, false),
			1, // limit
			0, // offset
		),
	)
	return scanNotebook(row)
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

	blocksJSON, err := json.Marshal(n.Blocks)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(
		ctx,
		sqlf.Sprintf(insertNotebookFmtStr, n.Title, blocksJSON, n.Public, nullInt32Column(n.CreatorUserID), sqlf.Join(notebookColumns, ",")),
	)
	return scanNotebook(row)
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}
