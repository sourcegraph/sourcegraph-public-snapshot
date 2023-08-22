package cloudflareprovider

import (
	cloudflare "github.com/sourcegraph/managed-services-platform-cdktf/gen/cloudflare/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

// With configures a stack to be able to use Cloudflare resources.
func With(cloudflareToken gsmsecret.DataConfig) stack.NewStackOption {
	return func(s stack.Stack) {
		_ = cloudflare.NewCloudflareProvider(s.Stack, pointer.Value("cloudflare"), &cloudflare.CloudflareProviderConfig{
			ApiToken: &gsmsecret.Get(s.Stack, resourceid.New("cloudflare-provider-token"), cloudflareToken).Value,
		})
	}
}
