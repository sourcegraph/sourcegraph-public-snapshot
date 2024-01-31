package nobl9provider

import (
	nobl9 "github.com/sourcegraph/managed-services-platform-cdktf/gen/nobl9/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	ClientID     string
	Organization string
	Nobl9Token   gsmsecret.DataConfig
}

// With configures a stack to be able to use nobl9 resources.
func With(config Config) stack.NewStackOption {
	return func(s stack.Stack) error {
		_ = nobl9.NewNobl9Provider(s.Stack, pointers.Ptr("nobl9"),
			&nobl9.Nobl9ProviderConfig{
				ClientId:     &config.ClientID,
				Organization: &config.Organization,
				ClientSecret: &gsmsecret.Get(s.Stack, resourceid.New("nobl9-provider-token"), config.Nobl9Token).Value,
			})
		return nil
	}
}
