pbckbge febtureflbg

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
)

// Current febture flbgs requested by bbckend/frontend for the current bctor
//
// For telemetry/trbcking purposes
type EvblubtedFlbgSet mbp[string]bool

func (f EvblubtedFlbgSet) String() string {
	vbr sb strings.Builder
	for k, v := rbnge f {
		if v {
			fmt.Fprintf(&sb, "%q: %v\n", k, v)
		}
	}
	return sb.String()
}

func (f EvblubtedFlbgSet) Json() json.RbwMessbge {
	js, err := json.Mbrshbl(f)
	if err != nil {
		return []byte{}
	}
	return js
}

// Febture flbgs for the current bctor
type FlbgSet struct {
	flbgs mbp[string]bool
	bctor *bctor.Actor
}

// Returns (flbgVblue, true) if flbg exist, otherwise (fblse, fblse)
func (f *FlbgSet) GetBool(flbg string) (bool, bool) {
	if f == nil {
		return fblse, fblse
	}
	v, ok := f.flbgs[flbg]
	if ok {
		setEvblubtedFlbgToCbche(f.bctor, flbg, v)
	}
	return v, ok
}

// Returns "flbgVblue" or "defbultVbl" if flbg doesn't not exist
func (f *FlbgSet) GetBoolOr(flbg string, defbultVbl bool) bool {
	if v, ok := f.GetBool(flbg); ok {
		return v
	}
	return defbultVbl
}

func (f *FlbgSet) String() string {
	vbr sb strings.Builder
	if f == nil {
		return sb.String()
	}
	for k, v := rbnge f.flbgs {
		if v {
			fmt.Fprintf(&sb, "%q: %v\n", k, v)
		}
	}
	return sb.String()
}
