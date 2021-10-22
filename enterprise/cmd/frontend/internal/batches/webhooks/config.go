package webhooks

import (
	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/schema"
)

// ConfigurationHasWebhooks checks if a configuration returned from
// ExternalService.Configuration() has one or more webhooks configured.
func ConfigurationHasWebhooks(cfg interface{}) (bool, error) {
	switch cfg := cfg.(type) {
	case *schema.BitbucketServerConnection:
		return cfg.WebhookSecret() != "", nil
	case *schema.GitHubConnection:
		return len(cfg.Webhooks) > 0, nil
	case *schema.GitLabConnection:
		return len(cfg.Webhooks) > 0, nil
	}

	return false, errors.Newf("external service configuration of type %T does not support webhooks", cfg)
}
