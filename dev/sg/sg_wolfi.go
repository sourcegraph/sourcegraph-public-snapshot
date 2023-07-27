package main

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/wolfi"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	wolfiCommand = &cli.Command{
		Name:        "wolfi",
		Usage:       "Automate Wolfi related tasks",
		Description: `Ideal for iterating quickly on new packages and base images, rather than relying on Continuous Integration`,
		UsageText: `
# Update base image hashes
sg wolfi update-hashes
# Build a specific package using a manifest from wolfi-packages/
sg wolfi package jaeger
sg wolfi package jaeger.yaml
# Build a base image using a manifest from wolfi-images/
sg wolfi image gitserver
sg wolfi image gitserver.yaml
`,
		Category: CategoryDev,
		Subcommands: []*cli.Command{{
			Name:      "package",
			ArgsUsage: "<package-manifest>",
			Usage:     "Build a package using a manifest from wolfi-packages/",
			Action: func(ctx *cli.Context) error {
				args := ctx.Args().Slice()
				if len(args) == 0 {
					return errors.New("no package manifest file provided")
				}
				packageName := args[0]

				// Set up package repo + keypair
				// TODO: Get location of sourcegraph directory
				c, err := wolfi.InitLocalPackageRepo()
				if err != nil {
					return err
				}

				// TODO: Sanitise .yaml input
				// TODO: Check file exists + copy to tempdir
				// TODO: Run docker command
				buildDir, err := wolfi.SetupPackageBuild(packageName)
				if err != nil {
					return err
				}

				err = c.DoPackageBuild(packageName, buildDir)
				if err != nil {
					return err
				}

				return nil
			},
		},
			{
				Name:      "image",
				ArgsUsage: "<base-image-manifest>",
				Usage:     "Build a base image using a manifest from wolfi-images/",
				Action: func(ctx *cli.Context) error {
					args := ctx.Args().Slice()
					if len(args) == 0 {
						return errors.New("no base image manifest file provided")
					}

					resolver, err := getTeamResolver(ctx.Context)
					if err != nil {
						return err
					}
					teammate, err := resolver.ResolveByName(ctx.Context, strings.Join(args, " "))
					if err != nil {
						return err
					}
					std.Out.Writef("Opening handbook link for %s: %s", teammate.Name, teammate.HandbookLink)
					return open.URL(teammate.HandbookLink)
				},
			},
			{
				Name:   "update-hashes",
				Usage:  "Update Wolfi dependency digests to the latest version",
				Action: wolfi.UpdateHashes,
			}},
	}
)
