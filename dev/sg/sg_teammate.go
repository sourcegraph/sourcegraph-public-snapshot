package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/slack"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getTeamResolver(ctx context.Context) (team.TeammateResolver, error) {
	slackClient, err := slack.NewClient(ctx)
	if err != nil {
		return nil, errors.Newf("slack.NewClient: %w", err)
	}
	githubClient := github.NewClient(http.DefaultClient)
	return team.NewTeammateResolver(githubClient, slackClient), nil
}

var (
	teammateFlagSet = flag.NewFlagSet("sg teammate", flag.ExitOnError)
	teammateCommand = &ffcli.Command{
		Name:       "teammate",
		ShortUsage: "sg teammate [time|handbook] <nickname>",
		ShortHelp:  "Get information about Sourcegraph teammates.",
		LongHelp:   `Get information about Sourcegraph teammates, such as their current time and handbook page!`,
		FlagSet:    teammateFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{{
			Name:       "time",
			ShortUsage: "sg teammate time <nickname>",
			ShortHelp:  "Get the current time of a Sourcegraph teammate.",
			Exec: func(ctx context.Context, args []string) error {
				if len(args) == 0 {
					return errors.New("no nickname provided")
				}
				resolver, err := getTeamResolver(ctx)
				if err != nil {
					return err
				}
				teammate, err := resolver.ResolveByName(ctx, strings.Join(args, " "))
				if err != nil {
					return err
				}
				stdout.Out.Writef("%s's current time is %s",
					teammate.Name, timeAtLocation(teammate.SlackTimezone))
				return nil
			},
		}, {
			Name:       "handbook",
			ShortUsage: "sg teammate handbook <nickname>",
			ShortHelp:  "Open the handbook page of a Sourcegraph teammate.",
			Exec: func(ctx context.Context, args []string) error {
				if len(args) == 0 {
					return errors.New("no nickname provided")
				}
				resolver, err := getTeamResolver(ctx)
				if err != nil {
					return err
				}
				teammate, err := resolver.ResolveByName(ctx, strings.Join(args, " "))
				if err != nil {
					return err
				}
				stdout.Out.Writef("Opening handbook link for %s: %s", teammate.Name, teammate.HandbookLink)
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
