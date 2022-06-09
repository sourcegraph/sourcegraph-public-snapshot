package featureflag

import (
	"fmt"
	"strings"
)

type FlagSet map[string]bool

func (f FlagSet) GetBool(flag string) (bool, bool) {
	v, ok := f[flag]
	return v, ok
}

func (f FlagSet) GetBoolOr(flag string, defaultVal bool) bool {
	if v, ok := f[flag]; ok {
		return v
	}
	return defaultVal
}

func (f FlagSet) String() string {
	var sb strings.Builder
	for k, v := range f {
		if v {
			fmt.Fprintf(&sb, "%q: %v\n", k, v)
		}
	}
	return sb.String()
}
