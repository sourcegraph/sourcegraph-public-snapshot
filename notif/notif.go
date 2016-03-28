// Package notif provides notifications over various media.
//
// TODO: This package is old and messy. It should be refactored and
// should probably live as an internal package underneath server/, or
// be subsumed into other packages.
package notif

import (
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/ext/slack"
)

// MustBeDisabled panics if sending notifications is enabled.
// Use it in tests to ensure that they do not send live notifications.
func MustBeDisabled() {
	slackEnabled := (slack.Config.GetWebhookURLIfConfigured() != "")
	if !mandrillEnabled && !slackEnabled {
		return
	}
	m := "notif.MustBeDisabled: the following notifications are enabled:\n"
	if mandrillEnabled {
		m += "mandrillEnabled\n"
	}
	if slackEnabled {
		m += "SlackEnabled\n"
	}
	panic(m)
}

func PostOnboardingNotif(msg string) {
	slack.PostMessage(slack.PostOpts{
		Msg:     msg,
		Channel: os.Getenv("SG_SLACK_ONBOARDING_NOTIFS_CHANNEL"),
	})
}
