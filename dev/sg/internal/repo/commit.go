package repo

import (
	"context"
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

func allLinesPrefixed(lines []string, match string) bool {
	for _, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), match) {
			return false
		}
	}
	return true
}
