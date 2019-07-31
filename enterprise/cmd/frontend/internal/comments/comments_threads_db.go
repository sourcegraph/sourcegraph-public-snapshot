package comments

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbCommentThread represents a thread's inclusion in a comment.
type dbCommentThread struct {
	Comment int64 // the ID of the comment
	Thread   int64 // the ID of the thread
}

type dbCommentsThreads struct{}

// AddThreadsToComment adds threads to the comment.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the comment and read
// the objects.
func (dbCommentsThreads) AddThreadsToComment(ctx context.Context, comment int64, threads []int64) error {
	if mocks.commentsThreads.AddThreadsToComment != nil {
		return mocks.commentsThreads.AddThreadsToComment(comment, threads)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		`
WITH insert_values AS (
  SELECT $1::bigint AS comment_id, unnest($2::bigint[]) AS thread_id
)
INSERT INTO comments_threads(comment_id, thread_id) SELECT * FROM insert_values
`,
		comment, pq.Array(threads),
	)
	return err
}

// RemoveThreadsFromComment removes the specified threads from the thread.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the thread and the
// threads.
func (dbCommentsThreads) RemoveThreadsFromComment(ctx context.Context, comment int64, threads []int64) error {
	if mocks.commentsThreads.RemoveThreadsFromComment != nil {
		return mocks.commentsThreads.RemoveThreadsFromComment(comment, threads)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		`
WITH delete_values AS (
  SELECT $1::bigint AS comment_id, unnest($2::bigint[]) AS thread_id
)
DELETE FROM comments_threads o USING delete_values d WHERE o.comment_id=d.comment_id AND o.thread_id=d.thread_id
`,
		comment, pq.Array(threads),
	)
	return err
}

// dbCommentsThreadsListOptions contains options for listing threads.
type dbCommentsThreadsListOptions struct {
	CommentID int64 // only list threads for this comment
	ThreadID   int64 // only list comments for this thread
	*db.LimitOffset
}

func (o dbCommentsThreadsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.CommentID != 0 {
		conds = append(conds, sqlf.Sprintf("comment_id=%d", o.CommentID))
	}
	if o.ThreadID != 0 {
		conds = append(conds, sqlf.Sprintf("thread_id=%d", o.ThreadID))
	}
	return conds
}

// List lists all comment-thread associations that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbCommentsThreads) List(ctx context.Context, opt dbCommentsThreadsListOptions) ([]*dbCommentThread, error) {
	if mocks.commentsThreads.List != nil {
		return mocks.commentsThreads.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbCommentsThreads) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbCommentThread, error) {
	q := sqlf.Sprintf(`
SELECT comment_id, thread_id FROM comments_threads
WHERE (%s) AND thread_id IS NOT NULL
ORDER BY comment_id ASC, thread_id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbCommentsThreads) query(ctx context.Context, query *sqlf.Query) ([]*dbCommentThread, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbCommentThread
	for rows.Next() {
		var t dbCommentThread
		if err := rows.Scan(&t.Comment, &t.Thread); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// mockCommentsThreads mocks the comments-threads-related DB operations.
type mockCommentsThreads struct {
	AddThreadsToComment      func(thread int64, threads []int64) error
	RemoveThreadsFromComment func(thread int64, threads []int64) error
	List                      func(dbCommentsThreadsListOptions) ([]*dbCommentThread, error)
}
