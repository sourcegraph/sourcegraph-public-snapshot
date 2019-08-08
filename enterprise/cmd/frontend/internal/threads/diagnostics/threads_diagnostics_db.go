package diagnostics

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbThreadDiagnostic represents a diagnostic's inclusion in a thread.
type dbThreadDiagnostic struct {
	ID       int64
	ThreadID int64
	//TODO!(sqs) LocationRepositoryID api.RepoID
	Type string
	Data json.RawMessage
}

// errThreadDiagnosticNotFound occurs when a database operation expects a specific thread diagnostic
// to exist but it does not exist.
var errThreadDiagnosticNotFound = errors.New("thread diagnostic not found")

type dbThreadDiagnosticEdges struct{}

const selectColumns = `id, thread_id, type, data`

// Create adds a diagnostic to the thread.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the thread.
func (dbThreadDiagnosticEdges) Create(ctx context.Context, threadDiagnostic dbThreadDiagnostic) (id int64, err error) {
	if mocks.threadsDiagnostics.Create != nil {
		return mocks.threadsDiagnostics.Create(threadDiagnostic)
	}

	args := []interface{}{
		threadDiagnostic.ThreadID,
		threadDiagnostic.Type,
		threadDiagnostic.Data,
	}
	query := sqlf.Sprintf(
		`INSERT INTO thread_diagnostic_edges(`+selectColumns+`) VALUES(DEFAULT`+strings.Repeat(", %v", len(args))+`) RETURNING id`,
		args...,
	)
	if err := dbconn.Global.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID retrieves the thread diagnostic (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this thread diagnostic.
func (s dbThreadDiagnosticEdges) GetByID(ctx context.Context, id int64) (*dbThreadDiagnostic, error) {
	if mocks.threadsDiagnostics.GetByID != nil {
		return mocks.threadsDiagnostics.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errThreadDiagnosticNotFound
	}
	return results[0], nil
}

// dbThreadDiagnosticEdgesListOptions contains options for listing threads.
type dbThreadDiagnosticEdgesListOptions struct {
	IDs        []int64
	ThreadID   int64 // only list diagnostics for this thread
	CampaignID int64 // only list diagnostics for threads in this campaign
	*db.LimitOffset
}

func (o dbThreadDiagnosticEdgesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.IDs != nil {
		conds = append(conds, sqlf.Sprintf("id = ANY(%v)", pq.Array(o.IDs)))
	}
	if o.ThreadID != 0 {
		conds = append(conds, sqlf.Sprintf("thread_id=%d", o.ThreadID))
	}
	if o.CampaignID != 0 {
		conds = append(conds, sqlf.Sprintf("thread_id IN (SELECT thread_id FROM campaigns_threads WHERE campaign_id=%d)", o.CampaignID))
	}
	return conds
}

// List lists all thread diagnostics that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
//options.
func (s dbThreadDiagnosticEdges) List(ctx context.Context, opt dbThreadDiagnosticEdgesListOptions) ([]*dbThreadDiagnostic, error) {
	if mocks.threadsDiagnostics.List != nil {
		return mocks.threadsDiagnostics.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbThreadDiagnosticEdges) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbThreadDiagnostic, error) {
	q := sqlf.Sprintf(`
SELECT `+selectColumns+` FROM thread_diagnostic_edges
WHERE (%s)
ORDER BY id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbThreadDiagnosticEdges) query(ctx context.Context, query *sqlf.Query) ([]*dbThreadDiagnostic, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbThreadDiagnostic
	for rows.Next() {
		t, err := dbThreadDiagnosticEdges{}.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}

func (dbThreadDiagnosticEdges) scanRow(row interface {
	Scan(dest ...interface{}) error
}) (*dbThreadDiagnostic, error) {
	var t dbThreadDiagnostic
	if err := row.Scan(
		&t.ID,
		&t.ThreadID,
		&t.Type,
		&t.Data,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// Count counts all thread diagnostics that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count.
func (dbThreadDiagnosticEdges) Count(ctx context.Context, opt dbThreadDiagnosticEdgesListOptions) (int, error) {
	if mocks.threadsDiagnostics.Count != nil {
		return mocks.threadsDiagnostics.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM thread_diagnostic_edges WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteByID removes a diagnostic from the thread.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the thread and the
// diagnostics.
func (s dbThreadDiagnosticEdges) DeleteByIDInThread(ctx context.Context, threadDiagnosticID, threadID int64) error {
	if mocks.threadsDiagnostics.DeleteByID != nil {
		return mocks.threadsDiagnostics.DeleteByID(threadDiagnosticID, threadID)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d AND thread_id=%d", threadDiagnosticID, threadID))
}

func (dbThreadDiagnosticEdges) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM thread_diagnostic_edges WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errThreadDiagnosticNotFound
	}
	return nil
}

// mockThreadsDiagnostics mocks the campaigns-threads-related DB operations.
type mockThreadsDiagnostics struct {
	Create     func(dbThreadDiagnostic) (int64, error)
	GetByID    func(int64) (*dbThreadDiagnostic, error)
	List       func(dbThreadDiagnosticEdgesListOptions) ([]*dbThreadDiagnostic, error)
	Count      func(dbThreadDiagnosticEdgesListOptions) (int, error)
	DeleteByID func(threadDiagnosticID, threadID int64) error
}
