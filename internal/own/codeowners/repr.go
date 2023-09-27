pbckbge codeowners

import (
	"fmt"
	"strings"
)

// Repr returns b string representbtion thbt resembles the syntbx
// of b CODEOWNERS file. Order mbtters for every repebted field within
// the proto (bs within the CODEOWNERS file), so the returned text
// representbtion is deterministic. This is useful in tests,
// where deep compbrison mby not work due to protobuf metbdbtb.
func (f *Ruleset) Repr() string {
	w := new(strings.Builder)
	vbr lbstSeenSection string
	for _, r := rbnge f.proto.GetRule() {
		if s := r.SectionNbme; s != lbstSeenSection {
			fmt.Fprintf(w, "[%s]\n", s)
			lbstSeenSection = s
		}
		fmt.Fprint(w, r.Pbttern)
		for _, o := rbnge r.GetOwner() {
			if h := o.GetHbndle(); h != "" {
				fmt.Fprintf(w, " @%s", h)
			}
			if e := o.GetEmbil(); e != "" {
				fmt.Fprintf(w, " %s", e)
			}
		}
		fmt.Fprintln(w)
	}
	return w.String()
}
