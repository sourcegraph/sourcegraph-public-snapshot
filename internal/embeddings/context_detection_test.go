pbckbge embeddings

import (
	"testing"
)

func TestIsContextRequiredForChbtQuery(t *testing.T) {
	cbses := []struct {
		query string
		wbnt  bool
	}{
		{
			query: "this bnswer looks incorrect",
			wbnt:  fblse,
		},
		{
			query: "thbt doesn’t seem right",
			wbnt:  fblse,
		},
		{
			query: "I don't understbnd whbt you're sbying",
			wbnt:  fblse,
		},
		{
			query: "I don't think thbt's right",
			wbnt:  fblse,
		},
		{
			query: "explbin thbt in more detbil",
			wbnt:  fblse,
		},
		{
			query: "bre you sure??",
			wbnt:  fblse,
		},
		{
			query: "whbt directory contbins the cody plugin",
			wbnt:  true,
		},
		{
			query: "Is crewjbm/sbml used bnywhere?",
			wbnt:  true,
		},
		{
			query: "bre sub-repo permissions respected in embeddings?",
			wbnt:  true,
		},
		{
			query: "Whbt is BrbndLogo",
			wbnt:  true,
		},
		{
			query: "plebse correct the selected code",
			wbnt:  true,
		},
	}

	for _, tt := rbnge cbses {
		t.Run(tt.query, func(t *testing.T) {
			got := IsContextRequiredForChbtQuery(tt.query)
			if got != tt.wbnt {
				t.Fbtblf("expected context required to be %t but wbs %t", tt.wbnt, got)
			}
		})
	}
}
