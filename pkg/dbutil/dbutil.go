package dbutil

import (
	"database/sql/driver"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/gorp.v1"
)

// Transact ensures that fn is called within a transaction. If dbh is a
// transaction (i.e., it has Begin, Commit, and Rollback methods),
// then just calls the function, otherwise, begins a transaction,
// rolling back on failure and committing on success.
func Transact(dbh gorp.SqlExecutor, fn func(dbh gorp.SqlExecutor) error) (err error) {
	dbh = GetUnderlyingSQLExecutor(dbh)

	type txLike interface {
		// transactions should have these methods no matter what their underlying type is
		Commit() error
		Rollback() error
	}
	type nonTxLike interface {
		// transactions can't have a Begin() method because they can only go 1 level deep
		Begin() (*gorp.Transaction, error)
	}

	tx, sharedTx := dbh.(txLike)
	_, isNonTxLike := dbh.(nonTxLike)
	beginNewTx := !sharedTx && isNonTxLike
	if beginNewTx {
		tx, err = dbh.(*gorp.DbMap).Begin() // enforce strict type here just to be cautious (what else has Begin?)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				if err := tx.Rollback(); err != nil {
					log.Println("dbutil.Transact Rollback failed:", err)
				}
			}
		}()
	}

	if tx == nil {
		panic(fmt.Sprintf("tx == nil; dbh type is %T", dbh))
	}

	if err = fn(tx.(gorp.SqlExecutor)); err != nil {
		return err
	}

	if beginNewTx {
		return tx.Commit()
	}
	return nil // don't commit the shared tx
}

// UnbindQuery interpolates the bind variables and returns the SQL query.
//
// SECURITY NOTE: It should only be used for logging; it is not secure to
// execute SQL that comes from UnbindQuery.
func UnbindQuery(query string, args ...interface{}) string {
	for i, arg := range args {
		// this is obviously a hack for logging; this SQL query should not ever
		// be executed because it would allow for SQL injection

		if dv, ok := arg.(driver.Valuer); ok {
			v, err := dv.Value()
			if err != nil {
				arg = fmt.Sprintf("'(Valuer).Value() error: %q'", err)
			}
			arg = v
		}
		switch a := arg.(type) {
		case int64:
			arg = fmt.Sprintf("%d", a)
		case int:
			arg = fmt.Sprintf("%d", a)
		case float64:
			arg = fmt.Sprintf("%f", a)
		case []byte:
			arg = fmt.Sprintf("'%s'", a)
		case time.Time:
			arg = fmt.Sprintf("'%s'", a)
		case string:
			arg = fmt.Sprintf("'%s'", a)
		}

		query = strings.Replace(query, fmt.Sprintf("$%d", i+1), fmt.Sprintf("%v", arg), -1)
	}
	return query
}
