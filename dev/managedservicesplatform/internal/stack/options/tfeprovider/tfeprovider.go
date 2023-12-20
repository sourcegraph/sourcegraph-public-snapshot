package tfeprovider

import (
	tfe "github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/terraformcloud"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// With configures a stack to be able to use Terraform Enterprise (TFC) resources.
func With(tfeToken gsmsecret.DataConfig) stack.NewStackOption {
	return func(s stack.Stack) error {
		_ = tfe.NewTfeProvider(s.Stack, pointers.Ptr("tfe"),
			&tfe.TfeProviderConfig{
				Hostname:     pointers.Ptr(terraformcloud.Hostname),
				Organization: pointers.Ptr(terraformcloud.Organization),
				Token:        &gsmsecret.Get(s.Stack, resourceid.New("tfe-provider-token"), tfeToken).Value,
			})
		return nil
	}
}
