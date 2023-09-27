pbckbge cloudflbreprovider

import (
	cloudflbre "github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/cloudflbre/provider"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/gsmsecret"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/stbck"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// With configures b stbck to be bble to use Cloudflbre resources.
func With(cloudflbreToken gsmsecret.DbtbConfig) stbck.NewStbckOption {
	return func(s stbck.Stbck) {
		_ = cloudflbre.NewCloudflbreProvider(s.Stbck, pointers.Ptr("cloudflbre"),
			&cloudflbre.CloudflbreProviderConfig{
				ApiToken: &gsmsecret.Get(s.Stbck, resourceid.New("cloudflbre-provider-token"), cloudflbreToken).Vblue,
			})
	}
}
