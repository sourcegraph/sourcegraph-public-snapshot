package labels

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbObjectLabel describes an association between a discussion thread and a label.
type dbObjectLabel struct {
	Label  int64 // the ID of the label
	Thread int64 // the ID of the thread
}

type dbLabelsObjects struct{}

// AddLabelsToThread labels the thread with the specified labels.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the thread and the
// labels.
func (dbLabelsObjects) AddLabelsToThread(ctx context.Context, thread int64, labels []int64) error {
	if mocks.labelsObjects.AddLabelsToThread != nil {
		return mocks.labelsObjects.AddLabelsToThread(thread, labels)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		// Include discussion_threads table query (with "FOR UPDATE") to ensure that the thread has not been
		// deleted. If it was deleted, the query will return an error.
		`
WITH labelable_object AS (
  SELECT id FROM discussion_threads WHERE id=$1 AND deleted_at IS NULL FOR UPDATE
),
insert_values AS (
  SELECT unnest($2::bigint[]) AS label_id, labelable_object.id AS thread_id
  FROM labelable_object
)
INSERT INTO labels_objects(label_id, thread_id) SELECT * FROM insert_values
`,
		thread, pq.Array(labels),
	)
	return err
}

// RemoveLabelsFromThread removes the specified labels from the thread.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the thread and the
// labels.
func (dbLabelsObjects) RemoveLabelsFromThread(ctx context.Context, thread int64, labels []int64) error {
	if mocks.labelsObjects.RemoveLabelsFromThread != nil {
		return mocks.labelsObjects.RemoveLabelsFromThread(thread, labels)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		// Include discussion_threads table query (with "FOR UPDATE") to ensure that the thread has not been
		// deleted. If it was deleted, the query will return an error.
		`
WITH labelable_object AS (
  SELECT id FROM discussion_threads WHERE id=$1 AND deleted_at IS NULL FOR UPDATE
),
delete_values AS (
  SELECT unnest($2::bigint[]) AS label_id, labelable_object.id AS thread_id
  FROM labelable_object
)
DELETE FROM labels_objects o USING delete_values d WHERE o.label_id=d.label_id AND o.thread_id=d.thread_id
`,
		thread, pq.Array(labels),
	)
	return err
}

// dbLabelsObjectsListOptions contains options for listing labels.
type dbLabelsObjectsListOptions struct {
	LabelID  int64 // only list threads for this label
	ThreadID int64 // only list labels for this thread
	*db.LimitOffset
}

func (o dbLabelsObjectsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.LabelID != 0 {
		conds = append(conds, sqlf.Sprintf("label_id=%d", o.LabelID))
	}
	if o.ThreadID != 0 {
		conds = append(conds, sqlf.Sprintf("thread_id=%d", o.ThreadID))
	}
	return conds
}

// List lists all label associations that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbLabelsObjects) List(ctx context.Context, opt dbLabelsObjectsListOptions) ([]*dbObjectLabel, error) {
	if mocks.labelsObjects.List != nil {
		return mocks.labelsObjects.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbLabelsObjects) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbObjectLabel, error) {
	q := sqlf.Sprintf(`
SELECT label_id, thread_id FROM labels_objects
WHERE (%s) AND thread_id IS NOT NULL
ORDER BY label_id ASC, thread_id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbLabelsObjects) query(ctx context.Context, query *sqlf.Query) ([]*dbObjectLabel, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbObjectLabel
	for rows.Next() {
		var t dbObjectLabel
		if err := rows.Scan(&t.Label, &t.Thread); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// mockLabelsObjects mocks the labels-objects-related DB operations.
type mockLabelsObjects struct {
	AddLabelsToThread      func(thread int64, labels []int64) error
	RemoveLabelsFromThread func(thread int64, labels []int64) error
	List                   func(dbLabelsObjectsListOptions) ([]*dbObjectLabel, error)
}
