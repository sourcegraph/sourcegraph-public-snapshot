package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func (s *Server) repoInfo(ctx context.Context, repo api.RepoName) (*protocol.RepoInfo, error) {
	repo = protocol.NormalizeRepo(repo)
	dir := path.Join(s.ReposDir, string(repo))
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
		if strings.ToLower(string(repo)) == "github.com/sourcegraphtest/alwayscloningtest" {
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

// TODO(slimsag): Remove this after 3.3 is released.
func (s *Server) handleDeprecatedRepoInfo(w http.ResponseWriter, r *http.Request) {
	var req protocol.DeprecatedRepoInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	repo := protocol.NormalizeRepo(req.Repo)
	dir := path.Join(s.ReposDir, string(repo))

	resp := protocol.DeprecatedRepoInfoResponse{
		Cloned: repoCloned(dir),
	}
	if resp.Cloned {
		remoteURL, err := repoRemoteURL(r.Context(), dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.URL = remoteURL
	}
	{
		resp.CloneProgress, resp.CloneInProgress = s.locker.Status(dir)
		if strings.ToLower(string(req.Repo)) == "github.com/sourcegraphtest/alwayscloningtest" {
			resp.CloneInProgress = true
			resp.CloneProgress = "This will never finish cloning"
		}
	}
	if resp.Cloned {
		if mtime, err := repoLastFetched(dir); err != nil {
			log15.Warn("error computing last-fetched date", "repo", req.Repo, "err", err)
		} else {
			resp.LastFetched = &mtime
		}

		if cloneTime, err := getRecloneTime(dir); err != nil {
			log15.Warn("error getting reclone time", "repo", req.Repo, "err", err)
		} else {
			resp.CloneTime = &cloneTime
		}

		if lastChanged, err := repoLastChanged(dir); err != nil {
			log15.Warn("error getting last changed", "repo", req.Repo, "err", err)
		} else {
			resp.LastChanged = &lastChanged
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	repo = protocol.NormalizeRepo(repo)
	dir := filepath.Join(s.ReposDir, string(repo))

	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		// New style, so we just delete the .git dir
		dir = filepath.Join(dir, ".git")
	} else {
		// Old style, ensure it actually is a git dir so we don't delete
		// multiple repos. We do not need to change dir.
		if _, err := os.Stat(filepath.Join(dir, "HEAD")); err != nil {
			return err
		}
	}

	return s.removeRepoDirectory(dir)
}
