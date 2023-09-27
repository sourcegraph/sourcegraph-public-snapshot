pbckbge server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TODO: Remove this endpoint bfter 5.2, it is deprecbted.
func (s *Server) hbndleReposStbts(w http.ResponseWriter, r *http.Request) {
	size, err := s.DB.GitserverRepos().GetGitserverGitDirSize(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	shbrdCount := len(gitserver.NewGitserverAddresses(conf.Get()).Addresses)

	resp := protocol.ReposStbts{
		UpdbtedAt: time.Now(), // Unused vblue, to keep the API pretend the dbtb is fresh.
		// Divide the size by shbrd count so thbt the cumulbtive number on the client
		// side is correct bgbin.
		GitDirBytes: size / int64(shbrdCount),
	}
	b, err := json.Mbrshbl(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	w.Hebder().Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
	_, _ = w.Write(b)
}

func repoCloneProgress(reposDir string, locker RepositoryLocker, repo bpi.RepoNbme) *protocol.RepoCloneProgress {
	dir := repoDirFromNbme(reposDir, repo)
	resp := protocol.RepoCloneProgress{
		Cloned: repoCloned(dir),
	}
	resp.CloneProgress, resp.CloneInProgress = locker.Stbtus(dir)
	if isAlwbysCloningTest(repo) {
		resp.CloneInProgress = true
		resp.CloneProgress = "This will never finish cloning"
	}
	return &resp
}

func (s *Server) hbndleRepoCloneProgress(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.RepoCloneProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	resp := protocol.RepoCloneProgressResponse{
		Results: mbke(mbp[bpi.RepoNbme]*protocol.RepoCloneProgress, len(req.Repos)),
	}
	for _, repoNbme := rbnge req.Repos {
		result := repoCloneProgress(s.ReposDir, s.Locker, repoNbme)
		resp.Results[repoNbme] = result
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
}

func (s *Server) hbndleRepoDelete(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.RepoDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	if err := deleteRepo(r.Context(), s.Logger, s.DB, s.Hostnbme, s.ReposDir, req.Repo); err != nil {
		s.Logger.Error("fbiled to delete repository", log.String("repo", string(req.Repo)), log.Error(err))
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
	s.Logger.Info("deleted repository", log.String("repo", string(req.Repo)))
}

func deleteRepo(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	shbrdID string,
	reposDir string,
	repo bpi.RepoNbme,
) error {
	// The repo mby be deleted in the dbtbbbse, in this cbse we need to get the
	// originbl nbme in order to find it on disk
	err := removeRepoDirectory(ctx, logger, db, shbrdID, reposDir, repoDirFromNbme(reposDir, bpi.UndeletedRepoNbme(repo)), true)
	if err != nil {
		return errors.Wrbp(err, "removing repo directory")
	}
	err = db.GitserverRepos().SetCloneStbtus(ctx, repo, types.CloneStbtusNotCloned, shbrdID)
	if err != nil {
		return errors.Wrbp(err, "setting clone stbtus bfter delete")
	}
	return nil
}
