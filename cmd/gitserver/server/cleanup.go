package server

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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

// CleanupRepos walks the repos directory and removes repositories that haven't been updated
// within a certain amount of time.
func (s *Server) CleanupRepos() {
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
			// Move the repo to a tmp directory before removing to prevent a
			// partially-deleted repo remaining in the event of a server-restart.
			tmp, err := ioutil.TempDir(s.ReposDir, "tmp-cleanup-")
			if err != nil {
				log15.Error("error removing inactive repo", "error", err, "repo", repoRoot)
				return
			}
			tmpRoot := path.Join(tmp, path.Base(repoRoot))
			err = os.Rename(repoRoot, tmpRoot)
			if err != nil {
				log15.Error("error removing inactive repo", "error", err, "repo", repoRoot)
				return
			}
			err = os.RemoveAll(tmp)
			if err != nil {
				log15.Error("error removing inactive repo", "error", err, "repo", repoRoot)
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
			uri := strings.TrimPrefix(repoRoot, s.ReposDir+"/")
			tmp, err := ioutil.TempDir(s.ReposDir, "tmp-clone-")
			tmpCloneRoot := path.Join(tmp, path.Base(repoRoot))
			if err != nil {
				log15.Error("error replacing repo", "error", err, "repo", repoRoot)
				return
			}
			defer os.RemoveAll(tmp)
			// Reclone the repo first to a tmp directory and hot-swap it into the old
			// repo's place when finished to minimize service disruption.
			//
			// TODO: this will not work for private repos which require authenticated
			// access.
			ctx, cancel := context.WithTimeout(context.Background(), longGitCommandTimeout)
			defer cancel()
			cmd := cloneCmd(ctx, OriginMap(uri), tmpCloneRoot)
			cloneLimiter.Acquire()
			defer cloneLimiter.Release()
			if output, err := cmd.CombinedOutput(); err != nil {
				log15.Error("reclone failed", "error", err, "repo", uri, "output", string(output))
				// Update the access time for the repo in the event of a clone failure.
				os.Chtimes(path.Join(repoRoot, "config"), time.Now(), time.Now())
				return
			}
			// Mark this repo as currently being cloned to prevent a race during the switchover.
			toDelete, err := ioutil.TempDir(s.ReposDir, "tmp-cleanup-")
			defer func() { go os.RemoveAll(toDelete) }()
			s.setCloneLock(repoRoot)
			defer s.releaseCloneLock(repoRoot)
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
