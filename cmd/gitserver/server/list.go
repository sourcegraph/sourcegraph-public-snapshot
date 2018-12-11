package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repos := make([]string, 0)

	q := r.URL.Query()
	query := func(name string) bool { _, ok := q[name]; return ok }
	switch {
	case r.URL.RawQuery == "":
		fallthrough // treat same as if the URL query was "gitolite" for backcompat
	case query("gitolite"):
		gitoliteHost := q.Get("gitolite")

		config, err := conf.GitoliteConfigs(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, gconf := range config {
			if gconf.Host != gitoliteHost {
				continue
			}
			rp, err := listGitoliteRepos(ctx, gconf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			repos = append(repos, rp...)
		}

	case query("cloned"):
		err := filepath.Walk(s.ReposDir, func(path string, info os.FileInfo, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if s.ignorePath(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if err != nil {
				return nil
			}

			// We only care about directories
			if !info.IsDir() {
				return nil
			}

			// New style git directory layout
			if filepath.Base(path) == ".git" {
				name, err := filepath.Rel(s.ReposDir, filepath.Dir(path))
				if err != nil {
					return err
				}
				repos = append(repos, name)
				return filepath.SkipDir
			}

			// For old-style directory layouts we need to do an extra extra
			// stat to check if this is a repo.
			if _, err := os.Stat(filepath.Join(path, "HEAD")); os.IsNotExist(err) {
				// HEAD doesn't exist, so keep recursing
				return nil
			} else if err != nil {
				return err
			}

			// path is an old style git repo since it contains HEAD
			name, err := filepath.Rel(s.ReposDir, path)
			if err != nil {
				return err
			}
			repos = append(repos, name)
			return filepath.SkipDir
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		// empty list response for unrecognized URL query
	}

	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func blacklistRegexp(gconf *schema.GitoliteConnection) (*regexp.Regexp, error) {
	if gconf.Blacklist == "" {
		return nil, nil
	}
	return regexp.Compile(gconf.Blacklist)
}

func listGitoliteRepos(ctx context.Context, gconf *schema.GitoliteConnection) ([]string, error) {
	blacklist, err := blacklistRegexp(gconf)
	if err != nil {
		log15.Error("Invalid regexp for Gitolite blacklist", "expr", gconf.Blacklist, "err", err)
		return nil, err
	}

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
			repoName := gconf.Prefix + name

			// Gitolite's internal rules for what a regex looks like exclude `+` from
			// consideration because of `gtk+`. The character list here is derived from
			// Gitolite's `$REPOPAT_PATT`. Note that even when these characters would
			// not have special meaning to a regex engine, Gitolite will treat them as
			// proof that a string is a pattern, not a literal name.
			if strings.ContainsAny(repoName, "\\^$|()[]*?{},") || (blacklist != nil && blacklist.MatchString(repoName)) {
				blacklistCount++
				continue
			}
			repos = append(repos, repoName)
		}
	}
	if blacklistCount > 0 {
		log15.Info("Excluded blacklisted Gitolite repositories", "num", blacklistCount)
	}

	return repos, nil
}
