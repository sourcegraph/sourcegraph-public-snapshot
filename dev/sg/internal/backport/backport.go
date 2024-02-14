package backport

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking for GitHub CLI..."))
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		p.Destroy()
		return errors.Wrap(err, "GitHub CLI (https://cli.github.com/) is required for installation")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Using GitHub CLI at %q", ghPath))

	p = std.Out.Pending(output.Styled(output.StylePending, "Checking GH auth status..."))
	_, err = ghExec(cmd.Context, "auth", "status")
	if err != nil {
		p.Destroy()
		return errors.Wrap(err, "GitHub CLI is not authenticated. Please run 'gh auth login' to authenticate")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "GH auth is authenticated"))

	p = std.Out.Pending(output.Styledf(output.StylePending, "Checking the existence of %s in remote...", version))
	_, err = ghExec(cmd.Context, "api", fmt.Sprintf("/repos/sourcegraph/sourcegraph/branches/%s", version))
	if err != nil {
		p.Destroy()
		return errors.Wrapf(err, "%q does not exist in sourcegraph/sourcegraph", version)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Found %q in remote", version))

	p = std.Out.Pending(output.Styled(output.StylePending, "Getting PR info ...."))
	rawPrInfo, err := ghExec(cmd.Context, "pr", "view", fmt.Sprintf("%d", prNumber), "--json", "mergeCommit,state,body,title")
	if err != nil {
		p.Destroy()
		return errors.Wrapf(err, "Unable to fetch information for pull request: %d", prNumber)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Fetched information for PR: %d", prNumber))

	var pr PRInfo
	if err := json.Unmarshal(rawPrInfo, &pr); err != nil {
		return errors.Wrap(err, "Unable to parse PR info")
	}

	if pr.State != "MERGED" {
		return errors.Newf("PR is not merged: %s. Only merged PRs can be backported", pr.State)
	}

	mergeCommit := pr.MergeCommit.Oid

	backportBranch := fmt.Sprintf("backport-%d-to-%s", prNumber, version)
	p = std.Out.Pending(output.Styledf(output.StylePending, "Creating backport branch %q...", backportBranch))
	if err := gitExec(cmd.Context, "checkout", "-b", backportBranch, fmt.Sprintf("origin/%s", version)); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "Unable to create backport branch: %q", backportBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Backport branch %q created", backportBranch))

	p = std.Out.Pending(output.Styledf(output.StylePending, "Cherry-picking merge commit for PR %d into backport branch...", prNumber))
	if err := gitExec(cmd.Context, "cherry-pick", mergeCommit); err != nil {
		p.Destroy()
		// If this fails looool, nothing we much we can do here lol.
		_ = gitExec(cmd.Context, "cherry-pick", "--abort")
		return errors.Wrapf(err, "Unable to cherry-pick merge commit: %s. This might be the result of a merge conflict. Manually run `git cherry-pick %s` and fix on your machine.", mergeCommit, mergeCommit)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Cherry-picked merge commit for PR %d into backport branch", prNumber))

	p = std.Out.Pending(output.Styledf(output.StylePending, "Pushing backport branch %q to remote...", backportBranch))
	if err := gitExec(cmd.Context, "push", "--set-upstream", "origin", backportBranch); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "Unable to push backport branch: %q", backportBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Backport branch %q pushed to remote", backportBranch))

	prBody := generatePRBody(pr.Body, mergeCommit, prNumber)
	prTitle := generatePRTitle(pr.Title, version)
	p = std.Out.Pending(output.Styledf(output.StylePending, "Creating pull request for backport branch %q...", backportBranch))
	if _, err := ghExec(cmd.Context, "pr", "create", "--fill", "--base", version, "--head", backportBranch, "--title", prTitle, "--body", prBody); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "Unable to create pull request for backport branch: %q", backportBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Pull request for backport branch %q created", backportBranch))

	return nil
}

type PRInfo struct {
	MergeCommit struct {
		Oid string `json:"oid"`
	} `json:"mergeCommit"`
	State string `json:"state"`
	Body  string `json:"body"`
	Title string `json:"title"`
}

func generatePRBody(body, mergeCommit string, prNumber int64) string {
	shortCommitSha := mergeCommit[:7]
	return fmt.Sprintf("%s\n\nBackport %s from #%d", body, shortCommitSha, prNumber)
}

func generatePRTitle(title, version string) string {
	return fmt.Sprintf("[Backport %s] %s", version, title)
}
