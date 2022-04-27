package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/images"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/docker/docker-credential-helpers/credentials"
)

var (
	opsCommand = &cli.Command{
		Name:        "ops",
		Usage:       "Commands used by operations teams to perform common tasks",
		Description: constructOpsCmdLongHelp(),
		Category:    CategoryCompany,
		Action:      suggestSubcommandsAction,
		Subcommands: []*cli.Command{opsUpdateImagesCommand},
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
				Usage:       "the `kind` of deployment (one of 'k8s', 'helm')",
				Value:       string(images.DeploymentTypeK8S),
				Destination: &opsUpdateImagesDeploymentKindFlag,
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
			&cli.StringFlag{
				Name:        "pin-tag",
				Usage:       "pin all images to a specific sourcegraph `tag` (e.g. 3.36.2, insiders)",
				Destination: &opsUpdateImagesPinTagFlag,
			},
		},
		Action: execAdapter(opsUpdateImage),
	}
)

func constructOpsCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Commands used by operations teams to perform common tasks")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "Supported subcommands")
	fmt.Fprintf(&out, "update-images -> Updates images when run from the root of a 'deploy-sourcegraph-*' repo")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "Supports internal deploy Sourcegraph repos (non-customer facing)")

	return out.String()
}

func opsUpdateImage(ctx context.Context, args []string) error {
	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No path provided"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "Multiple paths not currently supported"))
		return flag.ErrHelp
	}
	dockerCredentials := &credentials.Credentials{
		ServerURL: "https://index.docker.io/v1/",
		Username:  opsUpdateImagesContainerRegistryUsernameFlag,
		Secret:    opsUpdateImagesContainerRegistryPasswordFlag,
	}
	if opsUpdateImagesContainerRegistryUsernameFlag == "" || opsUpdateImagesContainerRegistryPasswordFlag == "" {
		if creds, err := docker.GetCredentialsFromStore(dockerCredentials.ServerURL); err != nil {
			// We do not want any error handling here, just fallback to anonymous requests
			writeWarningLinef("Registry credentials are not provided and could not be retrieved from docker config.")
			writeWarningLinef("You will be using anonymous requests and may be subject to rate limiting by Docker Hub.")
			dockerCredentials.Username = ""
			dockerCredentials.Secret = ""
		} else {
			writeFingerPointingLinef("Using credentials from docker credentials store (learn more https://docs.docker.com/engine/reference/commandline/login/#credentials-store)")
			dockerCredentials = creds
		}
	}

	if opsUpdateImagesPinTagFlag == "" {
		writeWarningLinef("No pin tag is provided.")
		writeWarningLinef("Falling back to the latest deveopment build available.")
	}

	return images.Parse(args[0], *dockerCredentials, images.DeploymentType(opsUpdateImagesDeploymentKindFlag), opsUpdateImagesPinTagFlag)
}
