package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/run"
)

// HasCommit returns true if and only if the given commit is successfully found in locally
// tracked remote branches from 'origin'.
func HasCommit(ctx context.Context, commit string) bool {
	remoteBranches, err := run.Cmd(ctx, "git branch --remotes --contains", commit).Run().Lines()
	if err != nil {
		return false
	}
	if len(remoteBranches) == 0 {
		return false
	}
	// All remote branches this commit exists in should be in 'origin/', which will most
	// likely be 'github.com/sourcegraph/sourcegraph'.
	return allLinesPrefixed(remoteBranches, "origin/")
}

func Push(ctx context.Context, branch string) (string, error) {
	if branch == "" {
		value, err := GetBranch(ctx)
		if err != nil {
			return "", err
		}
		branch = value
	}

	return run.Cmd(ctx, "git", "push", "origin", branch).Run().String()
}

func GetBranch(ctx context.Context) (string, error) {
	branch, err := run.Cmd(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD").Run().String()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(branch), nil
}

func GetHeadCommit(ctx context.Context) (string, error) {
	commit, err := run.Cmd(ctx, "git rev-parse HEAD").Run().String()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(commit), nil
}

func GetBranchHeadCommit(ctx context.Context, branch string) (string, error) {
	commit, err := run.Cmd(ctx, fmt.Sprintf("git rev-parse %s", branch)).Run().String()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(commit), nil
}

func allLinesPrefixed(lines []string, match string) bool {
	for _, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), match) {
			return false
		}
	}
	return true
}
