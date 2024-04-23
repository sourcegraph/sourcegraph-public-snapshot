package release

import (
	"fmt"
	"os/exec"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"

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

	// branchName := fmt.Sprintf("%s.%")

	fmt.Println("Cutting release branch", v.String())

	// if branch == "" {
	// 	// get current branch
	// 	// git rev-parse --abbrev-ref HEAD
	// 	err := execute.Git(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return errors.New("branch is required")
	// }

	// if err := execute.Git(ctx, "checkout", "-b", v, fmt.Sprintf("origin/%s", branch)); err != nil {
	// 	return err
	// }
	return nil
}
