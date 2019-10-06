package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"gopkg.in/inconshreveable/log15.v2"
)

func (s *Server) createCommitFromPatchTmpDir(ctx context.Context, repo, repoGitDir string, req protocol.CreateCommitFromPatchRequest) (commit string, err error) {
	// Ensure tmp directory exists
	tmpRepoDir, err := s.tempDir("patch-repo-")
	if err != nil {
		return "", errors.Wrap(err, "gitserver: make tmp repo")
	}
	defer cleanUpTmpRepo(tmpRepoDir)

	// Temporary logging command wrapper
	prefix := fmt.Sprintf("%d %s ", atomic.AddUint64(&patchID, 1), repo)
	run := func(cmd *exec.Cmd) ([]byte, error) {
		t := time.Now()
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("%scommand %s failed (%v): %v\nOUT: %s",
				prefix, cmd.Args, time.Since(t), err, string(out))
		} else {
			log.Printf("%sran successfully %s (%v)\nOUT: %s", prefix, cmd.Args, time.Since(t), string(out))
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

	if _, err := run(cmd); err != nil {
		return "", errors.Wrap(err, "gitserver: init tmp repo")
	}

	cmd = exec.CommandContext(ctx, "git", "reset", "-q", string(req.BaseCommit))
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)

	if out, err := run(cmd); err != nil {
		log15.Error("Failed to base the temporary repo on the base revision.", "ref", req.TargetRef, "base", req.BaseCommit, "output", string(out))
		return "", errors.Wrap(err, "gitserver: basing staging on base rev")
	}

	cmd = exec.CommandContext(ctx, "git", "apply", "--cached")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)
	cmd.Stdin = strings.NewReader(req.Patch)

	if out, err := run(cmd); err != nil {
		log15.Error("Failed to apply patch.", "ref", req.TargetRef, "output", string(out))
		return "", errors.Wrap(err, "gitserver: applying patch")
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", req.CommitInfo.Message)
	setGitEnvForCommit(cmd, req)
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)
	cmd.Dir = tmpRepoDir

	if out, err := run(cmd); err != nil {
		log15.Error("Failed to commit patch.", "ref", req.TargetRef, "output", out)
		return "", errors.Wrap(err, "gitserver: commiting patch")
	}

	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = tmpRepoDir
	cmd.Env = append(cmd.Env, tmpGitPathEnv, altObjectsEnv)

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "gitserver: retrieving new commit id")
	}

	commit = strings.TrimSpace(string(out))

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
		return "", errors.Wrap(err, "gitserver: copying git objects")
	}
	return commit, nil
}

func cleanUpTmpRepo(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Printf("unable to clean up tmp repo %s: %v", path, err)
	}
}
