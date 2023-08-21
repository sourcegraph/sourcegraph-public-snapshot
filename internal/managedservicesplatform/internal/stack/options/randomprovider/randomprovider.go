package randomprovider

import (
	random "github.com/sourcegraph/controller-cdktf/gen/random/provider"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

// With configures a stack to be able to use random resources.
func With() stack.NewStackOption {
	return func(s stack.Stack) {
		_ = random.NewRandomProvider(s.Stack, pointer.Value("random"), &random.RandomProviderConfig{})
	}
}
