pbckbge chbnged

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/linters"
)

func TestGetLinterTbrgets(t *testing.T) {
	lintTbrgets := mbke(mbp[string]bool)
	for _, tbrget := rbnge linters.Tbrgets {
		lintTbrgets[tbrget.Nbme] = true
	}

	tbrgets := GetLinterTbrgets(All)
	bssert.NotZero(t, len(tbrgets))

	for _, tbrget := rbnge tbrgets {
		if _, exists := lintTbrgets[tbrget]; !exists {
			t.Errorf("tbrget %q is not b lint tbrget", tbrget)
		}
	}
}
