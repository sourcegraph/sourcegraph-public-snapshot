package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const newDiscussionURL = "https://github.com/sourcegraph/sourcegraph/discussions/new"

// addFeedbackFlags adds a '--feedback' flag to each command to generate feedback
func addFeedbackFlags(commands []*cli.Command) {
	for _, command := range commands {
		if command.Action != nil {
			feedbackFlag := cli.BoolFlag{
				Name:  "feedback",
				Usage: "provide feedback about this command by opening up a GitHub discussion",
			}

			command.Flags = append(command.Flags, &feedbackFlag)
			action := command.Action
			command.Action = func(ctx *cli.Context) error {
				if feedbackFlag.Get(ctx) {
					return feedbackAction(ctx)
				}
				return action(ctx)
			}
		}

		addFeedbackFlags(command.Subcommands)
	}
}

var feedbackCommand = &cli.Command{
	Name:     "feedback",
	Usage:    "Provide feedback about sg",
	Category: CategoryUtil,
	Action:   feedbackAction,
}

func feedbackAction(ctx *cli.Context) error {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Gathering feedback for sg %s ...", ctx.Command.FullName()))
	title, body, err := gatherFeedback(ctx, std.Out, os.Stdin)
	if err != nil {
		return err
	}
	body = addSGInformation(ctx, body)

	if err := sendFeedback(ctx.Context, title, "developer-experience", body); err != nil {
		return err
	}
	return nil
}

func gatherFeedback(ctx *cli.Context, out *std.Output, in io.Reader) (string, string, error) {
	out.Promptf("Write your feedback below and press <CTRL+D> when you're done.\n")
	body, err := io.ReadAll(in)
	if err != nil && err != io.EOF {
		return "", "", err
	}

	out.Promptf("The title of your feedback is going to be \"sg %s\". Anything else you want to add? (press <Enter> to skip)", ctx.Command.FullName())
	reader := bufio.NewReader(in)
	userTitle, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	title := "sg " + ctx.Command.FullName()
	userTitle = strings.TrimSpace(userTitle)
	switch strings.ToLower(userTitle) {
	case "", "na", "no", "nothing", "nope":
		// if the userTitle matches anyone of these words, don't add it to the final title
		break
	default:
		title = title + " - " + userTitle
	}

	return title, strings.TrimSpace(string(body)), nil
}

func addSGInformation(ctx *cli.Context, body string) string {
	tplt := template.Must(template.New("SG").Funcs(template.FuncMap{
		"inline_code": func(s string) string { return fmt.Sprintf("`%s`", s) },
	}).Parse(`{{.Content}}


### {{ inline_code "sg" }} information

Commit: {{ inline_code .Commit}}
Command: {{ inline_code .Command}}
Flags: {{ inline_code .Flags}}
    `))

	flagPair := []string{}
	for _, f := range ctx.FlagNames() {
		if f == "feedback" {
			continue
		}
		flagPair = append(flagPair, fmt.Sprintf("%s=%v", f, ctx.Value(f)))
	}

	var buf bytes.Buffer
	data := struct {
		Content string
		Commit  string
		Command string
		Flags   string
	}{
		body,
		BuildCommit,
		"sg " + ctx.Command.FullName(),
		strings.Join(flagPair, " "),
	}
	_ = tplt.Execute(&buf, data)

	return buf.String()
}

func sendFeedback(ctx context.Context, title, category, body string) error {
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
