package main

import (
	"bytes"
	"context"
	"net/url"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	_ "github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const newDiscussionURL = "https://github.com/sourcegraph/sourcegraph/discussions/new"

var feedbackTitle string
var feedbackBody string
var feedbackEditor string

var feedbackCommand = &cli.Command{
	Name:     "feedback",
	Usage:    "opens up a Github disccussion page to provide feedback about sg",
	Category: CategoryCompany,
	Action:   feedbackExec,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "title",
			Usage:       "Title of the feedback discussion to be created",
			Required:    false,
			Destination: &feedbackTitle,
		},
		&cli.StringFlag{
			Name:        "message",
			Usage:       "The feedback you want to provide",
			Required:    false,
			Destination: &feedbackBody,
		},
	},
}

func feedbackExec(ctx *cli.Context) error {
	if err := sendFeedback(ctx.Context, feedbackTitle, "developer-experience", feedbackBody); err != nil {
		return err
	}
	return nil
}

func addSGInformation(content string) string {
	tplt := template.Must(template.New("SG").Parse(`{{.Content}}

### SG Information

Commit: {{.Tick}}{{.Commit}}{{.Tick}}
    `))

	var buf bytes.Buffer
	data := struct {
		Content string
		Tick    string
		Commit  string
	}{
		content,
		"`",
		BuildCommit,
	}
	_ = tplt.Execute(&buf, data)

	return buf.String()
}

func sendFeedback(ctx context.Context, title, category, body string) error {
	body = addSGInformation(body)
	values := make(url.Values)
	values["category"] = []string{category}
	values["title"] = []string{title}
	values["body"] = []string{body}
	values["labels"] = []string{"sg,team/devx"}

	feedbackURL, err := url.Parse(newDiscussionURL)
	if err != nil {
		return err
	}

	feedbackURL.RawQuery = values.Encode()
	std.Out.WriteNoticef("Launching your browser to complete feedback")

	if err := open.URL(feedbackURL.String()); err != nil {
		return errors.Wrapf(err, "failed to launch browser for url %q", feedbackURL.String())
	}

	return nil
}
