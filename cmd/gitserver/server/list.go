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

	// These two cases are for backcompat (can be removed in 3.4)
	case r.URL.RawQuery == "":
		fallthrough // treat same as if the URL query was "gitolite" for backcompat
	case query("gitolite"):
		defaultGitolite.listRepos(ctx, q.Get("gitolite"), w)
		return

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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_446(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
