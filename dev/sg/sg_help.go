package main

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/docgen"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const generatedSgReferenceHeader = "<!-- DO NOT EDIT: generated via: go generate ./dev/sg -->"

var helpCommand = &cli.Command{
	Name:            "help",
	ArgsUsage:       " ", // no args accepted for now
	Usage:           "Get help and docs about sg",
	Category:        category.Util,
	HideHelpCommand: true, // we don't want a "sg help help" :)
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "full",
			Aliases: []string{"f"},
			Usage:   "generate full markdown sg reference",
		},
		&cli.StringFlag{
			Name:      "output",
			TakesFile: true,
			Usage:     "write reference to `file`",
		},
	},
	Action: func(cmd *cli.Context) error {
		if cmd.NArg() != 0 {
			return errors.Newf("unexpected argument %s", cmd.Args().First())
		}
		if !cmd.IsSet("full") && !cmd.IsSet("output") {
			cli.ShowAppHelp(cmd)
			return nil
		}

		var doc string
		var err error
		if cmd.Bool("full") {
			doc, err = docgen.Markdown(cmd.App)
		} else {
			doc, err = docgen.Default(cmd.App)
		}
		if err != nil {
			return err
		}

		if output := cmd.String("output"); output != "" {
			rootDir, err := root.RepositoryRoot()
			if err != nil {
				return err
			}
			output = filepath.Join(rootDir, output)

			if err := os.WriteFile(output, []byte(generatedSgReferenceHeader+"\n\n"+doc), 0644); err != nil {
				return errors.Wrapf(err, "failed to write reference to %q", output)
			}
			return nil
		}

		return std.Out.WriteMarkdown(doc)
	},
}
