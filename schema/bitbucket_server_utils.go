package schema

// WebhookSecret returns the webhook secret from a BBS config
func (c *BitbucketServerConnection) WebhookSecret() string {
	switch {
	case c.Plugin != nil && c.Plugin.Webhooks != nil:
		return c.Plugin.Webhooks.Secret
	case c.Webhooks != nil:
		return c.Webhooks.Secret
	default:
		return ""
	}
}

func (c *BitbucketServerConnection) WebhookAutoCreationEnabled() bool {
	if c.Plugin == nil {
		return false
	}
	if c.Plugin.Webhooks == nil {
		return false
	}
	return c.Plugin.Webhooks.AutomaticCreation == "enabled"
}
