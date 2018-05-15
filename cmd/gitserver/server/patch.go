package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func (s *Server) handleCreateCommitFromPatch(w http.ResponseWriter, r *http.Request) {
	s.patchMu.Lock()
	defer s.patchMu.Unlock()
	var req protocol.CreatePatchFromPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo := string(protocol.NormalizeRepo(req.Repo))
	realDir := path.Join(s.ReposDir, repo)

	ref := req.TargetRef

	// Ensure tmp directory exists
	tmpRepoDir, err := ioutil.TempDir(s.ReposDir, "tmp-repo-")
	if err != nil {
		http.Error(w, "gitserver: make tmp repo - "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cleanUpTmpRepo(tmpRepoDir)

	ctx := r.Context()

	tmpGitPathEnv := fmt.Sprintf("GIT_DIR=%s/.git", tmpRepoDir)

	tmpObjectsDir := fmt.Sprintf("%s/.git/objects", tmpRepoDir)
	realObjectsDir := fmt.Sprintf("%s/.git/objects", realDir)

	altObjectsEnv := fmt.Sprintf("GIT_ALTERNATE_OBJECT_DIRECTORIES=%s", realObjectsDir)

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv)

	if err := cmd.Run(); err != nil {
		http.Error(w, "gitserver: init tmp repo - "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmd = exec.CommandContext(ctx, "git", "reset", "-q", string(req.BaseCommit))
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)

	if out, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to base the temporary repo on the base revision.", "ref", req.TargetRef, "base", req.BaseCommit, "output", string(out))

		http.Error(w, "gitserver: basing staging on base rev - "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmd = exec.CommandContext(ctx, "git", "apply", "--cached")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)
	cmd.Stdin = strings.NewReader(req.Patch)

	if out, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to apply patch.", "ref", req.TargetRef, "output", string(out))

		http.Error(w, "gitserver: applying patch - "+err.Error(), http.StatusInternalServerError)
		return
	}

	message := req.CommitInfo.Message
	if message == "" {
		message = "<Sourcegraph> Creating commit from patch"
	}
	authorName := req.CommitInfo.AuthorName
	if authorName == "" {
		authorName = "Sourcegraph"
	}
	authorEmail := req.CommitInfo.AuthorEmail
	if authorEmail == "" {
		authorEmail = "support@sourcegraph.com"
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, []string{
		tmpGitPathEnv,
		altObjectsEnv,
		"GIT_COMMITTER_NAME=sourcegraph-committer",
		"GIT_COMMITTER_EMAIL=support@sourcegraph.com",
		fmt.Sprintf("GIT_AUTHOR_NAME=%s", authorName),
		fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", authorEmail),
		fmt.Sprintf("GIT_COMMITTER_DATE=%v", req.CommitInfo.Date),
		fmt.Sprintf("GIT_AUTHOR_DATE=%v", req.CommitInfo.Date),
	}...)

	if out, err := cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to commit patch.", "ref", req.TargetRef, "output", out)

		http.Error(w, "gitserver: commiting patch - "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)

	out, err := cmd.Output()
	if err != nil {
		http.Error(w, "gitserver: retrieving new commit id - "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmtHash := strings.TrimSpace(string(out))

	cmd = exec.CommandContext(ctx, "cp", "-R", tmpObjectsDir+"/", realObjectsDir)
	cmd.Dir = tmpRepoDir

	if err := cmd.Run(); err != nil {
		http.Error(w, "gitserver: copying git objects - "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmd = exec.CommandContext(ctx, "git", "update-ref", req.TargetRef, cmtHash)
	cmd.Dir = realDir

	if out, err = cmd.CombinedOutput(); err != nil {
		log15.Error("Failed to create ref for commit.", "ref", req.TargetRef, "commit", cmtHash, "output", string(out))

		http.Error(w, "gitserver: creating ref - "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendResp(w, "refs/"+ref)
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

func cleanUpTmpRepo(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Printf("unable to clean up tmp repo %s: %v", path, err)
	}
}
