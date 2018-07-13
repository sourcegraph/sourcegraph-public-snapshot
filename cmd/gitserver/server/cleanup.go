package server

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"

	"github.com/prometheus/client_golang/prometheus"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	prometheus.MustRegister(reposRemoved)
	prometheus.MustRegister(reposRecloned)
}

// inactiveRepoTTL is the amount of time a repository will remain on a gitserver without being
// updated before it is removed.
const inactiveRepoTTL = time.Hour * 24 * 30
const repoTTL = time.Hour * 24 * 45

var reposRemoved = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "repos_removed",
	Help:      "number of repos removed during cleanup due to inactivity",
})
var reposRecloned = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "repos_recloned",
	Help:      "number of repos removed and recloned due to age",
})

// cleanupRepos walks the repos directory and removes repositories that haven't been updated
// within a certain amount of time.
func (s *Server) cleanupRepos() {
	bCtx, bCancel := s.serverContext()
	defer bCancel()

	filepath.Walk(s.ReposDir, func(gitDir string, fi os.FileInfo, fileErr error) (rtnErr error) {
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

		rtnErr = filepath.SkipDir
		// We rewrite the HEAD file whenever we update a repo, and repos are updated
		// in response to user traffic. Check to see the last time HEAD was rewritten
		// to determine whether to consider this repo inactive.
		head, err := os.Stat(filepath.Join(gitDir, "HEAD"))
		if err != nil {
			log15.Warn("GIT_DIR missing HEAD", "path", gitDir, "error", err)
			return
		}
		lastUpdated := head.ModTime()
		if time.Since(lastUpdated) > inactiveRepoTTL {
			log15.Info("removing inactive repo", "repo", gitDir)
			if err := s.removeAll(gitDir); err != nil {
				log15.Error("error removing inactive repo", "repo", gitDir, "error", err)
				return
			}
			reposRemoved.Inc()
			return
		}

		// Old git clones accumulate loose git objects that waste space and
		// slow down git operations. Periodically do a fresh clone to avoid
		// these problems. git gc is slow and resource intensive. It is
		// cheaper and faster to just reclone the repository.
		//
		// The config file of a repo shouldn't be edited after the repo is
		// cloned for the first time, so its modification time is used to
		// check for the repo's original clone date.
		cfg, err := os.Stat(filepath.Join(gitDir, "config"))
		if err != nil {
			return
		}
		initTime := cfg.ModTime()
		if time.Since(initTime) > repoTTL {
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
					log15.Error("error getting remote URL", "error", err, "repo", repo)
					return
				}
			}

			if _, err := s.cloneRepo(ctx, repo, remoteURL, &cloneOptions{Block: true, Overwrite: true}); err != nil {
				log15.Error("reclone failed", "repo", repo, "error", err)
				return
			}
			reposRecloned.Inc()
			return
		}
		return
	})
}

// removeAll removes the entire directory.
// It first moves the directory to a temporary location
// to avoid leaving partial state in the event of server
// restart or concurrent modifications to the directory.
func (s *Server) removeAll(dir string) error {
	tmpDir, err := s.tempDir("removeAll-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	tmpName := strings.Replace(dir, string(os.PathSeparator), "-", -1)
	tmpRoot := path.Join(tmpDir, tmpName)
	if err := os.Rename(dir, tmpRoot); err != nil {
		return err
	}
	return os.RemoveAll(tmpRoot)
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
	return
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
	go func() {
		files, err := ioutil.ReadDir(s.ReposDir)
		if err != nil {
			log15.Error("failed to do tmp cleanup", "error", err)
			return
		}

		for _, f := range files {
			// Remove older .tmp directories as well as our older tmp-
			// directories we would place into ReposDir. In September 2018 we
			// can remove support for removing tmp- directories.
			if !strings.HasPrefix(f.Name(), oldPrefix) && !strings.HasPrefix(f.Name(), "tmp-") {
				continue
			}
			path := filepath.Join(s.ReposDir, f.Name())
			if err := os.RemoveAll(path); err != nil {
				log15.Error("failed to remove old temporary directory", "path", path, "error", err)
			}
		}
	}()

	return dir, nil
}
