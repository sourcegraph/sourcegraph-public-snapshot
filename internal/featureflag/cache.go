package featureflag

import (
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	evalStore = redispool.Store
)

func getEvaluatedFlagSetFromCache(flagsSet *FlagSet) EvaluatedFlagSet {
	evaluatedFlagSet := EvaluatedFlagSet{}

	visitorID, err := getVisitorIDForActor(flagsSet.actor)

	if err != nil {
		return evaluatedFlagSet
	}

	for k := range flagsSet.flags {
		if value, err := evalStore.HGet(getFlagCacheKey(k), visitorID).Bool(); err == nil {
			evaluatedFlagSet[k] = value
		}
	}

	return evaluatedFlagSet
}

func setEvaluatedFlagToCache(a *actor.Actor, flagName string, value bool) {
	var visitorID string

	visitorID, err := getVisitorIDForActor(a)

	if err != nil {
		return
	}

	_ = evalStore.HSet(getFlagCacheKey(flagName), visitorID, strconv.FormatBool(value))
}

func getVisitorIDForActor(a *actor.Actor) (string, error) {
	if a.IsAuthenticated() {
		return fmt.Sprintf("uid_%d", a.UID), nil
	} else if a.AnonymousUID != "" {
		return "auid_" + a.AnonymousUID, nil
	} else {
		return "", errors.New("UID/AnonymousUID are empty for the given actor.")
	}
}

func getFlagCacheKey(name string) string {
	return "ff_" + name
}

// Clears stored evaluated feature flags from Redis
func ClearEvaluatedFlagFromCache(flagName string) {
	_ = evalStore.Del(getFlagCacheKey(flagName))
}
