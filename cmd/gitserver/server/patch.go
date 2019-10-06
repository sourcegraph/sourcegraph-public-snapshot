package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var patchID uint64

var useInPlaceCommitCreationFromPatch, _ = strconv.ParseBool(os.Getenv("USE_IN_PLACE_COMMIT_CREATION_FROM_PATCH"))

func (s *Server) handleCreateCommitFromPatch(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateCommitFromPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo := string(protocol.NormalizeRepo(req.Repo))
	repoGitDir := filepath.Join(s.ReposDir, repo, ".git")
	if _, err := os.Stat(repoGitDir); os.IsNotExist(err) {
		repoGitDir = filepath.Join(s.ReposDir, repo)
		if _, err := os.Stat(repoGitDir); os.IsNotExist(err) {
			http.Error(w, "gitserver: repo does not exist - "+err.Error(), http.StatusInternalServerError)
		}
	}

	ctx := r.Context()

	if req.CommitInfo.Message == "" {
		req.CommitInfo.Message = "<Sourcegraph> Creating commit from patch"
	}

	var (
		commit string
		err    error
	)
	if useInPlaceCommitCreationFromPatch {
		commit, err = createCommitFromPatch(ctx, repoGitDir, req)
	} else {
		commit, err = s.createCommitFromPatchTmpDir(ctx, repo, repoGitDir, req)
	}
	if err != nil {
		http.Error(w, "gitserver: CreateCommitFromPatch - "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmd := exec.CommandContext(ctx, "git", "update-ref", "--", req.TargetRef, commit)
	cmd.Dir = repoGitDir

	if out, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to create ref for commit.", "ref", req.TargetRef, "commit", commit, "output", string(out))
		http.Error(w, "gitserver: creating ref - "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendResp(w, "refs/"+req.TargetRef)
}

func sendResp(w http.ResponseWriter, commitID string) {
	resp := protocol.CreatePatchFromPatchResponse{
		Rev: commitID,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func setGitEnvForCommit(cmd *exec.Cmd, req protocol.CreateCommitFromPatchRequest) {
	authorName := req.CommitInfo.AuthorName
	if authorName == "" {
		authorName = "Sourcegraph"
	}
	authorEmail := req.CommitInfo.AuthorEmail
	if authorEmail == "" {
		authorEmail = "support@sourcegraph.com"
	}
	cmd.Env = append(cmd.Env,
		"GIT_COMMITTER_NAME=sourcegraph-committer",
		"GIT_COMMITTER_EMAIL=support@sourcegraph.com",
		fmt.Sprintf("GIT_AUTHOR_NAME=%s", authorName),
		fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", authorEmail),
		fmt.Sprintf("GIT_COMMITTER_DATE=%v", req.CommitInfo.Date),
		fmt.Sprintf("GIT_AUTHOR_DATE=%v", req.CommitInfo.Date),
	)
}
