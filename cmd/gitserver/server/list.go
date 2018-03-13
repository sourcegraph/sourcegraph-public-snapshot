package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

var (
	gitoliteBlacklists   map[string]*regexp.Regexp
	gitoliteBlacklistErr error
	gitoliteBlacklistMu  sync.Mutex
)

func init() {
	conf.Watch(func() {
		newBlacklists := make(map[string]*regexp.Regexp)
		for _, gconf := range conf.Get().Gitolite {
			if gconf.Blacklist == "" {
				continue
			}

			var err error
			newBlacklists[gconf.Host], err = regexp.Compile(gconf.Blacklist)
			if err != nil {
				gitoliteBlacklistErr = err
				log15.Error("Invalid regexp for Gitolite blacklist", "expr", gconf.Blacklist, "err", err)
				return
			}
		}
		gitoliteBlacklistMu.Lock()
		gitoliteBlacklists, gitoliteBlacklistErr = newBlacklists, nil
		gitoliteBlacklistMu.Unlock()
	})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	repos := make([]string, 0)

	q := r.URL.Query()
	query := func(name string) bool { _, ok := q[name]; return ok }
	switch {
	case r.URL.RawQuery == "":
		fallthrough // treat same as if the URL query was "gitolite" for backcompat
	case query("gitolite"):
		gitoliteHost := q.Get("gitolite")
		for _, gconf := range conf.Get().Gitolite {
			if gconf.Host != gitoliteHost {
				continue
			}
			rp, err := listGitoliteRepos(r.Context(), gconf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			repos = append(repos, rp...)
		}

	default:
		// empty list response for unrecognized URL query
	}

	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func listGitoliteRepos(ctx context.Context, gconf schema.GitoliteConnection) ([]string, error) {
	if gitoliteBlacklistErr != nil {
		return nil, gitoliteBlacklistErr
	}

	blacklist := gitoliteBlacklists[gconf.Host]

	out, err := exec.CommandContext(ctx, "ssh", gconf.Host, "info").CombinedOutput()
	if err != nil {
		log.Printf("listing gitolite failed: %s (Output: %q)", err, string(out))
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	blacklistCount := 0
	var repos []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "R" {
			name := fields[len(fields)-1]
			repoURI := gconf.Prefix + name

			if strings.Contains(repoURI, "..*") || (blacklist != nil && blacklist.MatchString(repoURI)) {
				blacklistCount++
				continue
			}
			repos = append(repos, repoURI)
		}
	}
	if blacklistCount > 0 {
		log15.Info("Excluded blacklisted Gitolite repositories", "num", blacklistCount)
	}

	return repos, nil
}
