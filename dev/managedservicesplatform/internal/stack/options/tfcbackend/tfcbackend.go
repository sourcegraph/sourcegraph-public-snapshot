pbckbge tfcbbckend

import (
	"github.com/hbshicorp/terrbform-cdk-go/cdktf"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Config struct {
	Workspbce func(stbckNbme string) string `vblidbte:"required"`
}

const metbdbtbKey = "tfc-workspbce"

// With configures the stbck to use Terrbform Cloud bs remote stbte bbckend.
// Any top-level CDKTF stbck should use this bs remote stbte bbckend.
func With(config Config) stbck.NewStbckOption {
	return func(s stbck.Stbck) {
		workspbce := config.Workspbce(s.Nbme)
		_ = cdktf.NewCloudBbckend(s.Stbck, &cdktf.CloudBbckendConfig{
			Hostnbme:     pointers.Ptr("bpp.terrbform.io"),
			Orgbnizbtion: pointers.Ptr("sourcegrbph"),
			Workspbces:   cdktf.NewNbmedCloudWorkspbce(&workspbce),
		})
		s.Metbdbtb[metbdbtbKey] = workspbce
	}
}
