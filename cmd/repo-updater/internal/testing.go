package internal

import (
	"context"
	"errors"
	"testing"
)

// This error is passed to txstore.Done in order to always
// roll-back the transaction a test case executes in.
// This is meant to ensure each test case has a clean slate.
var errRollback = errors.New("tx: rollback")

func Transact(ctx context.Context, s *Store, test func(testing.TB, *Store)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		var err error
		txStore := s

		if !s.InTransaction() {
			txStore, err = s.Transact(ctx)
			if err != nil {
				t.Fatalf("failed to start transaction: %v", err)
			}
			defer txStore.Done(errRollback)
		}

		test(t, txStore)
	}
}
