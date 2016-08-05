package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

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

func prepareDeps(update bool, workspace, repo, commit string) error {
	gopath := filepath.Join(workspace, "gopath")

	// TODO(slimsag): find a way to pass this information from the app instead
	// of hard-coding it here.
	if repo == "sourcegraph/sourcegraph" {
		repo = "sourcegraph.com/sourcegraph/sourcegraph"
	}

	// Clone the repository.
	repoDir := filepath.Join(gopath, "src", repo)
	var c *exec.Cmd
	if !update {
		c = langp.Cmd("go", "get", "-d", "./...")
	} else {
		// Note: we use -f flag because the sourcegraph repo isn't cloned from
		// the canonical source according go go-import meta tags:
		//
		//  package sourcegraph.com/sourcegraph/sourcegraph/vendor/sourcegraph.com/sourcegraph/go-diff/diff: sourcegraph.com/sourcegraph/sourcegraph is a custom import path for https://github.com/sourcegraph/sourcegraph, but /sourcegraph/workspace/go/sourcegraph/sourcegraph/042b3b4e624a6c291f1c3d0aebee413d0c8dd348/workspace/gopath/src/sourcegraph.com/sourcegraph/sourcegraph is checked out from ssh://git@github.com
		//
		c = langp.Cmd("go", "get", "-f", "-u", "-d", "./...")
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
