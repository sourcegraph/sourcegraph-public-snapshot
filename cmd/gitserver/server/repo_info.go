package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Server) repoInfo(ctx context.Context, repo api.RepoName) (*protocol.RepoInfo, error) {
	dir := s.dir(repo)
	resp := protocol.RepoInfo{}

	resp.CloneProgress, _ = s.locker.Status(dir)

	return &resp, nil
}

func (s *Server) handleRepoInfo(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := protocol.RepoInfoResponse{
		Results: make(map[api.RepoName]protocol.RepoInfoResult, len(req.Repos)),
	}
	for _, repoName := range req.Repos {
		result, err := s.repoInfo(r.Context(), repoName)
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		resp.Results[repoName] = protocol.RepoInfoResult{
			RepoInfo: result,
			Error:    errMsg,
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleReposStats(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile(filepath.Join(s.ReposDir, reposStatsName))
	if errors.Is(err, os.ErrNotExist) {
		// When a gitserver is new this file might not have been computed
		// yet. Clients are expected to handle this case by noticing UpdatedAt
		// is not set.
		b = []byte("{}")
	} else if err != nil {
		http.Error(w, fmt.Sprintf("failed to read %s: %v", reposStatsName, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(b)
}

func (s *Server) handleRepoDelete(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.deleteRepo(r.Context(), req.Repo); err != nil {
		s.Logger.Error("failed to delete repository", log.String("repo", string(req.Repo)), log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Logger.Info("deleted repository", log.String("repo", string(req.Repo)))
}

func (s *Server) deleteRepo(ctx context.Context, repo api.RepoName) error {
	// The repo may be deleted in the database, in this case we need to get the
	// original name in order to find it on disk
	err := s.removeRepoDirectory(s.dir(api.UndeletedRepoName(repo)), true)
	if err != nil {
		return errors.Wrap(err, "removing repo directory")
	}
	// TODO: This is duplicative?
	err = s.setCloneStatus(ctx, repo, types.CloneStatusNotCloned)
	if err != nil {
		return errors.Wrap(err, "setting clone status after delete")
	}
	return nil
}
