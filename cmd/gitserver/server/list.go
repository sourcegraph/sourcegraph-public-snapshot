package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repos := make([]string, 0)

	q := r.URL.Query()
	query := func(name string) bool { _, ok := q[name]; return ok }
	switch {

	// These two cases are for backcompat (can be removed in 3.4)
	case r.URL.RawQuery == "":
		fallthrough // treat same as if the URL query was "gitolite" for backcompat
	case query("gitolite"):
		defaultGitolite.listRepos(ctx, q.Get("gitolite"), w)
		return

	case query("cloned"):
		err := s.walkCloned(ctx, func(path string) error {
			name, err := filepath.Rel(s.ReposDir, path)
			if err != nil {
				return err
			}
			repos = append(repos, name)
			return nil
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

func (s *Server) walkCloned(ctx context.Context, handle func(path string) error) error {
	return filepath.Walk(s.ReposDir, func(path string, info os.FileInfo, err error) error {
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
			err := handle(filepath.Dir(path))
			if err != nil {
				return err
			}
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
		err = handle(path)
		if err != nil {
			return err
		}
		return filepath.SkipDir
	})
}
