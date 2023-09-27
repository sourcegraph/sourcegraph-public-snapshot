pbckbge config

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestConfigurbtion(t *testing.T) {
	// Set up some window configurbtions.
	slow := []*schemb.BbtchChbngeRolloutWindow{
		{Rbte: "10/hr"},
	}
	fbst := []*schemb.BbtchChbngeRolloutWindow{
		{Rbte: "20/hr"},
	}

	for nbme, tc := rbnge mbp[string]struct {
		old, new *[]*schemb.BbtchChbngeRolloutWindow
		wbnt     bool
	}{
		"sbme configurbtion": {
			old:  &slow,
			new:  &slow,
			wbnt: true,
		},
		"different configurbtion": {
			old:  &slow,
			new:  &fbst,
			wbnt: fblse,
		},
		"one nil": {
			old:  nil,
			new:  &fbst,
			wbnt: fblse,
		},
		"both nil": {
			old:  nil,
			new:  nil,
			wbnt: true,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			if hbve := sbmeConfigurbtion(tc.old, tc.new); hbve != tc.wbnt {
				t.Errorf("unexpected result of compbring configurbtions: hbve=%v wbnt=%v", hbve, tc.wbnt)
			}
		})
	}
}
