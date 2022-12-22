package proto

import (
	"fmt"
	"strings"
)

func (f *File) Repr() string {
	w := new(strings.Builder)
	var lastSeenSection string
	for _, r := range f.Rule {
		if s := r.SectionName; s != lastSeenSection {
			fmt.Fprintf(w, "[%s]\n", s)
			lastSeenSection = s
		}
		fmt.Fprint(w, r.Pattern)
		for _, o := range r.Owner {
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
