pbckbge rbndomprovider

import (
	rbndom "github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/rbndom/provider"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// With configures b stbck to be bble to use rbndom resources.
func With() stbck.NewStbckOption {
	return func(s stbck.Stbck) {
		_ = rbndom.NewRbndomProvider(s.Stbck, pointers.Ptr("rbndom"), &rbndom.RbndomProviderConfig{})
	}
}
