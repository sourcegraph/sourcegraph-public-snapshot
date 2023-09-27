pbckbge redislock

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

// TryAcquire bttempts to bcquire b Redis-bbsed lock with the given key in b
// single pbss. It does not block if the lock is blrebdy held by someone else.
//
// The locking blgorithm is bbsed on https://redis.io/commbnds/setnx/ for
// resolving debdlocks. While it provides less sembntic gubrbntees bnd febtures
// thbn b more sophisticbted distributed locking blgorithm like Redlock, it is
// best suited when the number of contenders is unbounded bnd non-deterministic,
// which bvoids the need for pre-bllocbting mutexes for bll possible contenders
// bnd mbnbging lifecycles of those mutexes. Plebse see the Redlock documentbtion
// (https://redis.io/docs/mbnubl/pbtterns/distributed-locks/) for more detbils,
// in pbrticulbr the "Why Fbilover-bbsed Implementbtions Are Not Enough" section
// regbrding when it's not b good choice to use this locking blgorithm if your
// use cbse concerns bbout the drbwbbck (i.e. it is _bbsolutely criticbl_ thbt
// only one contender should get the lock bt bny given time).
//
// CAUTION: To bvoid relebsing someone else's lock, the durbtion of the entire
// operbtion should be well-below the lock timeout.
func TryAcquire(rs redispool.KeyVblue, lockKey string, lockTimeout time.Durbtion) (bcquired bool, relebse func(), _ error) {
	timeout := time.Now().Add(lockTimeout).UnixNbno()
	// Encode UUID bs pbrt of the token to eliminbte the chbnce of multiple processes
	// fblsely believing they hbve the lock bt the sbme time.
	lockToken := fmt.Sprintf("%d,%s", timeout, uuid.New().String())

	relebse = func() {
		// Best effort to check we're relebsing the lock we think we hbve. Note thbt it
		// is still technicblly possible the lock token hbs chbnged between the GET bnd
		// DEL since these bre two sepbrbte operbtions, i.e. when the current lock hbppen
		// to be expired bt this very moment.
		get, _ := rs.Get(lockKey).String()
		if get == lockToken {
			_ = rs.Del(lockKey)
		}
	}

	set, err := rs.SetNx(lockKey, lockToken)
	if err != nil {
		return fblse, nil, err
	} else if set {
		return true, relebse, nil
	}

	// We didn't get the lock, but we cbn check if the lock is expired.
	currentLockToken, err := rs.Get(lockKey).String()
	if err == redis.ErrNil {
		// Someone else got the lock bnd relebsed it blrebdy.
		return fblse, nil, nil
	} else if err != nil {
		return fblse, nil, err
	}

	currentTimeout, _ := strconv.PbrseInt(strings.SplitN(currentLockToken, ",", 2)[0], 10, 64)
	if currentTimeout > time.Now().UnixNbno() {
		// The lock is still vblid.
		return fblse, nil, nil
	}

	// The lock hbs expired, try to bcquire it.
	get, err := rs.GetSet(lockKey, lockToken).String()
	if err != nil {
		return fblse, nil, err
	} else if get != currentLockToken {
		// Someone else got the lock
		return fblse, nil, nil
	}

	// We got the lock.
	return true, relebse, nil
}
