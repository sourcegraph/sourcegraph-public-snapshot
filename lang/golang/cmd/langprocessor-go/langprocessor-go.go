package main

import (
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

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numerals = "0123456789"
)

// safeRepoURI tells if r matches: a-zA-Z0-9/_-.
func safeRepoURI(r string) bool {
	for _, c := range alphabet + numerals + "/_-." {
		r = strings.Replace(r, string(c), "", -1)
	}
	return len(r) == 0
}

func prepareDeps(update bool, workspace, repo, commit string) error {
	gopath := filepath.Join(workspace, "gopath")

	// TODO(slimsag): find a way to pass this information from the app instead
	// of hard-coding it here.
	if repo == "sourcegraph/sourcegraph" {
		repo = "sourcegraph.com/sourcegraph/sourcegraph"
	}

	// Since we feed the repo into shell below, we sanitize it here. This
	// produces errors on first clone, rather than delaying them until later
	// on a workspace update.
	if !safeRepoURI(repo) {
		return fmt.Errorf("repo URI (%q) may only contain: a-zA-Z0-9/_-.", repo)
	}

	// Clone the repository.
	repoDir := filepath.Join(gopath, "src", repo)
	var c *exec.Cmd
	if !update {
		c = langp.Cmd("go", "get", "-d", "./...")
	} else {
		// Our repository is already updated by PrepareRepo, and go get would
		// fail on it anyway because `git pull --ff-only` doesn't work when the
		// repository is at a specific commit / not on a branch:
		//
		// 	fatal: Not possible to fast-forward, aborting.
		//
		// So instead we list the dependencies and update any not residing in
		// this repository.
		shCmd := fmt.Sprintf("$(go list ./... | grep -v %s) | xargs -r go get -u -d", repo)
		c = langp.Cmd("sh", "-c", shCmd)
	}
	c.Dir = repoDir
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "GOPATH=" + gopath}
	if err := c.Run(); err != nil {
		return err
	}
	return nil
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
