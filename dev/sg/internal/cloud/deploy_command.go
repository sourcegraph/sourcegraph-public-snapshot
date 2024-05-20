package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/ci/gitops"
	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var ErrDeploymentExists error = errors.New("deployment already exists")

var deployEphemeralCommand = cli.Command{
	Name:        "deploy",
	Usage:       "Deploy a new ephemeral instance from the current branch or specific version",
	Description: "Deploy a new ephemeral instance from the current branch or specific version",
	Action:      deployCloudEphemeral,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			Usage:       "name of the instance to update the lease expiry time for",
			DefaultText: "current branch name will be used",
		},
		&cli.StringFlag{
			Name:        "version",
			DefaultText: "deploys an ephemeral cloud Sourcegraph environment with the specified version. The version MUST exist and implies that no build will be created",
		},
	},
}

func deployUpgradeSuggestion(name, version string) string {
	var text = "You might want to try one of the following:\n" +
		"- Create a new deployment with a different name by running\n" +
		"\n```sg cloud deploy --name <new-name>```\n\n" +
		"- Upgrade the existing deployment with the new version once the build completes by running\n" +
		"\n```sg cloud upgrade --name \"%s\"--version \"%s\"```\n"
	return fmt.Sprintf(text, name, version)
}

func determineVersion(build *buildkite.Build, tag string) (string, error) {
	if tag == "" {
		t, err := gitops.GetLatestTag()
		if err != nil {
			if err != gitops.ErrNoTags {
				return "", err
				// if we get no tags then we use an empty string - this is how it is done in CI
			}
			t = ""
		}
		tag = t
	}

	return images.BranchImageTag(
		time.Now(),
		pointers.DerefZero(build.Commit),
		pointers.DerefZero(build.Number),
		pointers.DerefZero(build.Branch),
		tag,
	), nil
}

func getCloudEphemeralLicenseKey(ctx context.Context) (string, error) {
	store, err := secrets.FromContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to get secrets store")
	}

	sec := secrets.ExternalSecret{
		Project: secrets.LocalDevProject,
		Name:    "SG_CLOUD_EPHEMERAL_LICENSE_KEY",
	}
	return store.GetExternal(ctx, sec)
}

func createDeploymentForVersion(ctx context.Context, email, name, version string) error {
	cloudClient, err := NewClient(ctx, email, APIEndpoint)
	if err != nil {
		return err
	}

	cloudEmoji := "☁️"
	pending := std.Out.Pending(output.Linef(cloudEmoji, output.StylePending, "Starting deployment %q for version %q", name, version))

	// Check if the deployment already exists
	pending.Updatef("Checking if deployment %q already exists", name)
	_, err = cloudClient.GetInstance(ctx, name)
	if err != nil {
		if !errors.Is(err, ErrInstanceNotFound) {
			return errors.Wrapf(err, "failed to check if instance %q already exists", name)
		}
	} else {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Deployment of %q failed", name))
		// Deployment exists
		return ErrDeploymentExists
	}

	pending.Updatef("Fetching license key...")
	license, err := getCloudEphemeralLicenseKey(ctx)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Deployment of %q failed", name))
		return err
	}
	spec := NewDeploymentSpec(
		sanitizeInstanceName(name),
		version,
		license,
	)

	pending.Updatef("Creating deployment %q for version %q", spec.Name, spec.Version)
	inst, err := cloudClient.CreateInstance(ctx, spec)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Deployment of %q failed", spec.Name))
		return errors.Wrapf(err, "failed to deploy %q of version %s", spec.Name, spec.Version)
	}

	pending.Writef("Deploy instance details: \n%s", inst.String())
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Deployment %q created for version %q - access at: %s", spec.Name, spec.Version, inst.URL))
	return nil
}

func triggerEphemeralBuild(ctx context.Context, currRepo *repo.GitRepo) (*buildkite.Build, error) {
	pending := std.Out.Pending(output.Linef("🔨", output.StylePending, "Checking if branch %q is up to date with remote branch", currRepo.Branch))
	if isOutOfSync, err := currRepo.IsOutOfSync(ctx); err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to check if branch is out of sync with remote branch"))
		return nil, err
	} else if isOutOfSync {
		return nil, ErrBranchOutOfSync
	}

	client, err := bk.NewClient(ctx, std.Out)
	if err != nil {
		return nil, err
	}

	pending.Updatef("Starting cloud ephemeral build for %q on commit %q", currRepo.Branch, currRepo.Ref)
	build, err := client.TriggerBuild(ctx, "sourcegraph", currRepo.Branch, currRepo.Ref, bk.WithEnvVar("CLOUD_EPHEMERAL", "true"))
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to trigger build"))
		return nil, err
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Build %d created. Build progress can be viewed at %s", pointers.DerefZero(build.Number), pointers.DerefZero(build.WebURL)))

	return build, nil
}

func checkVersionExistsInRegistry(ctx context.Context, version string) error {
	ar, err := NewDefaultCloudEphemeralRegistry(ctx)
	if err != nil {
		std.Out.WriteFailuref("failed to create Cloud Ephemeral registry")
		return err
	}
	pending := std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Checking if version %q exists in Cloud ephemeral registry", version))
	if images, err := ar.FindDockerImageExact(ctx, "gitserver", version); err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to check if version %q exists in Cloud ephemeral registry", version))
		return err
	} else if len(images) == 0 {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "no version %q found in Cloud ephemeral registry!", version))
		return errors.Newf("no image with tag %q found", version)
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Version %q found in Cloud ephemeral registry", version))
	return nil
}

func createDeploymentName(originalName, version, email, branch string) string {
	var deploymentName string
	if originalName != "" {
		deploymentName = originalName
	} else if version != "" {
		// if a version is given we generate a name based on the email user and the given version
		// to make sure the deployment is unique
		user := strings.ReplaceAll(email[0:strings.Index(email, "@")], ".", "_")
		deploymentName = user[:min(12, len(user))] + "_" + version
	} else {
		deploymentName = branch
	}

	return deploymentName

}

func deployCloudEphemeral(ctx *cli.Context) error {
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}

	// TODO(burmudar): We need to handle tags
	var currRepo *repo.GitRepo
	// We are on the branch we want to deploy, so we use the current commit
	head, err := repo.GetHeadCommit(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current commit")
	}
	currRepo = repo.NewGitRepo(currentBranch, head)

	var build *buildkite.Build
	version := ctx.String("version")
	// if a version is specified we do not build anything and just trigger the cloud deployment
	if version == "" {
		b, err := triggerEphemeralBuild(ctx.Context, currRepo)
		if err != nil {
			if err == ErrBranchOutOfSync {
				std.Out.WriteWarningf(`Your branch %q is out of sync with remote.

Please make sure you have either pushed or pulled the latest changes before trying again`, currRepo.Branch)
			} else {
				std.Out.WriteFailuref("Cannot start deployment as there was problem with the ephemeral build")
			}
			return errors.Wrapf(err, "cloud ephemeral deployment failure")
		}

		build = b
		version, err = determineVersion(build, ctx.String("tag"))
		if err != nil {
			return err
		}
		std.Out.WriteMarkdown(fmt.Sprintf("The build will push images with the following tag/version: `%s`", version))
	} else if err = checkVersionExistsInRegistry(ctx.Context, version); err != nil {
		return err
	}
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		return err
	}

	deploymentName := createDeploymentName(ctx.String("name"), version, email, currRepo.Branch)
	err = createDeploymentForVersion(ctx.Context, email, deploymentName, version)
	if err != nil {
		if errors.Is(err, ErrDeploymentExists) {
			std.Out.WriteWarningf("Cannot create a new deployment since a deployment with name %q already exists", deploymentName)
			std.Out.WriteMarkdown(deployUpgradeSuggestion(deploymentName, version))
		}
		return err
	}
	return nil
}
