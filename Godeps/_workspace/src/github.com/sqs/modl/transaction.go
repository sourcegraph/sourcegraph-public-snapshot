package modl

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Transaction represents a database transaction.
// Insert/Update/Delete/Get/Exec operations will be run in the context
// of that transaction.  Transactions should be terminated with
// a call to Commit() or Rollback()
type Transaction struct {
	dbmap *DbMap
	Tx    *sqlx.Tx
}

// Insert has the same behavior as DbMap.Insert(), but runs in a transaction.
func (t *Transaction) Insert(list ...interface{}) error {
	return insert(t.dbmap, t, list...)
}

// Update has the same behavior as DbMap.Update(), but runs in a transaction.
func (t *Transaction) Update(list ...interface{}) (int64, error) {
	return update(t.dbmap, t, list...)
}

// Delete has the same behavior as DbMap.Delete(), but runs in a transaction.
func (t *Transaction) Delete(list ...interface{}) (int64, error) {
	return deletes(t.dbmap, t, list...)
}

// Get has the Same behavior as DbMap.Get(), but runs in a transaction.
func (t *Transaction) Get(dest interface{}, keys ...interface{}) error {
	return get(t.dbmap, t, dest, keys...)
}

// Select has the Same behavior as DbMap.Select(), but runs in a transaction.
func (t *Transaction) Select(dest interface{}, query string, args ...interface{}) error {
	return hookedselect(t.dbmap, t, dest, query, args...)
}

func (t *Transaction) SelectOne(dest interface{}, query string, args ...interface{}) error {
	return hookedget(t.dbmap, t, dest, query, args...)
}

// Exec has the same behavior as DbMap.Exec(), but runs in a transaction.
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	t.dbmap.trace(query, args)
	return t.Tx.Exec(query, args...)
}

// Commit commits the underlying database transaction.
func (t *Transaction) Commit() error {
	t.dbmap.trace("commit;")
	return t.Tx.Commit()
}

// Rollback rolls back the underlying database transaction.
func (t *Transaction) Rollback() error {
	t.dbmap.trace("rollback;")
	return t.Tx.Rollback()
}

func (t *Transaction) handle() handle {
	return &tracingHandle{h: t.Tx, d: t.dbmap}
}
