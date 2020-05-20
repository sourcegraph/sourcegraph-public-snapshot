package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (s *Server) repoInfo(ctx context.Context, repo api.RepoName) (*protocol.RepoInfo, error) {
	dir := s.dir(repo)
	resp := protocol.RepoInfo{
		Cloned: repoCloned(dir),
	}
	if resp.Cloned {
		remoteURL, err := repoRemoteURL(ctx, dir)
		if err != nil {
			return nil, err
		}
		resp.URL = remoteURL
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
			log15.Warn("error computing last-fetched date", "repo", repo, "err", err)
		} else {
			resp.LastFetched = &mtime
		}

		if cloneTime, err := getRecloneTime(dir); err != nil {
			log15.Warn("error getting reclone time", "repo", repo, "err", err)
		} else {
			resp.CloneTime = &cloneTime
		}

		if lastChanged, err := repoLastChanged(dir); err != nil {
			log15.Warn("error getting last changed", "repo", repo, "err", err)
		} else {
			resp.LastChanged = &lastChanged
		}
	}
	return &resp, nil
}

func (s *Server) repoCloneProgress(repo api.RepoName) (*protocol.RepoCloneProgress, error) {
	dir := s.dir(repo)
	resp := protocol.RepoCloneProgress{
		Cloned: repoCloned(dir),
	}
	resp.CloneProgress, resp.CloneInProgress = s.locker.Status(dir)
	if isAlwaysCloningTest(repo) {
		resp.CloneInProgress = true
		resp.CloneProgress = "This will never finish cloning"
	}
	return &resp, nil
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
		result, err := s.repoCloneProgress(repoName)
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

func (s *Server) handleRepoDelete(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.deleteRepo(req.Repo); err != nil {
		log15.Error("failed to delete repository", "repo", req.Repo, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log15.Info("deleted repository", "repo", req.Repo)
}

func (s *Server) deleteRepo(repo api.RepoName) error {
	return s.removeRepoDirectory(s.dir(repo))
}
