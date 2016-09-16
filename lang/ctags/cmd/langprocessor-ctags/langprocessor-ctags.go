package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

var (
	httpAddr = flag.String("http", ":4141", "HTTP address to listen on")
	lspAddr  = flag.String("lsp", ":2088", "LSP server address")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
	workDir  = flag.String("workspace", "$SGPATH/workspace/ctags", "where to create workspace directories")
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

	repo, cloneURI := langp.ResolveRepoAlias(repo)
	repoDir := filepath.Join(workspace, repo)
	// Clone the repository.
	return langp.Clone(update, cloneURI, repoDir, commit)
}

func prepareDeps(update bool, workspace, repo, commit string) error {
	return nil
}

func fileURI(repo, commit, file string) string {
	repo, _ = langp.ResolveRepoAlias(repo)
	return "file:///" + filepath.Join(repo, file)
}

func resolveFile(workspace, repo, commit, uri string) (*langp.File, error) {
	if !strings.HasPrefix(uri, "file:///") {
		return nil, fmt.Errorf("uri does not start with file:/// : %s", uri)
	}
	workspacePath := uri[8:]

	fullPath := filepath.Join("/", workspacePath)
	repoPath := filepath.Join(workspace, repo)
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
	langp.InitMetrics("ctags")

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
	}))
	http.ListenAndServe(*httpAddr, nil)
}
