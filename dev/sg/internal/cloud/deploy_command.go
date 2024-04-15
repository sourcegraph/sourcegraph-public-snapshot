package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/run"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var ErrUserCancelled = errors.New("user cancelled")

var DeployEphemeralCommand = cli.Command{
	Name:        "deploy",
	Usage:       "sg could deploy --branch <branch> --tag <tag>",
	Description: "Deploy the specified branch or tag to an ephemeral Sourcegraph Cloud environment",
	Action:      deployCloudEphemeral,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "branch",
		},
		&cli.StringFlag{
			Name: "tag",
		},
		&cli.StringFlag{
			Name:        "version",
			DefaultText: "deploys an ephemeral cloud Sourcegraph environment with the specified version. The version MUST exist and implies that no build will be created",
		},
	},
}

func determineVersion(build *buildkite.Build, tag string) string {
	return images.BranchImageTag(
		time.Now(),
		pointers.DerefZero(build.Commit),
		pointers.DerefZero(build.Number),
		pointers.DerefZero(build.Branch),
		tag,
	)
}

func ensureBranchIsSyncd(ctx context.Context, currRepo *repo.GitRepo) error {
	if ok, err := currRepo.IsOutOfSync(ctx); err != nil {
		return err
	} else if ok {
		return nil
	}

	var answer string
	oneOf := func(value string, i ...string) bool {
		for _, item := range i {
			if value == item {
				return true
			}
		}
		return false
	}
	ok, err := std.PromptAndScan(std.Out, fmt.Sprintf("Commit %q on branch %q does not exist remotely. Do you want to push it to origin? (yes/no)", currRepo.Ref, currRepo.Branch), &answer)
	if err != nil {
		return err
	}
	if !ok || !oneOf(answer, "yes", "y") {
		return ErrUserCancelled
	}

	std.Out.WriteNoticef("Pushing commit %q to origin/%s\n", currRepo.Ref, currRepo.Branch)
	out, err := currRepo.PushToRemote(ctx)
	if err != nil {
		std.Out.WriteCode("bash", out)
		return err
	}
	std.Out.WriteCode("bash", out)

	// if we pushed we wait a little bit otherwise follow up actions might not trigger properly
	time.Sleep(3 * time.Second)
	return nil
}

func getGcloudAccount(ctx context.Context) (string, error) {
	return run.Cmd(ctx, "gcloud", "config", "get", "account").Run().String()
}

func createEphemeralBuild(ctx context.Context, _, branch string) (*buildkite.Build, error) {
	currRepo, err := repo.NewWithBranch(ctx, branch)
	err = ensureBranchIsSyncd(ctx, currRepo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure current commit can be built")
	}

	// Check that branch has been pushed
	// offer to push branch
	//
	// 1. kick of a build so that we can get the images
	// 2. Once the build is kicked off we will need the build number so taht we can generate the version locally
	std.Out.WriteNoticef("Starting build for %q on commit %q\n", currRepo.Branch, currRepo.Ref)
	client, err := bk.NewClient(ctx, std.Out)
	if err != nil {
		return nil, err
	}
	build, err := client.TriggerBuild(ctx, "sourcegraph", currRepo.Branch, currRepo.Ref, bk.WithEnvVar("CLOUD_EPHEMERAL", "true"))
	if err != nil {
		return nil, err
	}
	std.Out.WriteSuccessf("Started build %d. Build progress can be viewed at %s\n", pointers.DerefZero(build.Number), pointers.DerefZero(build.WebURL))

	return build, nil
}

func deployCloudEphemeral(ctx *cli.Context) error {
	branch := ctx.String("branch")
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}
	// if the tag is set - we should prefer it over the branch
	tag := ctx.String("tag")

	if branch == "" && tag == "" {
		branch = currentBranch
	} else if branch != currentBranch {
		// we are not on the intended branch so we create a cloud-ephemeral branch so that we don't interfere with the branch specified
		account, err := getGcloudAccount(ctx.Context)
		if err != nil {
			return err
		}
		user := strings.ReplaceAll(account[:strings.Index(account, "@")], ".", "-")
		// create a branch of the format cloud-ephemeral/<user>_<branch>
		branch = fmt.Sprintf("cloud-ephemeral/%s_%s", user, strings.ReplaceAll(branch, "/", "-"))
		std.Out.Writef("currently not on %q branch - using %q as branch\n", currentBranch, branch)
	}

	version := ctx.String("version")
	// if a version is specified we do not build anything and just trigger the cloud deployment
	if version == "" {
		build, err := createEphemeralBuild(ctx.Context, tag, branch)
		if err != nil {
			return err
		}
		version = determineVersion(build, tag)
	}

	cloudClient, err := NewClient(ctx.Context, APIEndpoint)
	if err != nil {
		return err
	}

	std.Out.Writef("Starting cloud ephemeral deployment for version %q\n", version)
	inst, err := cloudClient.ListInstances(ctx.Context)
	if err != nil {
		return err
	}

	std.Out.Writef("Found %d instances\n", len(inst))
	// 3. Once we have the version we can kick off the cloud deploy so that it can start provisioning the environment

	return nil
}
