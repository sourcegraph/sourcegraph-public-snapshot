package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
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
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRepoDelete(w http.ResponseWriter, r *http.Request) {
	var req protocol.RepoInfoRequest
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

func (s *Server) deleteRepo(repo api.RepoURI) error {
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

	// Rename out of the location so we can atomically stop using the repo.
	tmp, err := ioutil.TempDir(s.ReposDir, "tmp-delete")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := os.Rename(dir, filepath.Join(tmp, "repo")); err != nil {
		return err
	}

	// Everything after this point is just cleanup, so any error that occurs
	// should not be returned, just logged.

	// Cleanup empty parent directories. We just attempt to remove and if we
	// have a failure we assume it's due to the directory having other
	// children. If we checked first we could race with someone else adding a
	// new clone.
	rootInfo, err := os.Stat(s.ReposDir)
	if err != nil {
		log15.Warn("Failed to stat ReposDir", "error", err)
		return nil
	}
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// This shouldn't happen, but protecting against escaping
			// ReposDir.
			break
		}
		dir = parent
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			// Someone else beat us to it.
			break
		}
		if err != nil {
			log15.Warn("failed to stat parent directory", "dir", dir, "error", err)
			return nil
		}
		if os.SameFile(rootInfo, info) {
			// Stop, we are at the parent.
			break
		}

		if err := os.Remove(dir); err != nil {
			// Stop, we assume remove failed due to dir not being empty.
			break
		}
	}

	// Delete the atomically renamed dir. We do this last since if it fails we
	// will rely on a janitor job to clean up for us.
	if err := os.RemoveAll(tmp); err != nil {
		log15.Warn("failed to cleanup after removing repo", "repo", repo, "error", err)
	}

	return nil
}
