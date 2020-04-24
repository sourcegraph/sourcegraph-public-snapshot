package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
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
	h, err := reposHandler(logger, ln.Addr().String(), repoDir)
	if err != nil {
		return errors.Wrap(err, "configuring server")
	}
	s := &http.Server{
		Handler: h,
	}
	if err := s.Serve(ln); err != nil {
		return errors.Wrap(err, "serving")
	}

	return nil
}

var indexHTML = template.Must(template.New("").Parse(`<html>
<head><title>src-expose</title></head>
<body>
<h2>src-expose</h2>
<pre>
{{.Explain}}
<ul>{{range .Links}}
<li><a href="{{.}}">{{.}}</a></li>
{{- end}}
</ul>
</pre>
</body>
</html>`))

type Repo struct {
	Name string
	URI  string
}

func reposHandler(logger *log.Logger, addr, reposRoot string) (http.Handler, error) {
	logger.Printf("serving git repositories from %s", reposRoot)
	configureRepos(logger, reposRoot)

	// Start the HTTP server.
	mux := &http.ServeMux{}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := indexHTML.Execute(w, map[string]interface{}{
			"Explain": explainAddr(addr),
			"Links": []string{
				"/v1/list-repos",
				"/repos/",
			},
		})
		if err != nil {
			log.Println(err)
		}
	})

	mux.HandleFunc("/v1/list-repos", func(w http.ResponseWriter, r *http.Request) {
		var repos []Repo
		var reposRootIsRepo bool
		for _, name := range configureRepos(logger, reposRoot) {
			if name == "." {
				reposRootIsRepo = true
			}

			repos = append(repos, Repo{
				Name: name,
				URI:  path.Join("/repos", name),
			})
		}

		if reposRootIsRepo {
			// Update all names to be relative to the parent of
			// reposRoot. This is to give a better name than "." for repos
			// root
			abs, err := filepath.Abs(reposRoot)
			if err != nil {
				http.Error(w, "failed to get the absolute path of reposRoot: "+err.Error(), http.StatusInternalServerError)
				return
			}
			rootName := filepath.Base(abs)
			for i := range repos {
				repos[i].Name = path.Join(rootName, repos[i].Name)
			}
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

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/.git/objects/") { // exclude noisy path
			logger.Printf("%s %s", r.Method, r.URL.Path)
		}
		mux.ServeHTTP(w, r)
	}), nil
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

		// We recurse into bare repositories to find subprojects. Prevent
		// recursing into .git
		if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		// Check whether a particular directory is a repository or not.
		//
		// A directory which also is a repository (have .git folder inside it)
		// will contain nil error. If it does, proceed to configure.
		gitdir := filepath.Join(path, ".git")
		if fi, err := os.Stat(gitdir); err != nil || !fi.IsDir() {
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
		gitDirs = append(gitDirs, filepath.ToSlash(subpath))

		// Check whether a repository is a bare repository or not.
		//
		// If it yields false, which means it is a non-bare repository,
		// skip the directory so that it will not recurse to the subdirectories.
		// If it is a bare repository, proceed to recurse.
		c := exec.Command("git", "rev-parse", "--is-bare-repository")
		c.Dir = gitdir
		out, _ := c.CombinedOutput()

		if string(out) == "false\n" {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		// Our WalkFunc doesn't return any errors, so neither should filepath.Walk
		panic(err)
	}

	return gitDirs
}

const postUpdateHook = `#!/bin/sh
#
# Added by Sourcegraph src-expose serve

exec git update-server-info
`

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
	postUpdatePath := filepath.Join(gitDir, "hooks", "post-update")
	if _, err := os.Stat(postUpdatePath); err != nil {
		if err := os.MkdirAll(filepath.Dir(postUpdatePath), 0755); err != nil {
			return errors.Wrapf(err, "create git hooks dir: %s", out)
		}
		if err := ioutil.WriteFile(postUpdatePath, []byte(postUpdatePath), 0755); err != nil {
			return errors.Wrapf(err, "setting post-update hook: %s", out)
		}
	}
	return nil
}
