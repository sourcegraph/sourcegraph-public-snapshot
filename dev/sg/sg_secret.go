package main

import (
	"encoding/json"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	secretListViewFlag bool

	secretCommand = &cli.Command{
		Name:  "secret",
		Usage: "Manipulate secrets stored in memory and in file",
		UsageText: `# List all secrets stored in your local configuration.
sg secret list

# Remove the secrets associated with buildkite (sg ci build) - (supports bash autocompletion).
sg secret reset buildkite
`,
		Category: category.Env,
		Subcommands: []*cli.Command{
			{
				Name:         "reset",
				ArgsUsage:    "key1 key2 ...",
				Usage:        "Remove a individual secret(s) from secrets file (see 'list' for getting the keys)",
				Action:       resetSecretExec,
				BashComplete: completions.CompleteArgs(bashCompleteSecrets),
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
				Action: listSecretExec,
			},
		},
	}
)

func resetSecretExec(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return errors.New("no key provided to reset")
	}

	secretsStore, err := secrets.FromContext(ctx.Context)
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

	std.Out.WriteSuccessf("Removed secret(s) %s.", strings.Join(args, ", "))

	return nil
}

func listSecretExec(ctx *cli.Context) error {
	secretsStore, err := secrets.FromContext(ctx.Context)
	if err != nil {
		return err
	}
	std.Out.WriteLine(output.Styled(output.StyleBold, "Secrets:"))
	keys := secretsStore.Keys()
	for _, key := range keys {
		std.Out.WriteLine(output.Styledf(output.StyleYellow, "- %s", key))

		// If we are just rendering the secret name, we are done
		if !secretListViewFlag {
			continue
		}

		// Otherwise render value
		var val map[string]any
		if err := secretsStore.Get(key, &val); err != nil {
			return errors.Newf("Get %q: %w", key, err)
		}
		data, err := json.MarshalIndent(val, "  ", "  ")
		if err != nil {
			return errors.Newf("Marshal %q: %w", key, err)
		}
		std.Out.WriteCode("json", "  "+string(data))
	}
	return nil
}

func bashCompleteSecrets() (options []string) {
	allSecrets, err := loadSecrets()
	if err != nil {
		return nil
	}
	return allSecrets.Keys()
}
