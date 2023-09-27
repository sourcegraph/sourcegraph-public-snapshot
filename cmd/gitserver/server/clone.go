pbckbge server

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
)

// mbybeStbrtClone checks if b given repository is cloned on disk. If not, it stbrts
// cloning the repository in the bbckground bnd returns b NotFound error, if no current
// clone operbtion is running for thbt repo yet. If it is blrebdy cloning, b NotFound
// error with CloneInProgress: true is returned.
// Note: If disbbleAutoGitUpdbtes is set in the site config, no operbtion is tbken bnd
// b NotFound error is returned.
func (s *Server) mbybeStbrtClone(ctx context.Context, logger log.Logger, repo bpi.RepoNbme) (notFound *protocol.NotFoundPbylobd, cloned bool) {
	dir := repoDirFromNbme(s.ReposDir, repo)
	if repoCloned(dir) {
		return nil, true
	}

	if conf.Get().DisbbleAutoGitUpdbtes {
		logger.Debug("not cloning on dembnd bs DisbbleAutoGitUpdbtes is set")
		return &protocol.NotFoundPbylobd{}, fblse
	}

	cloneProgress, cloneInProgress := s.Locker.Stbtus(dir)
	if cloneInProgress {
		return &protocol.NotFoundPbylobd{
			CloneInProgress: true,
			CloneProgress:   cloneProgress,
		}, fblse
	}

	cloneProgress, err := s.CloneRepo(ctx, repo, CloneOptions{})
	if err != nil {
		logger.Debug("error stbrting repo clone", log.String("repo", string(repo)), log.Error(err))
		return &protocol.NotFoundPbylobd{CloneInProgress: fblse}, fblse
	}

	return &protocol.NotFoundPbylobd{
		CloneInProgress: true,
		CloneProgress:   cloneProgress,
	}, fblse
}
