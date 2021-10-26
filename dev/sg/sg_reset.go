package main

import (
	"context"
	"errors"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	secretFlagSet = flag.NewFlagSet("sg secret", flag.ExitOnError)
	resetFlagSet = flag.NewFlagSet("sg secret reset", flag.ExitOnError)
	resetSubcommand = &ffcli.Command{
		Name:       "reset",
		ShortUsage: "sg secret reset <key>...",
		ShortHelp:  "Remove key value pair from secrets file",
		FlagSet:    resetFlagSet,
		Exec:       resetSecretExec,
	}
	secretCommand = &ffcli.Command{
		Name:        "secret",
		ShortUsage:  "sg secret <subcommand>...",
		ShortHelp:   "Manipulate secrets stored in memory and in file",
		FlagSet:     secretFlagSet,
		Subcommands: []*ffcli.Command{resetSubcommand},
	}
)

func resetSecretExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("no key provided to reset")
	}

	if err := loadSecrets(); err != nil {
		return err
	}

	for _, arg := range(args) {
		if err := secretsStore.RemoveAndSave(arg); err != nil {
			return err
		}
	}

	return nil
}
