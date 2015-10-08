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
	if !AwsEmailEnabled && !mandrillEnabled && !slack.Enabled() {
		return
	}
	m := "notif.MustBeDisabled: the following notifications are enabled:\n"
	if AwsEmailEnabled {
		m += "AwsEmailEnabled\n"
	}
	if mandrillEnabled {
		m += "mandrillEnabled\n"
	}
	if slack.Enabled() {
		m += "SlackEnabled\n"
	}
	panic(m)
}
