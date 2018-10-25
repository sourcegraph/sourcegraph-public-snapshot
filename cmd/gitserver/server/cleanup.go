package server

import (
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"

	"github.com/prometheus/client_golang/prometheus"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	prometheus.MustRegister(reposRemoved)
	prometheus.MustRegister(reposRecloned)
}

// inactiveRepoTTL is the amount of time a repository will remain on a
// gitserver without being updated before it is removed.
const inactiveRepoTTL = time.Hour * 24 * 20
const repoTTL = time.Hour * 24 * 45

var reposRemoved = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "repos_removed",
	Help:      "number of repos removed during cleanup",
})
var reposRecloned = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "repos_recloned",
	Help:      "number of repos removed and recloned due to age",
})

// cleanupRepos walks the repos directory and performs maintenance tasks:
//
// 1. Remove corrupt repos.
// 2. Remove stale lock files.
// 3. Remove inactive repos on sourcegraph.com
// 4. Reclone repos after a while. (simulate git gc)
func (s *Server) cleanupRepos() {
	bCtx, bCancel := s.serverContext()
	defer bCancel()

	maybeRemoveCorrupt := func(gitDir string) (done bool, err error) {
		// We treat repositories missing HEAD to be corrupt. Both our cloning
		// and fetching ensure there is a HEAD file.
		_, err = os.Stat(filepath.Join(gitDir, "HEAD"))
		if !os.IsNotExist(err) {
			return false, err
		}

		log15.Info("removing corrupt repo", "repo", gitDir)
		if err := s.removeRepoDirectory(gitDir); err != nil {
			return true, err
		}
		reposRemoved.Inc()
		return true, nil
	}

	maybeRemoveInactive := func(gitDir string) (done bool, err error) {
		// We rewrite the HEAD file whenever we update a repo, and repos are
		// updated in response to user traffic. Check to see the last time
		// HEAD was rewritten to determine whether to consider this repo
		// inactive. Note: This is only accurate for installations which set
		// disableAutoGitUpdates=true. This is true for sourcegraph.com and
		// maybeRemoveInactive should only be run for sourcegraph.com
		head, err := os.Stat(filepath.Join(gitDir, "HEAD"))
		if err != nil {
			return false, err
		}
		lastUpdated := head.ModTime()
		if time.Since(lastUpdated) <= inactiveRepoTTL {
			return false, nil
		}

		log15.Info("removing inactive repo", "repo", gitDir)
		if err := s.removeRepoDirectory(gitDir); err != nil {
			return true, err
		}
		reposRemoved.Inc()
		return true, nil
	}

	ensureGitAttributes := func(gitDir string) (done bool, err error) {
		return false, setGitAttributes(gitDir)
	}

	maybeReclone := func(gitDir string) (done bool, err error) {
		recloneTime, err := getRecloneTime(gitDir)
		if err != nil {
			return false, err
		}

		// Add a jitter to spread out recloning of repos cloned at the same
		// time.
		if time.Since(recloneTime) <= repoTTL+randDuration(repoTTL/4) {
			return false, nil
		}

		ctx, cancel := context.WithTimeout(bCtx, longGitCommandTimeout)
		defer cancel()

		// name is the relative path to ReposDir, but without the .git suffix.
		repo := protocol.NormalizeRepo(api.RepoURI(strings.TrimPrefix(filepath.Dir(gitDir), s.ReposDir+"/")))
		log15.Info("recloning expired repo", "repo", repo)

		remoteURL := OriginMap(repo)
		if remoteURL == "" {
			var err error
			remoteURL, err = repoRemoteURL(ctx, gitDir)
			if err != nil {
				return false, errors.Wrap(err, "failed to get remote URL")
			}
		}

		if _, err := s.cloneRepo(ctx, repo, remoteURL, &cloneOptions{Block: true, Overwrite: true}); err != nil {
			return true, err
		}
		reposRecloned.Inc()
		return true, nil
	}

	removeStaleLocks := func(gitDir string) (done bool, err error) {
		// if removing a lock fails, we still want to try the other locks.
		var multi error

		// config.lock should be held for a very short amount of time.
		if err := removeFileOlderThan(filepath.Join(gitDir, "config.lock"), time.Minute); err != nil {
			multi = multierror.Append(multi, err)
		}
		// packed-refs can be held for quite a while, so we are conservative
		// with the age.
		if err := removeFileOlderThan(filepath.Join(gitDir, "packed-refs.lock"), time.Hour); err != nil {
			multi = multierror.Append(multi, err)
		}
		// we use the same conservative age for locks inside of refs
		if err := filepath.Walk(filepath.Join(gitDir, "refs"), func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				// ignore
				return nil
			}

			if fi.IsDir() {
				return nil
			}

			if !strings.HasSuffix(path, ".lock") {
				return nil
			}

			return removeFileOlderThan(path, time.Hour)
		}); err != nil {
			multi = multierror.Append(multi, err)
		}

		return false, multi
	}

	type cleanupFn struct {
		Name string
		Do   func(string) (bool, error)
	}
	cleanups := []cleanupFn{
		// Do some sanity checks on the repository.
		{"maybe remove corrupt", maybeRemoveCorrupt},
		// If git is interrupted it can leave lock files lying around. It does
		// not clean these up, and instead fails commands.
		{"remove stale locks", removeStaleLocks},
		// We always want to have the same git attributes file at
		// info/attributes.
		{"ensure git attributes", ensureGitAttributes},
	}
	if s.DeleteStaleRepositories {
		// Sourcegraph.com can potentially clone all of github.com, so we
		// delete repos which have not been used for a period of
		// time. s.DeleteStaleRepositories should only be true for
		// sourcegraph.com.
		cleanups = append(cleanups, cleanupFn{"maybe remove inactive", maybeRemoveInactive})
	}
	// Old git clones accumulate loose git objects that waste space and
	// slow down git operations. Periodically do a fresh clone to avoid
	// these problems. git gc is slow and resource intensive. It is
	// cheaper and faster to just reclone the repository.
	cleanups = append(cleanups, cleanupFn{"maybe reclone", maybeReclone})

	filepath.Walk(s.ReposDir, func(gitDir string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			return nil
		}

		if s.ignorePath(gitDir) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for $GIT_DIR
		if !fi.IsDir() || fi.Name() != ".git" {
			return nil
		}

		for _, cfn := range cleanups {
			done, err := cfn.Do(gitDir)
			if err != nil {
				log15.Error("error running cleanup command", "name", cfn.Name, "repo", gitDir, "error", err)
			}
			if done {
				break
			}
		}
		return filepath.SkipDir
	})
}

// removeRepoDirectory atomically removes a directory from s.ReposDir.
//
// It first moves the directory to a temporary location to avoid leaving
// partial state in the event of server restart or concurrent modifications to
// the directory.
//
// Additionally it removes parent empty directories up until s.ReposDir.
func (s *Server) removeRepoDirectory(dir string) error {
	// Rename out of the location so we can atomically stop using the repo.
	tmp, err := s.tempDir("delete-repo")
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
	current := dir
	for {
		parent := filepath.Dir(current)
		if parent == current {
			// This shouldn't happen, but protecting against escaping
			// ReposDir.
			break
		}
		current = parent
		info, err := os.Stat(current)
		if os.IsNotExist(err) {
			// Someone else beat us to it.
			break
		}
		if err != nil {
			log15.Warn("failed to stat parent directory", "dir", current, "error", err)
			return nil
		}
		if os.SameFile(rootInfo, info) {
			// Stop, we are at the parent.
			break
		}

		if err := os.Remove(current); err != nil {
			// Stop, we assume remove failed due to current not being empty.
			break
		}
	}

	// Delete the atomically renamed dir. We do this last since if it fails we
	// will rely on a janitor job to clean up for us.
	if err := os.RemoveAll(filepath.Join(tmp, "repo")); err != nil {
		log15.Warn("failed to cleanup after removing dir", "dir", dir, "error", err)
	}

	return nil
}

// cleanTmpFiles tries to remove tmp_pack_* files from .git/objects/pack.
// These files can be created by an interrupted fetch operation,
// and would be purged by `git gc --prune=now`, but `git gc` is
// very slow. Removing these files while they're in use will cause
// an operation to fail, but not damage the repository.
func (s *Server) cleanTmpFiles(dir string) {
	now := time.Now()
	packdir := filepath.Join(dir, ".git", "objects", "pack")
	err := filepath.Walk(packdir, func(path string, info os.FileInfo, err error) error {
		if path != packdir && info.IsDir() {
			return filepath.SkipDir
		}
		file := filepath.Base(path)
		if strings.HasPrefix(file, "tmp_pack_") {
			if now.Sub(info.ModTime()) > longGitCommandTimeout {
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		log15.Error("error removing tmp_pack_* files", "error", err)
	}
}

// SetupAndClearTmp sets up the the tempdir for ReposDir as well as clearing it
// out. It returns the temporary directory location.
func (s *Server) SetupAndClearTmp() (string, error) {
	// Additionally we create directories with the prefix .tmp-old which are
	// asynchronously removed. We do not remove in place since it may be a
	// slow operation to block on. Our tmp dir will be ${s.ReposDir}/.tmp
	dir := filepath.Join(s.ReposDir, tempDirName) // .tmp
	oldPrefix := tempDirName + "-old"
	if _, err := os.Stat(dir); err == nil {
		// Rename the current tmp file so we can asynchronously remove it. Use
		// a consistent pattern so if we get interrupted, we can clean it
		// another time.
		oldTmp, err := ioutil.TempDir(s.ReposDir, oldPrefix)
		if err != nil {
			return "", err
		}
		// oldTmp dir exists, so we need to use a child of oldTmp as the
		// rename target.
		if err := os.Rename(dir, filepath.Join(oldTmp, tempDirName)); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	// Asynchronously remove old temporary directories
	files, err := ioutil.ReadDir(s.ReposDir)
	if err != nil {
		log15.Error("failed to do tmp cleanup", "error", err)
	} else {
		for _, f := range files {
			// Remove older .tmp directories as well as our older tmp-
			// directories we would place into ReposDir. In September 2018 we
			// can remove support for removing tmp- directories.
			if !strings.HasPrefix(f.Name(), oldPrefix) && !strings.HasPrefix(f.Name(), "tmp-") {
				continue
			}
			go func(path string) {
				if err := os.RemoveAll(path); err != nil {
					log15.Error("failed to remove old temporary directory", "path", path, "error", err)
				}
			}(filepath.Join(s.ReposDir, f.Name()))
		}
	}

	return dir, nil
}

// getRecloneTime returns an approximate time a repository is cloned. If the
// value is not stored in the repository, the reclone time for the repository
// is set to now.
func getRecloneTime(gitDir string) (time.Time, error) {
	// We store the time we recloned the repository. If the value is missing,
	// we store the current time. This decouples this timestamp from the
	// different ways a clone can appear in gitserver.
	update := func() (time.Time, error) {
		now := time.Now()
		cmd := exec.Command("git", "config", "--add", "sourcegraph.recloneTimestamp", strconv.FormatInt(time.Now().Unix(), 10))
		cmd.Dir = gitDir
		if _, err := cmd.Output(); err != nil {
			return now, errors.Wrap(wrapCmdError(cmd, err), "failed to update recloneTimestamp")
		}
		return now, nil
	}

	cmd := exec.Command("git", "config", "--get", "sourcegraph.recloneTimestamp")
	cmd.Dir = gitDir
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 means the key is not set.
		if ee, ok := err.(*exec.ExitError); ok && ee.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
			return update()
		}
		return time.Unix(0, 0), errors.Wrap(wrapCmdError(cmd, err), "failed to determine clone timestamp")
	}

	sec, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 0)
	if err != nil {
		// If the value is bad update it to the current time
		now, err2 := update()
		if err2 != nil {
			err = err2
		}
		return now, err
	}

	return time.Unix(sec, 0), nil
}

// randDuration returns a psuedo-random duration between [0, d)
func randDuration(d time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(d)))
}

// wrapCmdError will wrap errors for cmd to include the arguments. If the error
// is an exec.ExitError and cmd was invoked with Output(), it will also include
// the captured stderr.
func wrapCmdError(cmd *exec.Cmd, err error) error {
	if err == nil {
		return nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return errors.Wrapf(err, "%s %s failed with stderr: %s", cmd.Path, strings.Join(cmd.Args, " "), string(ee.Stderr))
	}
	return errors.Wrapf(err, "%s %s failed", cmd.Path, strings.Join(cmd.Args, " "))
}

// removeFileOlderThan removes path if its mtime is older than maxAge. If the
// file is missing, no error is returned.
func removeFileOlderThan(path string, maxAge time.Duration) error {
	fi, err := os.Stat(filepath.Join(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	age := time.Since(fi.ModTime())
	if age < maxAge {
		return nil
	}

	log15.Debug("removing stale lock file", "path", path, "age", age)
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
