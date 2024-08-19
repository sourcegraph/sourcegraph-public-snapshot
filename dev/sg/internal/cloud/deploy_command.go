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

const CloudEphemeralPipeline = "cloud-ephemeral"

var (
	ErrDeploymentExists        error = errors.New("deployment already exists")
	ErrVersionNotFoundRegistry error = errors.New("tag/version not in Cloud Ephemeral registry")
	ErrMainBranchBuild         error = errors.New("cannot trigger a Cloud Ephemeral build for main branch")
)

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
		"\n```sg cloud ephemeral deploy --name <new-name>```\n\n" +
		"- Upgrade the existing deployment with the new version once the build completes by running\n" +
		"\n```sg cloud ephemeral upgrade --name \"%s\" --version \"%s\"```\n"
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
	inst, err := cloudClient.GetInstance(ctx, name)
	if err != nil && !errors.Is(err, ErrInstanceNotFound) {
		return errors.Wrapf(err, "failed to check if instance %q already exists", name)
	}
	// non-empty reason means instance is not fully created yet, it's ok to re-create
	if inst != nil && inst.Status.Reason == "" {
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
		name,
		version,
		license,
	)

	pending.Updatef("Creating deployment %q for version %q", spec.Name, spec.Version)
	inst, err = cloudClient.CreateInstance(ctx, spec)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Deployment of %q failed", spec.Name))
		return errors.Wrapf(err, "failed to deploy %q of version %s", spec.Name, spec.Version)
	}

	pending.Writef("Deploy instance details: \n%s", inst.String())
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Deployment %q created for version %q - access at: %s", inst.Name, inst.Version, inst.URL))
	return nil
}

func triggerEphemeralBuild(ctx context.Context, currRepo *repo.GitRepo) (*buildkite.Build, error) {
	if currRepo.Branch == "main" {
		return nil, ErrMainBranchBuild
	} else if strings.HasPrefix(currRepo.Branch, "main-dry-run") {
		return nil, ErrMainDryRunBranch
	}
	pending := std.Out.Pending(output.Linef("🔨", output.StylePending, "Checking if branch %q is up to date with remote branch", currRepo.Branch))
	if isOutOfSync, err := currRepo.IsOutOfSync(ctx); err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Failed to check if branch is out of sync with remote branch"))
		return nil, err
	} else if isOutOfSync {
		return nil, ErrBranchOutOfSync
	}

	client, err := bk.NewClient(ctx, std.Out)
	if err != nil {
		return nil, err
	}

	pending.Updatef("Starting cloud ephemeral build for %q on commit %q", currRepo.Branch, currRepo.Ref)
	build, err := client.TriggerBuild(ctx, CloudEphemeralPipeline, currRepo.Branch, currRepo.Ref, bk.WithEnvVar("CLOUD_EPHEMERAL", "true"))
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Failed to trigger build"))
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
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Failed to check if version %q exists in Cloud ephemeral registry", version))
		return err
	} else if len(images) == 0 {
		pending.Complete(output.Linef(output.EmojiWarningSign, output.StyleYellow, "Whoops! Version %q seems to be missing from the Cloud ephemeral registry. Please ask in #discuss-dev-infra to get it added to the registry", version))
		return ErrVersionNotFoundRegistry
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Version %q found in Cloud ephemeral registry", version))
	return nil
}

func determineDeploymentName(originalName, version, email, branch string) string {
	var deploymentName string
	if originalName != "" {
		deploymentName = originalName
	} else if version != "" {
		// if a version is given we generate a name based on the email user and the given version
		// to make sure the deployment is unique
		user := strings.ReplaceAll(email[0:strings.Index(email, "@")], ".", "-")
		deploymentName = user[:min(12, len(user))] + "-" + version
	} else {
		deploymentName = branch
	}

	deploymentName = sanitizeInstanceName(deploymentName)
	// names can only be max 30 chars
	return deploymentName[:min(30, len(deploymentName))]
}

func deployCloudEphemeral(ctx *cli.Context) error {
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}

	if strings.HasPrefix(currentBranch, "main-dry-run") {
		return ErrMainDryRunBranch
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
			if errors.Is(err, ErrBranchOutOfSync) {
				std.Out.WriteWarningf(`Your branch %q is out of sync with remote.

Please make sure you have either pushed or pulled the latest changes before trying again`, currRepo.Branch)
			} else if errors.Is(err, ErrMainBranchBuild) {
				std.Out.WriteWarningf(`Triggering Cloud Ephemeral builds from "main" is not supported.`)
				steps := "1. create a new branch off main by running `git switch <branch-name>`\n"
				steps += "2. push the branch to the remote by running `git push -u origin <branch-name>`\n"
				steps += "3. trigger the build by running `sg cloud ephemeral build`\n"
				std.Out.WriteMarkdown(withFAQ(fmt.Sprintf("Alternatively, if you still want to deploy \"main\" you can do:\n%s", steps)))
			} else if errors.Is(err, ErrMainDryRunBranch) {
				msg := "Triggering Cloud Ephemeral builds from \"main-dry-run\" branches are not supported. Try renaming the branch to not have the \"main-dry-run\" prefix as it complicates the eventual pipeline that gets generated"
				suggestion := "To rename a branch and launch a cloud ephemeral deployment do:\n"
				suggestion += fmt.Sprintf("1. `git branch -m %q <my-new-name>`\n", currRepo.Branch)
				suggestion += "2. `git push --set-upstream origin <my-new-name>`\n"
				suggestion += "3. trigger the build by running `sg cloud ephemeral build`\n"

				std.Out.WriteWarningf(msg)
				std.Out.WriteMarkdown(withFAQ(suggestion))
			}
			return errors.Wrapf(err, "cloud ephemeral deployment failure")
		}

		build = b
		version, err = determineVersion(build, "")
		if err != nil {
			return err
		}
		std.Out.WriteMarkdown(fmt.Sprintf("The build will push images with the following tag/version: `%s`", version))
	} else if err = checkVersionExistsInRegistry(ctx.Context, version); err != nil {
		return err
	}
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		writeGCloudErrorSuggestion()
		return err
	}

	// note we do not use the version here, we use ORIGINAL version, since it if it is given we create a different deployment name
	deploymentName := determineDeploymentName(ctx.String("name"), ctx.String("version"), email, currRepo.Branch)
	if ctx.String("name") != "" && ctx.String("name") != deploymentName {
		std.Out.WriteNoticef("Your deployment name has been truncated to be %q", deploymentName)
	}
	err = createDeploymentForVersion(ctx.Context, email, deploymentName, version)
	if err != nil {
		if errors.Is(err, ErrDeploymentExists) {
			std.Out.WriteWarningf("Cannot create a new deployment since a deployment with name %q already exists", deploymentName)
			std.Out.WriteMarkdown(withFAQ(deployUpgradeSuggestion(deploymentName, version)))
		}
		return err
	}
	return nil
}
