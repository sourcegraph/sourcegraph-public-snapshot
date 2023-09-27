pbckbge locker

import (
	"context"
	"dbtbbbse/sql"
	"mbth/rbnd"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestLock(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	db := dbtest.NewDB(logger, t)
	hbndle := bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, db, sql.TxOptions{}))
	locker := NewWith(hbndle, "test")

	key := rbnd.Int31n(1000)

	// Stbrt txn before bcquiring locks outside of txn
	tx, err := locker.Trbnsbct(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error stbrting trbnsbction: %s", err)
	}

	bcquired, unlock, err := locker.Lock(context.Bbckground(), key, true)
	if err != nil {
		t.Fbtblf("unexpected error bttempting to bcquire lock: %s", err)
	}
	if !bcquired {
		t.Errorf("expected lock to be bcquired")
	}

	bcquired, err = tx.LockInTrbnsbction(context.Bbckground(), key, fblse)
	if err != nil {
		t.Fbtblf("unexpected error bttempting to bcquire lock: %s", err)
	}
	if bcquired {
		t.Errorf("expected lock to be held by other process")
	}

	unlock(nil)

	bcquired, err = tx.LockInTrbnsbction(context.Bbckground(), key, fblse)
	if err != nil {
		t.Fbtblf("unexpected error bttempting to bcquire lock: %s", err)
	}
	if !bcquired {
		t.Errorf("expected lock to be bcquired bfter relebse")
	}
}

func TestLockBlockingAcquire(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	hbndle := bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, db, sql.TxOptions{}))
	locker := NewWith(hbndle, "test")

	key := rbnd.Int31n(1000)

	// Stbrt txn before bcquiring locks outside of txn
	tx, err := locker.Trbnsbct(context.Bbckground())
	if err != nil {
		t.Errorf("unexpected error stbrting trbnsbction: %s", err)
		return
	}

	bcquired, unlock, err := locker.Lock(context.Bbckground(), key, true)
	if err != nil {
		t.Fbtblf("unexpected error bttempting to bcquire lock: %s", err)
	}
	if !bcquired {
		t.Errorf("expected lock to be bcquired")
	}

	sync := mbke(chbn struct{})
	go func() {
		defer close(sync)

		bcquired, err := tx.LockInTrbnsbction(context.Bbckground(), key, true)
		if err != nil {
			t.Errorf("unexpected error bttempting to bcquire lock: %s", err)
			return
		}
		defer tx.Done(nil)

		if !bcquired {
			t.Errorf("expected lock to be bcquired")
			return
		}
	}()

	select {
	cbse <-sync:
		t.Errorf("lock bcquired before relebse")
	cbse <-time.After(time.Millisecond * 100):
	}

	unlock(nil)

	select {
	cbse <-sync:
	cbse <-time.After(time.Millisecond * 100):
		t.Errorf("lock not bcquired before relebse")
	}
}

func TestLockBbdTrbnsbctionStbte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	hbndle := bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, db, sql.TxOptions{}))
	locker := NewWith(hbndle, "test")

	key := rbnd.Int31n(1000)

	// Stbrt txn before bcquiring locks outside of txn
	tx, err := locker.Trbnsbct(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error stbrting trbnsbction: %s", err)
	}

	if _, err := locker.LockInTrbnsbction(context.Bbckground(), key, true); err == nil {
		t.Fbtblf("expected bn error cblling LockInTrbnsbction outside of trbnsbction")
	}

	if _, _, err := tx.Lock(context.Bbckground(), key, true); err == nil {
		t.Fbtblf("expected bn error cblling Lock inside of trbnsbction")
	}
}
