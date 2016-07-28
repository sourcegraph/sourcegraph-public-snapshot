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

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

var (
	httpAddr = flag.String("http", ":4141", "HTTP address to listen on")
	lspAddr  = flag.String("lsp", ":2088", "LSP server address")
	profbind = flag.String("prof-http", ":6060", "net/http/pprof http bind address")
	workDir  = flag.String("workspace", "$SGPATH/workspace/go", "where to create workspace directories")
)

func cmd(name string, args ...string) *exec.Cmd {
	s := fmt.Sprintf("exec %s", name)
	for _, arg := range args {
		s = fmt.Sprintf("%s %q", s, arg)
	}
	log.Println(s)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func prepare(workspace, repo, commit string) error {
	gopath := filepath.Join(workspace, "gopath")

	// Clone the repository.
	repoDir := filepath.Join(gopath, "src", repo)
	c := cmd("git", "clone", "https://"+repo, repoDir)
	if err := c.Run(); err != nil {
		return err
	}

	// Reset to the specific revision.
	c = cmd("git", "reset", "--hard", commit)
	c.Dir = repoDir
	if err := c.Run(); err != nil {
		return err
	}

	c = cmd("go", "get", "-d", "./...")
	c.Dir = repoDir
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "GOPATH=" + gopath}
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func fileURI(repo, commit, file string) string {
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
	http.Handle("/", &langp.Translator{
		Addr:    *lspAddr,
		WorkDir: workDir,
		Prepare: prepare,
		FileURI: fileURI,
	})
	http.ListenAndServe(*httpAddr, nil)
}
