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
