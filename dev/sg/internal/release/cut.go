package release

import (
	"fmt"
	"os/exec"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/execute"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func cutReleaseBranch(cctx *cli.Context) error {
	p := std.Out.Pending(output.Styled(output.StylePending, "Checking for GitHub CLI..."))
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		p.Destroy()
		return errors.Wrap(err, "GitHub CLI (https://cli.github.com/) is required for installation")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Using GitHub CLI at %q", ghPath))

	version := cctx.String("version")
	v, err := semver.NewVersion(version)
	if err != nil {
		return errors.Newf("invalid version %q, must be semver", version)
	}

	ctx := cctx.Context

	releaseBranch := v.String()
	defaultBranch := "main"

	if _, err := execute.Git(ctx, "rev-parse", "--verify", releaseBranch); err == nil {
		return errors.Newf("release branch %q already exists", releaseBranch)
	}

	localCommitSHA, err := execute.Git(ctx, "rev-parse", defaultBranch)
	if err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to get local commit SHA")
	}

	remoteCommitSHA, err := execute.Git(ctx, "rev-parse", fmt.Sprintf("origin/%s", defaultBranch))
	if err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to get remote commit SHA")
	}

	if string(localCommitSHA) != string(remoteCommitSHA) {
		p.Destroy()
		return errors.New("local branch is not up to date with remote, please pull the latest changes")
	}

	if _, err := execute.Git(ctx, "checkout", "-b", releaseBranch); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to create release branch")
	}

	defer func() {
		if _, err = execute.Git(ctx, "checkout", "-"); err != nil {
			std.Out.WriteWarningf("Unable to checkout previous branch before branch cut. %s", err.Error())
		}
	}()

	if _, err := execute.Git(ctx, "push", "origin", releaseBranch); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to push release branch")
	}

	if _, err := execute.GH(
		ctx,
		"label",
		"create",
		fmt.Sprintf("backport %s", releaseBranch),
		"-d",
		fmt.Sprintf("label used to backport PRs to the %s release branch", releaseBranch),
	); err != nil {
		return errors.Wrap(err, "failed to create backport label")
	}

	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q created", releaseBranch))
	return nil
}
