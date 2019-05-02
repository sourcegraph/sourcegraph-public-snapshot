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
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

func main() {
	log.SetPrefix("")
	n := flag.Int("n", 1, "number of instances of each repo to make")
	addr := flag.String("addr", ":0", "address on which to serve (default picks an unused port)")
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: fakehub [opts] path/to/dir/containing/git/dirs

fakehub will serve any number (controlled with -n) of copies of the repo over
HTTP at /repo/1/.git, /repo/2/.git etc. These can be git cloned, and they can
be used as test data for sourcegraph. The easiest way to get them into
sourcegraph is to visit http://127.0.0.1:3434/config and paste the contents
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
		log.Fatalf("fakehub: %v", err)
	}
	if err := fakehub(*n, ln, repoDir); err != nil {
		log.Fatalf("fakehub: %v", err)
	}
}

func fakehub(n int, ln net.Listener, reposRoot string) error {
	gitDirs, err := configureRepos(reposRoot)
	if err != nil {
		return errors.Wrapf(err, "configuring repos under %s", reposRoot)
	}

	// Set up the template vars for pages.
	var relDirs []string
	for _, gd := range gitDirs {
		rd, err := filepath.Rel(reposRoot, gd)
		if err != nil {
			return errors.Wrap(err, "getting relative path of git dir")
		}
		relDirs = append(relDirs, rd)
	}
	addr := ln.Addr().String()
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	// Replace [::] with 127.0.0.1 because [::] breaks double-click link opening in Mac terminals.
	ipv6Rx, err := regexp.Compile(`^\[::\]`)
	if err != nil {
		return errors.Wrap(err, "compiling regexp for IPV6 localhost")
	}
	addr = ipv6Rx.ReplaceAllString(addr, "127.0.0.1")
	tvars := &templateVars{n, relDirs, addr}

	// Start the HTTP server.
	mux := &http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleDefault(tvars, w, r)
	})

	if n == 1 {
		mux.Handle("/repos/", http.StripPrefix("/repos/", http.FileServer(http.Dir(reposRoot))))
	} else {
		for i := 1; i <= n; i++ {
			pfx := fmt.Sprintf("/repos/%d/", i)
			mux.Handle(pfx, http.StripPrefix(pfx, http.FileServer(http.Dir(reposRoot))))
		}
	}

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		handleConfig(tvars, w, r)
	})
	s := http.Server{
		Addr:    addr,
		Handler: logger(mux),
	}
	log.Printf("listening on http://%s", s.Addr)
	return s.Serve(ln)
}

// configureRepos finds all .git directories and configures them to be served.
// It returns a slice of all the git directories it finds.
func configureRepos(root string) ([]string, error) {
	var gitDirs []string
	err := filepath.Walk(root, func(path string, fi os.FileInfo, fileErr error) error {
		if !(filepath.Base(path) == ".git" && fi.IsDir()) {
			return nil
		}
		if fileErr != nil {
			log.Printf("error encountered on %s: %v", path, fileErr)
			return nil
		}
		if err := configureOneRepo(path); err != nil {
			log.Printf("configuring repo at %s: %v", path, err)
			return nil
		}
		gitDirs = append(gitDirs, path)
		return nil
	})
	if err != nil {
		return gitDirs, errors.Wrap(err, "configuring repos and gathering git dirs")
	}
	return gitDirs, err
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
func handleDefault(tvars *templateVars, w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprintf(w, "%v", err.Error())
	}
}

// handleConfig shows the config for pasting into sourcegraph.
func handleConfig(tvars *templateVars, w http.ResponseWriter, r *http.Request) {
	t1 := `// Paste this into Site admin | External services | Add external service | Single Git repositories:
{
  "url": "http://127.0.0.1{{.Addr}}",
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
		fmt.Fprintf(w, "%v", err.Error())
	}
}

type templateVars struct {
	n       int
	RelDirs []string
	Addr    string
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
