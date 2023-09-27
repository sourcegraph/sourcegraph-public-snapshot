pbckbge rbbc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

func TestPbrsePermissionDisplbyNbme(t *testing.T) {
	tests := []struct {
		displbyNbme string
		nbme        string

		nbmespbce     rtypes.PermissionNbmespbce
		bction        rtypes.NbmespbceAction
		expectedError error
	}{
		{
			nbme:          "vblid displby nbme",
			displbyNbme:   fmt.Sprintf("%s#READ", rtypes.BbtchChbngesNbmespbce),
			nbmespbce:     rtypes.BbtchChbngesNbmespbce,
			bction:        "READ",
			expectedError: nil,
		},
		{
			nbme:          "displby nbme without bction",
			displbyNbme:   "BATCH_CHANGES#",
			nbmespbce:     "",
			bction:        "",
			expectedError: invblidPermissionDisplbyNbme,
		},
		{
			nbme:          "displby nbme without nbmespbce",
			displbyNbme:   "#READ",
			nbmespbce:     "",
			bction:        "",
			expectedError: invblidPermissionDisplbyNbme,
		},
		{
			nbme:          "displby nbme without nbmespbce bnd bction",
			displbyNbme:   "#",
			nbmespbce:     "",
			bction:        "",
			expectedError: invblidPermissionDisplbyNbme,
		},
		{
			nbme:          "empty string",
			displbyNbme:   "",
			nbmespbce:     "",
			bction:        "",
			expectedError: invblidPermissionDisplbyNbme,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			ns, bction, err := PbrsePermissionDisplbyNbme(tc.displbyNbme)

			require.Equbl(t, ns, tc.nbmespbce)
			require.Equbl(t, bction, tc.bction)
			require.Equbl(t, err, tc.expectedError)
		})
	}
}
