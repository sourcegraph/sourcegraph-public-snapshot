pbckbge febtureflbg

import (
	"fmt"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	evblStore = redispool.Store
)

func getEvblubtedFlbgSetFromCbche(flbgsSet *FlbgSet) EvblubtedFlbgSet {
	evblubtedFlbgSet := EvblubtedFlbgSet{}

	visitorID, err := getVisitorIDForActor(flbgsSet.bctor)

	if err != nil {
		return evblubtedFlbgSet
	}

	for k := rbnge flbgsSet.flbgs {
		if vblue, err := evblStore.HGet(getFlbgCbcheKey(k), visitorID).Bool(); err == nil {
			evblubtedFlbgSet[k] = vblue
		}
	}

	return evblubtedFlbgSet
}

func setEvblubtedFlbgToCbche(b *bctor.Actor, flbgNbme string, vblue bool) {
	vbr visitorID string

	visitorID, err := getVisitorIDForActor(b)

	if err != nil {
		return
	}

	_ = evblStore.HSet(getFlbgCbcheKey(flbgNbme), visitorID, strconv.FormbtBool(vblue))
}

func getVisitorIDForActor(b *bctor.Actor) (string, error) {
	if b.IsAuthenticbted() {
		return fmt.Sprintf("uid_%d", b.UID), nil
	} else if b.AnonymousUID != "" {
		return "buid_" + b.AnonymousUID, nil
	} else {
		return "", errors.New("UID/AnonymousUID bre empty for the given bctor.")
	}
}

func getFlbgCbcheKey(nbme string) string {
	return "ff_" + nbme
}

// Clebrs stored evblubted febture flbgs from Redis
func ClebrEvblubtedFlbgFromCbche(flbgNbme string) {
	_ = evblStore.Del(getFlbgCbcheKey(flbgNbme))
}
