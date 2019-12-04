package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func serveRepos(logger *log.Logger, addr, repoDir string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "listen")
	}
	logger.Printf("listening on http://%s", ln.Addr())
	s, err := serve(logger, ln, repoDir)
	if err != nil {
		return errors.Wrap(err, "configuring server")
	}
	if err := s.Serve(ln); err != nil {
		return errors.Wrap(err, "serving")
	}

	return nil
}

func serve(logger *log.Logger, ln net.Listener, reposRoot string) (*http.Server, error) {
	configureRepos(logger, reposRoot)

	// Start the HTTP server.
	mux := &http.ServeMux{}

	mux.HandleFunc("/v1/list-repos", func(w http.ResponseWriter, r *http.Request) {
		type Repo struct {
			Name string
			URI  string
		}
		var repos []Repo
		for _, path := range configureRepos(logger, reposRoot) {
			uri := "/repos/" + path
			repos = append(repos, Repo{
				Name: path,
				URI:  uri,
			})
		}

		resp := struct {
			Items []Repo
		}{
			Items: repos,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(&resp)
	})

	mux.Handle("/repos/", http.StripPrefix("/repos/", http.FileServer(httpDir{http.Dir(reposRoot)})))

	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/.git/objects/") { // exclude noisy path
				logger.Printf("%s %s", r.Method, r.URL.Path)
			}
			mux.ServeHTTP(w, r)
		}),
	}
	return s, nil
}

type httpDir struct {
	http.Dir
}

// Wraps the http.Dir to inject subdir "/.git" to the path.
func (d httpDir) Open(name string) (http.File, error) {
	// Backwards compatibility for old config, skip if name already contains "/.git/".
	if !strings.Contains(name, "/.git/") {
		// Loops over subpaths that are requested by Git client to find the insert point.
		// The order of slice matters, must try to match "/objects/" before "/info/"
		// because there is a path "/objects/info/" exists.
		for _, sp := range []string{"/objects/", "/info/", "/HEAD"} {
			if i := strings.LastIndex(name, sp); i > 0 {
				name = name[:i] + "/.git" + name[i:]
				break
			}
		}
	}
	return d.Dir.Open(name)
}

// configureRepos finds all .git directories and configures them to be served.
// It returns a slice of all the git directories it finds. The paths are
// relative to root.
func configureRepos(logger *log.Logger, root string) []string {
	var gitDirs []string
	err := filepath.Walk(root, func(path string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			logger.Printf("error encountered on %s: %v", path, fileErr)
			return nil
		}
		if !fi.IsDir() {
			return nil
		}
		// stat now to avoid recursing into the rest of path
		gitdir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitdir); os.IsNotExist(err) {
			return nil
		}
		if err := configureOneRepo(logger, gitdir); err != nil {
			logger.Printf("configuring repo at %s: %v", gitdir, err)
			return nil
		}

		subpath, err := filepath.Rel(root, path)
		if err != nil {
			// According to WalkFunc docs, path is always filepath.Join(root,
			// subpath). So Rel should always work.
			logger.Fatalf("filepath.Walk returned %s which is not relative to %s: %v", path, root, err)
		}
		gitDirs = append(gitDirs, subpath)
		return filepath.SkipDir
	})
	if err != nil {
		// Our WalkFunc doesn't return any errors, so neither should filepath.Walk
		panic(err)
	}
	return gitDirs
}

// configureOneRepos tweaks a .git repo such that it can be git cloned.
// See https://theartofmachinery.com/2016/07/02/git_over_http.html
// for background.
func configureOneRepo(logger *log.Logger, gitDir string) error {
	c := exec.Command("git", "update-server-info")
	c.Dir = gitDir
	out, err := c.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "updating server info: %s", out)
	}
	if _, err := os.Stat(filepath.Join(gitDir, "hooks", "post-update")); err != nil {
		logger.Printf("setting post-update hook on %s", gitDir)
		c = exec.Command("mv", "hooks/post-update.sample", "hooks/post-update")
		c.Dir = gitDir
		out, err = c.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "setting post-update hook: %s", out)
		}
	}
	return nil
}
