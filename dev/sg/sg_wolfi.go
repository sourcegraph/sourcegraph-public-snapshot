package main

import (
	"github.com/urfave/cli/v2"

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

				c, err := wolfi.InitLocalPackageRepo()
				if err != nil {
					return err
				}

				manifestBaseName, buildDir, err := wolfi.SetupPackageBuild(packageName)
				if err != nil {
					return err
				}

				err = c.DoPackageBuild(manifestBaseName, buildDir)
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

					baseImageName := args[0]

					c, err := wolfi.InitLocalPackageRepo()
					if err != nil {
						return err
					}

					manifestBaseName, buildDir, err := c.SetupBaseImageBuild(baseImageName)
					if err != nil {
						return err
					}

					if err = c.DoBaseImageBuild(manifestBaseName, buildDir); err != nil {
						return err
					}

					if err = c.LoadBaseImage(baseImageName); err != nil {
						return err
					}

					if err = c.CleanupBaseImageBuild(baseImageName); err != nil {
						return err
					}

					return nil

				},
			},
			{
				Name:   "update-hashes",
				Usage:  "Update Wolfi dependency digests to the latest version",
				Action: wolfi.UpdateHashes,
			}},
	}
)
