package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	secretListViewFlag bool

	secretCommand = &cli.Command{
		Name:      "secret",
		ArgsUsage: "<...subcommand>",
		Usage:     "Manipulate secrets stored in memory and in file",
		Category:  CategoryEnv,
		Subcommands: []*cli.Command{
			{
				Name:      "reset",
				ArgsUsage: "<...key>",
				Usage:     "Remove a specific secret from secrets file",
				Action:    execAdapter(resetSecretExec),
			},
			{
				Name:  "list",
				Usage: "List all stored secrets",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "view",
						Aliases:     []string{"v"},
						Usage:       "Display configured secrets when listing",
						Value:       false,
						Destination: &secretListViewFlag,
					},
				},
				Action: execAdapter(listSecretExec),
			},
		},
	}
)

func resetSecretExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("no key provided to reset")
	}

	secretsStore, err := secrets.FromContext(ctx)
	if err != nil {
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
	secretsStore, err := secrets.FromContext(ctx)
	if err != nil {
		return err
	}
	std.Out.WriteLine(output.Styled(output.StyleBold, "Secrets:"))
	keys := secretsStore.Keys()
	if secretListViewFlag {
		for _, key := range keys {
			var val map[string]any
			if err := secretsStore.Get(key, &val); err != nil {
				return errors.Newf("Get %q: %w", key, err)
			}
			data, err := json.MarshalIndent(val, "  ", "  ")
			if err != nil {
				return errors.Newf("Marshal %q: %w", key, err)
			}
			std.Out.WriteLine(output.Styledf(output.StyleYellow, "- %s:", key))
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "  %s", string(data)))
		}
	} else {
		std.Out.WriteLine(output.Styled(output.StyleYellow, strings.Join(keys, ", ")))
	}
	return nil
}
