package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	_ "os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	_ "github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	newDiscussionURL = "https://github.com/sourcegraph/sourcegraph/discussions/new"
)

var noEditorErr error = errors.New("$EDITOR environment variable was empty")
var emptyFeedbackErr error = errors.New("Feedback message has no content")
var feedbackGithubToken string
var feedbackTitle string
var feedbackContent string
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
			Destination: &feedbackContent,
		},
		&cli.StringFlag{
			Name:        "github.token",
			Usage:       "GitHub token with the 'write:discussion' permission to use when making API requests, defaults to $GITHUB_TOKEN",
			Destination: &feedbackGithubToken,
			Value:       os.Getenv("GITHUB_TOKEN"),
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
	if feedbackGithubToken == "" {
		std.Out.WriteWarningf("No Github Token value. Opening browser ...")
		err := openBrowserForFeedback()
		if err != nil {
			std.Out.WriteFailuref("failed to open browser to %s: %v", newDiscussionURL)
		}
		return err
	}
	if ctx.NumFlags() == 0 || feedbackTitle == "" {
		std.Out.WriteNoticef("Opening browser for feedback ...")
		return openBrowserForFeedback()
	}

	if err := newFeedback(ctx.Context, feedbackTitle, feedbackContent); err != nil {
		return err
	}
	return nil
}

func newFeedback(ctx context.Context, title, content string) error {
	// No feedback ? lets launch an editor to get some feedback
	if feedbackContent == "" {
		std.Out.WriteNoticef("launching editor with command %q to gather feedback ...", feedbackEditor)
		content, err := gatherFeedback(ctx)
		if err != nil {
			return err
		}

		feedbackContent = content
	}

	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: feedbackGithubToken},
	)))
	// we create the discussion using the github api
	url, err := createFeedbackDisccusion(ctx, ghc, feedbackTitle, feedbackContent)
	if err != nil {
		return err
	}

	std.Out.WriteSuccessf("Feedback created! See %s", url)
	return nil
}

func createFeedbackDisccusion(ctx context.Context, client *github.Client, title, content string) (string, error) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}

	fmt.Printf("User: %+v\n", user)

	team, _, err := client.Teams.GetTeamBySlug(ctx, "sourcegraph", "dev-experience")
	if err != nil {
		return "", errors.Wrapf(err, "failed to Github Team 'dev-experience'")
	}

	discussion := github.TeamDiscussion{
		Author: user,
		Title:  &title,
		Body:   &content,
	}

	discResult, _, err := client.Teams.CreateDiscussionBySlug(ctx, team.GetOrganization().GetName(), team.GetSlug(), discussion)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create Github discussion")
	}

	return discResult.GetHTMLURL(), err
}

func openBrowserForFeedback() error {
	if err := open.URL(newDiscussionURL); err != nil {
		return errors.Wrapf(err, "failed to launch browser for url %q", newDiscussionURL)
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
		return "", errors.Wrapf(err, "failed to launch editor using command %q", cmd)
	}

	content, err := ioutil.ReadFile(contentPath)
	if len(content) == 0 {
		return "", emptyFeedbackErr
	}
	return string(content), err
}

func launchEditor(ctx context.Context, editorCmd string) error {
	cmdParts := strings.Split(editorCmd, " ")

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
