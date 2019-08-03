package events

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbEventThread represents a thread's inclusion in a event.
type dbEventThread struct {
	Event int64 // the ID of the event
	Thread   int64 // the ID of the thread
}

type dbEventsThreads struct{}

// AddThreadsToEvent adds threads to the event.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the event and read
// the objects.
func (dbEventsThreads) AddThreadsToEvent(ctx context.Context, event int64, threads []int64) error {
	if mocks.eventsThreads.AddThreadsToEvent != nil {
		return mocks.eventsThreads.AddThreadsToEvent(event, threads)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		`
WITH insert_values AS (
  SELECT $1::bigint AS event_id, unnest($2::bigint[]) AS thread_id
)
INSERT INTO events_threads(event_id, thread_id) SELECT * FROM insert_values
`,
		event, pq.Array(threads),
	)
	return err
}

// RemoveThreadsFromEvent removes the specified threads from the thread.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the thread and the
// threads.
func (dbEventsThreads) RemoveThreadsFromEvent(ctx context.Context, event int64, threads []int64) error {
	if mocks.eventsThreads.RemoveThreadsFromEvent != nil {
		return mocks.eventsThreads.RemoveThreadsFromEvent(event, threads)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		`
WITH delete_values AS (
  SELECT $1::bigint AS event_id, unnest($2::bigint[]) AS thread_id
)
DELETE FROM events_threads o USING delete_values d WHERE o.event_id=d.event_id AND o.thread_id=d.thread_id
`,
		event, pq.Array(threads),
	)
	return err
}

// dbEventsThreadsListOptions contains options for listing threads.
type dbEventsThreadsListOptions struct {
	EventID int64 // only list threads for this event
	ThreadID   int64 // only list events for this thread
	*db.LimitOffset
}

func (o dbEventsThreadsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.EventID != 0 {
		conds = append(conds, sqlf.Sprintf("event_id=%d", o.EventID))
	}
	if o.ThreadID != 0 {
		conds = append(conds, sqlf.Sprintf("thread_id=%d", o.ThreadID))
	}
	return conds
}

// List lists all event-thread associations that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbEventsThreads) List(ctx context.Context, opt dbEventsThreadsListOptions) ([]*dbEventThread, error) {
	if mocks.eventsThreads.List != nil {
		return mocks.eventsThreads.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbEventsThreads) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbEventThread, error) {
	q := sqlf.Sprintf(`
SELECT event_id, thread_id FROM events_threads
WHERE (%s) AND thread_id IS NOT NULL
ORDER BY event_id ASC, thread_id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbEventsThreads) query(ctx context.Context, query *sqlf.Query) ([]*dbEventThread, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbEventThread
	for rows.Next() {
		var t dbEventThread
		if err := rows.Scan(&t.Event, &t.Thread); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// mockEventsThreads mocks the events-threads-related DB operations.
type mockEventsThreads struct {
	AddThreadsToEvent      func(thread int64, threads []int64) error
	RemoveThreadsFromEvent func(thread int64, threads []int64) error
	List                      func(dbEventsThreadsListOptions) ([]*dbEventThread, error)
}
