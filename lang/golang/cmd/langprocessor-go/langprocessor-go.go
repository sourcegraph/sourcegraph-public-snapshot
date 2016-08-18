package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

var (
	httpAddr = flag.String("http", ":4141", "HTTP address to listen on")
	lspAddr  = flag.String("lsp", ":2088", "LSP server address")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
	workDir  = flag.String("workspace", "$SGPATH/workspace/go", "where to create workspace directories")
)

func prepareRepo(update bool, workspace, repo, commit string) error {
	gopath := filepath.Join(workspace, "gopath")

	repo, cloneURI := langp.ResolveRepoAlias(repo)

	// Clone the repository.
	repoDir := filepath.Join(gopath, "src", repo)
	return langp.Clone(update, cloneURI, repoDir, commit)
}

// updateGoDependencies updates all Go dependencies in the repository. It is
// the same as:
//
//  go get -u -d ./...
//
// Except it does not update repoURI itself, because it has already been
// updated by PrepareRepo and go get would run `git pull --ff-only` which,
// because we're set to a specific commit and not a branch, fail:
//
// 	fatal: Not possible to fast-forward, aborting.
//
func updateGoDependencies(repoDir string, env []string, repoURI string) error {
	c := exec.Command("go", "list", "./...")
	c.Dir = repoDir
	c.Env = env
	out, err := c.Output()
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
	return langp.Cmd("go", args...).Run()
}

func prepareDeps(update bool, workspace, repo, commit string) error {
	gopath := filepath.Join(workspace, "gopath")
	repo, _ = langp.ResolveRepoAlias(repo)

	// Clone the repository.
	repoDir := filepath.Join(gopath, "src", repo)
	env := []string{"PATH=" + os.Getenv("PATH"), "GOPATH=" + gopath}
	var c *exec.Cmd
	if !update {
		c = langp.Cmd("go", "get", "-d", "./...")
		c.Dir = repoDir
		c.Env = env
		return c.Run()
	}
	return updateGoDependencies(repoDir, env, repo)
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
	out, err := cmd.Output()
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

	// HACK: copy files from /config to /root/.ssh because that is where our
	// configs are located. "_" (underscore) is not a valid key name in a
	// Kubernetes configmap, so we must do this or not use config maps. See:
	//
	// https://github.com/kubernetes/kubernetes/issues/13357#issuecomment-136554256
	// https://github.com/kubernetes/kubernetes/issues/16786#issue-115047222
	// https://github.com/kubernetes/kubernetes/issues/4789
	//
	move := map[string]string{
		"/config/idrsa":      "/root/.ssh/id_rsa",
		"/config/idrsa.pub":  "/root/.ssh/id_rsa.pub",
		"/config/knownhosts": "/root/.ssh/known_hosts",
	}
	for src, dst := range move {
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
			log.Println(err)
			continue
		}
		// We must copy the files because they live on separate devices and
		// renaming just gets us an "invalid cross-device link" error.
		srcFile, err := os.Open(src)
		if err != nil {
			log.Println(err)
			continue
		}
		dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			log.Println(err)
			continue
		}
		srcFile.Close()
		dstFile.Close()
	}

	workDir, err := langp.ExpandSGPath(*workDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Translating HTTP", *httpAddr, "to LSP", *lspAddr)
	http.Handle("/", langp.New(&langp.Translator{
		Addr:        *lspAddr,
		WorkDir:     workDir,
		PrepareRepo: prepareRepo,
		PrepareDeps: prepareDeps,
		ResolveFile: resolveFile,
		FileURI:     fileURI,
	}))
	http.ListenAndServe(*httpAddr, nil)
}
