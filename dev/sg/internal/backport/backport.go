package backport

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// # Fetch latest updates from GitHub
// git fetch
// # Create a new working tree
// git worktree add .worktrees/backport-5.3 5.3
// # Navigate to the new working tree
// cd .worktrees/backport-5.3
// # Create a new branch
// git switch --create backport-60100-to-5.3
// # Cherry-pick the merged commit of this pull request and resolve the conflicts
// git cherry-pick -x --mainline 1 ce5461792c70feb6d77408f4e6305e0c35bc984c
// # Push it to GitHub
// git push --set-upstream origin backport-60100-to-5.3
// # Go back to the original working tree
// cd ../..
// # Delete the working tree
// git worktree remove .worktrees/backport-5.3

// git cherry-pick --continue
// # Push it to GitHub
// git push --set-upstream origin backport-60100-to-5.3
// # Go back to the original working tree
// cd ../..
// # Delete the working tree
// git worktree remove .worktrees/backport-5.3

func Run(cmd *cli.Context, prNumber int64, version string) error {
	// Fetch latest change from remote
	p := std.Out.Pending(output.Styled(output.StylePending, "Fetching latest changes from remote..."))
	if err := gitExec(cmd.Context, "fetch"); err != nil {
		p.Destroy()
		return err
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Fetched latest changes from remote"))

	worktreeName := fmt.Sprintf(".worktrees/backport-%d", prNumber)
	p.Updatef("%s Creating new working tree %q...", output.EmojiHourglass, worktreeName)
	if err := gitExec(cmd.Context, "worktree", "add", worktreeName, version); err != nil {
		p.Destroy()
		return err
	}
	// confirm the branch
	return nil
}
