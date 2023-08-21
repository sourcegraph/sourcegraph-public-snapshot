package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Server) handleReposStats(w http.ResponseWriter, r *http.Request) {
	b, err := s.readReposStatsFile(filepath.Join(s.ReposDir, reposStatsName))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read %s: %v", reposStatsName, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(b)
}

func (s *Server) readReposStatsFile(filePath string) ([]byte, error) {
	b, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		// When a gitserver is new this file might not have been computed
		// yet. Clients are expected to handle this case by noticing UpdatedAt
		// is not set.
		b = []byte("{}")
	} else if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Server) repoCloneProgress(ctx context.Context, repo api.RepoName) *protocol.RepoCloneProgress {
	dir := s.dir(repo)
	resp := protocol.RepoCloneProgress{
		Cloned: repoCloned(dir),
	}
	resp.CloneProgress, resp.CloneInProgress = RepoCloningStatus(ctx, s.DB, repo)
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
		result := s.repoCloneProgress(r.Context(), repoName)
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
	err := removeRepoDirectory(ctx, s.Logger, s.DB, s.Hostname, s.ReposDir, s.dir(api.UndeletedRepoName(repo)), true)
	if err != nil {
		return errors.Wrap(err, "removing repo directory")
	}
	err = s.setCloneStatus(ctx, repo, types.CloneStatusNotCloned)
	if err != nil {
		return errors.Wrap(err, "setting clone status after delete")
	}
	return nil
}

func RepoCloningStatus(ctx context.Context, db database.DB, repoName api.RepoName) (status string, cloning bool) {
	paginationArgs := &database.PaginationArgs{
		First:     pointers.Ptr(1),
		OrderBy:   []database.OrderByOption{{Field: "repo_update_jobs.queued_at"}},
		Ascending: false,
	}
	opts := database.ListRepoUpdateJobOpts{RepoName: repoName, States: []string{"processing"}, PaginationArgs: paginationArgs}
	repoUpdateJobs, err := db.RepoUpdateJobs().List(ctx, opts)
	// If there is an error or there are no processing jobs for a given repo --
	// return empty string and false.
	if err != nil || len(repoUpdateJobs) != 1 {
		return
	}
	// If we have a status -- we're cloning and not fetching.
	status = repoUpdateJobs[0].CloningProgress
	cloning = status != ""
	return
}
