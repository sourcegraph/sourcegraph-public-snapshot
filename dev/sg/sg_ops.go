package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/run"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/images"
	sgrun "github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/docker/docker-credential-helpers/credentials"
)

var (
	opsCommand = &cli.Command{
		Name:        "ops",
		Usage:       "Commands used by operations teams to perform common tasks",
		Description: "Supports internal deploy-sourcegraph repos (non-customer facing)",
		Category:    CategoryCompany,
		Subcommands: []*cli.Command{
			opsUpdateImagesCommand,
			opsTagDetailsCommand,
		},
	}

	opsUpdateImagesDeploymentKindFlag            string
	opsUpdateImagesContainerRegistryUsernameFlag string
	opsUpdateImagesContainerRegistryPasswordFlag string
	opsUpdateImagesPinTagFlag                    string
	opsUpdateImagesCommand                       = &cli.Command{
		Name:        "update-images",
		ArgsUsage:   "<dir>",
		Usage:       "Updates images in given directory to latest published image",
		Description: "Updates images in given directory to latest published image.\nEx: in deploy-sourcegraph-cloud, run `sg ops update-images base/.`",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "kind",
				Aliases:     []string{"k"},
				Usage:       "the `kind` of deployment (one of 'k8s', 'helm', 'compose')",
				Value:       string(images.DeploymentTypeK8S),
				Destination: &opsUpdateImagesDeploymentKindFlag,
			},
			&cli.StringFlag{
				Name:        "pin-tag",
				Aliases:     []string{"t"},
				Usage:       "pin all images to a specific sourcegraph `tag` (e.g. '3.36.2', 'insiders') (default: latest main branch tag)",
				Destination: &opsUpdateImagesPinTagFlag,
			},
			&cli.StringFlag{
				Name:        "cr-username",
				Usage:       "`username` for the container registry",
				Destination: &opsUpdateImagesContainerRegistryUsernameFlag,
			},
			&cli.StringFlag{
				Name:        "cr-password",
				Usage:       "`password` or access token for the container registry",
				Destination: &opsUpdateImagesContainerRegistryPasswordFlag,
			},
		},
		Action: opsUpdateImage,
	}

	opsDeployImagesDeploymentKindFlag string
	opsDeployImagesNamespaceFlag      string
	opsDeployImagesCommand            = &cli.Command{
		Name:        "deploy-images",
		Usage:       "<helm chart name> <helm repository name> <path to values file>",
		Description: "",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "kind",
				Aliases:     []string{"k"},
				Usage:       "the `kind` of deployment (currently only 'helm' is supported)",
				Value:       string(images.DeploymentTypeHelm),
				Destination: &opsDeployImagesDeploymentKindFlag,
			},
			&cli.StringFlag{
				Name:        "namespace",
				Aliases:     []string{"n"},
				Usage:       "the namespace in which to deploy",
				Destination: &opsDeployImagesNamespaceFlag,
				Required:    true,
			},
		},
		Action: opsDeployImage,
	}

	opsTagDetailsCommand = &cli.Command{
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
				return errors.Wrap(err, "unable to understand tag")
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
)

func opsUpdateImage(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No path provided"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "Multiple paths not currently supported"))
		return flag.ErrHelp
	}
	dockerCredentials := &credentials.Credentials{
		ServerURL: "https://registry.hub.docker.com/v2",
		Username:  opsUpdateImagesContainerRegistryUsernameFlag,
		Secret:    opsUpdateImagesContainerRegistryPasswordFlag,
	}
	if opsUpdateImagesContainerRegistryUsernameFlag == "" || opsUpdateImagesContainerRegistryPasswordFlag == "" {
		if creds, err := docker.GetCredentialsFromStore(dockerCredentials.ServerURL); err != nil {
			// We do not want any error handling here, just fallback to anonymous requests
			std.Out.WriteWarningf("Registry credentials are not provided and could not be retrieved from docker config.")
			std.Out.WriteWarningf("You will be using anonymous requests and may be subject to rate limiting by Docker Hub.")
			dockerCredentials.Username = ""
			dockerCredentials.Secret = ""
		} else {
			std.Out.WriteSuccessf("Using credentials from docker credentials store (learn more https://docs.docker.com/engine/reference/commandline/login/#credentials-store)")
			dockerCredentials = creds
		}
	}

	if opsUpdateImagesPinTagFlag == "" {
		std.Out.WriteWarningf("No pin tag (-t) is provided - will fall back to latest main branch tag available.")
	}

	return images.Update(args[0], *dockerCredentials, images.DeploymentType(opsUpdateImagesDeploymentKindFlag), opsUpdateImagesPinTagFlag)
}

func opsDeployImage(ctx *cli.Context) error {
	if opsDeployImagesDeploymentKindFlag != images.DeploymentTypeHelm {
		return errors.New("")
	}
	args := ctx.Args().Slice()
	if len(args) != 3 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "Unexpected amount of arguments provided. Expected: <helm chart name> <helm repository name> <path to values file>"))
		return flag.ErrHelp
	}

	currentVersionCmd := run.Cmd(ctx.Context, fmt.Sprintf("helm list --filter '%s' -o json", args[0]))
	out, err := currentVersionCmd.Run().JQ(".[].chart")
	if err != nil {
		return err
	}
	version := strings.ReplaceAll(string(out), fmt.Sprintf("%s-", args[0]), "")

	defaultBranch, err := sgrun.TrimResult(sgrun.GitCmd("rev-parse", "--abbrev-ref", "origin/HEAD"))
	defaultBranch = strings.ReplaceAll(defaultBranch, "origin/", "")
	if err != nil {
		return err
	}
	branch, err := sgrun.TrimResult(sgrun.GitCmd("branch", "--show-current"))
	if err != nil {
		return err
	}
	if branch != defaultBranch {
		return errors.Newf("Can only deploy images from the default branch '%s'", defaultBranch)
	}

	helmCmdString := fmt.Sprintf("helm upgrade --install --values %s --version %s %s %s -n %s", args[2], version, args[0], fmt.Sprintf("%s/%s", args[0], args[1], opsDeployImagesNamespaceFlag))
	helmUpgradeCmd, err := run.Cmd(ctx.Context, ).Run().Lines()
	if err != nil {
		return err
	}

}
