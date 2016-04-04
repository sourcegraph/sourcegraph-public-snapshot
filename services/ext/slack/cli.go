package slack

import "sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

// Flags defines settings (in the form of CLI flags) for Slack integration.
type Flags struct {
	Disable bool `long:"slack.disable" description:"disable Slack integration"`

	WebhookURL string `long:"slack.webhook-url" description:"the URL to the Slack webhook endpoint for posting Slack notifications" env:"SG_SLACK_WEBHOOK_URL"`

	DefaultChannel string `long:"slack.default-channel" description:"the default channel to post notifications to" default:"dev-bot"`

	DefaultUsername string `long:"slack.default-username" description:"the default username to post notifications as" default:"Sourcegraph"`

	DefaultIcon string `long:"slack.default-icon" description:"the default icon to use in notifications" default:"https://sourcegraph.com/static/img/favicon.png"`

	DisableLinkNames bool `long:"slack.disable-link-names" description:"if true, @mentions won't notify users"`
}

// GetWebhookURLIfConfigured returns the Slack Webhook URL if Slack integration
// is enabled and a webhook URL is configured, otherwise it returns an empty string.
func (f *Flags) GetWebhookURLIfConfigured() string {
	if f.Disable {
		return ""
	}
	return f.WebhookURL
}

// Config is the currently active Slack integration config (as set by the CLI flags).
var Config Flags

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("Slack", "Slack", &Config)
	})
}
