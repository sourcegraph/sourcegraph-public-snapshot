package commitstatuses

import (
	"context"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/nnz"
)

// dbCommitStatusContext describes a commitStatus for a discussion thread.
type dbCommitStatusContext struct {
	ID           int64
	RepositoryID api.RepoID
	CommitOID    api.CommitID
	Context      string
	State        string
	Description  *string
	TargetURL    *string
	Actor        actor.DBColumns
	CreatedAt    time.Time
}

// errCommitStatusContextNotFound occurs when a database operation expects a specific commit status
// context to exist but it does not exist.
var errCommitStatusContextNotFound = errors.New("commit status context not found")

type dbCommitStatusContexts struct{}

const selectColumns = `id, repository_id, commit_oid, context, state, description, target_url, actor_user_id, external_actor_username, external_actor_url, created_at`

// Create creates a commit status context. The argument's ID field is ignored.
func (dbCommitStatusContexts) Create(ctx context.Context, commitStatusContext *dbCommitStatusContext) (*dbCommitStatusContext, error) {
	if mocks.commitStatuses.Create != nil {
		return mocks.commitStatuses.Create(commitStatusContext)
	}

	args := []interface{}{
		commitStatusContext.RepositoryID,
		commitStatusContext.CommitOID,
		commitStatusContext.Context,
		commitStatusContext.State,
		commitStatusContext.Description,
		commitStatusContext.TargetURL,
		nnz.Int32(commitStatusContext.Actor.UserID),
		nnz.String(commitStatusContext.Actor.ExternalActorUsername),
		nnz.String(commitStatusContext.Actor.ExternalActorURL),
	}
	query := sqlf.Sprintf(
		`INSERT INTO commit_status_contexts(`+selectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`) RETURNING `+selectColumns,
		args...,
	)
	return dbCommitStatusContexts{}.scanRow(dbconn.Global.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...))
}

// GetByID retrieves the commit status context (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this commit status context.
func (s dbCommitStatusContexts) GetByID(ctx context.Context, id int64) (*dbCommitStatusContext, error) {
	if mocks.commitStatuses.GetByID != nil {
		return mocks.commitStatuses.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errCommitStatusContextNotFound
	}
	return results[0], nil
}

// dbCommitStatusContextsListOptions contains options for listing commitStatuses.
type dbCommitStatusContextsListOptions struct {
	RepositoryID api.RepoID   // TODO!(sqs): enforce this is required
	CommitOID    api.CommitID // TODO!(sqs): enforce this is required
}

func (o dbCommitStatusContextsListOptions) sqlConditions() []*sqlf.Query {
	return []*sqlf.Query{
		sqlf.Sprintf("repository_id=%d", o.RepositoryID),
		sqlf.Sprintf("commit_oid=%s", o.CommitOID),
	}
}

// List lists all commit status contexts that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbCommitStatusContexts) List(ctx context.Context, opt dbCommitStatusContextsListOptions) ([]*dbCommitStatusContext, error) {
	if mocks.commitStatuses.List != nil {
		return mocks.commitStatuses.List(opt)
	}
	return s.list(ctx, opt.sqlConditions())
}

func (s dbCommitStatusContexts) list(ctx context.Context, conds []*sqlf.Query) ([]*dbCommitStatusContext, error) {
	q := sqlf.Sprintf(`
SELECT `+selectColumns+` FROM commit_status_contexts
WHERE (%s)
ORDER BY created_at ASC`,
		sqlf.Join(conds, ") AND ("),
	)
	return s.query(ctx, q)
}

func (dbCommitStatusContexts) query(ctx context.Context, query *sqlf.Query) ([]*dbCommitStatusContext, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbCommitStatusContext
	for rows.Next() {
		t, err := dbCommitStatusContexts{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (dbCommitStatusContexts) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*dbCommitStatusContext, error) {
	var t dbCommitStatusContext
	if err := row.Scan(
		&t.ID,
		&t.RepositoryID,
		&t.CommitOID,
		&t.Context,
		&t.State,
		&t.Description,
		&t.TargetURL,
		nnz.ToInt32(&t.Actor.UserID),
		(*nnz.String)(&t.Actor.ExternalActorUsername),
		(*nnz.String)(&t.Actor.ExternalActorURL),
		&t.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// mockCommitStatuses mocks the commitStatuses-related DB operations.
type mockCommitStatuses struct {
	Create  func(*dbCommitStatusContext) (*dbCommitStatusContext, error)
	GetByID func(int64) (*dbCommitStatusContext, error)
	List    func(dbCommitStatusContextsListOptions) ([]*dbCommitStatusContext, error)
}
