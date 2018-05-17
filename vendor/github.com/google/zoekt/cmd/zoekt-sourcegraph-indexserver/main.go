// Command zoekt-sourcegraph-indexserver periodically reindexes enabled
// repositories on sourcegraph
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/net/trace"

	"github.com/google/zoekt"
	"github.com/google/zoekt/build"
)

// Server is the main functionality of zoekt-sourcegraph-indexserver. It
// exists to conveniently use all the options passed in via func main.
type Server struct {
	// Root is the base URL for the Sourcegraph instance to index. Normally
	// http://sourcegraph-frontend-internal or http://localhost:3090.
	Root *url.URL

	// IndexDir is the index directory to use.
	IndexDir string

	// Interval is how often we sync with Sourcegraph.
	Interval time.Duration

	// CPUCount is the amount of parallelism to use when indexing a
	// repository.
	CPUCount int

	// Debug when true will output extra debug logs.
	Debug bool
}

func (s *Server) loggedRun(tr trace.Trace, cmd *exec.Cmd) error {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = errOut

	tr.LazyPrintf("%s", cmd.Args)
	if err := cmd.Run(); err != nil {
		outS := out.String()
		errS := errOut.String()
		tr.LazyPrintf("failed: %v", err)
		tr.LazyPrintf("stdout: %s", outS)
		tr.LazyPrintf("stderr: %s", errS)
		tr.SetError()
		return fmt.Errorf("command %s failed: %v\nOUT: %s\nERR: %s",
			cmd.Args, err, outS, errS)
	}
	tr.LazyPrintf("success")
	if s.Debug {
		log.Printf("ran successfully %s", cmd.Args)
	}
	return nil
}

// Run the sync loop. This blocks forever.
func (s *Server) Run() {
	queue := &Queue{}

	// Start a goroutine which updates the queue with commits to index.
	go func() {
		t := time.NewTicker(s.Interval)
		for {
			repos, err := listRepos(s.Root)
			if err != nil {
				log.Println(err)
				<-t.C
				continue
			}

			log.Printf("updating index queue with %d repositories", len(repos))

			// ResolveRevision is IO bound on the gitserver service. So we do
			// them concurrently.
			sem := newSemaphore(32)
			tr := trace.New("resolveRevisions", "")
			tr.LazyPrintf("resolving HEAD for %d repos", len(repos))
			for _, name := range repos {
				sem.Acquire()
				go func(name string) {
					defer sem.Release()
					commit, err := resolveRevision(s.Root, name, "HEAD")
					if err != nil && !os.IsNotExist(err) {
						tr.LazyPrintf("failed resolving HEAD for %v: %v", name, err)
						tr.SetError()
						return
					}
					queue.AddOrUpdate(name, commit)
				}(name)
			}
			sem.Wait()
			tr.Finish()

			// Only delete shards if we found repositories, to prevent strange
			// bugs in responses causing us to delete everything.
			if len(repos) > 0 {
				exists := make(map[string]bool)
				for _, name := range repos {
					exists[name] = true
				}
				s.deleteStaleIndexes(exists)
			}

			<-t.C
		}
	}()

	// In the current goroutine process the queue forever.
	for {
		name, commit, ok := queue.Pop()
		if !ok {
			time.Sleep(time.Second)
			continue
		}

		err := s.Index(name, commit)
		if err != nil {
			log.Printf("error indexing %s@%s: %s", name, commit, err)
			continue
		}
		queue.SetIndexed(name, commit)
	}
}

// Index starts an index job for repo name at commit.
func (s *Server) Index(name, commit string) error {
	tr := trace.New("index", name)
	defer tr.Finish()

	tr.LazyPrintf("commit: %v", commit)

	if commit == "" {
		return s.createEmptyShard(tr, name)
	}

	cmd := exec.Command("zoekt-archive-index",
		fmt.Sprintf("-parallelism=%d", s.CPUCount),
		"-index", s.IndexDir,
		"-incremental",
		"-branch", "HEAD",
		"-commit", commit,
		"-name", name,
		tarballURL(s.Root, name, commit))
	// Prevent prompting
	cmd.Stdin = &bytes.Buffer{}
	return s.loggedRun(tr, cmd)
}

func (s *Server) createEmptyShard(tr trace.Trace, name string) error {
	cmd := exec.Command("zoekt-archive-index",
		"-index", s.IndexDir,
		"-incremental",
		"-branch", "HEAD",
		// dummy commit
		"-commit", "404aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"-name", name,
		"-")
	// Empty archive
	cmd.Stdin = bytes.NewBuffer(bytes.Repeat([]byte{0}, 1024))
	return s.loggedRun(tr, cmd)
}

func (s *Server) deleteStaleIndexes(exists map[string]bool) {
	expr := s.IndexDir + "/*"
	fs, err := filepath.Glob(expr)
	if err != nil {
		log.Printf("Glob(%q): %v", expr, err)
	}

	for _, f := range fs {
		if err := deleteIfStale(exists, f); err != nil {
			log.Printf("deleteIfStale(%q): %v", f, err)
		}
	}
}

var repoTmpl = template.Must(template.New("name").Parse(`
<html><body>
<a href="debug/requests">Traces</a><br>
{{.IndexMsg}}<br />
<br />
<h3>Re-index repository</h3>
<form action="/" method="post">
{{range .Repos}}
<input type="submit" name="repo" value="{{ . }}" /> <br />
{{end}}
</form>
</body></html>
`))

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/debug/requests" {
		trace.Traces(w, r)
		return
	}

	var data struct {
		Repos    []string
		IndexMsg string
	}

	if r.Method == "POST" {
		r.ParseForm()
		name := r.Form.Get("repo")
		index := func() error {
			commit, err := resolveRevision(s.Root, name, "HEAD")
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			return s.Index(name, commit)
		}
		err := index()
		if err != nil {
			data.IndexMsg = fmt.Sprintf("Indexing %s failed: %s", name, err)
		} else {
			data.IndexMsg = "Indexed " + name
		}
	}

	var err error
	data.Repos, err = listRepos(s.Root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repoTmpl.Execute(w, data)
}

func listRepos(root *url.URL) ([]string, error) {
	u := root.ResolveReference(&url.URL{Path: "/.internal/repos/list"})
	resp, err := http.Post(u.String(), "application/json; charset=utf8", bytes.NewReader([]byte(`{"Enabled": true}`)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list repositories: status %s", resp.Status)
	}

	var data []struct {
		URI string
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	repos := make([]string, len(data))
	for i, r := range data {
		repos[i] = r.URI
	}
	return repos, nil
}

func resolveRevision(root *url.URL, repo, spec string) (string, error) {
	u := root.ResolveReference(&url.URL{Path: fmt.Sprintf("/.internal/git/%s/resolve-revision/%s", repo, spec)})
	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", os.ErrNotExist
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to resolve revision %s@%s: status %s", repo, spec, resp.Status)
	}

	var b bytes.Buffer
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func tarballURL(root *url.URL, repo, commit string) string {
	return root.ResolveReference(&url.URL{Path: fmt.Sprintf("/.internal/git/%s/tar/%s", repo, commit)}).String()
}

// deleteIfStale deletes the shard if its corresponding repo name is not in
// exists.
func deleteIfStale(exists map[string]bool, fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return nil
	}
	defer f.Close()

	ifile, err := zoekt.NewIndexFile(f)
	if err != nil {
		return nil
	}
	defer ifile.Close()

	repo, _, err := zoekt.ReadMetadata(ifile)
	if err != nil {
		return nil
	}

	if !exists[repo.Name] {
		log.Printf("%s no longer exists, deleting %s", repo.Name, fn)
		return os.Remove(fn)
	}

	return nil
}

func main() {
	root := flag.String("sourcegraph_url", "", "http://sourcegraph-frontend-internal or http://localhost:3090")
	interval := flag.Duration("interval", 10*time.Minute, "sync with sourcegraph this often")
	index := flag.String("index", build.DefaultDir, "set index directory to use")
	listen := flag.String("listen", "", "listen on this address.")
	cpuFraction := flag.Float64("cpu_fraction", 0.25,
		"use this fraction of the cores for indexing.")
	debug := flag.Bool("debug", false,
		"turn on more verbose logging.")
	flag.Parse()

	if *cpuFraction <= 0.0 || *cpuFraction > 1.0 {
		log.Fatal("cpu_fraction must be between 0.0 and 1.0")
	}
	if *index == "" {
		log.Fatal("must set -index")
	}
	if *root == "" {
		log.Fatal("must set -sourcegraph_url")
	}
	rootURL, err := url.Parse(*root)
	if err != nil {
		log.Fatalf("url.Parse(%v): %v", *root, err)
	}

	// Automatically prepend our own path at the front, to minimize
	// required configuration.
	if l, err := os.Readlink("/proc/self/exe"); err == nil {
		os.Setenv("PATH", filepath.Dir(l)+":"+os.Getenv("PATH"))
	}

	if _, err := os.Stat(*index); err != nil {
		if err := os.MkdirAll(*index, 0755); err != nil {
			log.Fatalf("MkdirAll %s: %v", *index, err)
		}
	}

	cpuCount := int(math.Round(float64(runtime.NumCPU()) * (*cpuFraction)))
	if cpuCount < 1 {
		cpuCount = 1
	}
	s := &Server{
		Root:     rootURL,
		IndexDir: *index,
		Interval: *interval,
		CPUCount: cpuCount,
		Debug:    *debug,
	}

	if *listen != "" {
		go func() {
			trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
				return true, true
			}
			log.Printf("serving HTTP on %s", *listen)
			log.Fatal(http.ListenAndServe(*listen, s))
		}()
	}

	s.Run()
}
