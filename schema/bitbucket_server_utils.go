package schema

// WebhookSecret returns the webhook secret from a BBS config
func (c *BitbucketServerConnection) WebhookSecret() string {
	if c == nil {
		return ""
	}
	switch {
	case c.Plugin != nil && c.Plugin.Webhooks != nil:
		return c.Plugin.Webhooks.Secret
	case c.Webhooks != nil:
		return c.Webhooks.Secret
	default:
		return ""
	}
}

// WebhookSyncDisabled returns true if no webhooks are configured or when webhook syncing is explicitly disabled.
func (c *BitbucketServerConnection) WebhookSyncDisabled() bool {
	if c == nil {
		return false
	}
	if c.Plugin == nil {
		return true
	}
	if c.Plugin.Webhooks == nil {
		return true
	}
	return c.Plugin.Webhooks.DisableSync
}
