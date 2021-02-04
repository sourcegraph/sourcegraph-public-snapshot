package servegit

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Serve struct {
	Addr  string
	Root  string
	Info  *log.Logger
	Debug *log.Logger
}

func (s *Serve) Start() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.Wrap(err, "listen")
	}

	// Update Addr to what listener actually used.
	s.Addr = ln.Addr().String()

	s.Info.Printf("listening on http://%s", s.Addr)
	s.Info.Printf("serving git repositories from %s", s.Root)

	if err := (&http.Server{Handler: s.handler()}).Serve(ln); err != nil {
		return errors.Wrap(err, "serving")
	}

	return nil
}

var indexHTML = template.Must(template.New("").Parse(`<html>
<head><title>src serve-git</title></head>
<body>
<h2>src serve-git</h2>
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

func (s *Serve) handler() http.Handler {
	mux := &http.ServeMux{}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := indexHTML.Execute(w, map[string]interface{}{
			"Explain": explainAddr(s.Addr),
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
		repos, err := s.Repos()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

	fs := http.FileServer(http.Dir(s.Root))
	svc := &gitServiceHandler{
		Dir: func(name string) string {
			return filepath.Join(s.Root, filepath.FromSlash(name))
		},
		Debug: s.Debug,
	}
	mux.Handle("/repos/", http.StripPrefix("/repos/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use git service if git is trying to clone. Otherwise show http.FileServer for convenience
		for _, suffix := range []string{"/info/refs", "/git-upload-pack"} {
			if strings.HasSuffix(r.URL.Path, suffix) {
				svc.ServeHTTP(w, r)
				return
			}
		}
		fs.ServeHTTP(w, r)
	})))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	})
}

// Repos returns a slice of all the git repositories it finds.
func (s *Serve) Repos() ([]Repo, error) {
	var repos []Repo
	var reposRootIsRepo bool

	root, err := filepath.EvalSymlinks(s.Root)
	if err != nil {
		s.Info.Printf("WARN: ignoring error searching %s: %v", root, err)
		return nil, nil
	}

	err = filepath.Walk(root, func(path string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			s.Info.Printf("WARN: ignoring error searching %s: %v", path, fileErr)
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
			s.Debug.Printf("not a repository root: %s", path)
			return nil
		}

		subpath, err := filepath.Rel(root, path)
		if err != nil {
			// According to WalkFunc docs, path is always filepath.Join(root,
			// subpath). So Rel should always work.
			s.Info.Fatalf("filepath.Walk returned %s which is not relative to %s: %v", path, root, err)
		}

		name := filepath.ToSlash(subpath)
		reposRootIsRepo = reposRootIsRepo || name == "."
		repos = append(repos, Repo{
			Name: name,
			URI:  pathpkg.Join("/repos", name),
		})

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
		return nil, err
	}

	if !reposRootIsRepo {
		return repos, nil
	}

	// Update all names to be relative to the parent of reposRoot. This is to
	// give a better name than "." for repos root
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to get the absolute path of reposRoot: %w", err)
	}
	rootName := filepath.Base(abs)
	for i := range repos {
		repos[i].Name = pathpkg.Join(rootName, repos[i].Name)
	}

	return repos, nil
}

func explainAddr(addr string) string {
	return fmt.Sprintf(`Serving the repositories at http://%s.

See https://docs.sourcegraph.com/admin/external_service/src_serve_git for
instructions to configure in Sourcegraph.
`, addr)
}
