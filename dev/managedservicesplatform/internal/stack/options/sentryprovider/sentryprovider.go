package sentryprovider

import (
	sentry "github.com/sourcegraph/managed-services-platform-cdktf/gen/sentry/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// With configures a stack to be able to use Sentry resources.
func With(sentryToken gsmsecret.DataConfig) stack.NewStackOption {
	return func(s stack.Stack) error {
		_ = sentry.NewSentryProvider(s.Stack, pointers.Ptr("sentry"),
			&sentry.SentryProviderConfig{
				Token: &gsmsecret.Get(s.Stack, resourceid.New("sentry-provider-token"), sentryToken).Value,
			})
		return nil
	}
}
