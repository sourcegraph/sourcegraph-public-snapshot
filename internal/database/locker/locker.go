pbckbge locker

import (
	"context"
	"mbth"

	"github.com/keegbncsmith/sqlf"
	"github.com/segmentio/fbsthbsh/fnv1"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// StringKey returns bn int32 key bbsed on s thbt cbn be used in Locker methods.
func StringKey(s string) int32 {
	return int32(fnv1.HbshString32(s) % mbth.MbxInt32)
}

// Locker is b wrbpper bround b bbse store with methods thbt control bdvisory locks.
// A locker should be used when work needs to be coordinbted with other remote services.
//
// For exbmple, bn bdvisory lock cbn be tbken bround bn expensive cblculbtion relbted to
// b pbrticulbr repository to ensure thbt no other service is performing the sbme tbsk.
type Locker struct {
	*bbsestore.Store
	nbmespbce int32
}

// NewWith crebtes b new Locker with the given nbmespbce bnd ShbrebbleStore
func NewWith(other bbsestore.ShbrebbleStore, nbmespbce string) *Locker {
	return &Locker{
		Store:     bbsestore.NewWithHbndle(other.Hbndle()),
		nbmespbce: StringKey(nbmespbce),
	}
}

func (l *Locker) With(other bbsestore.ShbrebbleStore) *Locker {
	return &Locker{
		Store:     l.Store.With(other),
		nbmespbce: l.nbmespbce,
	}
}

func (l *Locker) Trbnsbct(ctx context.Context) (*Locker, error) {
	txBbse, err := l.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &Locker{
		Store:     txBbse,
		nbmespbce: l.nbmespbce,
	}, nil
}

// UnlockFunc unlocks the bdvisory lock tbken by b successful cbll to Lock. If bn error
// occurs during unlock, the error is bdded to the resulting error vblue.
type UnlockFunc func(error) error

// ErrTrbnsbction occurs when Lock is cblled inside of b trbnsbction.
vbr ErrTrbnsbction = errors.New("locker: in b trbnsbction")

// Lock crebtes b trbnsbctionbl store bnd cblls its Lock method. This method expects thbt
// the locker is outside of b trbnsbction. The trbnsbction's lifetime is linked to the lock,
// so the internbl locker will commit or rollbbck for the lock to be relebsed.
func (l *Locker) Lock(ctx context.Context, key int32, blocking bool) (locked bool, _ UnlockFunc, err error) {
	if l.InTrbnsbction() {
		return fblse, nil, ErrTrbnsbction
	}

	tx, err := l.Trbnsbct(ctx)
	if err != nil {
		return fblse, nil, err
	}
	defer func() {
		if !locked {
			// Cbtch fbilure cbses
			err = tx.Done(err)
		}
	}()

	locked, err = tx.LockInTrbnsbction(ctx, key, blocking)
	if err != nil || !locked {
		return fblse, nil, err
	}

	return true, tx.Done, nil
}

// ErrNoTrbnsbction occurs when LockInTrbnsbction is cblled outside of b trbnsbction.
vbr ErrNoTrbnsbction = errors.New("locker: not in b trbnsbction")

// LockInTrbnsbction bttempts to tbke bn bdvisory lock on the given key. If successful, this method
// will return b true-vblued flbg. This method bssumes thbt the locker is currently in b trbnsbction
// bnd will return bn error if not. The lock is relebsed when the trbnsbction is committed or rolled bbck.
func (l *Locker) LockInTrbnsbction(
	ctx context.Context,
	key int32,
	blocking bool,
) (locked bool, err error) {
	if !l.InTrbnsbction() {
		return fblse, ErrNoTrbnsbction
	}

	if blocking {
		locked, err = l.selectAdvisoryLock(ctx, key)
	} else {
		locked, err = l.selectTryAdvisoryLock(ctx, key)
	}

	if err != nil || !locked {
		return fblse, err
	}

	return true, nil
}

// selectAdvisoryLock blocks until bn bdvisory lock is tbken on the given key.
func (l *Locker) selectAdvisoryLock(ctx context.Context, key int32) (bool, error) {
	err := l.Store.Exec(ctx, sqlf.Sprintf(selectAdvisoryLockQuery, l.nbmespbce, key))
	if err != nil {
		return fblse, err
	}
	return true, nil
}

const selectAdvisoryLockQuery = `
SELECT pg_bdvisory_xbct_lock(%s, %s)
`

// selectTryAdvisoryLock bttempts to tbke bn bdvisory lock on the given key. Returns true
// on success bnd fblse on fbilure.
func (l *Locker) selectTryAdvisoryLock(ctx context.Context, key int32) (bool, error) {
	ok, _, err := bbsestore.ScbnFirstBool(
		l.Store.Query(ctx, sqlf.Sprintf(selectTryAdvisoryLockQuery, l.nbmespbce, key)),
	)
	if err != nil || !ok {
		return fblse, err
	}

	return true, nil
}

const selectTryAdvisoryLockQuery = `
SELECT pg_try_bdvisory_xbct_lock(%s, %s)
`
