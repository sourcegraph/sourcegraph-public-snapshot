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
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
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

var (
	teammateCommand = &cli.Command{
		Name:        "teammate",
		Usage:       "Get information about Sourcegraph teammates",
		Description: `For example, you can check a teammate's current time and find their handbook bio!`,
		UsageText: `
# Get the current time of a team mate based on their slack handle (case insensitive).
sg teammate time @dax
sg teammate time dax
# or their full name (case insensitive)
sg teammate time thorsten ball

# Open their handbook bio
sg teammate handbook asdine
`,
		Category: category.Company,
		Subcommands: []*cli.Command{{
			Name:      "time",
			ArgsUsage: "<nickname>",
			Usage:     "Get the current time of a Sourcegraph teammate",
			Action: func(ctx *cli.Context) error {
				args := ctx.Args().Slice()
				if len(args) == 0 {
					return errors.New("no nickname provided")
				}
				resolver, err := getTeamResolver(ctx.Context)
				if err != nil {
					return err
				}
				teammate, err := resolver.ResolveByName(ctx.Context, strings.Join(args, " "))
				if err != nil {
					return err
				}
				std.Out.Writef("%s's current time is %s",
					teammate.Name, timeAtLocation(teammate.SlackTimezone))
				return nil
			},
		}, {
			Name:      "handbook",
			ArgsUsage: "<nickname>",
			Usage:     "Open the handbook page of a Sourcegraph teammate",
			Action: func(ctx *cli.Context) error {
				args := ctx.Args().Slice()
				if len(args) == 0 {
					return errors.New("no nickname provided")
				}
				resolver, err := getTeamResolver(ctx.Context)
				if err != nil {
					return err
				}
				teammate, err := resolver.ResolveByName(ctx.Context, strings.Join(args, " "))
				if err != nil {
					return err
				}
				std.Out.Writef("Opening handbook link for %s: %s", teammate.Name, teammate.HandbookLink)
				return open.URL(teammate.HandbookLink)
			},
		}},
	}
)

func timeAtLocation(loc *time.Location) string {
	t := time.Now().In(loc)
	t2 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	diff := t2.Sub(t) / time.Hour
	return fmt.Sprintf("%s (%dh from your local time)", t.Format(time.RFC822), diff)
}
