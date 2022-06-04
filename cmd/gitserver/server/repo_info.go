package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

func (s *Server) repoInfo(ctx context.Context, repo api.RepoName) (*protocol.RepoInfo, error) {
	dir := s.dir(repo)
	resp := protocol.RepoInfo{
		Cloned: repoCloned(dir),
	}
	if resp.Cloned {
		// TODO(keegancsmith,tsenart) the only user of this information is the site admin
		// settings page for a repo. That page should just ask the DB for the remote URL.
		//
		// We need an internal actor here since we query the repo table. We are trusting
		// the auth checks we already have in place that do not allow a site admin to
		// view private repos they don't own.
		remoteURL, err := s.getRemoteURL(actor.WithInternalActor(ctx), repo)
		if err != nil {
			return nil, err
		}
		resp.URL = remoteURL.String()
	}
	{
		resp.CloneProgress, resp.CloneInProgress = s.locker.Status(dir)
		if isAlwaysCloningTest(repo) {
			resp.CloneInProgress = true
			resp.CloneProgress = "This will never finish cloning"
		}
	}
	if resp.Cloned {
		if mtime, err := repoLastFetched(dir); err != nil {
			s.Logger.Warn("error computing last-fetched date", log.String("repo", string(repo)), log.Error(err))
		} else {
			resp.LastFetched = &mtime
		}

		if cloneTime, err := getRecloneTime(dir); err != nil {
			s.Logger.Warn("error getting re-clone time", log.String("repo", string(repo)), log.Error(err))
		} else {
			resp.CloneTime = &cloneTime
		}

		if lastChanged, err := repoLastChanged(dir); err != nil {
			s.Logger.Warn("error getting last changed", log.String("repo", string(repo)), log.Error(err))
		} else {
			resp.LastChanged = &lastChanged
		}
	}
	gitRepo, err := s.DB.GitserverRepos().GetByName(ctx, repo)
	if err == nil {
		resp.Size = gitRepo.RepoSizeBytes
	}
	return &resp, nil
}

func (s *Server) repoCloneProgress(repo api.RepoName) *protocol.RepoCloneProgress {
	dir := s.dir(repo)
	resp := protocol.RepoCloneProgress{
		Cloned: repoCloned(dir),
	}
	resp.CloneProgress, resp.CloneInProgress = s.locker.Status(dir)
	if isAlwaysCloningTest(repo) {
		resp.CloneInProgress = true
		resp.CloneProgress = "This will never finish cloning"
	}
	return &resp
}

func (s *Server) handleRepoInfo(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := protocol.RepoInfoResponse{
		Results: make(map[api.RepoName]*protocol.RepoInfo, len(req.Repos)),
	}
	for _, repoName := range req.Repos {
		result, err := s.repoInfo(r.Context(), repoName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Results[repoName] = result
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
		result := s.repoCloneProgress(repoName)
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
	err := s.removeRepoDirectory(s.dir(api.UndeletedRepoName(repo)))
	if err != nil {
		return errors.Wrap(err, "removing repo directory")
	}
	err = s.setCloneStatus(ctx, repo, types.CloneStatusNotCloned)
	if err != nil {
		return errors.Wrap(err, "setting clone status after delete")
	}
	return nil
}
