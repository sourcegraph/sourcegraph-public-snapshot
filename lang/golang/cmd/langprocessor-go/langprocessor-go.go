package main

import (
	"flag"
	"fmt"
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

	c = cmd("go", "get", "./...")
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
