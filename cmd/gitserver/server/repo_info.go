package server

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func (s *Server) handleRepoInfo(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	repo := protocol.NormalizeRepo(req.Repo)
	dir := path.Join(s.ReposDir, string(repo))

	resp := protocol.RepoInfoResponse{
		Cloned: repoCloned(dir),
		URL:    OriginMap(repo),
	}
	if resp.Cloned && resp.URL == "" {
		remoteURL, err := repoRemoteURL(r.Context(), dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.URL = remoteURL
	}
	{
		s.cloningMu.Lock()
		_, resp.CloneInProgress = s.cloning[dir]
		s.cloningMu.Unlock()
		if strings.ToLower(string(req.Repo)) == "github.com/sourcegraphtest/alwayscloningtest" {
			resp.CloneInProgress = true
		}
	}
	if resp.Cloned {
		if mtime, err := repoLastFetched(dir); err != nil {
			log15.Warn("error computing last-fetched date", "repo", req.Repo, "err", err)
		} else {
			resp.LastFetched = &mtime
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
