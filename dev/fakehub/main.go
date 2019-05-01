// Command fakehub serves multiple instances of a local git repository over HTTP,
// along with a pastable config for easier manual testing of sourcegraph.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

func main() {
	log.SetPrefix("")
	n := flag.Int("n", 10, "number of repos to specify in config")
	addr := flag.String("addr", ":9847", "address on which to serve")
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: fakehub [options] path/to/git/repo

fakehub will serve any number (controlled with -n) of copies of the repo over HTTP at /repo/1/.git, /repo/2/.git etc.
These can be git cloned, and they can be used as test data for sourcegraph. The easiest way to get them into
sourcegraph is to visit /config and paste the contents into the text box for adding single repos in sourcegraph
Site Admin.
`)
		flag.PrintDefaults()
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	repoDir := flag.Arg(0)
	if err := fakehub(*n, *addr, repoDir); err != nil {
		log.Fatalf("fakehub: %v", err)
	}
}

func fakehub(n int, addr, repoDir string) error {
	// Set up a copy of the repo in a temp dir, configuring the copy such that it
	// can be git cloned. See https://theartofmachinery.com/2016/07/02/git_over_http.html
	// for the idea behind this.
	d, err := ioutil.TempDir("", "fakehub")
	repoDir2 := filepath.Join(d, filepath.Base(repoDir))
	if err != nil {
		return errors.Wrap(err, "creating temp dir to contain template repo")
	}
	out, err := exec.Command("git", "--bare", "clone", repoDir, repoDir2).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "copying repo to temp dir: %s", out)
	}
	c := exec.Command("git", "--bare", "update-server-info")
	c.Dir = filepath.Join(repoDir2, ".git")
	out, err = c.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "updating server info: %s", out)
	}
	c = exec.Command("mv", "hooks/post-update.sample", "hooks/post-update")
	c.Dir = filepath.Join(repoDir2, ".git")
	out, err = c.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "setting post-update hook")
	}

	// Start the HTTP server.
	tvars := &templateVars{n, addr}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleDefault(tvars, w, r)
	})
	for i := 1; i <= n; i++ {
		pfx := fmt.Sprintf("/repo/%d/", i)
		mux.Handle(pfx, http.StripPrefix(pfx, http.FileServer(http.Dir(repoDir2))))
	}
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		handleConfig(tvars, w, r)
	})
	s := http.Server{
		Addr:    addr,
		Handler: logger(mux),
	}
	log.Printf("listening on http://localhost%s", s.Addr)
	return s.ListenAndServe()
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
<a href="/config">config</a>
<div>Example: git clone http://localhost{{.Addr}}/repo/1/.git</div>
<div>Repos:</div>
<div>
	{{range .Nums}}
		<a href="/repo/{{.}}">/repo/{{.}}</a>
	{{end}}
</div>
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
  "url": "http://localhost{{.Addr}}",
  "repos": [
	{{range .Nums}}
      "/repo/{{.}}/.git",
    {{end}}
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
	n    int
	Addr string
}

func (tv *templateVars) Nums() []int {
	var nums []int
	for i := 1; i <= tv.n; i++ {
		nums = append(nums, i)
	}
	return nums
}
