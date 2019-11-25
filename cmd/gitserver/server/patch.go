package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var patchID uint64

func (s *Server) handleCreateCommitFromPatch(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateCommitFromPatchRequest
	var resp protocol.CreateCommitFromPatchResponse
	var status int

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := new(protocol.CreateCommitFromPatchResponse)
		resp.SetError("", "", "", errors.Wrap(err, "decoding CreateCommitFromPatchRequest"))
		status = http.StatusBadRequest
	} else {
		status, resp = s.createCommitFromPatch(r.Context(), req)
	}

	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) createCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (int, protocol.CreateCommitFromPatchResponse) {
	var resp protocol.CreateCommitFromPatchResponse

	repo := string(protocol.NormalizeRepo(req.Repo))
	repoGitDir := filepath.Join(s.ReposDir, repo, ".git")
	if _, err := os.Stat(repoGitDir); os.IsNotExist(err) {
		repoGitDir = filepath.Join(s.ReposDir, repo)
		if _, err := os.Stat(repoGitDir); os.IsNotExist(err) {
			resp.SetError(repo, "", "", errors.Wrap(err, "gitserver: repo does not exist"))
			return http.StatusInternalServerError, resp
		}
	}

	ref := req.TargetRef

	// Ensure tmp directory exists
	tmpRepoDir, err := s.tempDir("patch-repo-")
	if err != nil {
		resp.SetError(repo, "", "", errors.Wrap(err, "gitserver: make tmp repo"))
		return http.StatusInternalServerError, resp
	}
	defer cleanUpTmpRepo(tmpRepoDir)

	argsToString := func(args []string) string {
		return strings.Join(args, " ")
	}

	// Temporary logging command wrapper
	prefix := fmt.Sprintf("%d %s ", atomic.AddUint64(&patchID, 1), repo)
	run := func(cmd *exec.Cmd, reason string) ([]byte, error) {
		t := time.Now()
		out, err := cmd.CombinedOutput()
		if err != nil {
			resp.SetError(repo, argsToString(cmd.Args), string(out), errors.Wrap(err, "gitserver: "+reason))
			log15.Info("command failed", "prefix", prefix, "command", argsToString(cmd.Args), "duration", time.Since(t), "error", err, "output", string(out))
		} else {
			log15.Info("command ran successfully", "prefix", prefix, "command", argsToString(cmd.Args), "duration", time.Since(t), "output", string(out))
		}
		return out, err
	}

	tmpGitPathEnv := "GIT_DIR=" + filepath.Join(tmpRepoDir, ".git")

	tmpObjectsDir := filepath.Join(tmpRepoDir, ".git", "objects")
	repoObjectsDir := filepath.Join(repoGitDir, "objects")

	altObjectsEnv := "GIT_ALTERNATE_OBJECT_DIRECTORIES=" + repoObjectsDir

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv)

	if _, err := run(cmd, "init tmp repo"); err != nil {
		return http.StatusInternalServerError, resp
	}

	cmd = exec.CommandContext(ctx, "git", "reset", "-q", string(req.BaseCommit))
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)

	if out, err := run(cmd, "basing staging on base rev"); err != nil {
		log15.Error("Failed to base the temporary repo on the base revision.", "ref", req.TargetRef, "base", req.BaseCommit, "output", string(out))
		return http.StatusInternalServerError, resp
	}

	applyArgs := append([]string{"apply", "--cached"}, req.GitApplyArgs...)
	cmd = exec.CommandContext(ctx, "git", applyArgs...)
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)
	cmd.Stdin = strings.NewReader(req.Patch)

	if out, err := run(cmd, "applying patch"); err != nil {
		log15.Error("Failed to apply patch.", "ref", req.TargetRef, "output", string(out))
		return http.StatusInternalServerError, resp
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

	if out, err := run(cmd, "commiting patch"); err != nil {
		log15.Error("Failed to commit patch.", "ref", req.TargetRef, "output", out)
		return http.StatusInternalServerError, resp
	}

	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)

	// We don't use 'run' here as we only want stdout
	out, err := cmd.Output()
	if err != nil {
		resp.SetError(repo, argsToString(cmd.Args), string(out), errors.Wrap(err, "gitserver: retrieving new commit id"))
		return http.StatusInternalServerError, resp
	}
	cmtHash := strings.TrimSpace(string(out))

	// Move objects from tmpObjectsDir to repoObjectsDir.
	err = filepath.Walk(tmpObjectsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(tmpObjectsDir, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(repoObjectsDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
			return err
		}
		// do the actual move. If dst exists we can ignore the error since it
		// will contain the same content (content addressable FTW).
		if err := os.Rename(path, dst); err != nil && !os.IsExist(err) {
			return err
		}
		return nil
	})
	if err != nil {
		resp.SetError(repo, "", "", errors.Wrap(err, "copying git objects"))
		return http.StatusInternalServerError, resp
	}

	cmd = exec.CommandContext(ctx, "git", "update-ref", "--", req.TargetRef, cmtHash)
	cmd.Dir = repoGitDir

	if out, err = run(cmd, "creating ref"); err != nil {
		log15.Error("Failed to create ref for commit.", "ref", req.TargetRef, "commit", cmtHash, "output", string(out))
		return http.StatusInternalServerError, resp
	}

	if req.Push {
		remoteURL, err := repoRemoteURL(ctx, GitDir(repoGitDir))
		if err != nil {
			log15.Error("Failed to get remote URL", "ref", req.TargetRef, "commit", cmtHash, "err", err)
			resp.SetError(repo, "", "", errors.Wrap(err, "repoRemoteURL"))
			return http.StatusInternalServerError, resp
		}

		cmd = exec.CommandContext(ctx, "git", "push", "--force", remoteURL, fmt.Sprintf("%s:refs/heads/%s", req.TargetRef, req.TargetRef))
		cmd.Dir = repoGitDir

		if out, err = run(cmd, "pushing ref"); err != nil {
			log15.Error("Failed to push", "ref", req.TargetRef, "commit", cmtHash, "output", string(out))
			return http.StatusInternalServerError, resp
		}
	}

	resp.Rev = "refs/" + ref
	return http.StatusOK, resp
}

func cleanUpTmpRepo(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log15.Info("unable to clean up tmp repo", "path", path, "err", err)
	}
}
