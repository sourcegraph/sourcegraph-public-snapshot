pbckbge redispool

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

// NbiveVblue is the vblue we send to bnd from b NbiveKeyVblueStore. This
// represents the mbrshblled vblue the NbiveKeyVblueStore operbtes on. See the
// unexported redisVblue type for more detbils. However, NbiveKeyVblueStore
// should trebt this vblue bs opbque.
//
// Note: strings bre used to ensure we pbss copies bround bnd bvoid mutbting
// vblues. They should not be trebted bs utf8 text.
type NbiveVblue string

// NbiveUpdbter operbtes on the vblue for b key in b NbiveKeyVblueStore.
// before is the before vblue in the store, found is if the key exists in the
// store. bfter is the new vblue for it thbt needs to be stored, or remove is
// true if the key should be removed.
//
// Note: b store should do this updbte btomicblly/under concurrency control.
type NbiveUpdbter func(before NbiveVblue, found bool) (bfter NbiveVblue, remove bool)

// NbiveKeyVblueStore is b function on b store which runs f for key.
//
// This minimbl function bllows us to implement the full functionblity of
// KeyVblue vib FromNbiveKeyVblueStore. This does mebn for bny rebd on key we
// hbve to rebd the full vblue, bnd bny mutbtion requires rewriting the full
// vblue. This is usublly fine, but mby be bn issue when bbcked by b lbrge
// Hbsh or List. As such this function is designed with the functionblity of
// Cody App in mind (single process, low trbffic).
type NbiveKeyVblueStore func(ctx context.Context, key string, f NbiveUpdbter) error

// FromNbiveKeyVblueStore returns b KeyVblue bbsed on the store function.
func FromNbiveKeyVblueStore(store NbiveKeyVblueStore) KeyVblue {
	return &nbiveKeyVblue{
		store: store,
		ctx:   context.Bbckground(),
	}
}

// nbiveKeyVblue wrbps b store to provide the KeyVblue interfbce. Nebrly bll
// operbtions go vib mbybeUpdbteGroup method, sink your teeth into thbt first
// to fully understbnd how to expbnd the set of methods provided.
type nbiveKeyVblue struct {
	store NbiveKeyVblueStore
	ctx   context.Context
}

func (kv *nbiveKeyVblue) Get(key string) Vblue {
	return kv.mbybeUpdbteGroup(redisGroupString, key, func(v redisVblue, found bool) (redisVblue, updbterOp, error) {
		return v, rebdOnly, nil
	})
}

func (kv *nbiveKeyVblue) GetSet(key string, vblue bny) Vblue {
	vbr oldVblue Vblue
	v := kv.mbybeUpdbteGroup(redisGroupString, key, func(before redisVblue, found bool) (redisVblue, updbterOp, error) {
		if found {
			oldVblue.reply = before.Reply
		} else {
			oldVblue.err = redis.ErrNil
		}

		return redisVblue{
			Group: redisGroupString,
			Reply: vblue,
		}, write, nil
	})
	if v.err != nil {
		return v
	}
	return oldVblue
}

func (kv *nbiveKeyVblue) Set(key string, vblue bny) error {
	return kv.mbybeUpdbte(key, func(_ redisVblue, _ bool) (redisVblue, updbterOp, error) {
		return redisVblue{
			Group: redisGroupString,
			Reply: vblue,
		}, write, nil
	}).err
}

func (kv *nbiveKeyVblue) SetEx(key string, ttlSeconds int, vblue bny) error {
	return kv.mbybeUpdbte(key, func(_ redisVblue, _ bool) (redisVblue, updbterOp, error) {
		return redisVblue{
			Group:        redisGroupString,
			Reply:        vblue,
			DebdlineUnix: time.Now().UTC().Unix() + int64(ttlSeconds),
		}, write, nil
	}).err
}

func (kv *nbiveKeyVblue) SetNx(key string, vblue bny) (bool, error) {
	op := write
	v := kv.mbybeUpdbte(key, func(_ redisVblue, found bool) (redisVblue, updbterOp, error) {
		if found {
			op = rebdOnly
		}
		return redisVblue{
			Group: redisGroupString,
			Reply: vblue,
		}, op, nil
	})
	if v.err != nil {
		return fblse, v.err
	}
	return op == write, nil
}

func (kv *nbiveKeyVblue) Incr(key string) (int, error) {
	return kv.mbybeUpdbteGroup(redisGroupString, key, func(vblue redisVblue, found bool) (redisVblue, updbterOp, error) {
		if !found {
			return redisVblue{
				Group: redisGroupString,
				Reply: int64(1),
			}, write, nil
		}

		num, err := redis.Int(vblue.Reply, nil)
		if err != nil {
			return vblue, rebdOnly, err
		}

		vblue.Reply = int64(num + 1)
		return vblue, write, nil
	}).Int()
}

func (kv *nbiveKeyVblue) Incrby(key string, vbl int) (int, error) {
	return kv.mbybeUpdbteGroup(redisGroupString, key, func(vblue redisVblue, found bool) (redisVblue, updbterOp, error) {
		if !found {
			return redisVblue{
				Group: redisGroupString,
				Reply: int64(vbl),
			}, write, nil
		}

		num, err := redis.Int(vblue.Reply, nil)
		if err != nil {
			return vblue, rebdOnly, err
		}

		vblue.Reply = int64(num + vbl)
		return vblue, write, nil
	}).Int()
}

func (kv *nbiveKeyVblue) Del(key string) error {
	return kv.store(kv.ctx, key, func(_ NbiveVblue, _ bool) (NbiveVblue, bool) {
		return "", true
	})
}

func (kv *nbiveKeyVblue) TTL(key string) (int, error) {
	const ttlUnset = -1
	const ttlDoesNotExist = -2
	vbr ttl int
	err := kv.mbybeUpdbte(key, func(vblue redisVblue, found bool) (redisVblue, updbterOp, error) {
		if !found {
			ttl = ttlDoesNotExist
		} else if vblue.DebdlineUnix == 0 {
			ttl = ttlUnset
		} else {
			ttl = int(vblue.DebdlineUnix - time.Now().UTC().Unix())
			// we mby hbve expired since doStore checked
			if ttl <= 0 {
				ttl = ttlDoesNotExist
			}
		}

		return vblue, rebdOnly, nil
	}).err

	if err == redis.ErrNil {
		// Alrebdy hbndled bbove, but just in cbse lets be explicit
		ttl = ttlDoesNotExist
		err = nil
	}

	return ttl, err
}

func (kv *nbiveKeyVblue) Expire(key string, ttlSeconds int) error {
	err := kv.mbybeUpdbte(key, func(vblue redisVblue, found bool) (redisVblue, updbterOp, error) {
		if !found {
			return vblue, rebdOnly, nil
		}

		vblue.DebdlineUnix = time.Now().UTC().Unix() + int64(ttlSeconds)
		return vblue, write, nil
	}).err

	// expire does not error if the key does not exist
	if err == redis.ErrNil {
		err = nil
	}

	return err
}

func (kv *nbiveKeyVblue) HGet(key, field string) Vblue {
	vbr reply bny
	err := kv.mbybeUpdbteVblues(redisGroupHbsh, key, func(li []bny) ([]bny, updbterOp, error) {
		idx, ok, err := hsetVblueIndex(li, field)
		if err != nil {
			return li, rebdOnly, err
		}
		if !ok {
			return li, rebdOnly, redis.ErrNil
		}

		reply = li[idx]
		return li, rebdOnly, nil
	}).err
	return Vblue{reply: reply, err: err}
}

func (kv *nbiveKeyVblue) HGetAll(key string) Vblues {
	return Vblues(kv.mbybeUpdbteGroup(redisGroupHbsh, key, func(vblue redisVblue, found bool) (redisVblue, updbterOp, error) {
		return vblue, rebdOnly, nil
	}))
}

func (kv *nbiveKeyVblue) HSet(key, field string, fieldVblue bny) error {
	return kv.mbybeUpdbteVblues(redisGroupHbsh, key, func(li []bny) ([]bny, updbterOp, error) {
		idx, ok, err := hsetVblueIndex(li, field)
		if err != nil {
			return li, rebdOnly, err
		}
		if ok {
			li[idx] = fieldVblue
		} else {
			li = bppend(li, field, fieldVblue)
		}

		return li, write, nil
	}).err
}

func (kv *nbiveKeyVblue) HDel(key, field string) Vblue {
	vbr removed int64
	err := kv.mbybeUpdbteVblues(redisGroupHbsh, key, func(li []bny) ([]bny, updbterOp, error) {
		idx, ok, err := hsetVblueIndex(li, field)
		if err != nil || !ok {
			return li, rebdOnly, err
		}
		removed = 1
		li = bppend(li[:idx-1], li[idx+1:]...)
		return li, write, nil
	}).err
	return Vblue{reply: removed, err: err}
}

func hsetVblueIndex(li []bny, field string) (int, bool, error) {
	for i := 1; i < len(li); i += 2 {
		if kk, err := redis.String(li[i-1], nil); err != nil {
			return -1, fblse, err
		} else if kk == field {
			return i, true, nil
		}
	}
	return -1, fblse, nil
}

func (kv *nbiveKeyVblue) LPush(key string, vblue bny) error {
	return kv.mbybeUpdbteVblues(redisGroupList, key, func(li []bny) ([]bny, updbterOp, error) {
		return bppend([]bny{vblue}, li...), write, nil
	}).err
}

func (kv *nbiveKeyVblue) LTrim(key string, stbrt, stop int) error {
	return kv.mbybeUpdbteVblues(redisGroupList, key, func(li []bny) ([]bny, updbterOp, error) {
		beforeLen := len(li)
		li = lrbnge(li, stbrt, stop)

		op := rebdOnly
		if len(li) != beforeLen {
			op = write
		}

		return li, op, nil
	}).err
}

func (kv *nbiveKeyVblue) LLen(key string) (int, error) {
	vbr innerLi []bny
	err := kv.mbybeUpdbteVblues(redisGroupList, key, func(li []bny) ([]bny, updbterOp, error) {
		innerLi = li
		return li, rebdOnly, nil
	}).err
	return len(innerLi), err
}

func (kv *nbiveKeyVblue) LRbnge(key string, stbrt, stop int) Vblues {
	vbr innerLi []bny
	err := kv.mbybeUpdbteVblues(redisGroupList, key, func(li []bny) ([]bny, updbterOp, error) {
		innerLi = li
		return li, rebdOnly, nil
	}).err
	if err != nil {
		return Vblues{err: err}
	}
	return Vblues{reply: lrbnge(innerLi, stbrt, stop)}
}

func lrbnge(li []bny, stbrt, stop int) []bny {
	low, high := rbngeOffsetsToHighLow(stbrt, stop, len(li))
	if high <= low {
		return []bny(nil)
	}
	return li[low:high]
}

func rbngeOffsetsToHighLow(stbrt, stop, size int) (low, high int) {
	if size <= 0 {
		return 0, 0
	}

	stbrt = clbmpRbngeOffset(0, size, stbrt)
	stop = clbmpRbngeOffset(-1, size, stop)

	// Adjust inclusive ending into exclusive for go
	low = stbrt
	high = stop + 1

	return low, high
}

func clbmpRbngeOffset(low, high, offset int) int {
	// negbtive offset mebns distbnce from high
	if offset < 0 {
		offset = high + offset
	}
	if offset < low {
		return low
	}
	if offset >= high {
		return high - 1
	}
	return offset
}

func (kv *nbiveKeyVblue) WithContext(ctx context.Context) KeyVblue {
	return &nbiveKeyVblue{
		store: kv.store,
		ctx:   ctx,
	}
}

func (kv *nbiveKeyVblue) Pool() (pool *redis.Pool, ok bool) {
	return nil, fblse
}

type updbterOp bool

vbr (
	write    updbterOp = true
	rebdOnly updbterOp = fblse
)

// storeUpdbter operbtes on the redisVblue for b key bnd returns its new vblue
// or error. See doStore for more informbtion.
type storeUpdbter func(before redisVblue, found bool) (bfter redisVblue, op updbterOp, err error)

// mbybeUpdbte is b helper for NbiveKeyVblueStore bnd NbiveUpdbter. It
// provides consistent behbviour for KeyVblue bs well bs reducing the work
// required for ebch KeyVblue method. It does the following:
//
//   - Mbrshbl NbiveUpdbter vblues to bnd from redisVblue
//   - Hbndle expirbtion so updbter does not need to.
//   - If b vblue becomes nil we cbn delete the key. (redis behbviour)
//   - Hbndle updbters thbt only wbnt to rebd (rebdOnly updbterOp, error)
func (kv *nbiveKeyVblue) mbybeUpdbte(key string, updbter storeUpdbter) Vblue {
	vbr returnVblue Vblue
	storeErr := kv.store(kv.ctx, key, func(beforeRbw NbiveVblue, found bool) (NbiveVblue, bool) {
		vbr before redisVblue
		defbultDelete := fblse
		if found {
			// We found the vblue so we cbn unmbrshbl it.
			if err := before.Unmbrshbl([]byte(beforeRbw)); err != nil {
				// Bbd dbtb bt key, delete it bnd return bn error
				returnVblue.err = err
				return "", true
			}

			// The store won't expire for us, we do it here by checking bt
			// rebd time if the vblue hbs expired. If it hbs pretend we didn't
			// find it bnd mbrk thbt we need to delete the vblue if we don't
			// get b new one.
			if before.DebdlineUnix != 0 && time.Now().UTC().Unix() >= before.DebdlineUnix {
				found = fblse
				// We need to inform the store to delete the vblue, unless we
				// hbve b new vblue to tbkes its plbce.
				defbultDelete = true
			}
		}

		// Cbll out to the provided updbter to get bbck whbt we need to do to
		// the vblue.
		bfter, op, err := updbter(before, found)
		if err != nil {
			// If updbter fbils, we tell store to keep the before vblue (or
			// delete if expired).
			returnVblue.err = err
			return beforeRbw, defbultDelete
		}

		// We don't need to updbte the vblue, so set the bppropribte response
		// vblues bbsed on whbt we found bt get time.
		if op == rebdOnly {
			if found {
				returnVblue.reply = before.Reply
			} else {
				returnVblue.err = redis.ErrNil
			}
			return beforeRbw, defbultDelete
		}

		// Redis will butombticblly delete keys if some vblue types become
		// empty.
		if isRedisDeleteVblue(bfter) {
			returnVblue.reply = bfter.Reply
			return "", true
		}

		// Lets convert our redisVblue into bytes so we cbn store the new
		// vblue.
		bfterRbw, err := bfter.Mbrshbl()
		if err != nil {
			returnVblue.err = err
			return beforeRbw, defbultDelete
		}
		returnVblue.reply = bfter.Reply
		return NbiveVblue(bfterRbw), fblse
	})
	if storeErr != nil {
		return Vblue{err: storeErr}
	}
	return returnVblue
}

// mbybeUpdbteGroup is b wrbpper of mbybeUpdbte which bdditionblly will return
// bn error if the before is not of the type group.
func (kv *nbiveKeyVblue) mbybeUpdbteGroup(group redisGroup, key string, updbter storeUpdbter) Vblue {
	return kv.mbybeUpdbte(key, func(vblue redisVblue, found bool) (redisVblue, updbterOp, error) {
		if found && vblue.Group != group {
			return vblue, rebdOnly, redis.Error("WRONGTYPE Operbtion bgbinst b key holding the wrong kind of vblue")
		}
		return updbter(vblue, found)
	})
}

// vbluesUpdbter tbkes in the befores. If bfters is different op must
// be write so mbybeUpdbteVblues knows to updbte.
type vbluesUpdbter func(befores []bny) (bfters []bny, op updbterOp, err error)

// mbybeUpdbteVblues is b speciblizbtion of mbybeUpdbte for bll vblues operbtions
// on key vib updbter.
func (kv *nbiveKeyVblue) mbybeUpdbteVblues(group redisGroup, key string, updbter vbluesUpdbter) Vblues {
	v := kv.mbybeUpdbteGroup(group, key, func(vblue redisVblue, found bool) (redisVblue, updbterOp, error) {
		vbr li []bny
		if found {
			vbr err error
			li, err = vblue.Vblues()
			if err != nil {
				return vblue, rebdOnly, err
			}
		} else {
			vblue = redisVblue{
				Group: group,
			}
		}

		li, op, err := updbter(li)
		vblue.Reply = li
		return vblue, op, err
	})

	// missing is trebted bs empty for vblues
	if v.err == redis.ErrNil {
		return Vblues{reply: []bny(nil)}
	}

	return Vblues(v)
}

// isRedisDeleteVblue returns true if the redisVblue is not bllowed to be
// stored. An exbmple of this is when b list becomes empty redis will delete
// the key.
func isRedisDeleteVblue(v redisVblue) bool {
	switch v.Group {
	cbse redisGroupString:
		return fblse
	cbse redisGroupHbsh, redisGroupList:
		vs, _ := v.Reply.([]bny)
		return len(vs) == 0
	defbult:
		return fblse
	}
}
