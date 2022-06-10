package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	secretListViewFlag bool

	secretCommand = &cli.Command{
		Name:  "secret",
		Usage: "Manipulate secrets stored in memory and in file",
		UsageText: `
# List all secrets stored in your local configuration.
sg secret list

# Remove the secrets associated with buildkite (sg ci build)
sg secret reset buildkite
`,
		Category: CategoryEnv,
		Subcommands: []*cli.Command{
			{
				Name:      "download-config",
				ArgsUsage: "",
				Usage:     "TODO",
				Action:    downloadDevPrivate,
			},
			{
				Name:      "reset",
				ArgsUsage: "<...key>",
				Usage:     "Remove a specific secret from secrets file",
				Action:    resetSecretExec,
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

var (
	siteConfig = secrets.ExternalSecret{
		Provider: secrets.ExternalProvider1Pass,
		Project:  "Shared",
		Name:     "DevPrivate",
		Field:    "site-config.json",
	}

	externalServicesConfig = secrets.ExternalSecret{
		Provider: secrets.ExternalProvider1Pass,
		Project:  "Shared",
		Name:     "DevPrivate",
		Field:    "external-services-config.json",
	}
)

func downloadDevPrivate(ctx *cli.Context) error {
	store, err := secrets.FromContext(ctx.Context)
	if err != nil {
		return err
	}

	root, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	s, err := store.GetExternal(ctx.Context, siteConfig)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(root, "enterprise", "dev", "site-config.json"), []byte(s), 0600)
	if err != nil {
		return err
	}
	s, err = store.GetExternal(ctx.Context, externalServicesConfig)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(root, "enterprise", "dev", "external-services-config.json"), []byte(s), 0600)
	if err != nil {
		return err
	}

	return nil
}

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

	return nil
}

func listSecretExec(ctx *cli.Context) error {
	secretsStore, err := secrets.FromContext(ctx.Context)
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
