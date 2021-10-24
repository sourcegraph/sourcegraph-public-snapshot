package webhooks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestConfigurationHasWebhooks(t *testing.T) {
	t.Run("supported external services", func(t *testing.T) {
		for name, tc := range map[string]struct {
			cfg  interface{}
			want bool
		}{
			"Bitbucket Server without webhooks": {
				cfg:  &schema.BitbucketServerConnection{},
				want: false,
			},
			"Bitbucket Server with webhooks": {
				cfg: &schema.BitbucketServerConnection{
					Webhooks: &schema.Webhooks{
						Secret: "secret",
					},
				},
				want: true,
			},
			"GitHub without webhooks": {
				cfg:  &schema.GitHubConnection{},
				want: false,
			},
			"GitHub with webhooks": {
				cfg: &schema.GitHubConnection{
					Webhooks: []*schema.GitHubWebhook{
						{},
					},
				},
				want: true,
			},
			"GitLab without webhooks": {
				cfg:  &schema.GitLabConnection{},
				want: false,
			},
			"GitLab with webhooks": {
				cfg: &schema.GitLabConnection{
					Webhooks: []*schema.GitLabWebhook{
						{},
					},
				},
				want: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				have, err := ConfigurationHasWebhooks(tc.cfg)
				assert.Nil(t, err)
				assert.Equal(t, tc.want, have)
			})
		}
	})

	t.Run("unsupported external services", func(t *testing.T) {
		for _, cfg := range []interface{}{
			&schema.AWSCodeCommitConnection{},
			&schema.BitbucketCloudConnection{},
			&schema.GitoliteConnection{},
			&schema.JVMPackagesConnection{},
			&schema.OtherExternalServiceConnection{},
			&schema.PerforceConnection{},
			&schema.PhabricatorConnection{},
		} {
			t.Run(fmt.Sprintf("%T", cfg), func(t *testing.T) {
				_, err := ConfigurationHasWebhooks(cfg)
				assert.NotNil(t, err)
			})
		}
	})
}
