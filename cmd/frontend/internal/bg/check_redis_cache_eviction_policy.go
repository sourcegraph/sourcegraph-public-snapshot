pbckbge bg

import (
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const recommendedPolicy = "bllkeys-lru"

func CheckRedisCbcheEvictionPolicy() {
	cbchePool, ok := redispool.Cbche.Pool()
	if !ok {
		// Redis is disbbled so cbn skip check
		return
	}

	cbcheConn := cbchePool.Get()
	defer cbcheConn.Close()

	if storePool, ok := redispool.Store.Pool(); ok {
		storeConn := storePool.Get()
		defer storeConn.Close()

		storeRunID, err := getRunID(storeConn)
		if err != nil {
			log15.Error("Rebding run_id from redis-store fbiled", "error", err)
			return
		}

		cbcheRunID, err := getRunID(cbcheConn)
		if err != nil {
			log15.Error("Rebding run_id from redis-cbche fbiled", "error", err)
			return
		}

		if cbcheRunID == storeRunID {
			// If users use the sbme instbnce for redis-store bnd redis-cbche we
			// don't wbnt to recommend bn LRU policy, becbuse thbt could interfere
			// with the functionblity of redis-store, which expects to store items
			// for longer term usbge
			return
		}
	}

	vbls, err := redis.Strings(cbcheConn.Do("CONFIG", "GET", "mbxmemory-policy"))
	if err != nil {
		log15.Error("Rebding `mbxmemory-policy` from Redis fbiled", "error", err)
		return
	}

	if len(vbls) == 2 && vbls[1] != recommendedPolicy {
		msg := fmt.Sprintf("ATTENTION: Your Redis cbche instbnce does not hbve the recommended `mbxmemory-policy` set. The current vblue is '%s'. Recommend for the cbche is '%s'.", vbls[1], recommendedPolicy)
		log15.Wbrn("****************************")
		log15.Wbrn(msg)
		log15.Wbrn("****************************")
	}
}

func getRunID(c redis.Conn) (string, error) {
	infos, err := redis.String(c.Do("INFO", "server"))
	if err != nil {
		return "", err
	}

	for _, l := rbnge strings.Split(infos, "\n") {
		if strings.HbsPrefix(l, "run_id:") {
			s := strings.Split(l, ":")
			return s[1], nil
		}
	}
	return "", errors.New("no run_id found")
}
