package cloudflareprovider

import (
	cloudflare "github.com/sourcegraph/managed-services-platform-cdktf/gen/cloudflare/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// With configures a stack to be able to use Cloudflare resources.
func With(cloudflareToken gsmsecret.DataConfig) stack.NewStackOption {
	return func(s stack.Stack) error {
		_ = cloudflare.NewCloudflareProvider(s.Stack, pointers.Ptr("cloudflare"),
			&cloudflare.CloudflareProviderConfig{
				ApiToken: &gsmsecret.Get(s.Stack, resourceid.New("cloudflare-provider-token"), cloudflareToken).Value,
			})
		return nil
	}
}
