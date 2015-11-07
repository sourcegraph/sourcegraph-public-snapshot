// Package notif provides notifications over various media.
//
// TODO: This package is old and messy. It should be refactored and
// should probably live as an internal package underneath server/, or
// be subsumed into other packages.
package notif

import "src.sourcegraph.com/sourcegraph/ext/slack"

// MustBeDisabled panics if sending notifications is enabled.
// Use it in tests to ensure that they do not send live notifications.
func MustBeDisabled() {
	slackEnabled := (slack.Config.GetWebhookURLIfConfigured() != "")
	if !AwsEmailEnabled && !mandrillEnabled && !slackEnabled {
		return
	}
	m := "notif.MustBeDisabled: the following notifications are enabled:\n"
	if AwsEmailEnabled {
		m += "AwsEmailEnabled\n"
	}
	if mandrillEnabled {
		m += "mandrillEnabled\n"
	}
	if slackEnabled {
		m += "SlackEnabled\n"
	}
	panic(m)
}
