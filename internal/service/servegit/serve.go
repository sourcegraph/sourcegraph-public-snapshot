package servegit

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/exp/slices"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/fastwalk"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/gitservice"
)

type ServeConfig struct {
	env.BaseConfig

	Addr string

	Timeout  time.Duration
	MaxDepth int
}

func (c *ServeConfig) Load() {
	url, err := url.Parse(c.Get("SRC_SERVE_GIT_URL", "http://127.0.0.1:3434", "URL that servegit should listen on."))
	if err != nil {
		c.AddError(errors.Wrapf(err, "failed to parse SRC_SERVE_GIT_URL"))
	} else if url.Scheme != "http" {
		c.AddError(errors.Errorf("only support http scheme for SRC_SERVE_GIT_URL got scheme %q", url.Scheme))
	} else {
		c.Addr = url.Host
	}

	c.Timeout = c.GetInterval("SRC_DISCOVER_TIMEOUT", "5s", "The maximum amount of time we spend looking for repositories.")
	c.MaxDepth = c.GetInt("SRC_DISCOVER_MAX_DEPTH", "10", "The maximum depth we will recurse when discovery for repositories.")
}

type Serve struct {
	ServeConfig

	Logger log.Logger
}

func (s *Serve) Start() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.Wrap(err, "listen")
	}

	// Update Addr to what listener actually used.
	s.Addr = ln.Addr().String()

	s.Logger.Info("serving git repositories", log.String("url", "http://"+s.Addr))

	srv := &http.Server{Handler: s.handler()}

	// We have opened the listener, now start serving connections in the
	// background.
	go func() {
		if err := srv.Serve(ln); err == http.ErrServerClosed {
			s.Logger.Info("http serve closed")
		} else {
			s.Logger.Error("http serve failed", log.Error(err))
		}
	}()

	// Also listen for shutdown signals in the background. We don't need
	// graceful shutdown since this only runs in app and the only clients of
	// the server will also be shutdown at the same time.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		<-c
		if err := srv.Close(); err != nil {
			s.Logger.Error("failed to Close http serve", log.Error(err))
		}
	}()

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
	Name        string
	URI         string
	ClonePath   string
	AbsFilePath string
}

func (s *Serve) handler() http.Handler {
	mux := &http.ServeMux{}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := indexHTML.Execute(w, map[string]interface{}{
			"Explain": explainAddr(s.Addr),
			"Links": []string{
				"/v1/list-repos-for-path",
				"/repos/",
			},
		})
		if err != nil {
			s.Logger.Debug("failed to return / response", log.Error(err))
		}
	})

	mux.HandleFunc("/v1/list-repos-for-path", func(w http.ResponseWriter, r *http.Request) {
		var req ListReposRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		repos, err := s.Repos(req.Root)
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

	svc := &gitservice.Handler{
		Dir: func(name string) string {
			// The cloneURL we generate is an absolute path. But gitservice
			// returns the name with the leading / missing. So we add it in before
			// calling FromSlash.
			return filepath.FromSlash("/" + name)
		},
		ErrorHook: func(err error, stderr string) {
			s.Logger.Error("git-service error", log.Error(err), log.String("stderr", stderr))
		},
		Trace: func(ctx context.Context, svc, repo, protocol string) func(error) {
			start := time.Now()
			return func(err error) {
				s.Logger.Debug("git service", log.String("svc", svc), log.String("protocol", protocol), log.String("repo", repo), log.Duration("duration", time.Since(start)), log.Error(err))
			}
		},
	}
	mux.Handle("/repos/", http.StripPrefix("/repos/", svc))

	return http.HandlerFunc(mux.ServeHTTP)
}

// Checks if git thinks the given path is a valid .git folder for a repository
func isBareRepo(path string) bool {
	c := exec.Command("git", "--git-dir", path, "rev-parse", "--is-bare-repository")
	c.Dir = path
	out, err := c.CombinedOutput()

	if err != nil {
		return false
	}

	return string(out) != "false\n"
}

// Check if git thinks the given path is a proper git checkout
func isGitRepo(path string) bool {
	// Executing git rev-parse --git-dir in the root of a worktree returns .git
	c := exec.Command("git", "rev-parse", "--git-dir")
	c.Dir = path
	out, err := c.CombinedOutput()

	if err != nil {
		return false
	}

	return string(out) == ".git\n"
}

// Returns a string of the git remote if it exists
func gitRemote(path string) string {
	// Executing git rev-parse --git-dir in the root of a worktree returns .git
	c := exec.Command("git", "remote", "get-url", "origin")
	c.Dir = path
	out, err := c.CombinedOutput()

	if err != nil {
		return ""
	}

	return convertGitCloneURLToCodebaseName(string(out))
}

// Converts a git clone URL to the codebase name that includes the slash-separated code host, owner, and repository name
// This should captures:
// - "github:sourcegraph/sourcegraph" a common SSH host alias
// - "https://github.com/sourcegraph/deploy-sourcegraph-k8s.git"
// - "git@github.com:sourcegraph/sourcegraph.git"
func convertGitCloneURLToCodebaseName(cloneURL string) string {
	cloneURL = strings.TrimSpace(cloneURL)
	if cloneURL == "" {
		return ""
	}
	uri, err := url.Parse(strings.Replace(cloneURL, "git@", "", 1))
	if err != nil {
		return ""
	}
	// Handle common Git SSH URL format
	match := regexp.MustCompile(`git@([^:]+):([\w-]+)\/([\w-]+)(\.git)?`).FindStringSubmatch(cloneURL)
	if strings.HasPrefix(cloneURL, "git@") && len(match) > 0 {
		host := match[1]
		owner := match[2]
		repo := match[3]
		return host + "/" + owner + "/" + repo
	}

	buildName := func(prefix string, uri *url.URL) string {
		name := uri.Path
		if name == "" {
			name = uri.Opaque
		}
		return prefix + strings.TrimSuffix(name, ".git")
	}

	// Handle GitHub URLs
	if strings.HasPrefix(uri.Scheme, "github") || strings.HasPrefix(uri.String(), "github") {
		return buildName("github.com/", uri)
	}
	// Handle GitLab URLs
	if strings.HasPrefix(uri.Scheme, "gitlab") || strings.HasPrefix(uri.String(), "gitlab") {
		return buildName("gitlab.com/", uri)
	}
	// Handle HTTPS URLs
	if strings.HasPrefix(uri.Scheme, "http") && uri.Host != "" && uri.Path != "" {
		return buildName(uri.Host, uri)
	}
	// Generic URL
	if uri.Host != "" && uri.Path != "" {
		return buildName(uri.Host, uri)
	}
	return ""
}

// Repos returns a slice of all the git repositories it finds. It is a wrapper
// around Walk which removes the need to deal with channels and sorts the
// response.
func (s *Serve) Repos(root string) ([]Repo, error) {
	var (
		repoC   = make(chan Repo, 4) // 4 is the same buffer size used in fastwalk
		walkErr error
	)
	go func() {
		defer close(repoC)
		walkErr = s.Walk(root, repoC)
	}()

	var repos []Repo
	for r := range repoC {
		repos = append(repos, r)
	}

	if walkErr != nil {
		return nil, walkErr
	}

	// walk is not deterministic due to concurrency, so introduce determinism
	// by sorting the results.
	slices.SortFunc(repos, func(a, b Repo) bool {
		return a.Name < b.Name
	})

	return repos, nil
}

// Walk is the core repos finding routine.
func (s *Serve) Walk(root string, repoC chan<- Repo) error {
	if root == "" {
		s.Logger.Warn("root path cannot be searched if it is not an absolute path", log.String("path", root))
		return nil
	}

	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		s.Logger.Warn("ignoring error searching", log.String("path", root), log.Error(err))
		return nil
	}
	root = filepath.Clean(root)

	if repo, ok, err := rootIsRepo(root); err != nil {
		return err
	} else if ok {
		repoC <- repo
		return nil
	}

	ctx := context.Background()
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}

	ignore := mkIgnoreSubPath(root, s.MaxDepth)

	// We use fastwalk since it is much faster. Notes for people used to
	// filepath.WalkDir:
	//
	//   - func is called concurrently
	//   - you can return fastwalk.ErrSkipFiles to avoid calling func on
	//     files (so will only get dirs)
	//   - filepath.SkipDir has the same meaning
	err = fastwalk.Walk(root, func(path string, typ os.FileMode) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !typ.IsDir() {
			return fastwalk.ErrSkipFiles
		}

		subpath, err := filepath.Rel(root, path)
		if err != nil {
			// According to WalkFunc docs, path is always filepath.Join(root,
			// subpath). So Rel should always work.
			return errors.Wrapf(err, "filepath.Walk returned %s which is not relative to %s", path, root)
		}

		if ignore(subpath) {
			s.Logger.Debug("ignoring path", log.String("path", path))
			return filepath.SkipDir
		}

		// Check whether a particular directory is a repository or not.
		//
		// Valid paths are either bare repositories or git worktrees.
		isBare := isBareRepo(path)
		isGit := isGitRepo(path)

		if !isGit && !isBare {
			s.Logger.Debug("not a repository root", log.String("path", path))
			return fastwalk.ErrSkipFiles
		}

		name := filepath.ToSlash(subpath)
		cloneURI := pathpkg.Join("/repos", filepath.ToSlash(path))
		clonePath := cloneURI

		// Regular git repos won't clone without the full path to the .git directory.
		if isGit {
			clonePath += "/.git"
		}

		// Use the remote as the name of repo if it exists
		remote := gitRemote(path)
		if remote != "" {
			name = remote
		}
		repoC <- Repo{
			Name:        name,
			URI:         cloneURI,
			ClonePath:   clonePath,
			AbsFilePath: path,
		}

		// At this point we know the directory is either a git repo or a bare git repo,
		// we don't need to recurse further to save time.
		// TODO: Look into whether it is useful to support git submodules
		return filepath.SkipDir
	})

	// If we timed out return what we found without an error
	if errors.Is(err, context.DeadlineExceeded) {
		err = nil
		s.Logger.Warn("stopped discovering repos since reached timeout", log.String("root", root), log.Duration("timeout", s.Timeout))
	}

	return err
}

// rootIsRepo is a special case when the root of our search is a repository.
func rootIsRepo(root string) (Repo, bool, error) {
	isBare := isBareRepo(root)
	isGit := isGitRepo(root)
	if !isGit && !isBare {
		return Repo{}, false, nil
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return Repo{}, false, errors.Errorf("failed to get the absolute path of reposRoot: %w", err)
	}

	cloneURI := pathpkg.Join("/repos", filepath.ToSlash(root))
	clonePath := cloneURI

	// Regular git repos won't clone without the full path to the .git directory.
	if isGit {
		clonePath += "/.git"
	}
	name := filepath.Base(abs)
	// Use the remote as the name if it exists
	remote := gitRemote(root)
	if remote != "" {
		name = remote
	}

	return Repo{
		Name:        name,
		URI:         cloneURI,
		ClonePath:   clonePath,
		AbsFilePath: abs,
	}, true, nil
}

// mkIgnoreSubPath which acts on subpaths to root. It returns true if the
// subpath should be ignored.
func mkIgnoreSubPath(root string, maxDepth int) func(string) bool {
	// A list of dirs which cause us trouble and are unlikely to contain
	// repos.
	ignoredSubPaths := ignoredPaths(root)

	// Heuristics on dirs which probably don't have useful source.
	ignoredSuffix := []string{
		// no point going into go mod dir.
		"/pkg/mod",

		// Source code should not be here.
		"/bin",

		// Downloaded code so ignore repos in it since it can be large.
		"/node_modules",
	}

	return func(subpath string) bool {
		if maxDepth > 0 && strings.Count(subpath, string(os.PathSeparator)) >= maxDepth {
			return true
		}

		// Previously we recursed into bare repositories which is why this check was here.
		// Now we use this as a sanity check to make sure we didn't somehow stumble into a .git dir.
		base := filepath.Base(subpath)
		if base == ".git" {
			return true
		}

		// skip hidden dirs
		if strings.HasPrefix(base, ".") && base != "." {
			return true
		}

		if slices.Contains(ignoredSubPaths, subpath) {
			return true
		}

		for _, suffix := range ignoredSuffix {
			if strings.HasSuffix(subpath, suffix) {
				return true
			}
		}

		return false
	}
}

// ignoredPaths returns paths relative to root which should be ignored.
//
// In particular this function returns the locations on Mac which trigger
// permission dialogs. If a user wanted to explore those directories they need
// to ensure root is the directory.
func ignoredPaths(root string) []string {
	if runtime.GOOS != "darwin" {
		return nil
	}

	// For simplicity we only trigger this code path if root is a homedir,
	// which is the most common mistake made. Note: Mac can be case
	// insensitive on the FS.
	if !strings.EqualFold("/Users", filepath.Dir(filepath.Clean(root))) {
		return nil
	}

	// Hard to find an actual list. This is based on error messages mentioned
	// in the Entitlement documentation followed by trial and error.
	// https://developer.apple.com/documentation/bundleresources/information_property_list/nsdocumentsfolderusagedescription
	return []string{
		"Applications",
		"Desktop",
		"Documents",
		"Downloads",
		"Library",
		"Movies",
		"Music",
		"Pictures",
		"Public",
	}
}

func explainAddr(addr string) string {
	return fmt.Sprintf(`Serving the repositories at http://%s.

See https://docs.sourcegraph.com/admin/external_service/src_serve_git for
instructions to configure in Sourcegraph.
`, addr)
}

type ListReposRequest struct {
	Root string `json:"root"`
}
