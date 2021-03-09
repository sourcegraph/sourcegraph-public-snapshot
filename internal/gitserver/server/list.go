package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/karrick/godirwalk"
)

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repos := make([]string, 0)

	q := r.URL.Query()
	query := func(name string) bool { _, ok := q[name]; return ok }
	switch {
	case query("cloned"):
		err := godirwalk.Walk(s.ReposDir, &godirwalk.Options{
			Callback: func(path string, de *godirwalk.Dirent) error {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				if s.ignorePath(path) {
					if de.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				// We only care about directories
				if !de.IsDir() {
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
			},
			ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
				// Ignore errors and simply continue with other nodes
				return godirwalk.SkipNode
			},
			Unsorted: true,
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
