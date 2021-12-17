package main

import (
	"context"
	"flag"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

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
	switch args[0] {
	case "time":
		if len(args) < 2 {
			return errors.New("no nickname given")
		}
		str, err := slack.QueryUserCurrentTime(ctx, strings.Join(args[1:], " "))
		if err != nil {
			return err
		}
		stdout.Out.Writef(str)
		return nil
	case "handbook":
		if len(args) < 2 {
			return errors.New("no nickname given")
		}
		str, err := slack.QueryUserHandbook(ctx, strings.Join(args[1:], " "))
		if err != nil {
			return err
		}
		open.URL(str)
		return nil
	default:
		return flag.ErrHelp
	}
}
