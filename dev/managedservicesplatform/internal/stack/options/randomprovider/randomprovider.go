package randomprovider

import (
	random "github.com/sourcegraph/managed-services-platform-cdktf/gen/random/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// With configures a stack to be able to use random resources.
func With() stack.NewStackOption {
	return func(s stack.Stack) error {
		_ = random.NewRandomProvider(s.Stack, pointers.Ptr("random"), &random.RandomProviderConfig{})
		return nil
	}
}
