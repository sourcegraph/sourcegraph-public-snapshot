package dbutil

import (
	"context"
	"database/sql"

	multierror "github.com/hashicorp/go-multierror"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Transaction calls f within a transaction, rolling back if any error is
// returned by the function.
func Transaction(ctx context.Context, db *sql.DB, f func(tx *sql.Tx) error) (err error) {
	finish := func(tx *sql.Tx) {
		if err != nil {
			if err2 := tx.Rollback(); err2 != nil {
				err = multierror.Append(err, err2)
			}
			return
		}
		err = tx.Commit()
	}

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
	defer finish(tx)
	return f(tx)
}
