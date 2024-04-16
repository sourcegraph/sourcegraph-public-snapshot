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
var ErrWrongBranch = errors.New("wrong current branch")
var ErrBranchOutOfSync = errors.New("branch is out of sync with remote")

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
		&cli.BoolFlag{
			Name: "skip-wip-notice",
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
func oneOfEquals(value string, i ...string) bool {
	for _, item := range i {
		if value == item {
			return true
		}
	}
	return false
}

func getGcloudAccount(ctx context.Context) (string, error) {
	return run.Cmd(ctx, "gcloud", "config", "get", "account").Run().String()
}

func triggerEphemeralBuild(ctx context.Context, currRepo *repo.GitRepo) (*buildkite.Build, error) {
	if isOutOfSync, err := currRepo.IsOutOfSync(ctx); err != nil {
		return nil, err
	} else if isOutOfSync {
		return nil, ErrBranchOutOfSync
	}

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

func printWIPNotice(ctx *cli.Context) error {
	if ctx.Bool("skip-wip-notice") {
		return nil
	}
	notice := "This is command is still a work in progress and it is not recommend for general use! 🚨 Do you want to continue? (yes/no)"

	var answer string
	if _, err := std.PromptAndScan(std.Out, notice, &answer); err != nil {
		return err
	}

	if oneOfEquals(answer, "yes", "y") {
		return nil
	}

	return ErrUserCancelled

}

func createDeploymentForVersion(ctx context.Context, version string) error {
	cloudClient, err := NewClient(ctx, APIEndpoint)
	if err != nil {
		return err
	}

	std.Out.Writef("Starting cloud ephemeral deployment for version %q\n", version)
	// Lets just list as a temporary sanity check that this works
	inst, err := cloudClient.ListInstances(ctx)
	if err != nil {
		return err
	}

	std.Out.Writef("Found %d instances\n", len(inst))
	return nil
}

func createEphemeralBranchName(ctx context.Context, branch string) (string, error) {
	account, err := getGcloudAccount(ctx)
	if err != nil {
		return "", err
	}
	user := strings.ReplaceAll(account[:strings.Index(account, "@")], ".", "-")
	// create a branch of the format cloud-ephemeral/<user>_<branch>
	return fmt.Sprintf("cloud-ephemeral/%s_%s", user, strings.ReplaceAll(branch, "/", "-")), nil
}

func deployCloudEphemeral(ctx *cli.Context) error {
	// while we work on this command we print a notice and ask to continue
	if err := printWIPNotice(ctx); err != nil {
		return err
	}
	branch := ctx.String("branch")
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}
	// if the tag is set - we should prefer it over the branch
	tag := ctx.String("tag")

	// if neither the branch nor the tag is set - then we assume the current branch
	if branch == "" && tag == "" {
		branch = currentBranch
	}

	var currRepo *repo.GitRepo
	if branch != currentBranch {
		wrongBranch := `You are currently on branch %q, but want to deploy %q.

We currently do not support deploying a branch that is different to the one you're currently on.

Please switch to branch %q and make sure it is up to date before trying again.`
		std.Out.WriteWarningf(wrongBranch, currentBranch, branch, branch)
		return ErrWrongBranch
	} else {
		// We are on the branch we want to deploy, so we use the current commit
		head, err := repo.GetHeadCommit(ctx.Context)
		if err != nil {
			return errors.Wrap(err, "failed to determine current commit")
		}
		currRepo = repo.NewGitRepo(currentBranch, head)
	}

	version := ctx.String("version")
	// if a version is specified we do not build anything and just trigger the cloud deployment
	if version == "" {
		build, err := triggerEphemeralBuild(ctx.Context, currRepo)
		if err != nil {
			if err == ErrBranchOutOfSync {
				std.Out.WriteWarningf(`Your branch %q is out of sync with remote.

Please make sure you have either pushed or pulled the latest changes before trying again`, currRepo.Branch)
			}
			return err
		}
		_ = determineVersion(build, tag)
	}
	// we could check if the version exists?

	return nil
	// trigger cloud depoyment here
}
