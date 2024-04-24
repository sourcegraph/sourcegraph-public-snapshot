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

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the release branch exists locally ..."))
	if _, err := execute.Git(ctx, "rev-parse", "--verify", releaseBranch); err == nil {
		p.Destroy()
		return errors.Newf("release branch %q already exists", releaseBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q does not exist locally", releaseBranch))

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the release branch exists in remote ..."))
	if _, err := execute.Git(ctx, "rev-parse", "--verify", fmt.Sprintf("origin/%s", releaseBranch)); err == nil {
		p.Destroy()
		return errors.Newf("release branch %q already exists", fmt.Sprintf("origin/%s", releaseBranch))
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q does not exist in remote", releaseBranch))

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the default branch is up to date with remote ..."))
	if _, err := execute.Git(ctx, "fetch"); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to fetch remote")
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
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Local branch is up to date with remote"))

	p = std.Out.Pending(output.Styled(output.StylePending, "Creating release branch..."))
	if _, err := execute.Git(ctx, "checkout", "-b", releaseBranch); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to create release branch")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q created", releaseBranch))

	defer func() {
		if _, err = execute.Git(ctx, "checkout", "-"); err != nil {
			std.Out.WriteWarningf("Unable to checkout previous branch before branch cut. %s", err.Error())
		}
	}()

	p = std.Out.Pending(output.Styled(output.StylePending, "Pushing release branch..."))
	if _, err := execute.Git(ctx, "push", "origin", releaseBranch); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to push release branch")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q pushed", releaseBranch))

	p = std.Out.Pending(output.Styled(output.StylePending, "Creating backport label..."))
	if _, err := execute.GH(
		ctx,
		"label",
		"create",
		fmt.Sprintf("backport %s", releaseBranch),
		"-d",
		fmt.Sprintf("label used to backport PRs to the %s release branch", releaseBranch),
	); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to create backport label")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Backport label created"))

	return nil
}
