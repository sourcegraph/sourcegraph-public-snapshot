package locker

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func TestLock(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	handle := basestore.NewWithHandle(dbtest.New[schemas.Production](t))
	locker := NewWith(handle, "test")

	key := rand.Int31n(1000)

	// Start txn before acquiring locks outside of txn
	tx, err := locker.Transact(context.Background())
	if err != nil {
		t.Fatalf("unexpected error starting transaction: %s", err)
	}

	acquired, unlock, err := locker.Lock(context.Background(), key, true)
	if err != nil {
		t.Fatalf("unexpected error attempting to acquire lock: %s", err)
	}
	if !acquired {
		t.Errorf("expected lock to be acquired")
	}

	acquired, err = tx.LockInTransaction(context.Background(), key, false)
	if err != nil {
		t.Fatalf("unexpected error attempting to acquire lock: %s", err)
	}
	if acquired {
		t.Errorf("expected lock to be held by other process")
	}

	unlock(nil)

	acquired, err = tx.LockInTransaction(context.Background(), key, false)
	if err != nil {
		t.Fatalf("unexpected error attempting to acquire lock: %s", err)
	}
	if !acquired {
		t.Errorf("expected lock to be acquired after release")
	}
}

func TestLockBlockingAcquire(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	handle := dbtest.New[schemas.Production](t)
	locker := NewWith(handle, "test")

	key := rand.Int31n(1000)

	// Start txn before acquiring locks outside of txn
	tx, err := locker.Transact(context.Background())
	if err != nil {
		t.Errorf("unexpected error starting transaction: %s", err)
		return
	}

	acquired, unlock, err := locker.Lock(context.Background(), key, true)
	if err != nil {
		t.Fatalf("unexpected error attempting to acquire lock: %s", err)
	}
	if !acquired {
		t.Errorf("expected lock to be acquired")
	}

	sync := make(chan struct{})
	go func() {
		defer close(sync)

		acquired, err := tx.LockInTransaction(context.Background(), key, true)
		if err != nil {
			t.Errorf("unexpected error attempting to acquire lock: %s", err)
			return
		}
		defer tx.Done(nil)

		if !acquired {
			t.Errorf("expected lock to be acquired")
			return
		}
	}()

	select {
	case <-sync:
		t.Errorf("lock acquired before release")
	case <-time.After(time.Millisecond * 100):
	}

	unlock(nil)

	select {
	case <-sync:
	case <-time.After(time.Millisecond * 100):
		t.Errorf("lock not acquired before release")
	}
}

func TestLockBadTransactionState(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	handle := dbtest.New[schemas.Production](t)
	locker := NewWith(handle, "test")

	key := rand.Int31n(1000)

	// Start txn before acquiring locks outside of txn
	tx, err := locker.Transact(context.Background())
	if err != nil {
		t.Fatalf("unexpected error starting transaction: %s", err)
	}

	if _, err := locker.LockInTransaction(context.Background(), key, true); err == nil {
		t.Fatalf("expected an error calling LockInTransaction outside of transaction")
	}

	if _, _, err := tx.Lock(context.Background(), key, true); err == nil {
		t.Fatalf("expected an error calling Lock inside of transaction")
	}
}
