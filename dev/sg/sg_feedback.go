package main

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	_ "os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	_ "github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	newDiscussionURL = "https://github.com/sourcegraph/sourcegraph/discussions/new"
)

var feedbackGithubToken string
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
			Usage:       "The feedback you want to provide. If empty then your $EDITOR will be launched to write your feedback in",
			Required:    false,
			Destination: &feedbackBody,
		},
		&cli.StringFlag{
			Name:        "editor",
			Usage:       "The editor command to use when launching your editor. Defaults to $EDITOR",
			Required:    false,
			Destination: &feedbackEditor,
			Value:       os.Getenv("EDITOR"),
		},
	},
}

func feedbackExec(ctx *cli.Context) error {
	if err := sendFeedback(ctx.Context, feedbackTitle, "developer-experience", feedbackBody); err != nil {
		return err
	}
	return nil
}

func sendFeedback(ctx context.Context, title, category, body string) error {
	values := make(url.Values)
	values["category"] = []string{category}
	values["title"] = []string{title}
	// No feedback ? lets launch an editor to get some feedback
	if body == "" {
		content, err := gatherFeedback(ctx)
		if err != nil {
			std.Out.WriteFailuref("A problem occured while gathering feedback but continuing: %s", err)

		}
		body = content
	}
	values["body"] = []string{body}

	feedbackURL, err := url.Parse(newDiscussionURL)
	if err != nil {
		return err
	}

	if len(values) > 0 {
		feedbackURL.RawQuery = values.Encode()
	}
	std.Out.Writef("Launching your browser to complete feedback", feedbackURL)

	if err := open.URL(feedbackURL.String()); err != nil {
		return errors.Wrapf(err, "failed to launch browser for url %q", feedbackURL.String())
	}

	return nil
}

func gatherFeedback(ctx context.Context) (string, error) {
	base, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	contentPath := filepath.Join(base, ".FEEDBACK_BODY")
	_, err = os.Create(contentPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create file .FEEDBACK_BODY")
	}
	defer func() {
		os.Remove(contentPath)
	}()

	cmd := feedbackEditor + " " + contentPath
	err = launchEditor(ctx, cmd)
	if err != nil {
		std.Out.WriteSuggestionf("There was a problem in launching your editor. To use a different command to launch your editor specify --editor. Alternatively you can provide your feedback with the --message flag")
		return "", errors.Wrapf(err, "failed to launch editor using command %q", cmd)
	}

	content, err := ioutil.ReadFile(contentPath)
	return string(content), err
}

func launchEditor(ctx context.Context, editorCmd string) error {
	std.Out.WriteNoticef("Launching your editor to gather feedback ...", editorCmd)
	cmdParts := strings.Split(editorCmd, " ")

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
