package server

import (
	"context"
	"io/ioutil"
	"log"
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

	filepath.Walk(s.ReposDir, func(p string, fi os.FileInfo, fileErr error) (rtnErr error) {
		if fileErr != nil {
			return nil
		}

		// Find each git repo root by looking for its HEAD file.
		if fi.IsDir() || !strings.HasSuffix(p, "/HEAD") {
			return nil
		}

		rtnErr = filepath.SkipDir
		// We rewrite the HEAD file whenever we update a repo, and repos are updated
		// in response to user traffic. Check to see the last time HEAD was rewritten
		// to determine whether to consider this repo inactive.
		repoRoot := path.Dir(p)
		lastUpdated := fi.ModTime()
		if time.Since(lastUpdated) > inactiveRepoTTL {
			log15.Info("removing inactive repo", "repo", repoRoot)
			if err := s.removeAll(repoRoot); err != nil {
				log15.Error("error removing inactive repo", "repo", repoRoot, "error", err)
				return
			}
			reposRemoved.Inc()
			return
		}

		// The config file of a repo shouldn't be edited after the repo is cloned for
		// the first time, so use this to check for the repo's original clone date
		// and remove ones that are past a certain age to prevent accumulation of
		// loose git objects.
		cfg, err := os.Stat(path.Join(repoRoot, "config"))
		if err != nil {
			return
		}
		initTime := cfg.ModTime()
		if time.Since(initTime) > repoTTL {
			log15.Info("recloning expired repo", "repo", repoRoot)
			uri := protocol.NormalizeRepo(api.RepoURI(strings.TrimPrefix(repoRoot, s.ReposDir+"/")))
			tmp, err := ioutil.TempDir(s.ReposDir, "tmp-clone-")
			tmpCloneRoot := path.Join(tmp, path.Base(repoRoot))
			if err != nil {
				log15.Error("error replacing repo", "error", err, "repo", repoRoot)
				return
			}
			defer os.RemoveAll(tmp)

			ctx, cancel1, err := s.acquireCloneLimiter(bCtx)
			if err != nil {
				log.Println("unexpected error while acquiring clone limiter:", err)
				return
			}
			defer cancel1()
			ctx, cancel2 := context.WithTimeout(ctx, longGitCommandTimeout)
			defer cancel2()

			remoteURL := OriginMap(uri)
			if remoteURL == "" {
				var err error
				remoteURL, err = repoRemoteURL(ctx, repoRoot)
				if err != nil {
					log15.Error("error getting remote URL", "error", err, "repo", repoRoot)
					return
				}
			}

			// Reclone the repo first to a tmp directory and hot-swap it into the old
			// repo's place when finished to minimize service disruption.
			//
			// TODO: this will not work for private repos which require authenticated
			// access.
			cmd := cloneCmd(ctx, remoteURL, tmpCloneRoot, false)
			if output, err := cmd.CombinedOutput(); err != nil {
				log15.Error("reclone failed", "error", err, "repo", uri, "output", string(output))
				// Update the access time for the repo in the event of a clone failure.
				os.Chtimes(path.Join(repoRoot, "config"), time.Now(), time.Now())
				return
			}
			// Mark this repo as currently being cloned to prevent a race during the switchover.
			toDelete, err := ioutil.TempDir(s.ReposDir, "tmp-cleanup-")
			defer func() { go os.RemoveAll(toDelete) }()
			lock, ok := s.locker.TryAcquire(repoRoot, "recloning")
			if !ok {
				log15.Warn("failed acquire repo lock when replacing repo", "repo", repoRoot)
				return
			}
			defer lock.Release()
			if err != nil {
				log15.Error("error replacing repo", "error", err, "repo", repoRoot)
				return
			}
			err = os.Rename(repoRoot, path.Join(toDelete, path.Base(repoRoot)))
			if err != nil {
				log15.Error("error replacing repo", "error", err, "repo", repoRoot)
				return
			}
			err = os.Rename(tmpCloneRoot, repoRoot)
			if err != nil {
				log15.Error("error replacing repo", "error", err, "repo", repoRoot)
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
	tmpDir, err := ioutil.TempDir(s.ReposDir, "tmp-cleanup-")
	if err != nil {
		return err
	}
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
