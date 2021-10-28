package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	secretFlagSet = flag.NewFlagSet("sg secret", flag.ExitOnError)
	secretResetFlagSet = flag.NewFlagSet("sg secret reset", flag.ExitOnError)
	secretListFlagSet = flag.NewFlagSet("sg secret list", flag.ExitOnError)
	secretCommand = &ffcli.Command{
		Name:        "secret",
		ShortUsage:  "sg secret <subcommand>...",
		ShortHelp:   "Manipulate secrets stored in memory and in file",
		FlagSet:     secretFlagSet,
		Subcommands: []*ffcli.Command{
			{
				Name:       "reset",
				ShortUsage: "sg secret reset <key>...",
				ShortHelp:  "Remove key value pair from secrets file",
				FlagSet:    secretResetFlagSet,
				Exec:       resetSecretExec,
			},
			{
				Name:       "list",
				ShortUsage: "sg secret list",
				ShortHelp:  "List all key value pairs from secrets file",
				FlagSet:    secretListFlagSet,
				Exec:       listSecretExec,
			},
		},
		Exec: secretExec,
	}
)

func secretExec(ctx context.Context, args []string) error {
	return flag.ErrHelp
}

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

func listSecretExec(ctx context.Context, args []string) error {
	if err := loadSecrets(); err != nil {
		return err
	}

	for key, value := range(secretsStore.List()) {
		fmt.Printf("%s: %s\n", key, value)
	}

	return nil
}
