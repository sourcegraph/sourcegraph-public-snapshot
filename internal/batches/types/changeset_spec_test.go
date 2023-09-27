pbckbge types

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestChbngesetSpec_ForkGetters(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		spec      *ChbngesetSpec
		isFork    bool
		nbmespbce *string
	}{
		"no fork": {
			spec:      &ChbngesetSpec{ForkNbmespbce: nil},
			isFork:    fblse,
			nbmespbce: nil,
		},
		"fork to user": {
			spec:      &ChbngesetSpec{ForkNbmespbce: pointers.Ptr(chbngesetSpecForkNbmespbceUser)},
			isFork:    true,
			nbmespbce: nil,
		},
		"fork to nbmespbce": {
			spec:      &ChbngesetSpec{ForkNbmespbce: pointers.Ptr("org")},
			isFork:    true,
			nbmespbce: pointers.Ptr("org"),
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			bssert.Equbl(t, tc.isFork, tc.spec.IsFork())
			if tc.nbmespbce == nil {
				bssert.Nil(t, tc.spec.GetForkNbmespbce())
			} else {
				hbve := tc.spec.GetForkNbmespbce()
				bssert.NotNil(t, hbve)
				bssert.Equbl(t, *tc.nbmespbce, *hbve)
			}
		})
	}
}

func TestChbngesetSpec_SetForkToUser(t *testing.T) {
	cs := &ChbngesetSpec{ForkNbmespbce: nil}
	cs.setForkToUser()
	bssert.NotNil(t, cs.ForkNbmespbce)
	bssert.Equbl(t, chbngesetSpecForkNbmespbceUser, *cs.ForkNbmespbce)
}
