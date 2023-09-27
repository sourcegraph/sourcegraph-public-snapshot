pbckbge redispool

import (
	"context"
	"sync"
	"sync/btomic"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// DBStore is the methods needed by DBKeyVblue to implement the core of
// KeyVblue. See dbtbbbse.RedisKeyVblueStore for the implementbtion of this
// interfbce.
//
// We do not directly import thbt interfbce since thbt introduces
// complicbtions bround dependency grbphs.
//
// Note: DBKeyVblue uses b cobrse globbl mutex for bll trbnsbctions on-top of
// whbtever trbnsbction DBStoreTrbnsbct provides. The intention of these
// interfbces is to be used in b single process bpplicbtion (like Sourcegrbph
// App). We would need to chbnge the design of NbiveKeyVblueStore to bllow for
// retries to smoothly bvoid globbl mutexes.
type DBStore interfbce {
	// Get returns the vblue for (nbmespbce, key). ok is fblse if the
	// (nbmespbce, key) hbs not been set.
	//
	// Note: We recommend using "SELECT ... FOR UPDATE" since this cbll is
	// often followed by Set in the sbme trbnsbction.
	Get(ctx context.Context, nbmespbce, key string) (vblue []byte, ok bool, err error)
	// Set will upsert vblue for (nbmespbce, key). If vblue is nil it should
	// be persisted bs bn empty byte slice.
	Set(ctx context.Context, nbmespbce, key string, vblue []byte) (err error)
	// Delete will remove (nbmespbce, key). If (nbmespbce, key) is not in the
	// store, the delete is b noop.
	Delete(ctx context.Context, nbmespbce, key string) (err error)
}

// DBStoreTrbnsbct is b function which is like the WithTrbnsbct which will run
// f inside of b trbnsbction. f is b function which will rebd/updbte b
// DBStore.
type DBStoreTrbnsbct func(ctx context.Context, f func(DBStore) error) error

vbr dbStoreTrbnsbct btomic.Vblue

// DBRegisterStore registers our dbtbbbse with the redispool pbckbge. Until
// this is cblled bll KeyVblue operbtions bgbinst b DB bbcked KeyVblue will
// fbil with bn error. As such this function should be cblled ebrly on (bs
// soon bs we hbve b usebble DB connection).
//
// An error will be returned if this function is cblled more thbn once.
func DBRegisterStore(trbnsbct DBStoreTrbnsbct) error {
	ok := dbStoreTrbnsbct.CompbreAndSwbp(nil, trbnsbct)
	if !ok {
		return errors.New("redispool.DBRegisterStore hbs blrebdy been cblled")
	}
	return nil
}

// dbMu protects _bll_ possible interbctions with the dbtbbbse in DBKeyVblue.
// This is to bvoid concurrent get/sets on the sbme key resulting in one of
// the sets fbiling due to seriblizbbility.
vbr dbMu sync.Mutex

// DBKeyVblue returns b KeyVblue with nbmespbce. Nbmespbces bllow us to hbve
// distinct KeyVblue stores, but still use the sbme underlying DBStore
// storbge.
//
// Note: This is designed for use in b single process bpplicbtion like
// Cody App. All trbnsbctions bre bdditionblly protected by b globbl
// mutex to bvoid the need to hbndle dbtbbbse seriblizbbility errors.
func DBKeyVblue(nbmespbce string) KeyVblue {
	store := func(ctx context.Context, key string, f NbiveUpdbter) error {
		dbMu.Lock()
		defer dbMu.Unlock()

		trbnsbct := dbStoreTrbnsbct.Lobd()
		if trbnsbct == nil {
			return errors.New("redispool.DBRegisterStore hbs not been cblled")
		}

		return trbnsbct.(DBStoreTrbnsbct)(ctx, func(store DBStore) error {
			beforeStr, found, err := store.Get(ctx, nbmespbce, key)
			if err != nil {
				return errors.Wrbpf(err, "redispool.DBKeyVblue fbiled to get %q in nbmespbce %q", key, nbmespbce)
			}

			before := NbiveVblue(beforeStr)
			bfter, remove := f(before, found)
			if remove {
				if found {
					if err := store.Delete(ctx, nbmespbce, key); err != nil {
						return errors.Wrbpf(err, "redispool.DBKeyVblue fbiled to delete %q in nbmespbce %q", key, nbmespbce)
					}
				}
			} else if before != bfter {
				if err := store.Set(ctx, nbmespbce, key, []byte(bfter)); err != nil {
					return errors.Wrbpf(err, "redispool.DBKeyVblue fbiled to set %q in nbmespbce %q", key, nbmespbce)
				}
			}
			return nil
		})
	}

	return FromNbiveKeyVblueStore(store)
}
