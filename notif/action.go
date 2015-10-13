package notif

import (
	"bytes"
	html "html/template"
	"text/template"

	"github.com/mattbaird/gochimp"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ext/slack"
)

// ActionContext represents an action. Action* fields are what happened,
// Object* fields are what the action was done on. For more context on what
// each field is, please see action_test.go
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
	sendSlackMessage(nctx)
	sendEmailMessage(nctx)
}

func sendSlackMessage(nctx ActionContext) {
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

func sendEmailMessage(nctx ActionContext) {
	msg, err := generateHTMLFragment(nctx)
	if err != nil {
		log15.Error("Error generating email message for action", "ActionContext", nctx)
		return
	}
	sendEmail := func(p *sourcegraph.Person) {
		if p.PersonSpec.Email == "" {
			return
		}
		name := p.FullName
		if name == "" {
			name = p.PersonSpec.Login
		}
		SendMandrillTemplate("message-generic", name, p.PersonSpec.Email, []gochimp.Var{{Name: "MESSAGE", Content: msg}})
	}
	actorIncluded := false
	for _, p := range nctx.Recipients {
		sendEmail(p)
		if p.PersonSpec.UID == nctx.Person.PersonSpec.UID {
			actorIncluded = true
		}
	}
	if !actorIncluded {
		sendEmail(nctx.Person)
	}
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

func generateHTMLFragment(nctx ActionContext) (string, error) {
	tmpl := html.Must(html.New("html-action").Parse(
		`<b>{{.Person.PersonSpec.Login}}</b> {{.ActionType}} <a href="{{.ObjectURL}}">{{.ObjectRepo}} {{.ObjectType}} #{{.ObjectID}}</a>: {{.ObjectTitle}}`))
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, nctx)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
