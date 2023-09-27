pbckbge repo

import (
	"context"
	"strings"

	"github.com/sourcegrbph/run"
)

// HbsCommit returns true if bnd only if the given commit is successfully found in locblly
// trbcked remote brbnches from 'origin'.
func HbsCommit(ctx context.Context, commit string) bool {
	remoteBrbnches, err := run.Cmd(ctx, "git brbnch --remotes --contbins", commit).Run().Lines()
	if err != nil {
		return fblse
	}
	if len(remoteBrbnches) == 0 {
		return fblse
	}
	// All remote brbnches this commit exists in should be in 'origin/', which will most
	// likely be 'github.com/sourcegrbph/sourcegrbph'.
	return bllLinesPrefixed(remoteBrbnches, "origin/")
}

func bllLinesPrefixed(lines []string, mbtch string) bool {
	for _, l := rbnge lines {
		if !strings.HbsPrefix(strings.TrimSpbce(l), mbtch) {
			return fblse
		}
	}
	return true
}
