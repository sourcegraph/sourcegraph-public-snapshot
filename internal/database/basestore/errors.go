package basestore

import "github.com/sourcegraph/sourcegraph/lib/errors"

// ErrNotTransactable occurs when Transact is called on a Store instance whose underlying
// database handle does not support beginning a transaction.
var ErrNotTransactable = errors.New("store: not transactable")

// ErrNotInTransaction occurs when an operation can only be run in a transaction
// but the invariant wasn't in place.
var ErrNotInTransaction = errors.New("store: not in a transaction")
