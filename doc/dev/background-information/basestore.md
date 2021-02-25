# Developing a store

Code that interacts directly with a Postgres instance should be grouped and hidden behind a `Store` abstraction. This enables a separation of concerns between our data access layer and the rest of the application logic. Consumers of such a store should rely on an _interface_ with the methods from the target store that are used, allowing a mock to be substituted into the concrete store's place in unit tests.

The following behaviors should be implemented over all stores:

#### Transactions

Stores should enable _transactional execution_ over a set of queries to ensure that if one query in a set fails that any observable effects are rolled back.

```go
func DoSomethingAtomic(ctx context.Context, store *MyStore) (err error) {
	// Create a new transaction context. tx should be the same type as myStore,
	// but with a new underlying transaction handle instead of a bare connection.
	// Nested transactions are implemented via savepoints.

	tx, err := myStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// On exit of the function, rollback the transaction if err != nil and commit
		// it otherwise. If an error occurs during transaction finalization, the given
		// error will be modified to reflect the additional error.
		err = tx.Done(err)
	}

	if err := tx.FirstOperation(ctx); err != nil {
		return err
	}
	if err := tx.SecondOperation(ctx); err != nil {
		return err
	}

	return nil
}
```

#### Sharing underlying handles

We implement several store implementations rather than combining all Postgres-related behavior into a single struct to reap several benefits:

1. Higher cohesion and better understandability/maintainability of each store
1. Better code isolation between teams/features (repo store vs codeintel store)
1. We target more than one physical database instance (main app / codeintel / codeinsights DBs are separate)

This creates a new problem around two stores _interacting_ when more than one store is required in a single code path. Stores should enable a way to borrow the underlying handle of another store instance so that they can operate within the same transaction context.

```go
func DoSomethingAtomicOverTwoStores(ctx context.Context, store *MyStore, otherStore *MyOtherStore) (err error) {
	tx, err := myStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}

	if err := tx.Operation(ctx); err != nil {
		return err
	}

	// OtherOperation executes with the underlying handle of tx
	if err := otherStore.With(tx).OtherOperation(ctx); err != nil {
		return err
	}

	return nil
}
```

**Note**: This is not well-defined over two stores targeting a different physical database.

## Using *basestore.Store

The [`Store`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v3.25.0/-/blob/internal/database/basestore/store.go#L37:6) struct defined in [github.com/sourcegraph/sourcegraph/internal/database/basestore](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v3.25.0/-/tree/internal/database/basestore) can be used to quickly bootstrap the base functionalities described above.

First, _embed_ a basestore pointer into your own store instance, as follows. Your store may need access to additional data for configuration or state - additional fields can be freely defined on this struct.

```go
import "github.com/sourcegraph/sourcegraph/internal/database/basestore"

type MyStore struct {
	*basestore.Store
	// ...
}

func NewStoreWithDB(db dbutil.DB) *Store {
	return &Store{Store: basestore.NewWithDB(db, sql.TxOptions{}) /*, ... */}
}
```

Next, ensure that your store enables the transaction behaviors as descried above. This functionality is already implemented by the basestore, but needs a bit of tweaking to ensure that the return types are correct.

Both the `With` and `Transact` methods need to be re-defined by your containing struct so that the methods return a `*MyStore` instead of a `*basestore.Store`. If you have any additional fields defined on your store that should exist across transaction boundaries, they must be assigned to the new store instance as well.

```go
// Wraps the basestore.With method to return the correct type.
func (s *MyStore) With(other basestore.ShareableStore) *MyStore {
	return &MyStore{Store: s.Store.With(other) /*, ... */}
}

// Wraps the basestore.Transact method to return the correct type.
func (s *MyStore) Transact(ctx context.Context) (*MyStore, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}

	return &MyStore{Store: txBase, /*, ... */}, nil
}
```

Lastly, implement your store's logic by adding additional methods to your store. By embedding a basestore, you gain access all of the helper methods defined to make common queries easier. The basestore package also provides a number of `Scan*` utility function to conveniently read over `*sql.Rows` result sets.

```go
func (s *MyStore) CountThingsForDomain(ctx context.Context, domain string) (int, error) {
	// Query and consume a single int from first row
	count, _, err := basestore.ScanFirstInt(s.Store.Query(sqlf.Sprintf("SELECT count(*) FROM things WHERE domain = %s", domain)))
	return count, err
}

func (s *MyStore) ThingsForDomain(ctx context.Context, domain string, limit, offset int) (_ []string, _ int, err error) {
	// Start txn so count and page results come from a consistent worldview
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = tx.Done(err) }()
	
	// Call count method defined above within current transaction
	totalCount, err := tx.CountThingsForDomain(ctx, domain)
	if err != nil {
		return nil, 0, err
	}

	// Actual get the currently requested page of values
	values, err := basestore.ScanStrings(tx.Store.Query(sqlf.Sprintf("SELECT value FROM things WHERE domain = %s ORDER BY value LIMIT %d OFFSET %d", domain, limit, offset)))
	if err != nil {
		return nil, 0, err
	}
	
	return values, totalCount, nil
}

func (s *MyStore) InsertThingForDomain(ctx context.Context, domain, value string) error {
	// Exec and throw away result
	return s.Store.Exec(sqlf.Sprintf("INSERT INTO thing (domain, value) VALUES (%s, %s)", domain, value))
}
```
