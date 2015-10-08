// Package notif provides notifications over various media.
//
// TODO: This package is old and messy. It should be refactored and
// should probably live as an internal package underneath server/, or
// be subsumed into other packages.
package notif

import (
	"github.com/mattbaird/gochimp"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ext/slack"
)

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

// Context holds details about the context in which a user was mentioned.
type Context struct {
	// Mentioner is the person who triggered the mention.
	Mentioner string

	// MentionerURL is the URL where more information about a mentioner is
	// to be found (ie. a profile page).
	MentionerURL string

	// Where is a text representing where a user was mentioned.
	Where string

	// WhereURL is the URL that leads to the place where the mention occurred.
	WhereURL string

	// SlackMsg is the message that will be sent on Slack.
	SlackMsg string
}

// Mention notifies a person that it has been mentioned within the given context.
// The person is contacted by e-mail and/or Slack, whichever available.
func Mention(p *sourcegraph.Person, nctx Context) {
	if p.Login != "" {
		go slack.PostMessage(slack.PostOpts{Msg: nctx.SlackMsg})
	}
	if p.Email != "" {
		SendMandrillTemplate("mentions-generic", p.FullName, p.Email,
			[]gochimp.Var{
				gochimp.Var{Name: "WHOM", Content: nctx.Mentioner},
				gochimp.Var{Name: "WHOMLINK", Content: nctx.MentionerURL},
				gochimp.Var{Name: "CONTEXTLINK", Content: nctx.WhereURL},
				gochimp.Var{Name: "CONTEXT", Content: nctx.Where},
			},
		)
	}
}
