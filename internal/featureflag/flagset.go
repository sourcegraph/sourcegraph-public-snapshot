package featureflag

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

type EvaluatedFlagSet map[string]bool

func (f EvaluatedFlagSet) String() string {
	var sb strings.Builder
	for k, v := range f {
		if v {
			fmt.Fprintf(&sb, "%q: %v\n", k, v)
		}
	}
	return sb.String()
}

type FlagSet struct {
	flags map[string]bool
	actor *actor.Actor
}

func (f FlagSet) GetBool(flag string) (bool, bool) {
	v, ok := f.flags[flag]
	if ok {
		setEvaluatedFlagToCache(f.actor, flag, v)
	}
	return v, ok
}

func (f FlagSet) GetBoolOr(flag string, defaultVal bool) bool {
	if v, ok := f.GetBool(flag); ok {
		return v
	}
	return defaultVal
}

func (f FlagSet) String() string {
	var sb strings.Builder
	for k, v := range f.flags {
		if v {
			fmt.Fprintf(&sb, "%q: %v\n", k, v)
		}
	}
	return sb.String()
}
