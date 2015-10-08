package notif

import (
	"github.com/mattbaird/gochimp"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ext/slack"
)

// Context holds details about the context in which a user was mentioned.
type MentionContext struct {
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
func Mention(p *sourcegraph.Person, nctx MentionContext) {
	if p.Login != "" {
		slack.PostMessage(slack.PostOpts{Msg: nctx.SlackMsg})
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
