package cli

import (
	"bytes"
	"fmt"
	"path/filepath"

	"os"
	"os/exec"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/util"
)

type Repo struct {
	RootDir  string // Root directory containing repository being analyzed
	VCSType  string // VCS type (git or hg)
	CommitID string // CommitID of current working directory
}

func OpenRepo(dir string) (*Repo, error) {
	if fi, err := os.Stat(dir); err != nil || !fi.Mode().IsDir() {
		return nil, fmt.Errorf("not a directory: %q", dir)
	}

	rc := new(Repo)

	// VCS and root directory
	var err error
	rc.RootDir, rc.VCSType, err = getRootDir(dir)
	if err != nil {
		return nil, fmt.Errorf("detecting git/hg repository in %s: %s", dir, err)
	}
	if rc.RootDir == "" {
		return nil, fmt.Errorf("no git/hg repository found in or above %s", dir)
	}

	rc.CommitID, err = resolveWorkingTreeRevision(rc.VCSType, rc.RootDir)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func resolveWorkingTreeRevision(vcsType string, dir string) (string, error) {
	var cmd *exec.Cmd
	switch vcsType {
	case "git":
		cmd = exec.Command("git", "rev-parse", "HEAD")
	case "hg":
		cmd = exec.Command("hg", "--config", "trusted.users=root", "identify", "--debug", "-i")
	default:
		return "", fmt.Errorf("unknown vcs type: %q", vcsType)
	}
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}
	// hg adds a "+" if the wd is dirty
	return strings.TrimSuffix(string(bytes.TrimSpace(out)), "+"), nil
}

func getRootDir(dir string) (rootDir string, vcsType string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	ancestors := util.AncestorDirs(dir, true)

	vcsTypes := []string{"git", "hg"}
	for i := len(ancestors) - 1; i >= 0; i-- {
		ancDir := ancestors[i]
		for _, vt := range vcsTypes {
			// Don't check that the FileInfo is a dir because git
			// submodules have a .git file.
			if _, err := os.Stat(filepath.Join(ancDir, "."+vt)); err == nil {
				return ancDir, vt, nil
			}
		}
	}
	return "", "", nil
}
