package codeowners

import (
	"fmt"
	"strings"
)

// Repr returns a string representation that resembles the syntax
// of a CODEOWNERS file. Order matters for every repeated field within
// the proto (as within the CODEOWNERS file), so the returned text
// representation is deterministic. This is useful in tests,
// where deep comparison may not work due to protobuf metadata.
func (f *Ruleset) Repr() string {
	w := new(strings.Builder)
	var lastSeenSection string
	for _, r := range f.proto.GetRule() {
		if s := r.SectionName; s != lastSeenSection {
			fmt.Fprintf(w, "[%s]\n", s)
			lastSeenSection = s
		}
		fmt.Fprint(w, r.Pattern)
		for _, o := range r.GetOwner() {
			if h := o.GetHandle(); h != "" {
				fmt.Fprintf(w, " @%s", h)
			}
			if e := o.GetEmail(); e != "" {
				fmt.Fprintf(w, " %s", e)
			}
		}
		fmt.Fprintln(w)
	}
	return w.String()
}
