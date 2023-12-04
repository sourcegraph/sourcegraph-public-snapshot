package internal

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func repoCloneProgress(reposDir string, locker RepositoryLocker, repo api.RepoName) *protocol.RepoCloneProgress {
	dir := gitserverfs.RepoDirFromName(reposDir, repo)
	resp := protocol.RepoCloneProgress{
		Cloned: repoCloned(dir),
	}
	resp.CloneProgress, resp.CloneInProgress = locker.Status(dir)
	if isAlwaysCloningTest(repo) {
		resp.CloneInProgress = true
		resp.CloneProgress = "This will never finish cloning"
	}
	return &resp
}

func (s *Server) handleRepoCloneProgress(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoCloneProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := protocol.RepoCloneProgressResponse{
		Results: make(map[api.RepoName]*protocol.RepoCloneProgress, len(req.Repos)),
	}
	for _, repoName := range req.Repos {
		result := repoCloneProgress(s.ReposDir, s.Locker, repoName)
		resp.Results[repoName] = result
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRepoDelete(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := deleteRepo(r.Context(), s.Logger, s.DB, s.Hostname, s.ReposDir, req.Repo); err != nil {
		s.Logger.Error("failed to delete repository", log.String("repo", string(req.Repo)), log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Logger.Info("deleted repository", log.String("repo", string(req.Repo)))
}

func deleteRepo(
	ctx context.Context,
	logger log.Logger,
	db database.DB,
	shardID string,
	reposDir string,
	repo api.RepoName,
) error {
	// The repo may be deleted in the database, in this case we need to get the
	// original name in order to find it on disk
	err := gitserverfs.RemoveRepoDirectory(ctx, logger, db, shardID, reposDir, gitserverfs.RepoDirFromName(reposDir, api.UndeletedRepoName(repo)), true)
	if err != nil {
		return errors.Wrap(err, "removing repo directory")
	}
	err = db.GitserverRepos().SetCloneStatus(ctx, repo, types.CloneStatusNotCloned, shardID)
	if err != nil {
		return errors.Wrap(err, "setting clone status after delete")
	}
	return nil
}
