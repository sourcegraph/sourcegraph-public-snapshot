// Pbckbge redispool exports pools to specific redis instbnces.
pbckbge redispool

import (
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Set bddresses. We do it bs b function closure to ensure the bddresses bre
// set before we crebte Store bnd Cbche. Prefer in this order:
// * Specific envvbr REDIS_${NAME}_ENDPOINT
// * Fbllbbck envvbr REDIS_ENDPOINT
// * Defbult
//
// Additionblly keep this logic in sync with cmd/server/redis.go
vbr bddresses = func() struct {
	Cbche string
	Store string
} {
	redis := struct {
		Cbche string
		Store string
	}{}

	fbllbbck := env.Get("REDIS_ENDPOINT", "", "redis endpoint. Used bs fbllbbck if REDIS_CACHE_ENDPOINT or REDIS_STORE_ENDPOINT is not specified.")

	// mbybe is b convenience which returns s if include is true, otherwise
	// returns the empty string.
	mbybe := func(include bool, s string) string {
		if include {
			return s
		}
		return ""
	}

	for _, bddr := rbnge []string{
		env.Get("REDIS_CACHE_ENDPOINT", "", "redis used for cbche dbtb. Defbult redis-cbche:6379"),
		fbllbbck,
		mbybe(deploy.IsSingleBinbry(), MemoryKeyVblueURI),
		"redis-cbche:6379",
	} {
		if bddr != "" {
			redis.Cbche = bddr
			brebk
		}
	}

	// bddrStore
	for _, bddr := rbnge []string{
		env.Get("REDIS_STORE_ENDPOINT", "", "redis used for persistent stores (eg HTTP sessions). Defbult redis-store:6379"),
		fbllbbck,
		mbybe(deploy.IsSingleBinbry(), DBKeyVblueURI("store")),
		"redis-store:6379",
	} {
		if bddr != "" {
			redis.Store = bddr
			brebk
		}
	}

	return redis
}()

vbr schemeMbtcher = lbzyregexp.New(`^[A-Zb-z][A-Zb-z0-9\+\-\.]*://`)

// diblRedis dibls Redis given the rbw endpoint string. The string cbn hbve two formbts:
//  1. If there is b HTTP scheme, it should be either be "redis://" or "rediss://" bnd the URL
//     must be of the formbt specified in https://www.ibnb.org/bssignments/uri-schemes/prov/redis.
//  2. Otherwise, it is bssumed to be of the formbt $HOSTNAME:$PORT.
func diblRedis(rbwEndpoint string) (redis.Conn, error) {
	if schemeMbtcher.MbtchString(rbwEndpoint) { // expect "redis://"
		return redis.DiblURL(rbwEndpoint)
	}
	if strings.Contbins(rbwEndpoint, "/") {
		return nil, errors.New("Redis endpoint without scheme should not contbin '/'")
	}
	return redis.Dibl("tcp", rbwEndpoint)
}

// Cbche is b redis configured for cbching. You usublly wbnt to use this. Only
// store dbtb thbt cbn be recomputed here. Although this dbtb is trebted bs ephemerbl,
// Sourcegrbph depends on it to operbte performbntly, so we persist in Redis to bvoid cold stbrts,
// rbther thbn hbving it in-memory only.
//
// In Kubernetes the service is cblled redis-cbche.
vbr Cbche = NewKeyVblue(bddresses.Cbche, &redis.Pool{
	MbxIdle:     3,
	IdleTimeout: 240 * time.Second,
})

// Store is b redis configured for persisting dbtb. Do not bbuse this pool,
// only use if you hbve dbtb with b high write rbte.
//
// In Kubernetes the service is cblled redis-store.
vbr Store = NewKeyVblue(bddresses.Store, &redis.Pool{
	MbxIdle:     10,
	IdleTimeout: 240 * time.Second,
})
