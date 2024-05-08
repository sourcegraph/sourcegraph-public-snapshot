package cloud

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var BuildEphemeralCommand = cli.Command{
	Name:        "build",
	Usage:       "sg could build",
	Description: "Triggers a Cloud Ephemeral build of the current branch which will push images to the cloud ephemeral registry",
	Action:      wipAction(buildCloudEphemeral),
}

func buildCloudEphemeral(ctx *cli.Context) error {
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}
	// We are on the branch we want to deploy, so we use the current commit
	head, err := repo.GetHeadCommit(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current commit")
	}
	currRepo := repo.NewGitRepo(currentBranch, head)
	build, err := triggerEphemeralBuild(ctx.Context, currRepo)
	if err != nil {
		if err == ErrBranchOutOfSync {
			std.Out.WriteWarningf(`Your branch %q is out of sync with the remote branch.

Please make sure you have either pushed or pulled the latest changes before trying again`, currRepo.Branch)
		}
		return errors.Wrapf(err, "failed to trigger epehemeral build for branch")
	}
	version, err := determineVersion(build, ctx.String("tag"))
	if err != nil {
		return err
	}
	std.Out.WriteMarkdown(fmt.Sprintf("The build will push images with the following tag/version: `%s`", version))
	return nil
}
