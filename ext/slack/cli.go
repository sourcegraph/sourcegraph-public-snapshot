package slack

import (
	"os"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags defines settings (in the form of CLI flags) for Slack integration.
type Flags struct {
	Disable bool `long:"slack.disable" description:"disable Slack integration"`

	WebhookURL string `long:"slack.webhook-url" description:"the URL to the Slack webhook endpoint for posting Slack notifications"`

	DefaultChannel string `long:"slack.default-channel" description:"the default channel to post notifications to" default:"dev-bot"`

	DefaultUsername string `long:"slack.default-username" description:"the default username to post notifications as" default:"Sourcegraph"`

	DefaultIcon string `long:"slack.default-icon" description:"the default icon to use in notifications" default:"https://sourcegraph.com/static/img/favicon.png"`

	DisableLinkNames bool `long:"slack.disable-link-names" description:"if true, @mentions won't notify users"`
}

// GetWebhookURLIfConfigured returns the Slack Webhook URL if Slack integration
// is enabled and a URL is configured, otherwise it returns an empty string.
// A URL can be configured by setting the --slack.webhook-url flag. If this
// flag is not set, then the SG_SLACK_WEBHOOK_URL env variable will be tried.
func (f *Flags) GetWebhookURLIfConfigured() string {
	if f.Disable {
		return ""
	}
	if f.WebhookURL == "" {
		return os.Getenv("SG_SLACK_WEBHOOK_URL")
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
