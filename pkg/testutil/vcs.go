package testutil

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os/exec"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/ssh"
)

// trivialGitRepoHandler is an HTTP handler that serves a dummy git repo with
// a single commit (with no files and the message 'hello').
//
// Note: the head commit ID of the repo is
// "bbf4e47c76299d42910d185c762a5b046299c651".
var trivialGitRepoHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/info/refs":
		fmt.Fprintln(w, "bbf4e47c76299d42910d185c762a5b046299c651\trefs/heads/master")
	case "/HEAD":
		fmt.Fprintln(w, "ref: refs/heads/master")
	case "/objects/bb/f4e47c76299d42910d185c762a5b046299c651":
		const data = "eAGdjUEKwjAQAD3nFfsBZRu3yS6I+AXxBUm6pcWkwZr+34o/8DQwMEyqpcwNOseHtqoCRbb9kBzZFJ1GCehQe4ojD04s8xhVSZBM2NpUV7hv87LAI4f0hMvr/eXth1Oq5Qodncl7EWI4okc0u92HTf9IzaQ5V/MBb/w2ig=="
		b, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	case "/objects/4b/825dc642cb6eb9a060e54bf8d69288fbee4904":
		const data = "eAErKUpNVTBgAAAKLAIB"
		b, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	}
})

// NewGitSSHServer starts a git SSH server to serve the repo in the
// given dir. Callers must call the server's Close() method when done
// to free up resources.
func NewGitSSHServer(dir string) (*ssh.Server, *vcs.RemoteOpts, error) {
	s, err := ssh.NewServer("git-shell", dir, ssh.PrivateKey(ssh.SamplePrivKey))
	if err != nil {
		return nil, nil, err
	}
	if err := s.Start(); err != nil {
		return nil, nil, err
	}

	opt := vcs.RemoteOpts{
		SSH: &vcs.SSHConfig{
			PrivateKey: ssh.SamplePrivKey,
		},
	}

	return s, &opt, nil
}

// InitTrivialGitRepo initializes a git repo (using `git init`) in the
// given dir and adds a dummy commit and file to it. It returns the
// commit ID of HEAD.
//
// Note: it is not intended to produce the exact same repo as that
// returned by trivialGitRepoHandler, even though they both have
// "trivial" in the name.
func InitTrivialGitRepo(dir string) (headCommitID string, err error) {
	err = runCmds(dir, []string{
		"git init",
		"touch a",
		"git add a",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit a -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	})
	if err != nil {
		return "", err
	}

	return "9a0754588840d4a6696b6793149290bdeffec7f2", nil
}

func runCmds(dir string, cmds []string) error {
	for _, c := range cmds {
		cmd := exec.Command("sh", "-c", c)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("exec %q in %s failed: %s\n\nOutput was:\n\n%s", c, dir, err, out)
		}
	}
	return nil
}
