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
	"strconv"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cache"
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

var gitRevParseCache = cache.Sync(lru.New(2000))

func gitRevParse(ctx context.Context, dir string) (repoPath, commit string, err error) {
	if v, found := gitRevParseCache.Get(dir); found {
		// This is cache to avoid running git rev-parse below
		lines := v.([]string)
		return lines[0], lines[1], nil
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel", "HEAD")
	cmd.Dir = dir
	out, err := langp.CmdOutput(ctx, cmd)
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) != 3 {
		return "", "", errors.New("unexpected number of lines from git rev-parse")
	}
	gitRevParseCache.Add(dir, lines)
	return lines[0], lines[1], nil
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

	// This could be a dependency that we have cloned via 'go get', so consult
	// git in order to find the repository (which is not always identical
	// to import path).
	repoPath, commit, err := gitRevParse(ctx, dir)
	if err != nil {
		return nil, err
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

func exportedSymbolsQuery(r *langp.RepoRev) string {
	importPath := langp.ResolveRepoAlias(r.Repo)
	return "is:exported " + importPath + "/..."
}

func exportedSymbol(r *langp.RepoRev, f *langp.File, s *lsp.SymbolInformation) *langp.Symbol {
	pkgParts := strings.Split(s.ContainerName, "/")
	var unit string
	if len(pkgParts) < 3 {
		// Hack for stdlib
		unit = s.ContainerName
	} else {
		unit = strings.Join(pkgParts, "/")
	}
	name := s.Name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	path := s.Name
	if pkg, typ := parseContainerName(unit); typ != "" {
		unit = pkg
		path = typ + "/" + path
	}
	if s.Kind == lsp.SKPackage {
		// Make our output match up with packages in srclib
		unit = path
		path = "."
	}
	// containerName may contain a type which we want as part of the path
	return &langp.Symbol{
		DefSpec: langp.DefSpec{
			Repo:     f.Repo,
			Commit:   f.Commit,
			UnitType: "GoPackage",
			Unit:     unit,
			Path:     path,
		},
		Name: name,
		File: f.Path,
		Kind: lspKindToSymbol(s.Kind),
	}
}

func externalRefsQuery(r *langp.RepoRev) string {
	importPath := langp.ResolveRepoAlias(r.Repo)
	return "is:external-ref " + importPath + "/..."
}

func externalRef(r *langp.RepoRev, f *langp.File, s *lsp.SymbolInformation) *langp.Ref {
	repo, pkg, filename, line, col := parseExternalRefContainerName(s.ContainerName)
	// containerName may contain a type which we want as part of the path
	return &langp.Ref{
		Def: &langp.DefSpec{
			Repo: repo,

			// Commit is intentionally omitted, as it has no use in the context of
			// external refs (all refs point to defs of repos at the default branch
			// only).
			Commit:   "",
			UnitType: "GoPackage",
			Unit:     pkg,
			Path:     s.Name,
		},
		File:   filename,
		Line:   line,
		Column: col,
	}
}

func parseExternalRefContainerName(containerName string) (repo, pkg, filename string, line, col int) {
	s := strings.Fields(containerName)
	if len(s) != 5 {
		panic(fmt.Sprintf("parseExternalRefContainerName: invalid container name %q", containerName))
	}
	l, _ := strconv.Atoi(s[3])
	c, _ := strconv.Atoi(s[4])
	return s[0], s[1], s[2], int(l), int(c)
}

func parseContainerName(containerName string) (pkg, typ string) {
	if containerName == "" {
		return containerName, ""
	}
	split := strings.Fields(containerName)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return split[0], ""
}

func lspKindToSymbol(kind lsp.SymbolKind) string {
	switch kind {
	case lsp.SKPackage:
		return "package"
	case lsp.SKField:
		return "field"
	case lsp.SKFunction:
		return "func"
	case lsp.SKMethod:
		// srclib stores this as func
		return "func"
	case lsp.SKVariable:
		return "var"
	case lsp.SKClass:
		return "type"
	case lsp.SKInterface:
		return "interface"
	case lsp.SKConstant:
		return "const"
	default:
		return "unknown"
	}
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
		SymbolsTranslator: &langp.SymbolsTranslator{
			ExportedSymbolsQuery: exportedSymbolsQuery,
			ExportedSymbol:       exportedSymbol,
			ExternalRefsQuery:    externalRefsQuery,
			ExternalRef:          externalRef,
		},
		ResolveFile: resolveFile,
		FileURI:     fileURI,
	}))
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
