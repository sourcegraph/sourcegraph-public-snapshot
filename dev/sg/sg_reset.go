package main

import (
	"context"
	"errors"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	resetFlagSet = flag.NewFlagSet("sg reset", flag.ExitOnError)
	resetCommand = &ffcli.Command{
		Name:       "reset",
		ShortUsage: "sg reset <key>...",
		ShortHelp:  "Resets the secrets for a given key",
		FlagSet:    resetFlagSet,
		Exec:       resetExec,
	}
)

func resetExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("no key given")
	}

	if err := loadSecrets(); err != nil {
		return err
	}

	for _, arg := range args {
		return secretsStore.RemoveAndSave(arg)
	}

	return nil
}
