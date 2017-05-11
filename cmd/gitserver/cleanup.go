package main

import (
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
}

// inactiveRepoTTL is the amount of time a repository will remain on a gitserver without being
// updated before it is removed.
const inactiveRepoTTL = time.Hour * 24 * 30

var reposRemoved = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "repos_removed",
	Help:      "number of repos removed during cleanup due to inactivity",
})

// removeInactiveRepos walks the repos directory and removes repositories that haven't been updated
// within a certain amount of time.
func removeInactiveRepos(reposDir string) {
	filepath.Walk(reposDir, func(p string, fi os.FileInfo, fileErr error) (rtnErr error) {
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
		lastUpdated := fi.ModTime()
		if time.Since(lastUpdated) > inactiveRepoTTL {
			repoRoot := path.Dir(p)
			log15.Info("removing inactive repo", "repo", repoRoot)
			// Move the repo to a tmp directory before removing to prevent a
			// partially-deleted repo remaining in the event of a server-restart.
			tmp, err := ioutil.TempDir(reposDir, "tmp-cleanup-")
			if err != nil {
				log15.Error("error removing inactive repo", "error", err, "repo", repoRoot)
				return
			}
			tmpRepoRoot := path.Join(tmp, path.Base(repoRoot))
			err = os.Rename(repoRoot, tmpRepoRoot)
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
		return
	})
}
