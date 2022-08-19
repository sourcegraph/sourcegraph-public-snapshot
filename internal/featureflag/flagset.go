package featureflag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

// Current feature flags requested by backend/frontend for the current actor
//
// For telemetry/tracking purposes
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

func (f EvaluatedFlagSet) Json() json.RawMessage {
	js, err := json.Marshal(f)
	if err != nil {
		return []byte{}
	}
	return js
}

// Feature flags for the current actor
type FlagSet struct {
	flags map[string]bool
	actor *actor.Actor
}

// Returns (flagValue, true) if flag exist, otherwise (false, false)
func (f *FlagSet) GetBool(flag string) (bool, bool) {
	if f == nil {
		return false, false
	}
	v, ok := f.flags[flag]
	if ok {
		setEvaluatedFlagToCache(f.actor, flag, v)
	}
	return v, ok
}

// Returns "flagValue" or "defaultVal" if flag doesn't not exist
func (f *FlagSet) GetBoolOr(flag string, defaultVal bool) bool {
	if v, ok := f.GetBool(flag); ok {
		return v
	}
	return defaultVal
}

func (f *FlagSet) String() string {
	var sb strings.Builder
	if f == nil {
		return sb.String()
	}
	for k, v := range f.flags {
		if v {
			fmt.Fprintf(&sb, "%q: %v\n", k, v)
		}
	}
	return sb.String()
}
