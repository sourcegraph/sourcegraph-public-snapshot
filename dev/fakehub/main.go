// Command fakehub serves git repositories within some directory over HTTP,
// along with a pastable config for easier manual testing of sourcegraph.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

func main() {
	log.SetPrefix("")
	n := flag.Int("n", 1, "number of instances of each repo to make")
	addr := flag.String("addr", "127.0.0.1:3434", "address on which to serve (end with : for unused port)")
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: fakehub [opts] path/to/dir/containing/git/dirs

fakehub will serve any number (controlled with -n) of copies of the repo over
HTTP at /repo/1/.git, /repo/2/.git etc. These can be git cloned, and they can
be used as test data for sourcegraph. The easiest way to get them into
sourcegraph is to visit the URL printed out on startup and paste the contents
into the text box for adding single repos in sourcegraph Site Admin.
`)
		flag.PrintDefaults()
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	repoDir := flag.Arg(0)
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("fakehub: listening: %v", err)
	}
	log.Printf("listening on http://%s", ln.Addr())
	s, err := fakehub(*n, ln, repoDir)
	if err != nil {
		log.Fatalf("fakehub: configuring server :%v", err)
	}
	if err := s.Serve(ln); err != nil {
		log.Fatalf("fakehub: serving: %v", err)
	}
}

func fakehub(n int, ln net.Listener, reposRoot string) (*http.Server, error) {
	configureRepos(reposRoot)

	// Start the HTTP server.
	mux := &http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tvars := &templateVars{n, configureRepos(reposRoot), ln.Addr()}

		handleConfig(tvars, w)
	})

	if n == 1 {
		mux.Handle("/repos/", http.StripPrefix("/repos/", http.FileServer(httpDir{http.Dir(reposRoot)})))
	} else {
		for i := 1; i <= n; i++ {
			pfx := fmt.Sprintf("/repos/%d/", i)
			mux.Handle(pfx, http.StripPrefix(pfx, http.FileServer(httpDir{http.Dir(reposRoot)})))
		}
	}

	s := &http.Server{
		Handler: logger(mux),
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
func configureRepos(root string) []string {
	var gitDirs []string
	err := filepath.Walk(root, func(path string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			log.Printf("error encountered on %s: %v", path, fileErr)
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
		if err := configureOneRepo(gitdir); err != nil {
			log.Printf("configuring repo at %s: %v", gitdir, err)
			return nil
		}

		subpath, err := filepath.Rel(root, path)
		if err != nil {
			// According to WalkFunc docs, path is always filepath.Join(root,
			// subpath). So Rel should always work.
			log.Fatalf("filepath.Walk returned %s which is not relative to %s: %v", path, root, err)
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
func configureOneRepo(gitDir string) error {
	c := exec.Command("git", "update-server-info")
	c.Dir = gitDir
	out, err := c.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "updating server info: %s", out)
	}
	if _, err := os.Stat(filepath.Join(gitDir, "hooks", "post-update")); err != nil {
		log.Printf("setting post-update hook on %s", gitDir)
		c = exec.Command("mv", "hooks/post-update.sample", "hooks/post-update")
		c.Dir = gitDir
		out, err = c.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "setting post-update hook: %s", out)
		}
	}
	return nil
}

// logger converts the given handler to one that will first log every request.
func logger(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

// handleDefault shows the root page with links to config and repos.
func handleDefault(tvars *templateVars, w http.ResponseWriter) {
	t1 := `
<p><a href="/config">config</a></p>
{{if .Repos}}
{{range .Repos}}
<div><a href="{{.}}">{{.}}</a></div>{{end}}
{{else}}
<div>No git repos found.</div>
{{end}}
`
	err := func() error {
		t2, err := template.New("linkspage").Parse(t1)
		if err != nil {
			return errors.Wrap(err, "parsing template for links page")
		}
		if err := t2.Execute(w, tvars); err != nil {
			return errors.Wrap(err, "executing links page template")
		}
		return nil
	}()
	if err != nil {
		log.Println(err)
		_, _ = w.Write([]byte(err.Error()))
	}
}

// handleConfig shows the config for pasting into sourcegraph.
func handleConfig(tvars *templateVars, w http.ResponseWriter) {
	t1 := `// Paste this into Site admin | External services | Add external service | Single Git repositories:
{
  "url": "http://{{.Addr}}",
  "repos": [{{range .Repos}}
      "{{.}}",{{end}}
  ]
}
`
	err := func() error {
		t2, err := template.New("config").Parse(t1)
		if err != nil {
			return errors.Wrap(err, "parsing config template")
		}
		if err := t2.Execute(w, tvars); err != nil {
			return errors.Wrap(err, "executing config template")
		}
		return nil
	}()
	if err != nil {
		log.Println(err)
		_, _ = w.Write([]byte(err.Error()))
	}
}

type templateVars struct {
	n       int
	RelDirs []string
	Addr    net.Addr
}

// Repos returns a slice of URL paths for all the repos, including any copies.
func (tv *templateVars) Repos() []string {
	var paths []string
	if tv.n == 1 {
		for _, rd := range tv.RelDirs {
			paths = append(paths, "/repos/"+rd)
		}
	} else {
		for i := 1; i <= tv.n; i++ {
			for _, rd := range tv.RelDirs {
				paths = append(paths, fmt.Sprint("/repos/", i, "/", rd))
			}
		}
	}
	return paths
}
