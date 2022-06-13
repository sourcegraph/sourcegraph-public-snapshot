package basestore

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store is an abstract Postgres-backed data access layer. Instances of this struct
// should not be used directly, but should be used compositionally by other stores
// that implement logic specific to a domain.
//
// The following is a minimal example of decorating the base store that preserves
// the correct behavior of the underlying base store. Note that `With` and `Transact`
// must be re-defined in the outer layer in order to create a useful return value.
// Failure to re-define these methods will result in `With` and `Transact` methods that
// return a modified base store with no methods from the outer layer. All other methods
// of the base store are available on the outer layer without needing to be re-defined.
//
//     type SprocketStore struct {
//         *basestore.Store
//     }
//
//     func NewWithDB(database dbutil.DB) *SprocketStore {
//         return &SprocketStore{Store: basestore.NewWithDB(database, sql.TxOptions{})}
//     }
//
//     func (s *SprocketStore) With(other basestore.ShareableStore) *SprocketStore {
//         return &SprocketStore{Store: s.Store.With(other)}
//     }
//
//     func (s *SprocketStore) Transact(ctx context.Context) (*SprocketStore, error) {
//         txBase, err := s.Store.Transact(ctx)
//         return &SprocketStore{Store: txBase}, err
//     }
type Store struct {
	handle TransactableHandle
}

// ShareableStore is implemented by stores to explicitly allow distinct store instances
// to reference the store's underlying handle. This is used to share transactions between
// multiple stores. See `Store.With` for additional details.
type ShareableStore interface {
	// Handle returns the underlying transactable database handle.
	Handle() TransactableHandle
}

var _ ShareableStore = &Store{}

// NewWithHandle returns a new base store using the given database handle.
func NewWithHandle(handle TransactableHandle) *Store {
	return &Store{handle: handle}
}

// Handle returns the underlying transactable database handle.
func (s *Store) Handle() TransactableHandle {
	return s.handle
}

// With creates a new store with the underlying database handle from the given store.
// This method should be used when two distinct store instances need to perform an
// operation within the same shared transaction.
//
//     txn1 := store1.Transact(ctx) // Creates a transaction
//     txn2 := store2.With(txn1)    // References the same transaction
//
//     txn1.A(ctx) // Occurs within shared transaction
//     txn2.B(ctx) // Occurs within shared transaction
//     txn1.Done() // closes shared transaction
//
// Note that once a handle is shared between two stores, committing or rolling back
// a transaction will affect the handle of both stores. Most notably, two stores that
// share the same handle are unable to begin independent transactions.
func (s *Store) With(other ShareableStore) *Store {
	return &Store{handle: other.Handle()}
}

// Query performs QueryContext on the underlying connection.
func (s *Store) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	rows, err := s.handle.DBUtilDB().QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return rows, s.wrapError(query, err)
}

// QueryRow performs QueryRowContext on the underlying connection.
func (s *Store) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return s.handle.DBUtilDB().QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// Exec performs a query without returning any rows.
func (s *Store) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := s.ExecResult(ctx, query)
	return err
}

// ExecResult performs a query without returning any rows, but includes the
// result of the execution.
func (s *Store) ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error) {
	res, err := s.handle.DBUtilDB().ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return res, s.wrapError(query, err)
}

// SetLocal performs the `SET LOCAL` query and returns a function to clear (aka to empty string) the setting.
// Calling this method only makes sense within a transaction, as the setting is unset after the transaction
// is either rolled back or committed. This does not perform argument parameterization.
func (s *Store) SetLocal(ctx context.Context, key, value string) (func(context.Context) error, error) {
	if !s.InTransaction() {
		return func(ctx context.Context) error { return nil }, ErrNotInTransaction
	}

	return func(ctx context.Context) error {
		return s.Exec(ctx, sqlf.Sprintf(fmt.Sprintf(`SET LOCAL "%s" TO ''`, key)))
	}, s.Exec(ctx, sqlf.Sprintf(fmt.Sprintf(`SET LOCAL "%s" TO "%s"`, key, value)))
}

// InTransaction returns true if the underlying database handle is in a transaction.
func (s *Store) InTransaction() bool {
	return s.handle.InTransaction()
}

// Transact returns a new store whose methods operate within the context of a new transaction
// or a new savepoint. This method will return an error if the underlying connection cannot be
// interface upgraded to a TxBeginner.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	handle, err := s.handle.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{handle: handle}, nil
}

// Done performs a commit or rollback of the underlying transaction/savepoint depending
// on the value of the error parameter. The resulting error value is a multierror containing
// the error parameter along with any error that occurs during commit or rollback of the
// transaction/savepoint. If the store does not wrap a transaction the original error value
// is returned unchanged.
func (s *Store) Done(err error) error {
	return s.handle.Done(err)
}

// if the code is run from within a test, wrapError wraps the given error
// with query information such as the SQL query and its arguments.
// If not, it returns the error as is.
func (s *Store) wrapError(query *sqlf.Query, err error) error {
	if err == nil {
		return nil
	}

	// if we are not in tests, return the error as is
	if flag.Lookup("test.v") == nil {
		return err
	}

	// in tests, return a wrapped error that includes the query information
	var b strings.Builder

	args := query.Args()
	if len(args) > 50 {
		args = args[:50]
	}

	for i, arg := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", arg)
	}

	if len(args) < len(query.Args()) {
		fmt.Fprintf(&b, ", ... (%d other arguments)", len(query.Args())-len(args))
	}

	return errors.Wrap(err, fmt.Sprintf("SQL Error\n----- Args: %#v\n----- SQL Query:\n%s\n-----\n", b.String(), query.Query(sqlf.PostgresBindVar)))
}
