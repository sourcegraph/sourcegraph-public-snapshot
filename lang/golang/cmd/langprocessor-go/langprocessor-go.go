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

	"sourcegraph.com/sourcegraph/sourcegraph/lang"

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

func prepareRepo(update bool, workspace, repo, commit string) error {
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

	repo, cloneURI := langp.ResolveRepoAlias(repo)

	// Clone the repository.
	repoDir := filepath.Join(gopath, "src", repo)
	return langp.Clone(update, cloneURI, repoDir, commit)
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
func goGetDependencies(repoDir string, env []string, repoURI string) error {
	c := exec.Command("go", "list", "./...")
	c.Dir = repoDir
	c.Env = env
	out, err := langp.CmdOutput(c)
	if err != nil {
		return err
	}
	var pkgs []string
	for _, line := range strings.Split(string(out), "\n") {
		// TODO(slimsag): prefix isn't 100% correct because in strange cases
		// you can have a package under the repo URI in a different repository
		// (e.g. azul3d.org did this for a while). Generally unlikely for most
		// packages though.
		if line != "" && !strings.HasPrefix(line, repoURI) {
			pkgs = append(pkgs, line)
		}
	}
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"get", "-u", "-d"}, pkgs...)
	return langp.CmdRun(exec.Command("go", args...))
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

func prepareDeps(update bool, workspace, repo, commit string) error {
	gopath := filepath.Join(workspace, "gopath")
	repo, _ = langp.ResolveRepoAlias(repo)
	repoDir := filepath.Join(gopath, "src", repo)
	env := []string{"PATH=" + os.Getenv("PATH"), "GOPATH=" + gopath}
	err := goGetDependencies(repoDir, env, repo)
	if err != nil {
		return err
	}

	// We don't want prepareLSPCache failing to signal that the workspace
	// is bad, since it is just for priming the symbol cache + is less
	// reliable.
	go prepareLSPCache(update, workspace, repo)

	return nil
}

func fileURI(repo, commit, file string) string {
	repo, _ = langp.ResolveRepoAlias(repo)
	return "file:///" + filepath.Join("gopath", "src", repo, file)
}

func resolveFile(workspace, _, _, uri string) (*langp.File, error) {
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
	cmd := exec.Command("git", "rev-parse", "--show-toplevel", "HEAD")
	cmd.Dir = dir
	out, err := langp.CmdOutput(cmd)
	if err != nil {
		return nil, err
	}
	lines := strings.Fields(string(out))
	if len(lines) != 2 {
		return nil, errors.New("unexpected number of lines from git rev-parse")
	}
	repoPath := lines[0]
	commit := lines[1]

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

	lang.PrepareKeys()

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
	http.ListenAndServe(*httpAddr, nil)
}
