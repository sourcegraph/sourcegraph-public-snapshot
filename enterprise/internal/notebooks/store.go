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
	After int32
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
    OR n.public
    -- Private notebooks are available only to its creator
    OR (n.creator_user_id IS NOT NULL AND n.creator_user_id = %d)
)`

func notebooksPermissionsCondition(ctx context.Context, db dbutil.DB) (*sqlf.Query, error) {
	a := actor.FromContext(ctx)
	authenticatedUserID := int32(0)
	bypassPermissionsCheck := a.Internal
	if !bypassPermissionsCheck && a.IsAuthenticated() {
		authenticatedUserID = a.UID
	}
	q := sqlf.Sprintf(notebooksPermissionsConditionFmtStr, bypassPermissionsCheck, authenticatedUserID)
	return q, nil
}

const listNotebooksFmtStr = `
SELECT n.id, n.title, n.blocks, n.public, n.creator_user_id, n.created_at, n.updated_at
FROM notebooks n
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
		return sqlf.Sprintf("n.created_at " + orderDirection)
	case NotebooksOrderByUpdatedAt:
		return sqlf.Sprintf("n.updated_at " + orderDirection)
	case NotebooksOrderByID:
		return sqlf.Sprintf("n.id " + orderDirection)
	}
	panic("invalid NotebooksOrderByOption option")
}

func scanSingleNotebook(rows *sql.Rows) (*Notebook, error) {
	notebooks, err := scanNotebooks(rows)
	if err != nil {
		return nil, err
	}
	if len(notebooks) != 1 {
		return nil, ErrNotebookNotFound
	}
	return notebooks[0], nil
}

func scanNotebooks(rows *sql.Rows) ([]*Notebook, error) {
	var out []*Notebook
	for rows.Next() {
		var blocksJSON json.RawMessage
		n := &Notebook{}
		err := rows.Scan(
			&n.ID,
			&n.Title,
			&blocksJSON,
			&n.Public,
			&dbutil.NullInt32{N: &n.CreatorUserID},
			&n.CreatedAt,
			&n.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		var blocks []NotebookBlock
		if err = json.Unmarshal(blocksJSON, &blocks); err != nil {
			return nil, err
		}
		n.Blocks = blocks
		out = append(out, n)
	}
	return out, nil
}

func (s *notebooksStore) GetNotebook(ctx context.Context, id int64) (*Notebook, error) {
	permissionsCond, err := notebooksPermissionsCondition(ctx, s.Handle().DB())
	if err != nil {
		return nil, err
	}
	cond := sqlf.Sprintf("n.id = %d", id)
	rows, err := s.Query(
		ctx,
		sqlf.Sprintf(
			listNotebooksFmtStr,
			permissionsCond,
			cond,
			getNotebooksOrderByClause(NotebooksOrderByID, false),
			1, // limit
			0, // offset
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSingleNotebook(rows)
}

const insertNotebookFmtStr = `
INSERT INTO notebooks (title, blocks, public, creator_user_id) VALUES (%s, %s, %s, %s)
RETURNING id
`

func (s *notebooksStore) CreateNotebook(ctx context.Context, n *Notebook) (*Notebook, error) {
	blocksJSON, err := json.Marshal(n.Blocks)
	if err != nil {
		return nil, err
	}
	var id int64
	row := s.QueryRow(ctx, sqlf.Sprintf(insertNotebookFmtStr, n.Title, blocksJSON, n.Public, nullInt32Column(n.CreatorUserID)))
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return s.GetNotebook(ctx, id)
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}
