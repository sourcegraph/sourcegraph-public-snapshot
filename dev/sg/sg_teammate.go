package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/slack"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getTeamResolver(ctx context.Context) (team.TeammateResolver, error) {
	slackClient, err := slack.NewClient(ctx, std.Out)
	if err != nil {
		return nil, errors.Newf("slack.NewClient: %w", err)
	}
	githubClient := github.NewClient(http.DefaultClient)
	return team.NewTeammateResolver(githubClient, slackClient), nil
}

var markdownFmt = `**Name**       : %s
**Local Time** : %s
**Slack**      : %s
**Email**      : %s
**GitHub**     : %s
**Location**   : %s
**Role**       : %s
**Description**: %s`

var (
	teammateCommand = &cli.Command{
		Name:     "teammate",
		Usage:    "Get information about Sourcegraph teammates",
		Category: category.Company,
		Action: func(ctx *cli.Context) error {
			args := ctx.Args().Slice()
			if len(args) == 0 {
				return errors.New("no nickname provided")
			}
			handle := strings.Join(args, " ")
			resolver, err := getTeamResolver(ctx.Context)
			if err != nil {
				return err
			}
			teammate, err := resolver.ResolveByName(ctx.Context, handle)
			if err != nil {
				return err
			}

			markdown := fmt.Sprintf(markdownFmt,
				teammate.Name,
				timeAtLocation(teammate.SlackTimezone),
				teammate.SlackName,
				teammate.Email,
				fmt.Sprintf("[%s](https://github.com/%s)", teammate.GitHub, teammate.GitHub),
				teammate.Location,
				teammate.Role,
				teammate.Description,
			)
			return std.Out.WriteMarkdown(markdown)
		},
	}
)

func timeAtLocation(loc *time.Location) string {
	t := time.Now().In(loc)
	t2 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	diff := t2.Sub(t) / time.Hour
	return fmt.Sprintf("%s (%dh from you)", t.Format(time.RFC822), diff)
}
