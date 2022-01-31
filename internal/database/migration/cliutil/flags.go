package cliutil

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type RunFunc func(ctx context.Context, options runner.Options) error

func Flags(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	rootFlagSet := flag.NewFlagSet(commandName, flag.ExitOnError)

	return &ffcli.Command{
		Name:       commandName,
		ShortUsage: fmt.Sprintf("%s <command>", commandName),
		ShortHelp:  "Modifies and runs database migrations",
		FlagSet:    rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			Up(commandName, run, out),
			UpTo(commandName, run, out),
			Undo(commandName, run, out),
			DownTo(commandName, run, out),
		},
	}
}
