package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/internal/team"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/slack"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
)

var (
	teammateFlagSet = flag.NewFlagSet("sg teammate", flag.ExitOnError)
	teammateCommand = &ffcli.Command{
		Name:       "teammate",
		ShortUsage: "sg teammate [time|handbook] nickname",
		ShortHelp:  "Run the given teammates command show infos about your teammates.",
		LongHelp:   `Display current time, handbook link of sourcegraphers.`,
		FlagSet:    teammateFlagSet,
		Exec:       teammateExec,
	}
)

func teammateExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	slackClient, err := slack.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("slack.NewClient: %w", err)
	}
	resolver := team.NewTeammateResolver(nil, slackClient)

	switch args[0] {
	case "time":
		if len(args) < 2 {
			return errors.New("no nickname given")
		}
		teammate, err := resolver.ResolveByName(ctx, strings.Join(args[1:], " "))
		if err != nil {
			return err
		}
		stdout.Out.Writef("%s's current time is %s",
			teammate.Name, timeAtLocation(teammate.SlackTimezone))
		return nil
	case "handbook":
		if len(args) < 2 {
			return errors.New("no nickname given")
		}
		teammate, err := resolver.ResolveByName(ctx, strings.Join(args[1:], " "))
		if err != nil {
			return err
		}
		stdout.Out.Writef("Opening handbook link for %s: %s", teammate.Name, teammate.HandbookLink)
		return open.URL(teammate.HandbookLink)
	default:
		return flag.ErrHelp
	}
}

func timeAtLocation(loc *time.Location) string {
	t := time.Now().In(loc)
	t2 := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	diff := t2.Sub(t) / time.Hour
	return fmt.Sprintf("%s (%dh from your local time)", t.Format(time.RFC822), diff)
}
