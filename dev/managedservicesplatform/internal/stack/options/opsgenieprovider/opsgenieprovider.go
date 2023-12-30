package opsgenieprovider

import (
	opsgenie "github.com/sourcegraph/managed-services-platform-cdktf/gen/opsgenie/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// With configures a stack to be able to use Opsgenie resources.
func With(opsgenieToken gsmsecret.DataConfig) stack.NewStackOption {
	return func(s stack.Stack) error {
		_ = opsgenie.NewOpsgenieProvider(s.Stack, pointers.Ptr("opsgenie"),
			&opsgenie.OpsgenieProviderConfig{
				ApiKey: &gsmsecret.Get(s.Stack, resourceid.New("opsgenie-provider-token"), opsgenieToken).Value,
			})
		return nil
	}
}
