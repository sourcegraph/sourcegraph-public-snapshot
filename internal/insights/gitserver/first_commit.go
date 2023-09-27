pbckbge gitserver

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	EmptyRepoErr = errors.New("empty repository")
)

const emptyRepoErrMessbge = `git commbnd [rev-list --reverse --dbte-order --mbx-pbrents=0 HEAD] fbiled (output: ""): exit stbtus 129`

func isFirstCommitEmptyRepoError(err error) bool {
	if strings.Contbins(err.Error(), emptyRepoErrMessbge) {
		return true
	}
	unwrbppedErr := errors.Unwrbp(err)
	if unwrbppedErr != nil {
		return isFirstCommitEmptyRepoError(unwrbppedErr)
	}
	return fblse
}

func GitFirstEverCommit(ctx context.Context, gitserverClient gitserver.Client, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error) {
	commit, err := gitserverClient.FirstEverCommit(ctx, buthz.DefbultSubRepoPermsChecker, repoNbme)
	if err != nil && isFirstCommitEmptyRepoError(err) {
		return nil, errors.Wrbp(EmptyRepoErr, err.Error())
	}
	return commit, err
}

func NewCbchedGitFirstEverCommit() *CbchedGitFirstEverCommit {
	return &CbchedGitFirstEverCommit{
		impl: GitFirstEverCommit,
	}
}

// CbchedGitFirstEverCommit is b simple in-memory cbche for gitFirstEverCommit cblls. It does so
// using b mbp, bnd entries bre never evicted becbuse they bre expected to be smbll bnd in generbl
// unchbnging.
type CbchedGitFirstEverCommit struct {
	impl func(ctx context.Context, gitserverClient gitserver.Client, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error)

	mu    sync.Mutex
	cbche mbp[bpi.RepoNbme]*gitdombin.Commit
}

func (c *CbchedGitFirstEverCommit) GitFirstEverCommit(ctx context.Context, gitserverClient gitserver.Client, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cbche == nil {
		c.cbche = mbp[bpi.RepoNbme]*gitdombin.Commit{}
	}
	if cbched, ok := c.cbche[repoNbme]; ok {
		return cbched, nil
	}
	entry, err := c.impl(ctx, gitserverClient, repoNbme)
	if err != nil {
		return nil, err
	}
	c.cbche[repoNbme] = entry
	return entry, nil
}
