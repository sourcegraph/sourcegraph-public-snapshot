package cloud

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var UpgradeEphemeralCommand = cli.Command{
	Name:        "upgrade",
	Usage:       "sg could upgrade --name <name> --version <version>",
	Description: "Upgrade the given Ephemeral deployment with the specified version",
	Action:      wipAction(upgradeCloudEphemeral),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			DefaultText: "the name of the ephemeral deployment. If not specified, the name will be derived from the branch name",
		},
		&cli.StringFlag{
			Name:        "version",
			DefaultText: "upgrades an ephemeral cloud Sourcegraph environment with the specified version. The version MUST exist and implies that no build will be created",
			Required:    true,
		},
	},
}

func upgradeDeploymentForVersion(ctx context.Context, email, name, version string) error {
	cloudClient, err := NewClient(ctx, email, APIEndpoint)
	if err != nil {
		return err
	}

	spec := NewDeploymentSpec(
		sanitizeInstanceName(name),
		version,
	)
	cloudEmoji := "☁️"
	pending := std.Out.Pending(output.Linef(cloudEmoji, output.StylePending, "Upgrading deployment %q to version %q", spec.Name, spec.Version))

	// Check if the deployment already exists
	_, err = cloudClient.GetInstance(ctx, spec.Name)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			return ErrInstanceNotFound
		} else {
			return errors.Wrapf(err, "failed to check if instance %q already exists", spec.Name)
		}
	}
	inst, err := cloudClient.UpgradeInstance(ctx, spec)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Upgrade of %q failed", spec.Name))
		return errors.Wrapf(err, "failed to issue upgrade of %q to version %s", spec.Name, spec.Version)
	}

	pending.Writef("Instance details: \n%s", inst.String())
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Upgrade of %q to version %q has been triggered - access at: %s", spec.Name, spec.Version, inst.URL))
	return nil
}

func upgradeCloudEphemeral(ctx *cli.Context) error {
	version := ctx.String("version")
	if version == "" {
		return errors.New("version is required")
	} else if err := checkVersionExistsInRegistry(ctx.Context, version); err != nil {
		return err
	}
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}
	// if a version is specified we do not build anything and just trigger the cloud deployment
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		return err
	}
	deploymentName := createDeploymentName(ctx.String("name"), version, email, currentBranch)

	err = upgradeDeploymentForVersion(ctx.Context, email, deploymentName, version)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			std.Out.WriteWarningf("Unable to upgrade %q since no deployment like that exists", deploymentName)
			std.Out.WriteMarkdown("You can check what deployments exist under your GCP account with `sg cloud list`, or you can see all deployments with `sg cloud list --all`")
		}
		return err
	}
	return nil
}
