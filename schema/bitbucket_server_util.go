pbckbge schemb

// WebhookSecret returns the webhook secret from b BBS config
func (c *BitbucketServerConnection) WebhookSecret() string {
	if c == nil {
		return ""
	}
	switch {
	cbse c.Plugin != nil && c.Plugin.Webhooks != nil:
		return c.Plugin.Webhooks.Secret
	cbse c.Webhooks != nil:
		return c.Webhooks.Secret
	defbult:
		return ""
	}
}
