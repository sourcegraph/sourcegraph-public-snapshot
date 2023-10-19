package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/images"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var opsCommand = &cli.Command{
	Name:        "ops",
	Usage:       "Commands used by operations teams to perform common tasks",
	Description: "Supports internal deploy-sourcegraph repos (non-customer facing)",
	Category:    category.Company,
	Subcommands: []*cli.Command{
		opsTagDetailsCommand,
		OpsUpdateImagesCommand,
	},
}

var OpsUpdateImagesCommand = &cli.Command{
	Name:      "update-images",
	Usage:     "Update images across a sourcegraph/deploy-sourcegraph/* manifests",
	ArgsUsage: "<dir>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "kind",
			Aliases: []string{"k"},
			Usage:   "the `kind` of deployment (one of 'k8s', 'helm', 'compose')",
			Value:   string(images.DeploymentTypeK8S),
		},
		&cli.StringFlag{
			Name:    "pin-tag",
			Aliases: []string{"t"},
			Usage:   "pin all images to a specific sourcegraph `tag` (e.g. '3.36.2', 'insiders') (default: latest main branch tag)",
		},
		&cli.StringFlag{
			Name:    "docker-username",
			Aliases: []string{"cr-username"}, // deprecated
			Usage:   "dockerhub username",
		},
		&cli.StringFlag{
			Name:    "docker-password",
			Aliases: []string{"cr-password"}, // deprecated
			Usage:   "dockerhub password",
		},
		&cli.StringFlag{
			Name:  "registry",
			Usage: "Sets the registry we want images to update to, public or internal.",
			Value: "public",
		},
		&cli.StringFlag{
			Name:    "skip",
			Aliases: []string{"skip-images"}, // deprecated
			Usage:   "List of comma separated images to skip updating, ex: --skip 'gitserver,indexed-server'",
		},
	},
	Action: func(ctx *cli.Context) error {
		// Ensure args are correct.
		args := ctx.Args().Slice()
		if len(args) == 0 {
			std.Out.WriteLine(output.Styled(output.StyleWarning, "No path provided"))
			return flag.ErrHelp
		}
		if len(args) != 1 {
			std.Out.WriteLine(output.Styled(output.StyleWarning, "Multiple paths not currently supported"))
			return flag.ErrHelp
		}

		return opsUpdateImages(
			ctx.Context,
			args[0],
			ctx.String("registry"),
			ctx.String("kind"),
			ctx.String("pin-tag"),
			ctx.String("docker-username"),
			ctx.String("docker-password"),
			strings.Split(ctx.String("skip"), ","),
		)
	},
}

var opsTagDetailsCommand = &cli.Command{
	Name:      "inspect-tag",
	ArgsUsage: "<image|tag>",
	Usage:     "Inspect main branch tag details from a image or tag",
	UsageText: `
# Inspect a full image
sg ops inspect-tag index.docker.io/sourcegraph/cadvisor:159625_2022-07-11_225c8ae162cc@sha256:foobar

# Inspect just the tag
sg ops inspect-tag 159625_2022-07-11_225c8ae162cc

# Get the build number
sg ops inspect-tag -p build 159625_2022-07-11_225c8ae162cc
`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "property",
			Aliases: []string{"p"},
			Usage:   "only output a specific `property` (one of: 'build', 'date', 'commit')",
		},
	},
	Action: func(cmd *cli.Context) error {
		input := cmd.Args().First()
		// trim out leading image
		parts := strings.SplitN(input, ":", 2)
		if len(parts) > 1 {
			input = parts[1]
		}
		// trim out shasum
		parts = strings.SplitN(input, "@sha256", 2)
		if len(parts) > 1 {
			input = parts[0]
		}

		std.Out.Verbosef("inspecting %q", input)

		tag, err := images.ParseMainBranchImageTag(input)
		if err != nil {
			return errors.Wrap(err, "unable to parse tag")
		}

		selectProperty := cmd.String("property")
		if len(selectProperty) == 0 {
			std.Out.WriteMarkdown(fmt.Sprintf("# %s\n- Build: `%d`\n- Date: %s\n- Commit: `%s`", input, tag.Build, tag.Date, tag.ShortCommit))
			return nil
		}

		properties := map[string]string{
			"build":  strconv.Itoa(tag.Build),
			"date":   tag.Date,
			"commit": tag.ShortCommit,
		}
		v, exists := properties[selectProperty]
		if !exists {
			return errors.Newf("unknown property %q", selectProperty)
		}
		std.Out.Write(v)
		return nil
	},
}

func opsUpdateImages(
	ctx context.Context,
	path string,
	registryType string,
	deploymentType string,
	pinTag string,
	dockerUsername string,
	dockerPassword string,
	skipImages []string,
) error {
	{
		// Select the registry we're going to work with.
		var registry images.Registry
		switch registryType {
		case "internal":
			gcr := images.NewGCR("us.gcr.io", "sourcegraph-dev")
			if err := gcr.LoadToken(); err != nil {
				return err
			}
			registry = gcr
		case "public":
			registry = images.NewDockerHub("sourcegraph", dockerUsername, dockerPassword)
		default:
			std.Out.WriteLine(output.Styled(output.StyleWarning, "Registry is either 'internal' or 'public'"))
		}

		// Select the type of operation we're performing.
		var op images.UpdateOperation
		// Keep track of the tags we updated, they should all be the same one after performing the update.
		foundTags := []string{}

		shouldSkip := func(r *images.Repository) bool {
			for _, img := range skipImages {
				if r.Name() == img {
					return true
				}
			}
			return false
		}

		if pinTag != "" {
			std.Out.WriteNoticef("pinning images to tag %q", pinTag)
			// We're pinning a tag.
			op = func(registry images.Registry, r *images.Repository) (*images.Repository, error) {
				if !images.IsSourcegraph(r) || shouldSkip(r) {
					return nil, images.ErrNoUpdateNeeded
				}

				newR, err := registry.GetByTag(r.Name(), pinTag)
				if err != nil {
					return nil, err
				}
				announce(r.Name(), r.Ref(), newR.Ref())
				return newR, nil
			}
		} else {
			std.Out.WriteNoticef("updating images to latest")
			// We're updating to the latest found tag.
			op = func(registry images.Registry, r *images.Repository) (*images.Repository, error) {
				if !images.IsSourcegraph(r) || shouldSkip(r) {
					return nil, images.ErrNoUpdateNeeded
				}

				newR, err := registry.GetLatest(r.Name(), images.FindLatestMainTag)
				if err != nil {
					return nil, err
				}
				// store this new tag we found for further inspection.
				foundTags = append(foundTags, newR.Tag())
				announce(r.Name(), r.Ref(), newR.Ref())
				return newR, nil
			}
		}

		// Apply the updates.
		switch images.DeploymentType(deploymentType) {
		case images.DeploymentTypeK8S:
			if err := images.UpdateK8sManifest(ctx, registry, path, op); err != nil {
				return err
			}
		case images.DeploymentTypeHelm:
			if err := images.UpdateHelmManifest(ctx, registry, path, op); err != nil {
				return err
			}
		case images.DeploymentTypeCompose:
			if err := images.UpdateComposeManifests(ctx, registry, path, op); err != nil {
				return err
			}
		}

		// Ensure the updates were correct.
		if len(foundTags) > 0 {
			t := foundTags[0]
			for _, tag := range foundTags {
				if tag != t {
					std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("expected all tags to be the same after updating, but found %q != %q\nTree left intact for inspection", t, tag)))
					return errors.New("tag mistmatch detected")
				}
			}
		}

		return nil
	}
}

func announce(name string, before string, after string) {
	std.Out.Writef("Updated %s", name)
	std.Out.Writef("  - %s", before)
	std.Out.Writef("  + %s", after)
}
