package threads

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbThread describes a thread.
type dbThread struct {
	ID           int64
	RepositoryID api.RepoID // the repository associated with this thread
	Title        string
	ExternalURL  *string
}

// errThreadNotFound occurs when a database operation expects a specific thread to exist but it does
// not exist.
var errThreadNotFound = errors.New("thread not found")

type dbThreads struct{}

// Create creates a thread. The thread argument's (Thread).ID field is ignored. The database ID of
// the new thread is returned.
func (dbThreads) Create(ctx context.Context, thread *dbThread) (*dbThread, error) {
	if mocks.threads.Create != nil {
		return mocks.threads.Create(thread)
	}

	var id int64
	if err := dbconn.Global.QueryRowContext(ctx,
		`INSERT INTO threads(repository_id, title, external_url) VALUES($1, $2, $3) RETURNING id`,
		thread.RepositoryID, thread.Title, thread.ExternalURL,
	).Scan(&id); err != nil {
		return nil, err
	}
	created := *thread
	created.ID = id
	return &created, nil
}

type dbThreadUpdate struct {
	Title       *string
	ExternalURL *string
}

// Update updates a thread given its ID.
func (s dbThreads) Update(ctx context.Context, id int64, update dbThreadUpdate) (*dbThread, error) {
	if mocks.threads.Update != nil {
		return mocks.threads.Update(id, update)
	}

	var setFields []*sqlf.Query
	if update.Title != nil {
		setFields = append(setFields, sqlf.Sprintf("title=%s", *update.Title))
	}
	if update.ExternalURL != nil {
		// Treat empty string as meaning "set to null". Otherwise there is no way to express that
		// intent.
		var value *string
		if *update.ExternalURL != "" {
			value = update.ExternalURL
		}
		setFields = append(setFields, sqlf.Sprintf("external_url=%s", value))
	}

	if len(setFields) == 0 {
		return nil, nil
	}

	results, err := s.query(ctx, sqlf.Sprintf(`UPDATE threads SET %v WHERE id=%s RETURNING id, repository_id, title, external_url`, sqlf.Join(setFields, ", "), id))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errThreadNotFound
	}
	return results[0], nil
}

// GetByID retrieves the thread (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this thread.
func (s dbThreads) GetByID(ctx context.Context, id int64) (*dbThread, error) {
	if mocks.threads.GetByID != nil {
		return mocks.threads.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errThreadNotFound
	}
	return results[0], nil
}

// dbThreadsListOptions contains options for listing threads.
type dbThreadsListOptions struct {
	Query        string     // only list threads matching this query (case-insensitively)
	RepositoryID api.RepoID // only list threads in this repository
	ThreadIDs    []int64
	*db.LimitOffset
}

func (o dbThreadsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf("title LIKE %s", "%"+o.Query+"%"))
	}
	if o.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("repository_id=%d", o.RepositoryID))
	}
	if o.ThreadIDs != nil {
		if len(o.ThreadIDs) > 0 {
			conds = append(conds, sqlf.Sprintf("id = ANY(%v)", pq.Array(o.ThreadIDs)))
		} else {
			conds = append(conds, sqlf.Sprintf("FALSE"))
		}
	}
	return conds
}

// List lists all threads that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbThreads) List(ctx context.Context, opt dbThreadsListOptions) ([]*dbThread, error) {
	if mocks.threads.List != nil {
		return mocks.threads.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbThreads) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbThread, error) {
	q := sqlf.Sprintf(`
SELECT id, repository_id, title, external_url FROM threads
WHERE (%s)
ORDER BY title ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbThreads) query(ctx context.Context, query *sqlf.Query) ([]*dbThread, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbThread
	for rows.Next() {
		var t dbThread
		if err := rows.Scan(&t.ID, &t.RepositoryID, &t.Title, &t.ExternalURL); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all threads that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the threads.
func (dbThreads) Count(ctx context.Context, opt dbThreadsListOptions) (int, error) {
	if mocks.threads.Count != nil {
		return mocks.threads.Count(opt)
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM threads WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete deletes a thread given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the thread.
func (s dbThreads) DeleteByID(ctx context.Context, id int64) error {
	if mocks.threads.DeleteByID != nil {
		return mocks.threads.DeleteByID(id)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d", id))
}

func (dbThreads) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("TRUE")}
	q := sqlf.Sprintf("DELETE FROM threads WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errThreadNotFound
	}
	return nil
}

// mockThreads mocks the threads-related DB operations.
type mockThreads struct {
	Create     func(*dbThread) (*dbThread, error)
	Update     func(int64, dbThreadUpdate) (*dbThread, error)
	GetByID    func(int64) (*dbThread, error)
	List       func(dbThreadsListOptions) ([]*dbThread, error)
	Count      func(dbThreadsListOptions) (int, error)
	DeleteByID func(int64) error
}

// TestCreateThread creates a thread in the DB, for use in tests only.
func TestCreateThread(ctx context.Context, title string, repositoryID api.RepoID) (id int64, err error) {
	thread, err := dbThreads{}.Create(ctx, &dbThread{
		RepositoryID: repositoryID,
		Title:        title,
	})
	if err != nil {
		return 0, err
	}
	return thread.ID, nil
}
