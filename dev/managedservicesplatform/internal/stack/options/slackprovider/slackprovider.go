package slackprovider

import (
	slack "github.com/sourcegraph/managed-services-platform-cdktf/gen/slack/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// With configures a stack to be able to use Slack resources.
func With(slackToken gsmsecret.DataConfig) stack.NewStackOption {
	return func(s stack.Stack) error {
		_ = slack.NewSlackProvider(s.Stack, pointers.Ptr("slack"),
			&slack.SlackProviderConfig{
				Token: &gsmsecret.Get(s.Stack, resourceid.New("slack-provider-token"), slackToken).Value,
			})
		return nil
	}
}
