package notif

import (
	"bytes"
	"text/template"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ext/slack"
)

// ActionContext represents an action. Action* fields are what happened,
// Object* fields are what the action was done on.
type ActionContext struct {
	Person     *sourcegraph.Person
	Recipients []*sourcegraph.Person

	ActionContent string
	ActionType    string
	ObjectID      int64
	ObjectRepo    string
	ObjectTitle   string
	ObjectType    string
	ObjectURL     string

	// SlackOpts specifies what to post to Slack. If empty it will be generated
	SlackOpts slack.PostOpts
}

// Message is a generic way to notify users about an event happening
func Action(nctx ActionContext) {
	if nctx.SlackOpts.Msg == "" {
		msg, err := generateSlackMessage(nctx)
		if err != nil {
			log15.Error("Error generating slack message for action", "ActionContext", nctx)
			return
		}
		nctx.SlackOpts.Msg = msg

	}
	slack.PostMessage(nctx.SlackOpts)
}

func generateSlackMessage(nctx ActionContext) (string, error) {
	tmpl := template.Must(template.New("slack-action").Parse(
		"*{{.Person.PersonSpec.Login}}* {{.ActionType}} <{{.ObjectURL}}|{{.ObjectRepo}} {{.ObjectType}} #{{.ObjectID}}>: {{.ObjectTitle}}{{if .Recipients}} /cc{{end}}{{range .Recipients}} @{{.PersonSpec.Login}}{{end}}{{if .ActionContent}}\n\n{{.ActionContent}}{{end}}"))
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, nctx)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
