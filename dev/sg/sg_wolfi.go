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
		Description: `Build Wolfi packages and images locally, and update base image hashes`,
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
			Usage:     "Build a package locally using a manifest from sourcegraph/wolfi-packages/",
			UsageText: `
Build a Wolfi package locally by running Melange against a provided Melange manifest file, which can be found in sourcegraph/wolfi-packages.

This is convenient for testing package changes locally before pushing to the Wolfi registry.
Base images containing locally-built packages can then be built using 'sg wolfi image'.
`,
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

				defer wolfi.RemoveBuildDir(buildDir)

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
				Usage:     "Build a base image locally using a manifest from sourcegraph/wolfi-images/",
				UsageText: `
Build a base image locally by running apko against a provided apko manifest file, which can be found in sourcegraph/wolfi-images.

Any packages built locally using 'sg wolfi package' can be included in the base image using the 'package@local' syntax in the base image manifest.
This is convenient for testing package changes locally before publishing them.

Once built, the base image is loaded into Docker and can be run locally.
It can also be used for local development by updating its path and hash in the 'dev/oci_deps.bzl' file.
`,
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
				Name:  "update-hashes",
				Usage: "Update Wolfi base images hashes to the latest versions",
				UsageText: `
Update the hash references for all Wolfi base images in the 'dev/oci_deps.bzl' file.

This is done by fetching the ':latest' tag for each base image from the registry, and updating the corresponding hash in 'dev/oci_deps.bzl'.
`,
				Action: wolfi.UpdateHashes,
			}},
	}
)
