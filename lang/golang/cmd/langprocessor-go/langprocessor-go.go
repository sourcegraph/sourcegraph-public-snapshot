package main

import (
	"flag"
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

	// TODO(slimsag): find a way to pass this information from the app instead
	// of hard-coding it here.
	cloneURI := "https://" + repo
	if repo == "sourcegraph/sourcegraph" {
		cloneURI = "git@github.com:sourcegraph/sourcegraph"
		repo = "sourcegraph.com/sourcegraph/sourcegraph"
	}

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

	// TODO(slimsag): find a way to pass this information from the app instead
	// of hard-coding it here.
	if repo == "sourcegraph/sourcegraph" {
		repo = "sourcegraph.com/sourcegraph/sourcegraph"
	}

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
	// TODO(slimsag): find a way to pass this information from the app instead
	// of hard-coding it here.
	if repo == "sourcegraph/sourcegraph" {
		repo = "sourcegraph.com/sourcegraph/sourcegraph"
	}
	return filepath.Join("gopath", "src", repo, file)
}

func main() {
	flag.Parse()

	if *profbind != "" {
		go debugserver.Start(*profbind)
	}

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
		FileURI:     fileURI,
	}))
	http.ListenAndServe(*httpAddr, nil)
}
