package main

import (
	"context"
	"flag"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/slack"
)

var (
	teammatesFlagSet = flag.NewFlagSet("sg teammates", flag.ExitOnError)
	teammatesCommand = &ffcli.Command{
		Name:       "teammates",
		ShortUsage: "sg teammates [time|handbook]",
		ShortHelp:  "Run the given teammates command show informations about teammates",
		LongHelp:   `Display current time, handbook link of sourcegraphers`,
		FlagSet:    teammatesFlagSet,
		Exec:       teammatesExec,
	}
)

func teammatesExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("no teammates command given")
	}
	switch args[0] {
	case "time":
		if len(args) != 2 {
			return errors.New("no nickname given")
		}
		str, err := slack.QueryUserCurrentTime(ctx, args[1])
		if err != nil {
			return err
		}
		out.Writef(str)
		return nil
	case "handbook":
		if len(args) != 2 {
			return errors.New("no nickname given")
		}
		str, err := slack.QueryUserHandbook(ctx, args[1])
		if err != nil {
			return err
		}
		open.URL(str)
		return nil
	default:
		return nil
	}
}
