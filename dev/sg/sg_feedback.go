package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	_ "os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	_ "github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

const (
	newDiscussionURL = "https://github.com/sourcegraph/sourcegraph/discussions/new"
)

var noEditorErr error = errors.New("$EDITOR environment variable was empty")
var emptyFeedbackErr error = errors.New("Feedback message has no content")
var feedbackGithubToken string
var feedbackTitle string
var feedbackBody string

var feedbackCommand = &cli.Command{
	Name:     "feedback",
	Usage:    "opens up a Github disccussion page to provide feedback about sg",
	Category: CategoryCompany,
	Action:   feedbackExec,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "title",
			Usage:       "title of the feedback discussion to be created",
			Required:    false,
			Destination: &feedbackTitle,
		},
		&cli.StringFlag{
			Name:        "description",
			Usage:       "the feedback you want to provide",
			Required:    false,
			Destination: &feedbackBody,
		},
		&cli.StringFlag{
			Name:        "github.token",
			Usage:       "GitHub token to use when making API requests, defaults to $GITHUB_TOKEN.",
			Destination: &feedbackGithubToken,
			Value:       os.Getenv("GITHUB_TOKEN"),
		},
	},
}

func feedbackExec(ctx *cli.Context) error {
	if ctx.Args().Len() == 0 || feedbackGithubToken == "" {
		err := openBrowserForFeedback()
		if err != nil {
			std.Out.WriteFailuref("failed to open browser to %s: %v", newDiscussionURL)
		}
		return err
	}
	// If we have a Github Token, then we should have a title!
	if feedbackTitle == "" {
		std.Out.WriteWarningf("cannot create a discussion without a title")
		return errors.New("cannot create a discussion without a title")
	}
	// when all else fails, open the browser for a user to provide feedback
	if feedbackBody == "" {
		content, err := gatherFeedback(ctx.Context)
		if err != nil {
			if err == noEditorErr || err == emptyFeedbackErr {
				return openBrowserForFeedback()
			}
		}

		feedbackBody = content
	}

	ghc := github.NewClient(oauth2.NewClient(ctx.Context, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: feedbackGithubToken},
	)))
	// we create the discussion using the github api
	println("contacting github")
	url, err := createFeedbackDisccusion(ctx.Context, ghc, "test from sg", "LOTS OF CONTENTINO")
	if err != nil {
		println(err)
		return err
	}

	fmt.Println(url)

	return nil
}

func createFeedbackDisccusion(ctx context.Context, client *github.Client, title, content string) (string, error) {
	if content == "" {
		c, err := gatherFeedback(ctx)
		if err != nil {
			return "", err
		}
		content = c
	}
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}

	fmt.Printf("User: %+v\n", user)

	team, _, err := client.Teams.GetTeamBySlug(ctx, "sourcegraph", "dev-experience")
	if err != nil {
		return "", err
	}

	fmt.Printf("Team: %+v\n", team)

	discussion := github.TeamDiscussion{
		Author: user,
		Title:  &title,
		Body:   &content,
	}

	discResult, _, err := client.Teams.CreateDiscussionBySlug(ctx, "sourcegraph", team.GetSlug(), discussion)
	if err != nil {
		return "", err
	}

	return discResult.GetHTMLURL(), err
}

func openBrowserForFeedback() error {
	err := open.URL(newDiscussionURL)
	if err != nil {
	}
	return err
}

func gatherFeedback(ctx context.Context) (string, error) {
	base, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	contentPath := filepath.Join(base, ".FEEDBACK_BODY")
	_, err = os.Create(contentPath)
	if err != nil {
		return "", err
	}
	defer func() {
		os.Remove(contentPath)
	}()

	err = launchEditorForFile(ctx, contentPath)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadFile(contentPath)
	if len(content) == 0 {
		return "", emptyFeedbackErr
	}
	return string(content), err
}

func editorFlags(editor string) string {
	switch editor {
	case "code":
		return "code -w"
	default:
		return editor
	}
}

func launchEditorForFile(ctx context.Context, file string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return noEditorErr
	}

	uctx, _ := usershell.Context(ctx)
	cmd := usershell.Cmd(uctx, fmt.Sprintf("%s %s", editorFlags(editor), file))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
