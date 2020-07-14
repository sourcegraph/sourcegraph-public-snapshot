package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
)

type dequeueScanner func(rows *sql.Rows, err error) (interface{}, bool, error)

// dequeueRecord selects the record from the given table in a queued state according to the given sort expression.
// This record is locked in a transaction. The record and the transaction in which it is locked are both returned.
// This transaction must be closed by the caller. If there is no such unlocked record, a nil record and a nil store
// will be returned along with a false-valued flag. This method must not be called from within a transaction.
//
// Assumptions: The table name describes a record with an `id`, `state`, `started_at`, `process_after`, and
// `repository_id` column, where state can be one of (at least) 'queued' or 'processing' and repository_id refers
// to the PK of the repo table..
func (s *store) dequeueRecord(
	ctx context.Context,
	viewName string,
	tableName string,
	columnExpressions []*sqlf.Query,
	sortExpression *sqlf.Query,
	additionalConditions []*sqlf.Query,
	scan dequeueScanner,
) (interface{}, Store, bool, error) {
	conditions := []*sqlf.Query{
		sqlf.Sprintf("state = 'queued'"),
		sqlf.Sprintf("(process_after IS NULL OR process_after >= NOW())"),
	}
	conditions = append(conditions, additionalConditions...)

	for {
		// First, we try to select an eligible record outside of a transaction. This will skip
		// any rows that are currently locked inside of a transaction of another dequeue process.
		id, ok, err := scanFirstInt(s.query(ctx, sqlf.Sprintf(`
			WITH candidate AS (
				SELECT id FROM `+tableName+`
					WHERE %s
					ORDER BY %s
					FOR UPDATE SKIP LOCKED
					LIMIT 1
			)
			UPDATE `+tableName+` SET state = 'processing', started_at = now() WHERE id IN (SELECT id FROM candidate)
			RETURNING id
		`, sqlf.Join(conditions, " AND "), sortExpression)))
		if err != nil || !ok {
			return nil, nil, false, err
		}

		record, tx, ok, err := s.dequeueByID(ctx, viewName, tableName, columnExpressions, scan, id)
		if err != nil {
			// This will occur if we selected an ID that raced with another dequeue process. If both
			// dequeue processes select the same ID and the other process begins its transaction first,
			// this condition will occur. We'll re-try the process by selecting a fresh ID.
			if err == ErrDequeueRace {
				continue
			}

			return nil, nil, false, errors.Wrap(err, "store.dequeueByID")
		}

		return record, tx, ok, nil
	}
}

// dequeueByID begins a transaction to lock an record for updating. This marks the record as ineligible
// to other dequeue processes. All updates to the database while this record is being processes should
// happen through returned transactional store, which must be explicitly closed (via Done) at the end of
// processing by the caller.
func (s *store) dequeueByID(
	ctx context.Context,
	viewName string,
	tableName string,
	columnExpressions []*sqlf.Query,
	scan dequeueScanner,
	id int,
) (_ interface{}, _ Store, _ bool, err error) {
	if s.InTransaction() {
		return nil, nil, false, ErrDequeueTransaction
	}
	tx, err := s.transact(ctx)
	if err != nil {
		return nil, nil, false, err
	}

	// SKIP LOCKED is necessary not to block on this select. We allow the database driver to return
	// sql.ErrNoRows on this condition so we can determine if we need to select a new record to process
	// on race conditions with other dequeue attempts.
	_, exists, err := scanFirstInt(tx.query(
		ctx,
		sqlf.Sprintf(`SELECT 1 FROM `+tableName+` u WHERE u.id = %s FOR UPDATE SKIP LOCKED LIMIT 1`, id),
	))
	if err != nil {
		return nil, nil, false, tx.Done(err)
	}
	if !exists {
		return nil, nil, false, tx.Done(ErrDequeueRace)
	}

	record, _, err := scan(tx.query(
		ctx,
		sqlf.Sprintf(`SELECT %s FROM `+viewName+` u WHERE u.id = %s LIMIT 1`, sqlf.Join(columnExpressions, ","), id),
	))
	if err != nil {
		return nil, nil, false, tx.Done(err)
	}

	return record, tx, true, nil
}
