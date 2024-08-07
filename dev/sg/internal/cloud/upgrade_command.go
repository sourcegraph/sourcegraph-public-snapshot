package cloud

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var upgradeEphemeralCommand = cli.Command{
	Name:        "upgrade",
	Usage:       "upgrade an ephemeral instance",
	Description: "upgrade the given ephemeral instance  with the specified version",
	Action:      upgradeCloudEphemeral,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			Usage:       "name of the ephemeral instance",
			DefaultText: "current branch name will be used",
		},
		&cli.StringFlag{
			Name:        "version",
			DefaultText: "upgrades an ephemeral instance with the specified version. The version MUST exist in the cloud ephemeral registry  and implies that no build will be created",
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
		name,
		version,
		"", // we don't need a license during upgrade
	)
	cloudEmoji := "☁️"
	pending := std.Out.Pending(output.Linef(cloudEmoji, output.StylePending, "Upgrading deployment %q to version %q", spec.Name, spec.Version))

	// Check if the deployment already exists
	inst, err := cloudClient.GetInstance(ctx, spec.Name)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			return ErrInstanceNotFound
		} else {
			return errors.Wrapf(err, "failed to check if instance %q already exists", spec.Name)
		}
	}
	// Do various checks before upgrading the instance
	if !inst.HasStatus(InstanceStatusCompleted) {
		std.Out.WriteWarningf("Cannot upgrade instance with status %q - if this issue persists, please reach out to #discuss-dev-infra", inst.Status.Status)
		return ErrInstanceStatusNotComplete
	}
	if !inst.IsEphemeral() {
		std.Out.WriteWarningf("Cannot upgrade non-ephemeral instance %q", name)
		return ErrNotEphemeralInstance
	}
	if inst.IsExpired() {
		std.Out.WriteWarningf("Cannot upgrade expired instance %q", name)
		return ErrExpiredInstance
	}

	// All checks OK, we can issue the upgrade
	inst, err = cloudClient.UpgradeInstance(ctx, spec)
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
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		writeGCloudErrorSuggestion()
		return err
	}
	deploymentName := determineDeploymentName(ctx.String("name"), version, email, currentBranch)
	if ctx.String("name") != "" && ctx.String("name") != deploymentName {
		std.Out.WriteNoticef("Your deployment name has been truncated to be %q", deploymentName)
	}

	err = upgradeDeploymentForVersion(ctx.Context, email, deploymentName, version)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			std.Out.WriteWarningf("Unable to upgrade %q since no deployment like that exists", deploymentName)
			std.Out.WriteMarkdown("You can check what deployments exist under your GCP account with `sg cloud ephemeral list`, or you can see all deployments with `sg cloud ephemeral list --all`")
		}
		return err
	}
	return nil
}
