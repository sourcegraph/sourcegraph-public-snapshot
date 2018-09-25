package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/keegancsmith/sqlf"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// LimitOffset specifies SQL LIMIT and OFFSET counts. A pointer to it is typically embedded in other options
// structs that need to perform SQL queries with LIMIT and OFFSET.
type LimitOffset struct {
	Limit  int // SQL LIMIT count
	Offset int // SQL OFFSET count
}

// SQL returns the SQL query fragment ("LIMIT %d OFFSET %d") for use in SQL queries.
func (o *LimitOffset) SQL() *sqlf.Query {
	if o == nil {
		return &sqlf.Query{}
	}
	return sqlf.Sprintf("LIMIT %d OFFSET %d", o.Limit, o.Offset)
}

// Transaction calls f within a transaction, rolling back if any error is
// returned by the function.
func Transaction(ctx context.Context, db *sql.DB, f func(tx *sql.Tx) error) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Transaction")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if err2 := tx.Rollback(); err2 != nil {
				log.Println("Transaction Rollback failed:", err2)
			}
		}
	}()
	if err := f(tx); err != nil {
		return err
	}
	return tx.Commit()
}
