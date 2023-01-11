package stores

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type noopDB struct{}

type noopHandle struct {
	noopDB
}

var (
	NoopDB     = noopDB{}
	NoopHandle = noopHandle{}
	ErrNoop    = errors.New("this service is initialized without a connection to CodeIntelDB")
)

func (n noopDB) Handle() basestore.TransactableHandle                               { return NoopHandle }
func (n noopHandle) Transact(context.Context) (basestore.TransactableHandle, error) { return n, nil }
func (n noopDB) InTransaction() bool                                                { return false }
func (n noopDB) Transact(context.Context) (CodeIntelDB, error)                      { return n, nil }
func (n noopDB) Done(err error) error                                               { return err }

func (n noopDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return nil, ErrNoop
}

func (n noopDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, ErrNoop
}

func (n noopDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	// Unfortunately, can't do much about this one as it's a concrete type
	// with no exported fields or constructors in the defining package.

	return nil
}
