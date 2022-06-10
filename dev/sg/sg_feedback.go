package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"

	_ "github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

var feedbackCommand = &cli.Command{
	Name:     "feedback",
	Usage:    "opens up a Github disccussion page to provide feedback about sg",
	Category: CategoryCompany,
	Action:   feedbackExec,
}

func feedbackExec(ctx *cli.Context) error {
	editor := os.Getenv("EDITOR")

	base, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	feedbackPath := filepath.Join(base, ".FEEDBACK_BODY")
	_, err = os.Create(feedbackPath)
	if err != nil {
		return nil
	}
	defer func() {
		os.Remove(feedbackPath)
	}()
	fmt.Println(editor + " " + feedbackPath)
	cmd := exec.Command(editor, feedbackPath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		fmt.Println("Awwwww")
		return err
	}

	return nil
}

/* func postFeedback(ctx context.Context, client github.Client, fk Feedback) (string, error) {
	teamResp, err := client.Teams.GetTeamBySlug("dev-experience")

	user := client.Users.Get(ctx, "")
	discussion := github.TeamDiscussion{
		Author: user,
		Body:   fk.Body,
	}
	client.Teams.CreateDiscussionBySlug(ctx, "sourcegraph", "dev-experience")

} */
