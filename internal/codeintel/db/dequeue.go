package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
)

type dequeueScanner func(rows *sql.Rows, err error) (interface{}, bool, error)

// dequeueRecord selects the record from the given table in a queued state according to the given sort expression.
// This record is locked in a transaction. The record and the transaction in which it is locked are both returned.
// This transaction must be closed by the caller. If there is no such unlocked record, a nil record and a nil DB
// will be returned along with a false-valued flag. This method must not be called from within a transaction.
//
// Assumptions: The table name describes a record with an `id`, `state`, and `started_at` column, where state can
// be one of (at least) 'queued' or 'processing'.
func (db *dbImpl) dequeueRecord(
	ctx context.Context,
	tableName string,
	columnExpressions []*sqlf.Query,
	sortExpression *sqlf.Query,
	scan dequeueScanner,
) (interface{}, DB, bool, error) {
	for {
		// First, we try to select an eligible record outside of a transaction. This will skip
		// any rows that are currently locked inside of a transaction of another dequeue process.
		id, ok, err := scanFirstInt(db.query(ctx, sqlf.Sprintf(`
			UPDATE `+tableName+` SET state = 'processing', started_at = now() WHERE id = (
				SELECT id FROM `+tableName+` WHERE state = 'queued' ORDER BY %s
				FOR UPDATE SKIP LOCKED LIMIT 1
			)
			RETURNING id
		`, sortExpression)))
		if err != nil || !ok {
			return nil, nil, false, err
		}

		record, tx, ok, err := db.dequeueByID(ctx, tableName, columnExpressions, scan, id)
		if err != nil {
			// This will occur if we selected an ID that raced with another dequeue process. If both
			// dequeue processes select the same ID and the other process begins its transaction first,
			// this condition will occur. We'll re-try the process by selecting a fresh ID.
			if err == ErrDequeueRace {
				continue
			}

			return nil, nil, false, errors.Wrap(err, "db.dequeue")
		}

		return record, tx, ok, nil
	}
}

// dequeueByID begins a transaction to lock an record for updating. This marks the record as ineligible
// to other dequeue processes. All updates to the database while this record is being processes should
// happen through returned transactional DB, which must be explicitly closed (via Done) at the end of
// processing by the caller.
func (db *dbImpl) dequeueByID(
	ctx context.Context,
	tableName string,
	columnExpressions []*sqlf.Query,
	scan dequeueScanner,
	id int,
) (_ interface{}, _ DB, _ bool, err error) {
	tx, started, err := db.transact(ctx)
	if err != nil {
		return nil, nil, false, err
	}
	if !started {
		return nil, nil, false, ErrDequeueTransaction
	}

	// SKIP LOCKED is necessary not to block on this select. We allow the database driver to return
	// sql.ErrNoRows on this condition so we can determine if we need to select a new record to process
	// on race conditions with other dequeue attempts.
	record, exists, err := scan(tx.query(
		ctx,
		sqlf.Sprintf(`SELECT %s FROM `+tableName+` WHERE id = %s FOR UPDATE SKIP LOCKED LIMIT 1`, sqlf.Join(columnExpressions, ","), id),
	))
	if err != nil {
		return nil, nil, false, tx.Done(err)
	}
	if !exists {
		return nil, nil, false, tx.Done(ErrDequeueRace)
	}
	return record, tx, true, nil
}
