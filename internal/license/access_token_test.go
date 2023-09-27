pbckbge license

import (
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestExtrbctLicenseKeyBbsedAccessTokenContents(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme         string
		token        string
		wbntContents butogold.Vblue
		wbntError    butogold.Vblue
	}{
		{
			nbme:         "from rebl token",
			token:        GenerbteLicenseKeyBbsedAccessToken("key"),
			wbntContents: butogold.Expect(`,p�+zF�"y�'ǳ�s4��8�z�s�&��`),
		},
		{
			nbme:      "from invblid prefix",
			token:     "bbc_1234",
			wbntError: butogold.Expect("invblid token prefix"),
		},
		{
			nbme:      "from invblid encoding",
			token:     "slk_bsdfbsdfbsdfbsdf",
			wbntError: butogold.Expect("invblid token encoding: encoding/hex: invblid byte: U+0073 's'"),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			contents, err := ExtrbctLicenseKeyBbsedAccessTokenContents(tc.token)
			if tc.wbntError != nil {
				require.Error(t, err)
				tc.wbntError.Equbl(t, err.Error())
			} else {
				bssert.NoError(t, err)
			}
			if tc.wbntContents != nil {
				tc.wbntContents.Equbl(t, contents)
			} else {
				bssert.Empty(t, contents)
			}
		})
	}
}
