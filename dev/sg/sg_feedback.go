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
)

const newDiscussionURL = "https://github.com/sourcegraph/sourcegraph/discussions/new"

type stopReadFunc func(lastRead string, err error) bool

func addFeedbackSubcommand(commands []*cli.Command) {
	giveFeedback := false
	feedbackFlag := cli.BoolFlag{
		Name:        "feedback",
		Usage:       "provide feedback about this command by opening up a Github discussion",
		Destination: &giveFeedback,
	}

	for _, command := range commands {
		command.Flags = append(command.Flags, &feedbackFlag)
		action := command.Action
		command.Action = func(ctx *cli.Context) error {
			if giveFeedback {
				return feedbackExec(ctx)
			}

			return action(ctx)
		}

		addFeedbackSubcommand(command.Subcommands)
	}
}

var feedbackCommand = &cli.Command{
	Name:     "feedback",
	Usage:    "opens up a Github discussion page to provide feedback about sg",
	Category: CategoryCompany,
	Action:   feedbackExec,
}

func feedbackExec(ctx *cli.Context) error {
	title, body, err := gatherFeedback(ctx)
	if err != nil {
		return err
	}
	body = addSGInformation(ctx, body)

	if err := sendFeedback(ctx.Context, title, "developer-experience", body); err != nil {
		return err
	}
	return nil
}

func gatherFeedback(ctx *cli.Context) (string, string, error) {
	std.Out.WriteNoticef("Gathering feedback for sg %s", ctx.Command.FullName())

	fmt.Println("What is the title of your feedback ?")
	title, err := readUntilDelim(os.Stdin, '\n')

	fmt.Println("Write your feedback below and press <CTRL+D> when you're done.")
	body, err := readUntilEOF(os.Stdin)
	if err != nil {
		return "", "", err
	}

	return title, body, nil
}

func readUntilEOF(r io.Reader) (string, error) {
	reader := bufio.NewReader(r)

	readFunc := func() (string, error) { return reader.ReadString('\n') }

	var eofFunc stopReadFunc = func(data string, err error) bool {
		if err != nil {
			return true
		}
		return false
	}

	return readUntil(readFunc, eofFunc)
}

func readUntilDelim(r io.Reader, delim byte) (string, error) {
	reader := bufio.NewReader(r)

	readFunc := func() (string, error) { return reader.ReadString(delim) }
	var firstReadStop stopReadFunc = func(data string, err error) bool {
		return true
	}

	return readUntil(readFunc, firstReadStop)

}

func readUntil(readFunc func() (string, error), stopRead stopReadFunc) (string, error) {
	var data string
	for {
		line, err := readFunc()
		data = data + line

		if stopRead(data, err) {
			break
		}
	}

	return data, nil
}

func addSGInformation(ctx *cli.Context, body string) string {
	tplt := template.Must(template.New("SG").Funcs(template.FuncMap{
		"inline_code": func(s string) string { return fmt.Sprintf("`%s`", s) },
	}).Parse(`{{.Content}}


### {{ inline_code "sg" }} Information

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
		Tick    string
		Commit  string
		Command string
		Flags   string
	}{
		body,
		"`",
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
