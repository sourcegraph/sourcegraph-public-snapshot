package main

import (
	"context"
	"flag"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	secretFlagSet      = flag.NewFlagSet("sg secret", flag.ExitOnError)
	secretResetFlagSet = flag.NewFlagSet("sg secret reset", flag.ExitOnError)
	secretListFlagSet  = flag.NewFlagSet("sg secret list", flag.ExitOnError)
	secretCommand      = &ffcli.Command{
		Name:       "secret",
		ShortUsage: "sg secret <subcommand>...",
		ShortHelp:  "Manipulate secrets stored in memory and in file",
		FlagSet:    secretFlagSet,
		Subcommands: []*ffcli.Command{
			{
				Name:       "reset",
				ShortUsage: "sg secret reset <key>...",
				ShortHelp:  "Remove a specific secret from secrets file",
				FlagSet:    secretResetFlagSet,
				Exec:       resetSecretExec,
			},
			{
				Name:       "list",
				ShortUsage: "sg secret list",
				ShortHelp:  "List all stored secrets",
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

	for _, arg := range args {
		if err := secretsStore.Remove(arg); err != nil {
			return err
		}
	}
	if err := secretsStore.SaveFile(); err != nil {
		return err
	}

	return nil
}

func listSecretExec(ctx context.Context, args []string) error {
	if err := loadSecrets(); err != nil {
		return err
	}
	stdout.Out.WriteLine(output.Linef("", output.StyleBold, "Secrets:"))
	keys := secretsStore.Keys()
	stdout.Out.WriteLine(output.Linef("", output.StyleWarning, strings.Join(keys, ", ")))
	return nil
}
