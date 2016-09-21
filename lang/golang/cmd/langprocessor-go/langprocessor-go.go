package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cmdutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

var (
	httpAddr = flag.String("http", ":4141", "HTTP address to listen on")
	lspAddr  = flag.String("lsp", ":2088", "LSP server address")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
	workDir  = flag.String("workspace", "$SGPATH/workspace/go", "where to create workspace directories")
)

func prepareRepo(ctx context.Context, update bool, workspace, repo, commit string) error {
	// We check if there is an existing cache directory. Our LSP server
	// does not incrementally change that, so we need to delete it to
	// prevent it serving old data.
	cache := filepath.Join(workspace, "cache")
	if _, err := os.Stat(cache); err == nil {
		err = os.RemoveAll(cache)
		if err != nil {
			return err
		}
	}

	gopath := filepath.Join(workspace, "gopath")

	cloneURI := langp.RepoCloneURL(ctx, repo)
	repo = langp.ResolveRepoAlias(repo)

	// Clone the repository.
	repoDir := filepath.Join(gopath, "src", repo)
	return langp.Clone(ctx, update, cloneURI, repoDir, commit)
}

var (
	goStdlibPackages = make(map[string]struct{})
	listGoStdlibOnce sync.Once
)

func listGoStdlibPackages(ctx context.Context) map[string]struct{} {
	listGoStdlibOnce.Do(func() {
		// Just so that we don't have to hard-code a list of stdlib packages.
		out, err := langp.CmdOutput(ctx, exec.Command("go", "list", "std"))
		if err != nil {
			// Not fatal because this list is not 100% important.
			log.Println("WARNING:", err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if line != "" {
				goStdlibPackages[line] = struct{}{}
			}
		}
	})
	return goStdlibPackages
}

// goGetDependencies gets all Go dependencies in the repository. It is the
// same as:
//
//  go get -u -d ./...
//
// Except it does not update repoURI itself, because it has already been
// updated by PrepareRepo and go get would run `git pull --ff-only` which,
// because we're set to a specific commit and not a branch, fail:
//
// 	fatal: Not possible to fast-forward, aborting.
//
func goGetDependencies(ctx context.Context, repoDir string, env []string, repoURI string) error {
	if repoURI == "github.com/golang/go" {
		return nil // updating dependencies for stdlib does not make sense.
	}
	c := exec.Command("go", "list", "-f", `{{join .Deps "\n"}}`, "./...")
	c.Dir = repoDir
	c.Env = env
	out, err := langp.CmdOutput(ctx, c)
	if err != nil {
		// Note: We do not consider `go list` failures here to be real
		// failures. Because otherwise repositories that contain Go code in an
		// invalid (non-Go) format, like for example:
		//
		// https://github.com/kubernetes/kubernetes/tree/master/staging
		//
		// would simply cause:
		//
		//  can't load package: package k8s.io/kubernetes/staging/src/k8s.io/client-go/1.4/tools/cache: code in directory /Users/stephen/.sourcegraph/workspace/go/github.com/kubernetes/kubernetes/32ba5815ee48cdac110137d36631fe4d7f9458c4/workspace/gopath/src/k8s.io/kubernetes/staging/src/k8s.io/client-go/1.4/tools/cache expects import \"k8s.io/client-go/1.4/tools/cache\"
		//
		// because of the Go import path checking logic. In general 'go list'
		// should not fail under normal circumstances.
		if v, ok := err.(*cmdutil.ExitError); ok {
			if !strings.Contains(string(v.Stderr), "expects import") {
				// Assume it is a real error then.
				return err
			}
		}
		log.Println("ignoring 'go list' failure intentionally")
		err = nil
	}
	var pkgs []string
	stdlib := listGoStdlibPackages(ctx)
	for _, line := range strings.Split(string(out), "\n") {
		// We remove stdlib packages from the list, `go get -u` would be no-op
		// on them, but it would pollute logs and hurt their readability /
		// debuggability.
		_, inStdlib := stdlib[line]

		// TODO(slimsag): prefix isn't 100% correct because in strange cases
		// you can have a package under the repo URI in a different repository
		// (e.g. azul3d.org did this for a while). Generally unlikely for most
		// packages though.
		if line != "" && !strings.HasPrefix(line, repoURI) && !inStdlib {
			pkgs = append(pkgs, line)
		}
	}
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"get", "-u", "-d"}, pkgs...)
	c = exec.Command("go", args...)
	c.Env = env
	return langp.CmdRun(ctx, c)
}

// prepareLSPCache sends a workspace/symbols request. The underlying LSP
// implementation should cache the data it calculates, so future requests to
// it should respond quickly.
func prepareLSPCache(update bool, workspace, repo string) error {
	conn, err := net.Dial("tcp", *lspAddr)
	if err != nil {
		return err
	}
	ctx := context.Background()
	c := jsonrpc2.NewConn(ctx, conn, nil)
	defer func() {
		if err := c.Close(); err != nil {
			log.Println(err)
		}
	}()

	if err := c.Call(ctx, "initialize", lsp.InitializeParams{RootPath: workspace}, nil); err != nil {
		return err
	}
	if err := c.Call(ctx, "workspace/symbol", lsp.WorkspaceSymbolParams{Query: "external " + repo + "/..."}, nil); err != nil {
		return err
	}
	if err := c.Call(ctx, "shutdown", nil, nil); err != nil {
		return err
	}

	return nil
}

func prepareDeps(ctx context.Context, update bool, workspace, repo, commit string) error {
	gopath := filepath.Join(workspace, "gopath")
	repo = langp.ResolveRepoAlias(repo)
	repoDir := filepath.Join(gopath, "src", repo)
	env := []string{"PATH=" + os.Getenv("PATH"), "GOPATH=" + gopath}
	err := goGetDependencies(ctx, repoDir, env, repo)
	if err != nil {
		return err
	}

	// We don't want prepareLSPCache failing to signal that the workspace
	// is bad, since it is just for priming the symbol cache + is less
	// reliable.
	go prepareLSPCache(update, workspace, repo)

	return nil
}

func fileURI(ctx context.Context, repo, commit, file string) string {
	repo = langp.ResolveRepoAlias(repo)
	return "file:///" + filepath.Join("gopath", "src", repo, file)
}

func resolveFile(ctx context.Context, workspace, mainRepo, mainRepoCommit, uri string) (*langp.File, error) {
	if strings.HasPrefix(uri, "stdlib://") {
		// We don't have stdlib checked out as a dep, so LSP returns a
		// special URI for them.
		p := uri[9:]
		i := strings.Index(p, "/")
		if i < 0 {
			return nil, fmt.Errorf("invalid stdlib URI: %s", uri)
		}
		return &langp.File{
			Repo:   "github.com/golang/go",
			Commit: p[:i],
			Path:   p[i+1:],
		}, nil
	}

	if !strings.HasPrefix(uri, "file:///") {
		return nil, fmt.Errorf("uri does not start with file:/// : %s", uri)
	}
	workspacePath := uri[8:]

	// Query the repo and commit containing workspacePath
	fullPath := filepath.Join(workspace, workspacePath)
	dir := fullPath
	if fi, err := os.Stat(fullPath); err != nil || !fi.IsDir() {
		dir = filepath.Dir(dir)
	}
	mainRepoDir := filepath.Join(workspace, "gopath/src", mainRepo)
	var repoPath, commit string
	if dir == mainRepoDir {
		// We already have the information we need (this is the repo that the
		// user is browsing), so no need to consult git.
		repoPath = mainRepoDir
		commit = mainRepoCommit
	} else {
		// This is a dependency that we have cloned via 'go get', so consult
		// git in order to find the repository (which is not always identical
		// to import path).
		cmd := exec.Command("git", "rev-parse", "--show-toplevel", "HEAD")
		cmd.Dir = dir
		out, err := langp.CmdOutput(ctx, cmd)
		if err != nil {
			return nil, err
		}
		lines := strings.Fields(string(out))
		if len(lines) != 2 {
			return nil, errors.New("unexpected number of lines from git rev-parse")
		}
		repoPath = lines[0]
		commit = lines[1]
	}

	// Repo is repoPath relative to our GOPATH/src
	repo, err := filepath.Rel(filepath.Join(workspace, "gopath", "src"), repoPath)
	if err != nil {
		return nil, err
	}
	repo = langp.UnresolveRepoAlias(repo)
	// Path is fullPath relative to repoPath
	path, err := filepath.Rel(repoPath, fullPath)
	if err != nil {
		return nil, err
	}

	return &langp.File{
		Repo:   repo,
		Commit: commit,
		Path:   path,
	}, nil
}

func main() {
	flag.Parse()

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}
	langp.InitMetrics("go")

	workDir, err := langp.ExpandSGPath(*workDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Translating HTTP", *httpAddr, "to LSP", *lspAddr)
	http.Handle("/", langp.New(&langp.Translator{
		Addr: *lspAddr,
		Preparer: langp.NewPreparer(&langp.PreparerOpts{
			WorkDir:     workDir,
			PrepareRepo: prepareRepo,
			PrepareDeps: prepareDeps,
		}),
		ResolveFile: resolveFile,
		FileURI:     fileURI,
	}))
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
