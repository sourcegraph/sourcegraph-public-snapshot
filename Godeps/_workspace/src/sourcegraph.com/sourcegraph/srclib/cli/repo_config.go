package cli

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"code.google.com/p/rog-go/parallel"

	"os"
	"os/exec"
	"strings"
	"syscall"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

type Repo struct {
	RootDir  string // Root directory containing repository being analyzed
	VCSType  string // VCS type (git or hg)
	CommitID string // CommitID of current working directory
	CloneURL string // CloneURL of repo.
}

// URI returns the Repo's URI. It returns the empty string if the
// Repo's CloneURL is malformed or empty.
func (c *Repo) URI() string {
	uri, _ := graph.TryMakeURI(c.CloneURL)
	return uri
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

	par := parallel.NewRun(4)
	par.Do(func() error {
		// Current commit ID
		var err error
		rc.CommitID, err = resolveWorkingTreeRevision(rc.VCSType, rc.RootDir)
		return err
	})
	par.Do(func() error {
		// Get repo URI from clone URL.
		rc.CloneURL = getVCSCloneURL(rc.VCSType, rc.RootDir)
		return nil
	})
	return rc, par.Wait()
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
	ancestors := ancestorDirsAndSelfExceptRoot(dir)

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

// ancestorDirsAndSelfExceptRoot returns a list of p's ancestor
// directories (including itself but excluding the root ("." or "/")).
func ancestorDirsAndSelfExceptRoot(p string) []string {
	if p == "" {
		return nil
	}
	if len(p) == 1 && (p[0] == '.' || p[0] == '/') {
		return nil
	}

	var dirs []string
	for i, c := range p {
		if c == '/' {
			dirs = append(dirs, p[:i])
		}
	}
	dirs = append(dirs, p)
	return dirs
}

// getVCSCloneURL gets the primary remote url. getVCSCloneURL returns
// the empty string if it fails or if there is no primary remote.
func getVCSCloneURL(vcsType string, repoDir string) string {
	logWarning := func(err error) {
		if GlobalOpt.Verbose {
			log.Printf("Warning: getVCSCloneURL: %s", err)
		}
	}
	run := func(args ...string) (string, error) {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		cloneURL := strings.TrimSpace(string(out))
		return cloneURL, nil
	}
	switch vcsType {
	case "git":
		// Try to get the "srclib" remote first.
		url, err := run("git", "config", "remote.srclib.url")
		if err == nil {
			return url
		}

		url, err = run("git", "config", "remote.origin.url")
		if code, _ := exitStatus(err); code == 1 {
			// `git config --get` returns exit code 1 if the config key doesn't exist.
			logWarning(ErrNoVCSCloneURL)
			return ""
		} else if err != nil {
			logWarning(ErrNoVCSCloneURL)
			return ""
		}
		return url
	case "hg":
		cloneURL, err := run("hg", "--config", "trusted.users=root", "paths", "default")
		if err != nil {
			logWarning(err)
			return ""
		}
		return cloneURL
	default:
		logWarning(fmt.Errorf("unrecognized VCS %v", vcsType))
		return ""
	}
}

var ErrNoVCSCloneURL = errors.New("Could not determine remote clone URL for the current repository. For git repositories, srclib checks for remotes named 'srclib' or 'origin' (in that order). Run 'git remote add NAME URL' to add a remote, where NAME is either 'srclib' or 'origin' and URL is a git clone URL (e.g. https://example.com/repo.git).' to add a remote. For hg repositories, srclib checks the 'default' remote.")

func exitStatus(err error) (uint32, error) {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// There is no platform independent way to retrieve
			// the exit code, but the following will work on Unix
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return uint32(status.ExitStatus()), nil
			}
		}
		return 0, err
	}
	return 0, nil
}
