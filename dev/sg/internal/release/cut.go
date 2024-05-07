package release

import (
	"fmt"
	"os/exec"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/execute"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
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

	releaseBranch := v.String()
	if v.Patch() == 0 {
		releaseBranch = fmt.Sprintf("%d.%d", v.Major(), v.Minor()) // 5.3 not 5.3.0
	}

	defaultBranch := cctx.String("branch")

	ctx := cctx.Context
	releaseGitRepoBranch := repo.NewGitRepo(releaseBranch, releaseBranch)
	defaultGitRepoBranch := repo.NewGitRepo(defaultBranch, defaultBranch)

	if ok, err := defaultGitRepoBranch.IsDirty(ctx); err != nil {
		return errors.Wrap(err, "check if current branch is dirty")
	} else if ok {
		return errors.Newf("current branch is dirty. please commit your unstaged changes")
	}

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the release branch exists locally ..."))
	if ok, err := releaseGitRepoBranch.HasLocalBranch(ctx); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking if %q branch exists localy", releaseBranch)
	} else if ok {
		p.Destroy()
		return errors.Newf("branch %q exists locally", releaseBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q does not exist locally", releaseBranch))

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the release branch exists in remote ..."))
	if ok, err := releaseGitRepoBranch.HasRemoteBranch(ctx); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking if %q branch exists in remote repo", releaseBranch)
	} else if ok {
		p.Destroy()
		return errors.Newf("release branch %q exists in remote repo", releaseBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q does not exist in remote", releaseBranch))

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the default branch is up to date with remote ..."))
	if _, err := defaultGitRepoBranch.FetchOrigin(ctx); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "fetching origin for %q", defaultBranch)
	}

	if err := defaultGitRepoBranch.Checkout(ctx); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking out %q", defaultBranch)
	}

	if ok, err := defaultGitRepoBranch.IsOutOfSync(ctx); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking if %q branch is up to date with remote", defaultBranch)
	} else if ok {
		p.Destroy()
		return errors.Newf("local branch %q is not up to date with remote", defaultBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Local branch is up to date with remote"))

	p = std.Out.Pending(output.Styled(output.StylePending, "Creating release branch..."))
	if err := releaseGitRepoBranch.CheckoutNewBranch(ctx); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to create release branch")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q created", releaseBranch))

	p = std.Out.Pending(output.Styled(output.StylePending, "Pushing release branch..."))
	if _, err := releaseGitRepoBranch.Push(ctx); err != nil {
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
