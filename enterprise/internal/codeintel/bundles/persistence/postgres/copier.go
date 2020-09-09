package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type copyWriter struct {
	ch   chan []interface{}
	errs chan error
}

type preparer interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

func NewCopyInserter(ctx context.Context, db dbutil.DB, tableName string, columns ...string) (*copyWriter, error) {
	preparer, ok := db.(preparer)
	if !ok {
		return nil, fmt.Errorf("db is not a preparer")
	}

	stmt, err := preparer.PrepareContext(ctx, pq.CopyIn(tableName, columns...))
	if err != nil {
		return nil, err
	}

	ch := make(chan []interface{}, 512)
	errs := make(chan error, 3)
	ctx = disableTrace(ctx)

	go func() {
		defer close(errs)

		if err := func() (err error) {
			defer func() {
				if closeErr := stmt.Close(); closeErr != nil {
					err = multierror.Append(err, closeErr)
				}
			}()

			for row := range ch {
				if _, err := stmt.ExecContext(ctx, row...); err != nil {
					for range ch {
					}

					return err
				}
			}

			if _, err := stmt.ExecContext(ctx); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			errs <- err
		}
	}()

	return &copyWriter{ch: ch, errs: errs}, nil
}

func (w *copyWriter) Insert(ctx context.Context, values ...interface{}) error {
	w.ch <- values
	return nil
}

func (w *copyWriter) Flush(ctx context.Context) error {
	close(w.ch)
	return <-w.errs
}

func disableTrace(ctx context.Context) context.Context {
	return ot.WithShouldTrace(ctx, false)
}
